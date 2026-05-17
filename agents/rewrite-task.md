# Rewrite Task — coraza-dashboard v2

## Kontext: Was in v1 falsch war
Die erste Version hat angenommen, dass coraza-spoa HTTP-Webhooks sendet (POST /api/log).
Das ist **falsch**. Basierend auf dem echten Source Code:

### Wie coraza-spoa WIRKLICH Logs schreibt
- coraza-spoa schreibt **zerolog JSON Lines** auf stdout oder in eine Datei
- Konfiguriert via `log_file` + `log_format: json` in der coraza-spoa.yaml
- Jede gematchte Rule erzeugt eine Log-Zeile:

```json
{
  "level": "error",
  "match": {
    "client": "1.2.3.4",
    "file": "/path/to/rule.conf",
    "line": 123,
    "rule_id": 941100,
    "revision": "",
    "msg": "XSS Attack Detected via libinjection",
    "data": "...",
    "severity": "CRITICAL",
    "severity_id": 2,
    "version": "...",
    "maturity": 0,
    "accuracy": 0,
    "tags": ["attack-xss", "OWASP_CRS"],
    "server": "10.0.0.1",
    "uri": "/search?q=<script>",
    "unique_id": "ABCDEFGHIJKLMNOP",
    "disruptive": true,
    "phase_id": 2,
    "phase": "request-body"
  },
  "time": "2024-01-01T00:00:00Z"
}
```

- `disruptive: true` = Block, `disruptive: false` = Detect (das ist der echte Indikator!)
- coraza-spoa hat optional einen Prometheus `/metrics` Endpoint via `--metrics-addr` Flag
  - Metric: `coraza_handle_spoe_duration_seconds` (Histogram)

### Korrekte Architektur

```
[HAProxy] --SPOE--> [coraza-spoa]
                        |
                   JSON log lines → /var/log/coraza/coraza.log (shared volume)
                        |
                   optional: --metrics-addr → /metrics (Prometheus)
                        |
                   [Backend: log-tailer + Prometheus scraper]
                        |
                   [PostgreSQL]
                        |
                   [Frontend Dashboard]
```

## Deine Aufgabe: Komplettes Rewrite

Lösche alles unter `services/backend/` und `services/frontend/` und schreibe alles neu.
Docker-compose.yml und .env.example ebenfalls neu.

## Arbeitsverzeichnis
`/home/ubuntu/.openclaw/workspace/coraza-dashboard/`

---

## BACKEND REWRITE (services/backend/)

### Tech Stack (unverändert)
- Go, `chi` Router, `pgx/v5`, Multi-stage Dockerfile

### Struktur
```
services/backend/
  main.go
  go.mod / go.sum
  Dockerfile
  internal/
    models/models.go      — WAFEvent, Stats structs
    db/db.go              — pgxpool connect + migrate
    db/queries.go         — Insert, List, Stats queries
    api/handlers.go       — REST Handler
    tailer/tailer.go      — Log-File Tailer (NEU, ersetzt collector)
    scraper/scraper.go    — Prometheus Metrics Scraper (NEU, optional)
```

### models.go — WAFEvent
Basierend auf dem echten `matchedRuleErrorJson` Struct:
```go
type WAFEvent struct {
    ID         string    `json:"id"`
    Timestamp  time.Time `json:"timestamp"`
    // aus "match" Objekt:
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
    Disruptive bool      `json:"disruptive"`   // true=block, false=detect
    Tags       []string  `json:"tags"`
    Data       string    `json:"data"`
    UniqueID   string    `json:"unique_id"`
    RawLog     []byte    `json:"raw_log"`      // original JSON line
}
```

### PostgreSQL Schema (waf_events Tabelle)
```sql
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
    raw_log     JSONB
);
CREATE INDEX IF NOT EXISTS idx_waf_events_timestamp ON waf_events(timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_waf_events_client_ip ON waf_events(client_ip);
CREATE INDEX IF NOT EXISTS idx_waf_events_disruptive ON waf_events(disruptive);
CREATE INDEX IF NOT EXISTS idx_waf_events_rule_id ON waf_events(rule_id);
```

