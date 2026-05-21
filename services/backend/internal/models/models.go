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
	UniqueID     string    `json:"unique_id"`
	RawLog       []byte    `json:"raw_log,omitempty"`
	AnomalyScore int       `json:"anomaly_score"`  // 0 = no block event
	BlockType    string    `json:"block_type"`     // "inbound", "outbound", "" = not a block
}

// Stats aggregates statistics for the dashboard.
type Stats struct {
	TotalBlocks         int64         `json:"total_blocks"`
	TotalInboundBlocks  int64         `json:"total_inbound_blocks"`
	TotalOutboundBlocks int64         `json:"total_outbound_blocks"`
	TotalDetections     int64         `json:"total_detections"`
	AvgAnomalyScore     float64       `json:"avg_anomaly_score"`
	MaxAnomalyScore     int           `json:"max_anomaly_score"`
	ScoreDistribution   []ScoreBucket `json:"score_distribution"`
	TopIPs          []TopEntry    `json:"top_ips"`
	TopIPsBlocked   []TopEntry    `json:"top_ips_blocked"`
	TopRules        []TopRule     `json:"top_rules"`
	TopTags         []TopEntry    `json:"top_tags"`
	TopPhases       []TopEntry    `json:"top_phases"`
	EventsPerHour   []HourBucket  `json:"events_per_hour"`
}

// ScoreBucket represents a range of anomaly scores and how many events fall in it.
type ScoreBucket struct {
	Range string `json:"range"`
	Count int64  `json:"count"`
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

// RulesConfig represents the WAF rules configuration.
type RulesConfig struct {
	EngineMode           string   `json:"engine_mode"`
	ParanoiaLevel        int      `json:"paranoia_level"`
	ParanoiaLevelEnabled bool     `json:"paranoia_level_enabled"`
	InboundThreshold     int      `json:"inbound_threshold"`
	OutboundThreshold    int      `json:"outbound_threshold"`
	DisabledRuleIds      []string `json:"disabled_rule_ids"`
	DisabledTags         []string `json:"disabled_tags"`
	ResponseCheck        bool     `json:"response_check"`
	CRSVersion           string   `json:"crs_version,omitempty"`
}

// RuleCategory represents a CRS rule category.
type RuleCategory struct {
	Tag         string `json:"tag"`
	Label       string `json:"label"`
	Description string `json:"description"`
}

type HourBucket struct {
	Hour  time.Time `json:"hour"`
	Count int64     `json:"count"`
}
