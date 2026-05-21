package db

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Connect(ctx context.Context) (*pgxpool.Pool, error) {
	host := getEnv("DB_HOST", "postgres")
	port := getEnv("DB_PORT", "5432")
	name := getEnv("DB_NAME", "coraza_dashboard")
	user := getEnv("DB_USER", "coraza")
	pass := getEnv("DB_PASSWORD", "changeme")

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", user, pass, host, port, name)
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}
	return pgxpool.NewWithConfig(ctx, cfg)
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

const schema = `
CREATE TABLE IF NOT EXISTS rules_config (
    id                    SERIAL PRIMARY KEY,
    updated_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    engine_mode           TEXT NOT NULL DEFAULT 'On',
    paranoia_level        INT NOT NULL DEFAULT 1,
    paranoia_level_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    inbound_threshold     INT NOT NULL DEFAULT 5,
    outbound_threshold    INT NOT NULL DEFAULT 4,
    disabled_rule_ids     TEXT[] NOT NULL DEFAULT '{}',
    disabled_tags         TEXT[] NOT NULL DEFAULT '{}',
    response_check        BOOLEAN NOT NULL DEFAULT FALSE
);
INSERT INTO rules_config (id) VALUES (1) ON CONFLICT DO NOTHING;
ALTER TABLE rules_config ADD COLUMN IF NOT EXISTS paranoia_level_enabled BOOLEAN NOT NULL DEFAULT TRUE;
ALTER TABLE rules_config ADD COLUMN IF NOT EXISTS response_check BOOLEAN NOT NULL DEFAULT FALSE;

CREATE TABLE IF NOT EXISTS waf_events (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    timestamp   TIMESTAMPTZ NOT NULL,
    client_ip   TEXT,
    server      TEXT,
    uri         TEXT,
    rule_id     INTEGER,
    rule_msg    TEXT,
    rule_file   TEXT,
    severity    TEXT,
    severity_id INTEGER,
    phase       TEXT,
    phase_id    INTEGER,
    disruptive  BOOLEAN NOT NULL DEFAULT FALSE,
    tags        TEXT[],
    data        TEXT,
    unique_id   TEXT,
    raw_log     JSONB,
    anomaly_score INTEGER NOT NULL DEFAULT 0,
    block_type    VARCHAR(10) NOT NULL DEFAULT ''
);
CREATE INDEX IF NOT EXISTS idx_waf_events_timestamp  ON waf_events(timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_waf_events_client_ip  ON waf_events(client_ip);
CREATE INDEX IF NOT EXISTS idx_waf_events_disruptive ON waf_events(disruptive);
CREATE INDEX IF NOT EXISTS idx_waf_events_rule_id    ON waf_events(rule_id);
ALTER TABLE waf_events ADD COLUMN IF NOT EXISTS anomaly_score INTEGER NOT NULL DEFAULT 0;
ALTER TABLE waf_events ADD COLUMN IF NOT EXISTS block_type VARCHAR(10) NOT NULL DEFAULT '';
DO $$ BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'waf_events_unique_id_key'
    ) THEN
        -- Remove duplicates before adding unique constraint (keep lowest id)
        DELETE FROM waf_events w1
        USING waf_events w2
        WHERE w1.id > w2.id AND w1.unique_id = w2.unique_id AND w1.unique_id IS NOT NULL AND w1.unique_id != '';
        ALTER TABLE waf_events ADD CONSTRAINT waf_events_unique_id_key UNIQUE (unique_id);
    END IF;
END $$;
`

func Migrate(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, schema)
	return err
}
