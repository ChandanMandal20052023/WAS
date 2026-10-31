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

	"was/backend/internal/config"
	"was/backend/internal/database"
	"was/backend/internal/middleware"
	"was/backend/internal/utils"
	"was/backend/migrations"
)

func main() {
	cfg := config.Load()

	log.Printf("[WAS] Starting %s in %s mode on port %s", cfg.AppName, cfg.AppEnv, cfg.Port)

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

		// Placeholder route groups ready for next phases
		v1.Route("/auth", func(r chi.Router) {
			r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
				utils.JSON(w, http.StatusOK, "auth module initialized")
			})
		})

		v1.Route("/channels", func(r chi.Router) {
			r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
				utils.JSON(w, http.StatusOK, "channels module initialized")
			})
		})

		v1.Route("/contacts", func(r chi.Router) {
			r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
				utils.JSON(w, http.StatusOK, "contacts module initialized")
			})
		})

		v1.Route("/conversations", func(r chi.Router) {
			r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
				utils.JSON(w, http.StatusOK, "conversations module initialized")
			})
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

