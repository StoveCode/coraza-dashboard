package db

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/corazawaf/coraza-dashboard/internal/models"
)

// InsertEvent inserts a WAFEvent into the database.
func InsertEvent(ctx context.Context, pool *pgxpool.Pool, e *models.WAFEvent) error {
	_, err := pool.Exec(ctx, `
		INSERT INTO waf_events
			(timestamp, client_ip, server, uri, rule_id, rule_msg, rule_file,
			 severity, severity_id, phase, phase_id, disruptive, tags, data, unique_id, raw_log,
			 anomaly_score, block_type)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18)
		ON CONFLICT (unique_id) DO NOTHING
	`, e.Timestamp, e.Client, e.Server, e.URI, e.RuleID, e.RuleMsg, e.RuleFile,
		e.Severity, e.SeverityID, e.Phase, e.PhaseID, e.Disruptive,
		e.Tags, e.Data, e.UniqueID, e.RawLog, e.AnomalyScore, e.BlockType)
	return err
}

// ListFilter holds filter params for ListEvents.
type ListFilter struct {
	Limit      int
	Offset     int
	From       time.Time
	To         time.Time
	Disruptive *bool
	ClientIP   string
	RuleID     int    // 0 = no filter
	Tag        string // "" = no filter
	BlockType  string // "" = no filter
}

// ListResult is the paginated response.
type ListResult struct {
	Total  int64
	Events []models.WAFEvent
}

