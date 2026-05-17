package db

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/stovecode/coraza-dashboard/backend/internal/models"
)

// InsertEvent inserts a WAFEvent into the database.
func InsertEvent(ctx context.Context, e *models.WAFEvent) error {
	if e.ID == uuid.Nil {
		e.ID = uuid.New()
	}
	if e.Timestamp.IsZero() {
		e.Timestamp = time.Now().UTC()
	}

	rawJSON, err := json.Marshal(e.RawLog)
	if err != nil {
		rawJSON = []byte("{}")
	}

	_, err = Pool.Exec(ctx, `
		INSERT INTO waf_events (id, timestamp, client_ip, method, uri, rule_id, rule_msg, severity, action, raw_log)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`, e.ID, e.Timestamp, e.ClientIP, e.Method, e.URI, e.RuleID, e.RuleMsg, e.Severity, e.Action, rawJSON)
	return err
}

// ListEvents returns a paginated, filtered list of WAF events.
func ListEvents(ctx context.Context, f models.EventFilter) ([]models.WAFEvent, int64, error) {
	where := []string{}
	args := []interface{}{}
	idx := 1

	if f.From != nil {
		where = append(where, fmt.Sprintf("timestamp >= $%d", idx))
		args = append(args, *f.From)
		idx++
	}
	if f.To != nil {
		where = append(where, fmt.Sprintf("timestamp <= $%d", idx))
		args = append(args, *f.To)
		idx++
	}
	if f.Action != "" {
		where = append(where, fmt.Sprintf("action = $%d", idx))
		args = append(args, f.Action)
		idx++
	}
	if f.ClientIP != "" {
		where = append(where, fmt.Sprintf("client_ip = $%d", idx))
		args = append(args, f.ClientIP)
		idx++
	}

	whereClause := ""
	if len(where) > 0 {
		whereClause = "WHERE " + strings.Join(where, " AND ")
	}

	// Count total
	var total int64
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM waf_events %s", whereClause)
	if err := Pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count events: %w", err)
	}

	// Pagination args
	limit := f.Limit
	if limit <= 0 || limit > 1000 {
		limit = 50
	}
	offset := f.Offset
	if offset < 0 {
		offset = 0
	}

	query := fmt.Sprintf(`
		SELECT id, timestamp, client_ip, method, uri, rule_id, rule_msg, severity, action, raw_log
		FROM waf_events %s
		ORDER BY timestamp DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, idx, idx+1)
	args = append(args, limit, offset)

	rows, err := Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list events: %w", err)
	}
	defer rows.Close()

	events := make([]models.WAFEvent, 0, limit)
	for rows.Next() {
		var e models.WAFEvent
		var rawJSON []byte
		if err := rows.Scan(&e.ID, &e.Timestamp, &e.ClientIP, &e.Method, &e.URI,
			&e.RuleID, &e.RuleMsg, &e.Severity, &e.Action, &rawJSON); err != nil {
			return nil, 0, err
		}
		if err := json.Unmarshal(rawJSON, &e.RawLog); err != nil {
			e.RawLog = map[string]interface{}{}
		}
		events = append(events, e)
	}

	return events, total, rows.Err()
}

// GetStats returns aggregated WAF statistics.
func GetStats(ctx context.Context) (*models.Stats, error) {
	stats := &models.Stats{}

	// Totals
	err := Pool.QueryRow(ctx, `
		SELECT
			COUNT(*) FILTER (WHERE action = 'block'),
			COUNT(*) FILTER (WHERE action = 'detect')
		FROM waf_events
	`).Scan(&stats.TotalBlocks, &stats.TotalDetections)
	if err != nil {
		return nil, fmt.Errorf("totals: %w", err)
	}

	// Top IPs (blocked)
	rows, err := Pool.Query(ctx, `
		SELECT client_ip, COUNT(*) AS cnt
		FROM waf_events
		WHERE action = 'block'
		GROUP BY client_ip
		ORDER BY cnt DESC
		LIMIT 10
	`)
	if err != nil {
		return nil, fmt.Errorf("top ips: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var ip models.IPCount
		if err := rows.Scan(&ip.IP, &ip.Count); err != nil {
			return nil, err
		}
		stats.TopIPs = append(stats.TopIPs, ip)
	}
	rows.Close()

	// Top Rules
	rows, err = Pool.Query(ctx, `
		SELECT rule_id, rule_msg, COUNT(*) AS cnt
		FROM waf_events
		WHERE rule_id != ''
		GROUP BY rule_id, rule_msg
		ORDER BY cnt DESC
		LIMIT 10
	`)
	if err != nil {
		return nil, fmt.Errorf("top rules: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var r models.RuleCount
		if err := rows.Scan(&r.RuleID, &r.RuleMsg, &r.Count); err != nil {
			return nil, err
		}
		stats.TopRules = append(stats.TopRules, r)
	}
	rows.Close()

	// Events per hour (last 24h)
	rows, err = Pool.Query(ctx, `
		SELECT date_trunc('hour', timestamp) AS hour, COUNT(*) AS cnt
		FROM waf_events
		WHERE timestamp >= NOW() - INTERVAL '24 hours'
		GROUP BY hour
		ORDER BY hour ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("events per hour: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var h models.HourCount
		if err := rows.Scan(&h.Hour, &h.Count); err != nil {
			return nil, err
		}
		stats.EventsPerHour = append(stats.EventsPerHour, h)
	}

	return stats, rows.Err()
}
