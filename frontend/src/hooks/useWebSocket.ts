import { useEffect, useRef, useState, useCallback } from 'react';
import { useQueryClient } from '@tanstack/react-query';
import { useAuthStore } from '../stores/authStore';

export interface WebSocketEvent {
  type: string;
  organization_id: string;
  payload: any;
  timestamp: string;
}

export function useWebSocket() {
  const queryClient = useQueryClient();
  const token = useAuthStore((state) => state.accessToken);
  const [isConnected, setIsConnected] = useState(false);
  const socketRef = useRef<WebSocket | null>(null);
  const reconnectTimeoutRef = useRef<NodeJS.Timeout | null>(null);

  const connect = useCallback(() => {
    if (!token) return;

    // Determine WS protocol and URL
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const host = window.location.host;
    const wsUrl = `${protocol}//${host}/api/v1/ws?token=${encodeURIComponent(token)}`;

    try {
      const ws = new WebSocket(wsUrl);
      socketRef.current = ws;

      ws.onopen = () => {
        setIsConnected(true);
      };

      ws.onmessage = (event) => {
        try {
          const wsEvent: WebSocketEvent = JSON.parse(event.data);

          switch (wsEvent.type) {
            case 'message.new':
              // Invalidate conversations list & specific conversation message history
              queryClient.invalidateQueries({ queryKey: ['conversations'] });
              queryClient.invalidateQueries({ queryKey: ['messages', wsEvent.payload?.conversation_id] });
              queryClient.invalidateQueries({ queryKey: ['dashboard-metrics'] });
              break;

            case 'conversation.status_changed':
            case 'conversation.assigned':
              queryClient.invalidateQueries({ queryKey: ['conversations'] });
              queryClient.invalidateQueries({ queryKey: ['dashboard-metrics'] });
              break;

            case 'channel.status_changed':
              queryClient.invalidateQueries({ queryKey: ['channels'] });
              queryClient.invalidateQueries({ queryKey: ['dashboard-metrics'] });
              break;
          }
        } catch {
          // ignore non-json messages
        }
      };

      ws.onclose = () => {
        setIsConnected(false);
        // Attempt reconnect in 3s
        reconnectTimeoutRef.current = setTimeout(connect, 3000);
      };

      ws.onerror = () => {
        ws.close();
      };
    } catch {
      setIsConnected(false);
    }
  }, [token, queryClient]);

  useEffect(() => {
    connect();

    return () => {
      if (reconnectTimeoutRef.current) {
        clearTimeout(reconnectTimeoutRef.current);
      }
      if (socketRef.current) {
        socketRef.current.close();
      }
    };
  }, [connect]);

  return { isConnected };
}
