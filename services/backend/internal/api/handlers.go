package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/stovecode/coraza-dashboard/backend/internal/db"
	"github.com/stovecode/coraza-dashboard/backend/internal/models"
)

// HealthHandler returns a simple health check response.
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// EventsHandler returns a paginated list of WAF events.
func EventsHandler(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	filter := models.EventFilter{
		Limit:    parseIntParam(q.Get("limit"), 50),
		Offset:   parseIntParam(q.Get("offset"), 0),
		Action:   q.Get("action"),
		ClientIP: q.Get("client_ip"),
	}

	if from := q.Get("from"); from != "" {
		if t, err := time.Parse(time.RFC3339, from); err == nil {
			filter.From = &t
		} else {
			writeError(w, http.StatusBadRequest, "invalid 'from' timestamp, use RFC3339")
			return
		}
	}
	if to := q.Get("to"); to != "" {
		if t, err := time.Parse(time.RFC3339, to); err == nil {
			filter.To = &t
		} else {
			writeError(w, http.StatusBadRequest, "invalid 'to' timestamp, use RFC3339")
			return
		}
	}

	events, total, err := db.ListEvents(r.Context(), filter)
	if err != nil {
		log.Printf("api: list events: %v", err)
		writeError(w, http.StatusInternalServerError, "database error")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"total":  total,
		"limit":  filter.Limit,
		"offset": filter.Offset,
		"events": events,
	})
}

// StatsHandler returns aggregated WAF statistics.
func StatsHandler(w http.ResponseWriter, r *http.Request) {
	stats, err := db.GetStats(r.Context())
	if err != nil {
		log.Printf("api: get stats: %v", err)
		writeError(w, http.StatusInternalServerError, "database error")
		return
	}
	writeJSON(w, http.StatusOK, stats)
}

// --- helpers ---

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("api: encode response: %v", err)
	}
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func parseIntParam(s string, fallback int) int {
	if s == "" {
		return fallback
	}
	v, err := strconv.Atoi(s)
	if err != nil || v < 0 {
		return fallback
	}
	return v
}
