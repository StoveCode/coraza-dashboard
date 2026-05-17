// Package collector receives and parses Coraza WAF audit log JSON.
// Coraza audit log format reference: https://coraza.io/docs/audit-log/
package collector

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/stovecode/coraza-dashboard/backend/internal/db"
	"github.com/stovecode/coraza-dashboard/backend/internal/models"
)

// CorazaAuditLog represents the top-level Coraza JSON audit log entry.
// Fields follow the Coraza audit log v2 format.
type CorazaAuditLog struct {
	Transaction *CorazaTransaction `json:"transaction"`
	// Coraza also nests things under "messages" / "rules"
	Messages []CorazaMessage `json:"messages"`
}

// CorazaTransaction holds request/response transaction details.
type CorazaTransaction struct {
	ID        string              `json:"id"`
	Timestamp string              `json:"timestamp"`
	ClientIP  string              `json:"client_ip"`
	Request   *CorazaRequest      `json:"request"`
	Response  *CorazaResponse     `json:"response"`
}

// CorazaRequest holds HTTP request details.
type CorazaRequest struct {
	Method  string            `json:"method"`
	URI     string            `json:"uri"`
	Headers map[string]string `json:"headers"`
}

// CorazaResponse holds HTTP response details.
type CorazaResponse struct {
	Status int `json:"status"`
}

// CorazaMessage holds rule match details.
type CorazaMessage struct {
	Message string      `json:"message"`
	Details *RuleDetail `json:"details"`
}

// RuleDetail holds the matched rule metadata.
type RuleDetail struct {
	RuleID   string `json:"ruleId"`
	Severity string `json:"severity"`
	Tags     []string `json:"tags"`
}

// Handler returns an http.HandlerFunc that accepts Coraza audit log JSON.
func Handler(w http.ResponseWriter, r *http.Request) {
	// Always respond 200 to avoid blocking Coraza's log pipeline.
	defer func() {
		if rec := recover(); rec != nil {
			log.Printf("collector: panic recovered: %v", rec)
		}
	}()

	var raw map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
		log.Printf("collector: failed to decode body: %v", err)
		w.WriteHeader(http.StatusOK)
		return
	}

	event, err := parseAuditLog(raw)
	if err != nil {
		log.Printf("collector: parse error: %v", err)
		w.WriteHeader(http.StatusOK)
		return
	}

	if err := db.InsertEvent(context.Background(), event); err != nil {
		log.Printf("collector: db insert error: %v", err)
	}

	w.WriteHeader(http.StatusOK)
}

// parseAuditLog converts a raw Coraza audit log map into a WAFEvent.
func parseAuditLog(raw map[string]interface{}) (*models.WAFEvent, error) {
	// Re-marshal for typed parsing
	data, err := json.Marshal(raw)
	if err != nil {
		return nil, fmt.Errorf("re-marshal: %w", err)
	}

	var audit CorazaAuditLog
	if err := json.Unmarshal(data, &audit); err != nil {
		return nil, fmt.Errorf("unmarshal audit: %w", err)
	}

	event := &models.WAFEvent{
		Action: "detect",
		RawLog: raw,
	}

	if audit.Transaction != nil {
		tx := audit.Transaction
		event.ClientIP = tx.ClientIP
		if tx.Request != nil {
			event.Method = tx.Request.Method
			event.URI = tx.Request.URI
		}
		if ts, err := time.Parse(time.RFC3339Nano, tx.Timestamp); err == nil {
			event.Timestamp = ts.UTC()
		}
	}

	// Extract first matched rule info
	for _, msg := range audit.Messages {
		event.RuleMsg = msg.Message
		if msg.Details != nil {
			event.RuleID = msg.Details.RuleID
			event.Severity = msg.Details.Severity
			// If severity is CRITICAL or ERROR → treat as block
			if event.Severity == "CRITICAL" || event.Severity == "ERROR" {
				event.Action = "block"
			}
		}
		break // only first rule
	}

	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now().UTC()
	}

	return event, nil
}