### tailer/tailer.go — Log-File Tailer
- Liest `/var/log/coraza/coraza.log` (oder Pfad via `LOG_FILE` Env-Var)
- Nutzt `tail` package (`github.com/nxadm/tail`) oder einfaches polling mit `os.Seek`
- Empfehlung: `github.com/nxadm/tail` mit `tail.Config{Follow: true, ReOpen: true, MustExist: false}`
- Parsed jede Zeile: `{"level":"...","match":{...},"time":"..."}`
- Nur Zeilen mit `"match"` Feld verarbeiten (andere sind normale Info/Debug Logs)
- Mapped `match` auf `WAFEvent` Struct
- Inserted in PostgreSQL via `db.InsertEvent()`
- Bei Parse-Fehler: loggen, weiter machen — niemals crashen

```go
type logLine struct {
    Level string          `json:"level"`
    Match *matchedRule    `json:"match"`
    Time  time.Time       `json:"time"`
}

type matchedRule struct {
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
```

### scraper/scraper.go — Prometheus Scraper (optional)
- Scrapt `CORAZA_METRICS_URL` Env-Var (z.B. `http://coraza:9090/metrics`) alle 15s
- Parsed `coraza_handle_spoe_duration_seconds` Bucket-Daten
- Speichert als `scrape_metrics` Tabelle ODER gibt direkt via `/api/metrics` weiter
- Wenn `CORAZA_METRICS_URL` leer → Scraper deaktiviert (graceful degradation)

### REST API (unverändert, aber `action` field angepasst)
- `GET /api/health`
- `GET /api/events?limit=50&offset=0&from=&to=&disruptive=true|false&client_ip=`
  - `disruptive` statt `action` als Filter-Parameter
  - Response: `{ "total": int, "events": [WAFEvent] }`
- `GET /api/stats`
  - `total_blocks` (disruptive=true), `total_detections` (disruptive=false)
  - `top_ips`, `top_rules` (rule_id + msg + count)
  - `events_per_hour` (letzte 24h)
  - `top_tags` (welche CRS-Tags am häufigsten, z.B. "attack-xss", "attack-sqli")
  - `top_phases` (request-headers, request-body, etc.)
- `GET /api/metrics` — rohe Prometheus-Daten von coraza-spoa weiterleiten (wenn Scraper aktiv)

### Env-Vars
```
# Database
DB_HOST=postgres
DB_PORT=5432
DB_NAME=coraza_dashboard
DB_USER=coraza
DB_PASSWORD=changeme

# Server
SERVER_PORT=8080
CORS_ALLOWED_ORIGINS=http://localhost:3000

# Log Tailer
LOG_FILE=/var/log/coraza/coraza.log

# Prometheus Scraper (optional)
CORAZA_METRICS_URL=http://coraza-spoa:9090/metrics

LOG_LEVEL=info
```

---

## FRONTEND REWRITE (services/frontend/)

### Tech Stack (unverändert)
Vue 3 + TypeScript + Vite + Pinia + Tailwind CSS + Chart.js

### Was sich ändert
- Filter: `disruptive` statt `action` — "Blocked" (disruptive=true) / "Detected" (disruptive=false)
- Neue Charts: **Top Tags** (attack-xss, attack-sqli etc.) + **Top Phases**
- EventsTable: neue Columns (Phase, Tags, Data)
- Stats: `total_blocks` + `total_detections` kommen aus dem Backend

### Struktur (gleich wie v1, aber angepasste Komponenten)
```
services/frontend/src/
  api/
    client.ts
    events.ts    — GET /api/events mit disruptive param
    stats.ts     — GET /api/stats
  stores/
    events.ts
    stats.ts
  views/
    Dashboard.vue   — 4 StatCards + 4 Charts
    Events.vue      — Tabelle + Filter
  components/
    StatCard.vue
    EventsTable.vue — angepasste Columns
    FilterBar.vue   — disruptive filter
    TimelineChart.vue
    TopIPsChart.vue
    TopRulesChart.vue
    TopTagsChart.vue   — NEU: OWASP CRS Tag-Kategorien
    PhaseChart.vue     — NEU: Pie/Donut - Request/Response Phase Verteilung
```

### Dashboard Layout
```
[Blocked: 1234]  [Detected: 567]  [Unique IPs: 89]  [Top Rule: 941100]
[                Timeline (Events/Stunde letzte 24h)                  ]
[   Top IPs (Bar)    ]  [  Top Rules (Horizontal Bar)   ]
[  Top Tags (Bar)    ]  [  Phase Distribution (Donut)   ]
```

---

## DOCKER COMPOSE REWRITE

