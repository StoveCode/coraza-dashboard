package parser

import (
	"encoding/json"
	"time"

	"github.com/corazawaf/coraza-dashboard/internal/models"
)

// LogLine is the top-level structure of each coraza-spoa JSON log entry.
type LogLine struct {
	Level string       `json:"level"`
	Match *MatchedRule `json:"match"`
	Time  time.Time    `json:"time"`
}

// MatchedRule maps to the "match" field in coraza-spoa log output.
type MatchedRule struct {
	Client     string   `json:"client"`
	File       string   `json:"file"`
	Line       int      `json:"line"`
	RuleID     int      `json:"rule_id"`
	Msg        string   `json:"msg"`
	Data       string   `json:"data"`
	Severity   string   `json:"severity"`
	SeverityID int      `json:"severity_id"`
	Tags       []string `json:"tags"`
	Server     string   `json:"server"`
	URI        string   `json:"uri"`
	UniqueID   string   `json:"unique_id"`
	Disruptive bool     `json:"disruptive"`
	PhaseID    int      `json:"phase_id"`
	Phase      string   `json:"phase"`
}

// ParseLine parses a raw JSON log line and returns a WAFEvent if it's a rule match.
// Returns nil if the line is not a rule match event or cannot be parsed.
func ParseLine(raw []byte) *models.WAFEvent {
	var ll LogLine
	if err := json.Unmarshal(raw, &ll); err != nil {
		return nil
	}

	if ll.Match == nil {
		return nil
	}

	m := ll.Match
	event := &models.WAFEvent{
		Timestamp:  ll.Time,
		Client:     m.Client,
		Server:     m.Server,
		URI:        m.URI,
		RuleID:     m.RuleID,
		RuleMsg:    m.Msg,
		RuleFile:   m.File,
		Severity:   m.Severity,
		SeverityID: m.SeverityID,
		Phase:      m.Phase,
		PhaseID:    m.PhaseID,
		Disruptive: m.Disruptive,
		Tags:       m.Tags,
		Data:       m.Data,
		UniqueID:   m.UniqueID,
		RawLog:     raw,
	}

	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now().UTC()
	}

	return event
}
