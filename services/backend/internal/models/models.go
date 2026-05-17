package models

import (
	"time"

	"github.com/google/uuid"
)

// WAFEvent represents a single WAF audit log entry stored in the database.
type WAFEvent struct {
	ID        uuid.UUID              `json:"id"`
	Timestamp time.Time              `json:"timestamp"`
	ClientIP  string                 `json:"client_ip"`
	Method    string                 `json:"method"`
	URI       string                 `json:"uri"`
	RuleID    string                 `json:"rule_id"`
	RuleMsg   string                 `json:"rule_msg"`
	Severity  string                 `json:"severity"`
	Action    string                 `json:"action"`
	RawLog    map[string]interface{} `json:"raw_log"`
}

// EventFilter holds query parameters for filtering events.
type EventFilter struct {
	Limit    int
	Offset   int
	From     *time.Time
	To       *time.Time
	Action   string
	ClientIP string
}

// Stats holds aggregated WAF statistics.
type Stats struct {
	TotalBlocks     int64        `json:"total_blocks"`
	TotalDetections int64        `json:"total_detections"`
	TopIPs          []IPCount    `json:"top_ips"`
	TopRules        []RuleCount  `json:"top_rules"`
	EventsPerHour   []HourCount  `json:"events_per_hour"`
}

// IPCount is a client IP with its event count.
type IPCount struct {
	IP    string `json:"ip"`
	Count int64  `json:"count"`
}

// RuleCount is a rule ID with its event count.
type RuleCount struct {
	RuleID  string `json:"rule_id"`
	RuleMsg string `json:"rule_msg"`
	Count   int64  `json:"count"`
}

// HourCount is a time bucket with its event count.
type HourCount struct {
	Hour  time.Time `json:"hour"`
	Count int64     `json:"count"`
}
