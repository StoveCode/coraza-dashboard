package db

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/corazawaf/coraza-dashboard/internal/models"
)

// InsertEvent inserts a WAFEvent into the database.
func InsertEvent(ctx context.Context, pool *pgxpool.Pool, e *models.WAFEvent) error {
	_, err := pool.Exec(ctx, `
		INSERT INTO waf_events
			(timestamp, client_ip, server, uri, rule_id, rule_msg, rule_file,
			 severity, severity_id, phase, phase_id, disruptive, tags, data, unique_id, raw_log)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)
	`, e.Timestamp, e.Client, e.Server, e.URI, e.RuleID, e.RuleMsg, e.RuleFile,
		e.Severity, e.SeverityID, e.Phase, e.PhaseID, e.Disruptive,
		e.Tags, e.Data, e.UniqueID, e.RawLog)
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
		where += " AND timestamp >= $" + itoa(i)
		args = append(args, f.From)
		i++
	}
	if !f.To.IsZero() {
		where += " AND timestamp <= $" + itoa(i)
		args = append(args, f.To)
		i++
	}
	if f.Disruptive != nil {
		where += " AND disruptive = $" + itoa(i)
		args = append(args, *f.Disruptive)
		i++
	}
	if f.ClientIP != "" {
		where += " AND client_ip = $" + itoa(i)
		args = append(args, f.ClientIP)
		i++
	}
	if f.RuleID != 0 {
		where += " AND rule_id = $" + itoa(i)
		args = append(args, f.RuleID)
		i++
	}
	if f.Tag != "" {
		where += " AND $" + itoa(i) + " = ANY(tags)"
		args = append(args, f.Tag)
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
		       severity, severity_id, phase, phase_id, disruptive, tags, data, unique_id
		FROM waf_events `+where+`
		ORDER BY timestamp DESC
		LIMIT $`+itoa(i)+` OFFSET $`+itoa(i+1), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []models.WAFEvent
	for rows.Next() {
		var e models.WAFEvent
		if err := rows.Scan(&e.ID, &e.Timestamp, &e.Client, &e.Server, &e.URI,
			&e.RuleID, &e.RuleMsg, &e.RuleFile, &e.Severity, &e.SeverityID,
			&e.Phase, &e.PhaseID, &e.Disruptive, &e.Tags, &e.Data, &e.UniqueID); err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	return &ListResult{Total: total, Events: events}, nil
}

// GetStats returns aggregated statistics.
func GetStats(ctx context.Context, pool *pgxpool.Pool) (*models.Stats, error) {
	var s models.Stats

	// Counts
	_ = pool.QueryRow(ctx, "SELECT COUNT(DISTINCT unique_id) FROM waf_events WHERE rule_id IN (949110, 949111)").Scan(&s.TotalBlocks)
	_ = pool.QueryRow(ctx, "SELECT COUNT(DISTINCT unique_id) FROM waf_events WHERE rule_id NOT IN (949110, 949111)").Scan(&s.TotalDetections)

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
	}

	return &s, nil
}

func itoa(i int) string {
	return strconv(i)
}

func strconv(n int) string {
	if n == 0 {
		return "0"
	}
	b := []byte{}
	for n > 0 {
		b = append([]byte{byte('0' + n%10)}, b...)
		n /= 10
	}
	return string(b)
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
		UPDATE rules_config
		SET engine_mode=$1, paranoia_level=$2, paranoia_level_enabled=$3, inbound_threshold=$4, outbound_threshold=$5,
		    disabled_rule_ids=$6, disabled_tags=$7, response_check=$8, updated_at=NOW()
		WHERE id=1
	`, cfg.EngineMode, cfg.ParanoiaLevel, cfg.ParanoiaLevelEnabled, cfg.InboundThreshold, cfg.OutboundThreshold,
		cfg.DisabledRuleIds, cfg.DisabledTags, cfg.ResponseCheck)
	return err
}
