package db

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Pool is the global database connection pool.
var Pool *pgxpool.Pool

// Connect establishes the database connection pool.
func Connect(ctx context.Context) error {
	dsn := fmt.Sprintf(
		"host=%s port=%s dbname=%s user=%s password=%s sslmode=disable",
		getenv("DB_HOST", "localhost"),
		getenv("DB_PORT", "5432"),
		getenv("DB_NAME", "coraza_dashboard"),
		getenv("DB_USER", "coraza"),
		getenv("DB_PASSWORD", "changeme"),
	)

	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return fmt.Errorf("parse db config: %w", err)
	}

	config.MaxConns = 20
	config.MinConns = 2
	config.MaxConnIdleTime = 5 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return fmt.Errorf("create pool: %w", err)
	}

	// Verify connectivity
	if err := pool.Ping(ctx); err != nil {
		return fmt.Errorf("ping db: %w", err)
	}

	Pool = pool
	log.Println("database connected")
	return nil
}

// Migrate runs the database migration to create required tables.
func Migrate(ctx context.Context) error {
	_, err := Pool.Exec(ctx, schema)
	if err != nil {
		return fmt.Errorf("migrate: %w", err)
	}
	log.Println("database migration complete")
	return nil
}

// Close closes the connection pool.
func Close() {
	if Pool != nil {
		Pool.Close()
	}
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

const schema = `
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE IF NOT EXISTS waf_events (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    timestamp   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    client_ip   TEXT NOT NULL DEFAULT '',
    method      TEXT NOT NULL DEFAULT '',
    uri         TEXT NOT NULL DEFAULT '',
    rule_id     TEXT NOT NULL DEFAULT '',
    rule_msg    TEXT NOT NULL DEFAULT '',
    severity    TEXT NOT NULL DEFAULT '',
    action      TEXT NOT NULL DEFAULT 'detect',
    raw_log     JSONB NOT NULL DEFAULT '{}'
);

CREATE INDEX IF NOT EXISTS idx_waf_events_timestamp  ON waf_events (timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_waf_events_client_ip  ON waf_events (client_ip);
CREATE INDEX IF NOT EXISTS idx_waf_events_action     ON waf_events (action);
CREATE INDEX IF NOT EXISTS idx_waf_events_rule_id    ON waf_events (rule_id);
`
