package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"
	"github.com/stovecode/coraza-dashboard/backend/internal/api"
	"github.com/stovecode/coraza-dashboard/backend/internal/collector"
	"github.com/stovecode/coraza-dashboard/backend/internal/db"
)

func main() {
	// Load .env if present (non-fatal)
	_ = godotenv.Load()

	port := getenv("SERVER_PORT", "8080")
	allowedOrigins := strings.Split(getenv("CORS_ALLOWED_ORIGINS", "http://localhost:3000"), ",")

	// Connect DB with retry
	ctx := context.Background()
	if err := connectWithRetry(ctx, 30, 2*time.Second); err != nil {
		log.Fatalf("database connect: %v", err)
	}
	defer db.Close()

	if err := db.Migrate(ctx); err != nil {
		log.Fatalf("database migrate: %v", err)
	}

	// Router
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Request-ID"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	r.Get("/api/health", api.HealthHandler)
	r.Get("/api/events", api.EventsHandler)
	r.Get("/api/stats", api.StatsHandler)
	r.Post("/api/log", collector.Handler)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", port),
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("backend listening on :%s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v", err)
		}
	}()

	<-quit
	log.Println("shutting down...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("shutdown error: %v", err)
	}
	log.Println("stopped")
}

func connectWithRetry(ctx context.Context, attempts int, delay time.Duration) error {
	for i := 0; i < attempts; i++ {
		if err := db.Connect(ctx); err == nil {
			return nil
		} else {
			log.Printf("db connect attempt %d/%d failed: %v — retrying in %s", i+1, attempts, err, delay)
		}
		time.Sleep(delay)
	}
	return fmt.Errorf("could not connect to database after %d attempts", attempts)
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
