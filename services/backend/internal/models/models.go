package models

import (
	"time"
)

// WAFEvent represents a single WAF rule match from coraza-spoa JSON logs.
type WAFEvent struct {
	ID         string    `json:"id"`
	Timestamp  time.Time `json:"timestamp"`
	Client     string    `json:"client_ip"`
	Server     string    `json:"server"`
	URI        string    `json:"uri"`
	RuleID     int       `json:"rule_id"`
	RuleMsg    string    `json:"rule_msg"`
	RuleFile   string    `json:"rule_file"`
	Severity   string    `json:"severity"`
	SeverityID int       `json:"severity_id"`
	Phase      string    `json:"phase"`
	PhaseID    int       `json:"phase_id"`
	Disruptive bool      `json:"disruptive"`
	Tags       []string  `json:"tags"`
	Data       string    `json:"data"`
	UniqueID   string    `json:"unique_id"`
	RawLog     []byte    `json:"raw_log,omitempty"`
}

// Stats aggregates statistics for the dashboard.
type Stats struct {
	TotalBlocks     int64         `json:"total_blocks"`
	TotalDetections int64         `json:"total_detections"`
	TopIPs          []TopEntry    `json:"top_ips"`
	TopRules        []TopRule     `json:"top_rules"`
	TopTags         []TopEntry    `json:"top_tags"`
	TopPhases       []TopEntry    `json:"top_phases"`
	EventsPerHour   []HourBucket  `json:"events_per_hour"`
}

type TopEntry struct {
	Label string `json:"label"`
	Count int64  `json:"count"`
}

type TopRule struct {
	RuleID int    `json:"rule_id"`
	Msg    string `json:"msg"`
	Count  int64  `json:"count"`
}

type HourBucket struct {
	Hour  time.Time `json:"hour"`
	Count int64     `json:"count"`
}
