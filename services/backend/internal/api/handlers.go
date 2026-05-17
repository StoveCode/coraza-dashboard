package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"

	"github.com/corazawaf/coraza-dashboard/internal/db"
	"github.com/corazawaf/coraza-dashboard/internal/models"
	"github.com/corazawaf/coraza-dashboard/internal/scraper"
)

type Handler struct {
	pool    *pgxpool.Pool
	scraper *scraper.Scraper
}

func NewHandler(pool *pgxpool.Pool, sc *scraper.Scraper) *Handler {
	return &Handler{pool: pool, scraper: sc}
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	if err := h.pool.Ping(r.Context()); err != nil {
		jsonError(w, "db unavailable", http.StatusServiceUnavailable)
		return
	}
	jsonOK(w, map[string]string{"status": "ok"})
}

func (h *Handler) Events(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	limit, _ := strconv.Atoi(q.Get("limit"))
	if limit == 0 {
		limit = 50
	}
	offset, _ := strconv.Atoi(q.Get("offset"))

	f := db.ListFilter{
		Limit:    limit,
		Offset:   offset,
		ClientIP: q.Get("client_ip"),
	}

	if v := q.Get("from"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			f.From = t
		}
	}
	if v := q.Get("to"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			f.To = t
		}
	}
	if v := q.Get("disruptive"); v != "" {
		b := v == "true"
		f.Disruptive = &b
	}

	result, err := db.ListEvents(r.Context(), h.pool, f)
	if err != nil {
		log.Error().Err(err).Msg("ListEvents failed")
		jsonError(w, "internal error", http.StatusInternalServerError)
		return
	}
	if result.Events == nil {
		result.Events = []models.WAFEvent{}
	}
	jsonOK(w, map[string]interface{}{
		"total":  result.Total,
		"events": result.Events,
	})
}

func (h *Handler) Stats(w http.ResponseWriter, r *http.Request) {
	stats, err := db.GetStats(r.Context(), h.pool)
	if err != nil {
		log.Error().Err(err).Msg("GetStats failed")
		jsonError(w, "internal error", http.StatusInternalServerError)
		return
	}
	jsonOK(w, stats)
}

func (h *Handler) Metrics(w http.ResponseWriter, r *http.Request) {
	if h.scraper == nil {
		jsonError(w, "metrics scraper not configured", http.StatusServiceUnavailable)
		return
	}
	latest := h.scraper.Latest()
	if latest == "" {
		jsonError(w, "no metrics available yet", http.StatusServiceUnavailable)
		return
	}
	w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	_, _ = w.Write([]byte(latest))
}

func jsonOK(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Error().Err(err).Msg("json encode error")
	}
}

func jsonError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
