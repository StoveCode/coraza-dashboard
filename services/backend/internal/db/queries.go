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
	_ = pool.QueryRow(ctx, "SELECT COUNT(*) FROM waf_events WHERE disruptive=true").Scan(&s.TotalBlocks)
	_ = pool.QueryRow(ctx, "SELECT COUNT(*) FROM waf_events WHERE disruptive=false").Scan(&s.TotalDetections)

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
