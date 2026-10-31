package websocket

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"was/backend/internal/middleware"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow configured origins
	},
}

type Event struct {
	Type           string      `json:"type"`
	OrganizationID uuid.UUID   `json:"organization_id"`
	Payload        interface{} `json:"payload"`
	Timestamp      time.Time   `json:"timestamp"`
}

type Client struct {
	hub            *Hub
	conn           *websocket.Conn
	send           chan []byte
	organizationID uuid.UUID
	userID         uuid.UUID
}

type Hub struct {
	mu         sync.RWMutex
	orgClients map[uuid.UUID]map[*Client]bool
	register   chan *Client
	unregister chan *Client
	broadcast  chan *Event
}

func NewHub() *Hub {
	return &Hub{
		orgClients: make(map[uuid.UUID]map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan *Event, 256),
	}
}

func (h *Hub) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return

		case client := <-h.register:
			h.mu.Lock()
			if _, ok := h.orgClients[client.organizationID]; !ok {
				h.orgClients[client.organizationID] = make(map[*Client]bool)
			}
			h.orgClients[client.organizationID][client] = true
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if clients, ok := h.orgClients[client.organizationID]; ok {
				if _, exists := clients[client]; exists {
					delete(clients, client)
					close(client.send)
					if len(clients) == 0 {
						delete(h.orgClients, client.organizationID)
					}
				}
			}
			h.mu.Unlock()

		case event := <-h.broadcast:
			h.mu.RLock()
			clients, ok := h.orgClients[event.OrganizationID]
			if ok {
				eventBytes, err := json.Marshal(event)
				if err == nil {
					for client := range clients {
						select {
						case client.send <- eventBytes:
						default:
							close(client.send)
							delete(clients, client)
						}
					}
				}
			}
			h.mu.RUnlock()
		}
	}
}

// BroadcastEvent delivers a real-time event to all connected agents within the tenant
func (h *Hub) BroadcastEvent(eventType string, orgID uuid.UUID, payload interface{}) {
	event := &Event{
		Type:           eventType,
		OrganizationID: orgID,
		Payload:        payload,
		Timestamp:      time.Now(),
	}
	select {
	case h.broadcast <- event:
	default:
		log.Printf("[websocket] Warning: event queue full, dropping event: %s", eventType)
	}
}

func (h *Hub) HandleWebSocket(validator middleware.TokenValidator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenStr := r.URL.Query().Get("token")
		if tokenStr == "" {
			tokenStr = r.Header.Get("Sec-WebSocket-Protocol")
		}

		if tokenStr == "" {
			http.Error(w, "missing authentication token", http.StatusUnauthorized)
			return
		}

		claims, err := validator.ValidateToken(tokenStr)
		if err != nil || claims.TokenType != "access" {
			http.Error(w, "invalid or expired token", http.StatusUnauthorized)
			return
		}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("[websocket] Upgrade failed: %v", err)
			return
		}

		client := &Client{
			hub:            h,
			conn:           conn,
			send:           make(chan []byte, 256),
			organizationID: claims.OrganizationID,
			userID:         claims.UserID,
		}

		h.register <- client

		go client.writePump()
		go client.readPump()
	}
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(512)
	_ = c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		_ = c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			_ = c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			_, _ = w.Write(message)

			// Drain queued messages to avoid fragmentation
			n := len(c.send)
			for i := 0; i < n; i++ {
				_, _ = w.Write([]byte{'\n'})
				_, _ = w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