```yaml
services:

  postgres:
    image: postgres:16-alpine
    environment:
      POSTGRES_DB: ${DB_NAME:-coraza_dashboard}
      POSTGRES_USER: ${DB_USER:-coraza}
      POSTGRES_PASSWORD: ${DB_PASSWORD:-changeme}
    volumes:
      - postgres_data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U ${DB_USER:-coraza}"]
      interval: 5s
      timeout: 5s
      retries: 10

  coraza-spoa:
    image: ghcr.io/corazawaf/coraza-spoa:latest
    command: ["-config", "/etc/coraza-spoa/coraza-spoa.yaml"]
    volumes:
      - ./services/coraza/coraza-spoa.yaml:/etc/coraza-spoa/coraza-spoa.yaml:ro
      - coraza_logs:/var/log/coraza    # shared log volume!
    ports:
      - "${CORAZA_SPOE_PORT:-9000}:9000"
    # Prometheus metrics optional:
    # command: ["-config", "/etc/coraza-spoa/coraza-spoa.yaml", "-metrics-addr", "0.0.0.0:9090"]

  haproxy:
    image: haproxy:2.9-alpine
    volumes:
      - ./services/haproxy/haproxy.cfg:/usr/local/etc/haproxy/haproxy.cfg:ro
      - ./services/haproxy/coraza-spoe.cfg:/etc/haproxy/coraza.cfg:ro
    ports:
      - "${HAPROXY_PORT:-80}:80"
      - "${HAPROXY_STATS_PORT:-8404}:8404"
    depends_on:
      - coraza-spoa

  backend:
    build: ./services/backend
    environment:
      DB_HOST: postgres
      DB_PORT: 5432
      DB_NAME: ${DB_NAME:-coraza_dashboard}
      DB_USER: ${DB_USER:-coraza}
      DB_PASSWORD: ${DB_PASSWORD:-changeme}
      SERVER_PORT: 8080
      CORS_ALLOWED_ORIGINS: ${CORS_ALLOWED_ORIGINS:-http://localhost:3000}
      LOG_FILE: /var/log/coraza/coraza.log
      CORAZA_METRICS_URL: ${CORAZA_METRICS_URL:-http://coraza-spoa:9090/metrics}
      LOG_LEVEL: ${LOG_LEVEL:-info}
    volumes:
      - coraza_logs:/var/log/coraza    # same shared log volume!
    ports:
      - "${BACKEND_PORT:-8080}:8080"
    depends_on:
      postgres:
        condition: service_healthy

  frontend:
    build: ./services/frontend
    ports:
      - "${FRONTEND_PORT:-3000}:3000"
    depends_on:
      - backend

volumes:
  postgres_data:
  coraza_logs:
```

### coraza-spoa.yaml anpassen
```yaml
bind: 0.0.0.0:9000
log_level: info
log_file: /var/log/coraza/coraza.log   # WICHTIG: in shared volume!
log_format: json                         # WICHTIG: JSON format!

applications:
  - name: sample_app
    directives: |
      Include @coraza.conf-recommended
      Include @crs-setup.conf.example
      Include @owasp_crs/*.conf
      SecRuleEngine On
    response_check: false
    transaction_ttl_ms: 60000
    log_level: info
    log_file: /var/log/coraza/coraza.log
    log_format: json
```

### .env.example
```
# PostgreSQL
DB_NAME=coraza_dashboard
DB_USER=coraza
DB_PASSWORD=changeme

# Ports
HAPROXY_PORT=80
HAPROXY_STATS_PORT=8404
CORAZA_SPOE_PORT=9000
BACKEND_PORT=8080
FRONTEND_PORT=3000

# Backend
CORS_ALLOWED_ORIGINS=http://localhost:3000
LOG_LEVEL=info

# Optional: Prometheus scraping von coraza-spoa
# coraza-spoa mit --metrics-addr 0.0.0.0:9090 starten
# CORAZA_METRICS_URL=http://coraza-spoa:9090/metrics
```

---

## Git
- `git rm -r services/backend services/frontend` (alles weg)
- Neu committen
- Commit Message: `feat: rewrite backend/frontend based on actual coraza-spoa log format`

## Output-Format (PFLICHT)
```
STATUS: done | failed
FILES: alle erstellten/geänderten Dateien
PROBLEMS: keine | Liste
NOTES: Was geändert wurde vs v1, wie der Log-Tailer funktioniert
FEEDBACK: Verbesserungsvorschläge
```
