package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"

	"github.com/corazawaf/coraza-dashboard/internal/configwriter"
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

var ruleCategories = []models.RuleCategory{
	{Tag: "attack-sqli", Label: "SQL Injection", Description: "SQLi attacks via libinjection and pattern matching"},
	{Tag: "attack-xss", Label: "Cross-Site Scripting", Description: "XSS attacks via libinjection and pattern matching"},
	{Tag: "attack-rce", Label: "Remote Code Execution", Description: "OS command injection and RCE attempts"},
	{Tag: "attack-lfi", Label: "Local File Inclusion", Description: "Path traversal and LFI attempts"},
	{Tag: "attack-rfi", Label: "Remote File Inclusion", Description: "RFI attempts"},
	{Tag: "attack-ssrf", Label: "SSRF", Description: "Server-Side Request Forgery"},
	{Tag: "attack-injection", Label: "Generic Injection", Description: "Generic injection attacks"},
	{Tag: "attack-scanner", Label: "Scanner Detection", Description: "Web application scanner user-agents"},
	{Tag: "attack-protocol", Label: "Protocol Attacks", Description: "HTTP protocol violations"},
	{Tag: "attack-reputation", Label: "IP Reputation", Description: "Known malicious IP addresses"},
}

func (h *Handler) GetRulesConfig(w http.ResponseWriter, r *http.Request) {
	cfg, err := db.GetRulesConfig(r.Context(), h.pool)
	if err != nil {
		log.Error().Err(err).Msg("GetRulesConfig failed")
		jsonError(w, "internal error", http.StatusInternalServerError)
		return
	}
	jsonOK(w, cfg)
}

func (h *Handler) PutRulesConfig(w http.ResponseWriter, r *http.Request) {
	var cfg models.RulesConfig
	if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
		jsonError(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate
	switch cfg.EngineMode {
	case "On", "DetectionOnly", "Off":
	default:
		jsonError(w, "engine_mode must be On, DetectionOnly, or Off", http.StatusBadRequest)
		return
	}
	if cfg.ParanoiaLevel < 1 || cfg.ParanoiaLevel > 4 {
		jsonError(w, "paranoia_level must be 1-4", http.StatusBadRequest)
		return
	}
	if cfg.InboundThreshold <= 0 || cfg.OutboundThreshold <= 0 {
		jsonError(w, "thresholds must be > 0", http.StatusBadRequest)
		return
	}
	if cfg.DisabledRuleIds == nil {
		cfg.DisabledRuleIds = []string{}
	}
	if cfg.DisabledTags == nil {
		cfg.DisabledTags = []string{}
	}

	if err := db.SaveRulesConfig(r.Context(), h.pool, &cfg); err != nil {
		log.Error().Err(err).Msg("SaveRulesConfig failed")
		jsonError(w, "internal error", http.StatusInternalServerError)
		return
	}

	if err := configwriter.WriteConfig(&cfg); err != nil {
		log.Warn().Err(err).Msg("configwriter.WriteConfig failed (CORAZA_CONFIG_PATH not set?)")
	}

	jsonOK(w, map[string]string{"status": "ok", "message": "Config updated, coraza-spoa reloading..."})
}

func (h *Handler) GetRuleCategories(w http.ResponseWriter, r *http.Request) {
	jsonOK(w, ruleCategories)
}
