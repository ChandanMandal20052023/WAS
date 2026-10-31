package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"

	"was/backend/internal/auth"
	"was/backend/internal/channels"
	"was/backend/internal/config"
	"was/backend/internal/contacts"
	"was/backend/internal/conversations"
	"was/backend/internal/dashboard"
	"was/backend/internal/database"
	"was/backend/internal/evolution"
	"was/backend/internal/messages"
	"was/backend/internal/middleware"
	"was/backend/internal/models"
	"was/backend/internal/organizations"
	"was/backend/internal/utils"
	"was/backend/migrations"
)

func main() {
	cfg := config.Load()

	log.Printf("[WAS] Starting %s in %s mode on port %s", cfg.AppName, cfg.AppEnv, cfg.Port)

	// Token Manager for JWT Auth
	tokenManager := auth.NewTokenManager(cfg.JWTSecret)

	// Evolution Go Gateway selection (Live vs Mock)
	var evolutionGateway evolution.EvolutionGateway
	if cfg.UseMockEvolution {
		log.Println("[evolution] Using Mock Evolution Adapter for testing/development")
		evolutionGateway = evolution.NewMockAdapter()
	} else {
		log.Printf("[evolution] Initializing Live Evolution Client pointing to %s", cfg.EvolutionURL)
		evolutionGateway = evolution.NewLiveClient(cfg.EvolutionURL, cfg.EvolutionKey)
	}

	// Database initialization
	pg, err := database.NewPostgres(cfg.DatabaseURL)
	if err != nil {
		log.Printf("[database] Warning: Failed to connect to PostgreSQL: %v", err)
	} else {
		defer pg.Close()

		// Execute database schema migrations
		log.Println("[database] Running database schema migrations...")
		if err := migrations.RunMigrations(cfg.DatabaseURL); err != nil {
			log.Printf("[database] Warning: Schema migration failed: %v", err)
		}
	}

	// Redis initialization
	rdb, err := database.NewRedis(cfg.RedisURL)
	if err != nil {
		log.Printf("[redis] Warning: Failed to connect to Redis: %v", err)
	} else {
		defer rdb.Close()
	}

	// Instantiate Services and Handlers if DB is available
	var authHandler *auth.Handler
	var orgHandler *organizations.Handler
	var channelHandler *channels.Handler
	var webhookHandler *evolution.WebhookHandler
	var contactHandler *contacts.Handler
	var convHandler *conversations.Handler
	var msgHandler *messages.Handler
	var dashHandler *dashboard.Handler

	if pg != nil && pg.Pool != nil {
		authRepo := auth.NewRepository(pg.Pool)
		authService := auth.NewService(authRepo, tokenManager)
		authHandler = auth.NewHandler(authService)

		orgRepo := organizations.NewRepository(pg.Pool)
		orgHandler = organizations.NewHandler(orgRepo)

		channelRepo := channels.NewRepository(pg.Pool)
		channelService := channels.NewService(channelRepo, evolutionGateway, cfg.WebhookBaseURL)
		channelHandler = channels.NewHandler(channelService)

		contactRepo := contacts.NewRepository(pg.Pool)
		contactService := contacts.NewService(contactRepo)
		contactHandler = contacts.NewHandler(contactService)

		convRepo := conversations.NewRepository(pg.Pool)
		convService := conversations.NewService(convRepo)
		convHandler = conversations.NewHandler(convService)

		msgRepo := messages.NewRepository(pg.Pool)
		msgService := messages.NewService(msgRepo, convRepo, contactRepo, channelRepo, evolutionGateway)
		msgHandler = messages.NewHandler(msgService)

		webhookHandler = evolution.NewWebhookHandler(channelRepo, msgService)

		dashRepo := dashboard.NewRepository(pg.Pool)
		dashHandler = dashboard.NewHandler(dashRepo)
	}

	r := chi.NewRouter()

	// Base Middlewares
	r.Use(middleware.CORS())
	r.Use(middleware.Logger)
	r.Use(middleware.Recovery)

	// Root Healthcheck
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		utils.JSON(w, http.StatusOK, map[string]string{
			"app":     cfg.AppName,
			"version": "1.0.0",
			"status":  "running",
		})
	})

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		healthStatus := map[string]string{
			"status":   "healthy",
			"database": "connected",
			"redis":    "connected",
		}

		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		if pg == nil || pg.Ping(ctx) != nil {
			healthStatus["database"] = "disconnected"
			healthStatus["status"] = "degraded"
		}

		if rdb == nil || rdb.Ping(ctx) != nil {
			healthStatus["redis"] = "disconnected"
			healthStatus["status"] = "degraded"
		}

		statusCode := http.StatusOK
		if healthStatus["status"] != "healthy" {
			statusCode = http.StatusServiceUnavailable
		}

		utils.JSON(w, statusCode, healthStatus)
	})

	// API v1 Namespace
	r.Route("/api/v1", func(v1 chi.Router) {
		v1.Get("/health", func(w http.ResponseWriter, r *http.Request) {
			utils.JSON(w, http.StatusOK, map[string]string{
				"api":     "v1",
				"service": "WAS API",
				"status":  "operational",
			})
		})

		// Public Inbound Evolution Go Webhook Callback
		v1.Post("/webhooks/evolution/{token}", func(w http.ResponseWriter, r *http.Request) {
			if webhookHandler != nil {
				webhookHandler.HandleWebhook(w, r)
			} else {
				utils.Error(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "Database not connected")
			}
		})

		// Authentication Routes
		v1.Route("/auth", func(authRouter chi.Router) {
			if authHandler != nil {
				// Public routes
				authRouter.Post("/register", authHandler.Register)
				authRouter.Post("/login", authHandler.Login)
				authRouter.Post("/refresh", authHandler.Refresh)
				authRouter.Post("/forgot-password", authHandler.ForgotPassword)
				authRouter.Post("/reset-password", authHandler.ResetPassword)
				authRouter.Post("/invites/accept", authHandler.AcceptInvite)

				// Protected routes
				authRouter.Group(func(protected chi.Router) {
					protected.Use(middleware.Authenticate(tokenManager))

					protected.Get("/me", authHandler.GetMe)
					protected.Post("/logout", authHandler.Logout)
					protected.With(middleware.RequireRoles(models.RoleOwner, models.RoleAdmin)).Post("/invite", authHandler.Invite)
				})
			} else {
				authRouter.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
					utils.JSON(w, http.StatusOK, "auth module initialized (database pending)")
				})
			}
		})

		// Organization Routes
		v1.Route("/organizations", func(orgRouter chi.Router) {
			if orgHandler != nil {
				orgRouter.Use(middleware.Authenticate(tokenManager))

				orgRouter.Get("/current", orgHandler.GetCurrent)
				orgRouter.With(middleware.RequireRoles(models.RoleOwner, models.RoleAdmin)).Put("/current", orgHandler.UpdateCurrent)
				orgRouter.Get("/members", orgHandler.GetMembers)
			}
		})

		// WhatsApp Channels Routes
		v1.Route("/channels", func(chRouter chi.Router) {
			if channelHandler != nil {
				chRouter.Use(middleware.Authenticate(tokenManager))

				chRouter.Get("/", channelHandler.List)
				chRouter.With(middleware.RequireRoles(models.RoleOwner, models.RoleAdmin)).Post("/", channelHandler.Create)
				chRouter.Get("/{id}", channelHandler.Get)
				chRouter.Get("/{id}/qr", channelHandler.GetQRCode)
				chRouter.Post("/{id}/connect", channelHandler.Connect)
				chRouter.Get("/{id}/status", channelHandler.SyncStatus)
				chRouter.Post("/{id}/disconnect", channelHandler.Disconnect)
				chRouter.Post("/{id}/reconnect", channelHandler.Reconnect)
				chRouter.With(middleware.RequireRoles(models.RoleOwner, models.RoleAdmin)).Delete("/{id}", channelHandler.Delete)
				chRouter.Post("/{id}/pair", channelHandler.PairPhone)
			} else {
				chRouter.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
					utils.JSON(w, http.StatusOK, "channels module initialized (database pending)")
				})
			}
		})

		// Dashboard Metrics
		v1.Route("/dashboard", func(dashRouter chi.Router) {
			if dashHandler != nil {
				dashRouter.Use(middleware.Authenticate(tokenManager))
				dashHandler.RegisterRoutes(dashRouter)
			}
		})

		// Contact Routes
		v1.Route("/contacts", func(cRouter chi.Router) {
			if contactHandler != nil {
				cRouter.Use(middleware.Authenticate(tokenManager))
				contactHandler.RegisterRoutes(cRouter)
			}
		})

		// Tag Routes
		v1.Route("/tags", func(tRouter chi.Router) {
			if contactHandler != nil {
				tRouter.Use(middleware.Authenticate(tokenManager))
				contactHandler.RegisterTagRoutes(tRouter)
			}
		})

		// Conversations & Messaging Routes
		v1.Route("/conversations", func(cvRouter chi.Router) {
			if convHandler != nil && msgHandler != nil {
				cvRouter.Use(middleware.Authenticate(tokenManager))
				convHandler.RegisterRoutes(cvRouter)
				msgHandler.RegisterRoutes(cvRouter)
			}
		})
	})

	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Port),
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful Shutdown
	go func() {
		log.Printf("[WAS] Server listening on :%s", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("[WAS] Server error: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	<-stop
	log.Println("[WAS] Shutting down gracefully...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("[WAS] Server forced to shutdown: %v", err)
	}

	log.Println("[WAS] Server exited cleanly.")
}