// ListEvents returns a paginated list of events.
func ListEvents(ctx context.Context, pool *pgxpool.Pool, f ListFilter) (*ListResult, error) {
	args := []interface{}{}
	where := "WHERE 1=1"
	i := 1

	if !f.From.IsZero() {
	where += " AND timestamp >= $" + strconv.Itoa(i)
		args = append(args, f.From)
		i++
	}
	if !f.To.IsZero() {
	where += " AND timestamp <= $" + strconv.Itoa(i)
		args = append(args, f.To)
		i++
	}
	if f.Disruptive != nil {
	where += " AND disruptive = $" + strconv.Itoa(i)
		args = append(args, *f.Disruptive)
		i++
	}
	if f.ClientIP != "" {
	where += " AND client_ip = $" + strconv.Itoa(i)
		args = append(args, f.ClientIP)
		i++
	}
	if f.RuleID != 0 {
	where += " AND rule_id = $" + strconv.Itoa(i)
		args = append(args, f.RuleID)
		i++
	}
	if f.Tag != "" {
	where += " AND $" + strconv.Itoa(i) + " = ANY(tags)"
		args = append(args, f.Tag)
		i++
	}
	if f.BlockType != "" {
	where += " AND block_type = $" + strconv.Itoa(i)
		args = append(args, f.BlockType)
		i++
	}

	countArgs := make([]interface{}, len(args))
	copy(countArgs, args)

	var total int64
	if err := pool.QueryRow(ctx, "SELECT COUNT(*) FROM waf_events "+where, countArgs...).Scan(&total); err != nil {
		return nil, err
	}

	limit := f.Limit
	if limit <= 0 {
		limit = 50
	}
	args = append(args, limit, f.Offset)

	rows, err := pool.Query(ctx, `
		SELECT id, timestamp, client_ip, server, uri, rule_id, rule_msg, rule_file,
		       severity, severity_id, phase, phase_id, disruptive, tags, data, unique_id,
		       anomaly_score, block_type
		FROM waf_events `+where+`
		ORDER BY timestamp DESC
		LIMIT $`+strconv.Itoa(i)+` OFFSET $`+strconv.Itoa(i+1), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []models.WAFEvent
	for rows.Next() {
		var e models.WAFEvent
		if err := rows.Scan(&e.ID, &e.Timestamp, &e.Client, &e.Server, &e.URI,
			&e.RuleID, &e.RuleMsg, &e.RuleFile, &e.Severity, &e.SeverityID,
			&e.Phase, &e.PhaseID, &e.Disruptive, &e.Tags, &e.Data, &e.UniqueID,
			&e.AnomalyScore, &e.BlockType); err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}
	return &ListResult{Total: total, Events: events}, nil
}

// GetStats returns aggregated statistics.
func GetStats(ctx context.Context, pool *pgxpool.Pool) (*models.Stats, error) {
	var s models.Stats
	// Initialize slices to empty (not nil) so JSON encodes as [] not null
	s.EventsPerHour = []models.HourBucket{}
	s.TopIPs = []models.TopEntry{}
	s.TopIPsBlocked = []models.TopEntry{}
	s.TopRules = []models.TopRule{}
	s.TopTags = []models.TopEntry{}
	s.TopPhases = []models.TopEntry{}
	s.ScoreDistribution = []models.ScoreBucket{}

	// Counts
	_ = pool.QueryRow(ctx, "SELECT COUNT(DISTINCT unique_id) FROM waf_events WHERE rule_id IN (949110, 949111)").Scan(&s.TotalInboundBlocks)
	_ = pool.QueryRow(ctx, "SELECT COUNT(DISTINCT unique_id) FROM waf_events WHERE rule_id = 959100").Scan(&s.TotalOutboundBlocks)
	s.TotalBlocks = s.TotalInboundBlocks + s.TotalOutboundBlocks
	_ = pool.QueryRow(ctx, "SELECT COUNT(DISTINCT unique_id) FROM waf_events WHERE rule_id NOT IN (949110, 949111, 959100)").Scan(&s.TotalDetections)

	// Top IPs
	rows, err := pool.Query(ctx, `
		SELECT client_ip, COUNT(*) as cnt FROM waf_events
		WHERE client_ip IS NOT NULL AND client_ip != ''
		GROUP BY client_ip ORDER BY cnt DESC LIMIT 10`)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var e models.TopEntry
			_ = rows.Scan(&e.Label, &e.Count)
			s.TopIPs = append(s.TopIPs, e)
		}
		_ = rows.Err()
	}

	// Top IPs Blocked
	rowsBlocked, err := pool.Query(ctx, `
		SELECT client_ip, COUNT(DISTINCT unique_id) as cnt 
		FROM waf_events WHERE rule_id IN (949110, 949111) 
		GROUP BY client_ip ORDER BY cnt DESC LIMIT 10`)
	if err == nil {
		defer rowsBlocked.Close()
		for rowsBlocked.Next() {
			var e models.TopEntry
			_ = rowsBlocked.Scan(&e.Label, &e.Count)
			s.TopIPsBlocked = append(s.TopIPsBlocked, e)
		}
		_ = rowsBlocked.Err()
	}

	// Top Rules
	rows2, err := pool.Query(ctx, `
		SELECT rule_id, COALESCE(MAX(rule_msg),''), COUNT(*) as cnt FROM waf_events
		WHERE rule_id IS NOT NULL
		GROUP BY rule_id ORDER BY cnt DESC LIMIT 10`)
	if err == nil {
		defer rows2.Close()
		for rows2.Next() {
			var e models.TopRule
			_ = rows2.Scan(&e.RuleID, &e.Msg, &e.Count)
			s.TopRules = append(s.TopRules, e)
		}
		_ = rows2.Err()
	}

	// Top Tags
	rows3, err := pool.Query(ctx, `
		SELECT tag, COUNT(*) as cnt FROM waf_events, UNNEST(tags) AS tag
		GROUP BY tag ORDER BY cnt DESC LIMIT 10`)
	if err == nil {
		defer rows3.Close()
		for rows3.Next() {
			var e models.TopEntry
			_ = rows3.Scan(&e.Label, &e.Count)
			s.TopTags = append(s.TopTags, e)
		}
		_ = rows3.Err()
	}

	// Top Phases
	rows4, err := pool.Query(ctx, `
		SELECT phase, COUNT(*) as cnt FROM waf_events
		WHERE phase IS NOT NULL AND phase != ''
		GROUP BY phase ORDER BY cnt DESC`)
	if err == nil {
		defer rows4.Close()
		for rows4.Next() {
			var e models.TopEntry
			_ = rows4.Scan(&e.Label, &e.Count)
			s.TopPhases = append(s.TopPhases, e)
		}
		_ = rows4.Err()
	}

	// Anomaly Score stats
	_ = pool.QueryRow(ctx, "SELECT COALESCE(AVG(anomaly_score::float), 0), COALESCE(MAX(anomaly_score), 0) FROM waf_events WHERE anomaly_score > 0").Scan(&s.AvgAnomalyScore, &s.MaxAnomalyScore)

	// Score distribution
	type scoreRange struct {
		label string
		min   int
		max   int // -1 = no upper bound
	}
	ranges := []scoreRange{
		{"0-5", 1, 5},
		{"6-10", 6, 10},
		{"11-15", 11, 15},
		{"16-25", 16, 25},
		{"25+", 26, -1},
	}
	for _, sr := range ranges {
		var cnt int64
		var qErr error
		if sr.max == -1 {
			qErr = pool.QueryRow(ctx, "SELECT COUNT(*) FROM waf_events WHERE anomaly_score >= $1", sr.min).Scan(&cnt)
		} else {
			qErr = pool.QueryRow(ctx, "SELECT COUNT(*) FROM waf_events WHERE anomaly_score >= $1 AND anomaly_score <= $2", sr.min, sr.max).Scan(&cnt)
		}
		if qErr == nil {
			s.ScoreDistribution = append(s.ScoreDistribution, models.ScoreBucket{Range: sr.label, Count: cnt})
		}
	}

	// Events per hour (last 24h)
	rows5, err := pool.Query(ctx, `
		SELECT date_trunc('hour', timestamp) as hr, COUNT(*) as cnt FROM waf_events
		WHERE timestamp >= NOW() - INTERVAL '24 hours'
		GROUP BY hr ORDER BY hr`)
	if err == nil {
		defer rows5.Close()
		for rows5.Next() {
			var b models.HourBucket
			_ = rows5.Scan(&b.Hour, &b.Count)
			s.EventsPerHour = append(s.EventsPerHour, b)
		}
		_ = rows5.Err()
	}

	return &s, nil
}

