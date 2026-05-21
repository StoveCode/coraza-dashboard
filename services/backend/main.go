package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/corazawaf/coraza-dashboard/internal/api"
	"github.com/corazawaf/coraza-dashboard/internal/configwriter"
	"github.com/corazawaf/coraza-dashboard/internal/db"
	"github.com/corazawaf/coraza-dashboard/internal/models"
	"github.com/corazawaf/coraza-dashboard/internal/scraper"
	"github.com/corazawaf/coraza-dashboard/internal/tailer"
)

func main() {
	// Configure zerolog
	logLevel := zerolog.InfoLevel
	if os.Getenv("LOG_LEVEL") == "debug" {
		logLevel = zerolog.DebugLevel
	}
	zerolog.SetGlobalLevel(logLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})

	// Connect to DB
	pool, err := db.Connect(context.Background())
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to database")
	}
	defer pool.Close()

	// Run migrations
	if err := db.Migrate(context.Background(), pool); err != nil {
		log.Fatal().Err(err).Msg("failed to run migrations")
	}

	// Initialize coraza config file if not present
	corazaConfigPath := os.Getenv("CORAZA_CONFIG_PATH")
	if corazaConfigPath != "" {
		if _, err := os.Stat(corazaConfigPath); os.IsNotExist(err) {
			defaultCfg := &models.RulesConfig{
				EngineMode:        "DetectionOnly",
				ParanoiaLevel:     1,
				InboundThreshold:  5,
				OutboundThreshold: 4,
				DisabledRuleIds:   []string{},
				DisabledTags:      []string{},
			}
			if werr := configwriter.WriteConfig(defaultCfg); werr != nil {
				log.Warn().Err(werr).Msg("could not write default coraza config")
			} else {
				log.Info().Str("path", corazaConfigPath).Msg("wrote default coraza config")
			}
		}
	}

	// Start log tailer (disabled when LOG_INGEST_MODE=true, i.e. Fluent Bit is used)
	if os.Getenv("LOG_INGEST_MODE") != "true" {
		logFile := os.Getenv("LOG_FILE")
		if logFile == "" {
			logFile = "/var/log/coraza/coraza.log"
		}
		t := tailer.New(logFile, pool)
		go t.Run(context.Background())
	} else {
		log.Info().Msg("LOG_INGEST_MODE=true: log tailer disabled, using Fluent Bit ingest endpoint")
	}

	// Start metrics scraper (optional)
	metricsURL := os.Getenv("CORAZA_METRICS_URL")
	var sc *scraper.Scraper
	if metricsURL != "" {
		sc = scraper.New(metricsURL)
		go sc.Run(context.Background())
	}

	// Setup router
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// CORS
	corsOrigins := os.Getenv("CORS_ALLOWED_ORIGINS")
	if corsOrigins == "" {
		corsOrigins = "http://localhost:3000"
	}
	r.Use(corsMiddleware(corsOrigins))

	h := api.NewHandler(pool, sc)
	r.Get("/api/health", h.Health)
	r.Get("/api/events", h.Events)
	r.Get("/api/stats", h.Stats)
	r.Get("/api/metrics", h.Metrics)
	r.Post("/api/ingest", h.Ingest)
	r.Get("/api/rules/config", h.GetRulesConfig)
	r.Put("/api/rules/config", h.PutRulesConfig)
	r.Get("/api/rules/categories", h.GetRuleCategories)
	r.Get("/api/rules/catalog", h.GetRuleCatalog)
	r.Get("/api/system/versions", h.GetSystemVersions)

	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	go func() {
		log.Info().Str("port", port).Msg("starting HTTP server")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("server error")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
	log.Info().Msg("server stopped")
}

func corsMiddleware(allowedOrigins string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", allowedOrigins)
			w.Header().Set("Access-Control-Allow-Methods", "GET, PUT, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
