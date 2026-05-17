# Backend Agent Task — coraza-dashboard

## Deine Rolle
Du bist Senior Backend Engineer. Du schreibst Code, committest, und meldest dich strukturiert zurück.
Du sprichst NUR mit Rusty (dem PM). Kein direkter Kontakt zu Stefan.

## Projekt
Logging & Monitoring Backend für OWASP Coraza WAF + HAProxy.

## Tech Stack
- **Sprache:** Go (aktuell stable)
- **DB:** PostgreSQL (via `pgx` oder `database/sql`)
- **Framework:** Minimalistisch — `net/http` oder `chi` Router
- **Config:** Alle Werte via Umgebungsvariablen (kein Hardcoding!)
- **Docker:** Multi-stage Build, non-root User

## Arbeitsverzeichnis
`/home/ubuntu/.openclaw/workspace/coraza-dashboard/`

## Was zu tun ist

### 1. Go-Modul initialisieren
```
services/backend/
  main.go
  go.mod / go.sum
  internal/
    api/       — REST Handler
    collector/ — HTTP Log-Collector (empfängt Coraza Audit Logs)
    db/        — PostgreSQL Verbindung + Migrations
    models/    — Event, Stats Structs
  Dockerfile
```

### 2. Datenbank-Schema (PostgreSQL)
Tabelle `waf_events`:
- id (UUID, PK)
- timestamp (timestamptz)
- client_ip (text)
- method (text)
- uri (text)
- rule_id (text)
- rule_msg (text)
- severity (text)
- action (text: block/detect)
- raw_log (jsonb)

Tabelle `waf_stats_cache` (optional, für schnelle Dashboard-Queries):
- oder alternativ direkt via SQL-Aggregation

### 3. REST API Endpoints
- `GET /api/health` — Health Check
- `GET /api/events` — Liste Events (Query-Params: limit, offset, from, to, action, client_ip)
- `GET /api/stats` — Aggregierte Stats:
  - Top-IPs (blockiert)
  - Top-Rule-IDs
  - Events pro Stunde (letzte 24h)
  - Total blocks / detections

### 4. Log-Collector Endpoint
- `POST /api/log` — Empfängt Coraza Audit Log JSON
- Parsed das JSON und speichert in waf_events
- Coraza Audit Log Format: https://coraza.io/docs/audit-log/

**Wichtig:** Der Collector muss robust sein — bei Parse-Fehlern loggen und trotzdem 200 zurückgeben (damit Coraza nicht blockiert)

### 5. CORS
- CORS-Header setzen damit das Frontend (anderer Port) zugreifen kann
- Origin via Env-Var konfigurierbar: `CORS_ALLOWED_ORIGINS`

### 6. Env-Vars (alle zwingend via .env / OS Env)
```
DB_HOST=postgres
DB_PORT=5432
DB_NAME=coraza_dashboard
DB_USER=coraza
DB_PASSWORD=changeme
SERVER_PORT=8080
CORS_ALLOWED_ORIGINS=http://localhost:3000
LOG_LEVEL=info
```

### 7. Dockerfile
```dockerfile
# Multi-stage
FROM golang:1.23-alpine AS builder
...
FROM alpine:latest
# non-root user
USER nobody
CMD ["/app/backend"]
```

### 8. docker-compose.yml im Projekt-Root
Erstelle `docker-compose.yml` und `.env.example` im Root:
```
coraza-dashboard/
  docker-compose.yml
  .env.example
  services/
    backend/
    frontend/   (Placeholder, kommt später)
    haproxy/    (Demo-Config)
    coraza/     (Demo-Config)
```

docker-compose.yml soll enthalten:
- `postgres` Service (image: postgres:16-alpine)
- `backend` Service (build: ./services/backend)
- Volumes für postgres-data
- healthchecks
- alle Ports + Env-Vars aus .env

### 9. Demo HAProxy + Coraza Config
In `services/haproxy/haproxy.cfg` eine funktionierende Demo-Config mit SPOE-Verweis.
In `services/coraza/` das SPOE-Konfigurationsfile und eine Beispiel-Coraza-Config (coraza.conf).
Diese sollten zeigen WIE man Coraza mit HAProxy verbindet und Logs an das Backend sendet.

## Output-Format (PFLICHT am Ende)
```
STATUS: done | failed
FILES: alle erstellten/geänderten Dateien
PROBLEMS: keine | Liste
NOTES: Was der Frontend-Agent wissen muss (API-Endpoints, Datenstrukturen)
FEEDBACK: Verbesserungsvorschläge an Rusty
```

## Git
- Initialisiere ein Git-Repo in `coraza-dashboard/` falls nicht vorhanden
- Committe am Ende alles mit sinnvollen Commit-Messages
- Kein Push — Rusty macht das