// GetRulesConfig fetches the current rules configuration (row id=1).
func GetRulesConfig(ctx context.Context, pool *pgxpool.Pool) (*models.RulesConfig, error) {
	cfg := &models.RulesConfig{}
	err := pool.QueryRow(ctx, `
		SELECT engine_mode, paranoia_level, paranoia_level_enabled, inbound_threshold, outbound_threshold, disabled_rule_ids, disabled_tags, response_check
		FROM rules_config WHERE id = 1
	`).Scan(&cfg.EngineMode, &cfg.ParanoiaLevel, &cfg.ParanoiaLevelEnabled, &cfg.InboundThreshold, &cfg.OutboundThreshold,
		&cfg.DisabledRuleIds, &cfg.DisabledTags, &cfg.ResponseCheck)
	if err != nil {
		return nil, err
	}
	if cfg.DisabledRuleIds == nil {
		cfg.DisabledRuleIds = []string{}
	}
	if cfg.DisabledTags == nil {
		cfg.DisabledTags = []string{}
	}
	return cfg, nil
}

// SaveRulesConfig saves the rules configuration (upsert on id=1).
func SaveRulesConfig(ctx context.Context, pool *pgxpool.Pool, cfg *models.RulesConfig) error {
	_, err := pool.Exec(ctx, `
		INSERT INTO rules_config (id, engine_mode, paranoia_level, paranoia_level_enabled, inbound_threshold, outbound_threshold,
		    disabled_rule_ids, disabled_tags, response_check, updated_at)
		VALUES (1, $1, $2, $3, $4, $5, $6, $7, $8, NOW())
		ON CONFLICT (id) DO UPDATE
		SET engine_mode=$1, paranoia_level=$2, paranoia_level_enabled=$3, inbound_threshold=$4, outbound_threshold=$5,
		    disabled_rule_ids=$6, disabled_tags=$7, response_check=$8, updated_at=NOW()
	`, cfg.EngineMode, cfg.ParanoiaLevel, cfg.ParanoiaLevelEnabled, cfg.InboundThreshold, cfg.OutboundThreshold,
		cfg.DisabledRuleIds, cfg.DisabledTags, cfg.ResponseCheck)
	return err
}
