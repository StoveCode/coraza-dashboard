# coraza-dashboard

Logging & Monitoring Dashboard für **OWASP Coraza WAF** + **HAProxy**.

Zeigt Block-Events, Traffic-Statistiken und WAF-Aktivitäten in einem Dark-Theme Web-Dashboard.

## Architektur

```
[Client]
   │
   ▼
[HAProxy :80]  ──── SPOE ────▶  [coraza-spoa :9000]
                                       │
                              JSON log lines (zerolog)
                                       │
                              /var/log/coraza/coraza.log
                              (shared Docker Volume)
                                       │
                                       ▼
                              [Backend :8080] (Go)
                              Log-Tailer + REST API
                                       │
                              [PostgreSQL :5432]
                                       │
                                       ▼
                              [Frontend :3000] (Vue 3)
```

## Services

| Service       | Port  | Beschreibung                                  |
|---------------|-------|-----------------------------------------------|
| `haproxy`     | 80    | Reverse Proxy mit Coraza SPOE-Filter          |
| `haproxy`     | 8404  | HAProxy Stats-Seite (admin/changeme)          |
| `coraza-spoa` | 9000  | OWASP Coraza WAF Agent (SPOE)                 |
| `backend`     | 8080  | Go REST API + Log-Tailer                      |
| `frontend`    | 3000  | Vue 3 Dashboard                               |
| `httpbin`     | 8081  | Echo-Backend (für Tests)                      |
| `postgres`    | 5432  | PostgreSQL Datenbank                          |

## Quick Start

```bash
# 1. Repo klonen
git clone https://github.com/StoveCode/coraza-dashboard.git
cd coraza-dashboard

# 2. Env-Datei anlegen
cp .env.example .env
# Optional: Passwörter in .env anpassen

# 3. coraza-spoa lokal bauen (kein öffentliches Image verfügbar)
git clone https://github.com/corazawaf/coraza-spoa.git /tmp/coraza-spoa-source
docker build -f /tmp/coraza-spoa-source/ftw/Dockerfile.coraza_spoa \
  -t coraza-spoa:local /tmp/coraza-spoa-source

# 4. Starten
docker compose up -d --build

# 5. Dashboard öffnen
open http://localhost:3000
```

## URLs

| URL                              | Beschreibung                   |
|----------------------------------|--------------------------------|
| http://localhost:3000            | WAF Dashboard                  |
| http://localhost:8080/api/health | Backend Health Check           |
| http://localhost:8080/api/stats  | Aggregierte Stats (JSON)       |
| http://localhost:8404/stats      | HAProxy Stats (admin/changeme) |
| http://localhost:8081            | httpbin Echo-Backend           |

## Traffic-Generator

Test-Script das zufälligen legitimen und bösartigen Traffic generiert:

```bash
# Requests mit 50% Angriffen
python3 scripts/traffic-gen.py --rate 1 --ratio 0.5

# Optionen
python3 scripts/traffic-gen.py --help
#   --target  Ziel-URL (default: http://localhost:80)
#   --rate    Requests pro Sekunde (default: 1.0)
#   --ratio   Anteil Angriffs-Requests (default: 0.4)
```

Angriffsvektoren: SQLi, XSS, LFI, Path Traversal, RCE, SSRF, Scanner-UAs, Recon

## API-Referenz (Backend)

| Method | Pfad            | Beschreibung                                        |
|--------|-----------------|-----------------------------------------------------|
| GET    | /api/health     | Health Check                                        |
| GET    | /api/events     | WAF-Events (paginated, filterbar)                   |
| GET    | /api/stats      | Aggregierte Statistiken                             |
| GET    | /api/metrics    | Prometheus-Metriken von coraza-spoa (wenn aktiv)    |

### GET /api/events Parameter

| Parameter    | Typ     | Beschreibung                        |
|--------------|---------|-------------------------------------|
| `limit`      | int     | Anzahl Einträge (default: 50)       |
| `offset`     | int     | Paginierung-Offset                  |
| `from`       | RFC3339 | Von-Zeitstempel                     |
| `to`         | RFC3339 | Bis-Zeitstempel                     |
| `disruptive` | bool    | true=Blocks, false=Detections       |
| `client_ip`  | string  | Filter nach Client-IP               |

### GET /api/stats Response

```json
{
  "total_blocks": 1234,
  "total_detections": 56,
  "top_ips": [{"label": "1.2.3.4", "count": 42}],
  "top_rules": [{"rule_id": 941100, "msg": "XSS Attack...", "count": 30}],
  "top_tags": [{"label": "attack-xss", "count": 152}],
  "top_phases": [{"label": "request-body", "count": 535}],
  "events_per_hour": [{"hour": "2024-01-01T00:00:00Z", "count": 10}]
}
```

## Konfiguration (.env)

```env
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

# Optional: Prometheus-Scraping von coraza-spoa
# CORAZA_METRICS_URL=http://coraza-spoa:9090/metrics
```

## OWASP CRS v4

coraza-spoa lädt automatisch das komplette **OWASP Core Rule Set v4** (`@owasp_crs/*.conf`).
Konfiguration in `services/coraza/coraza-spoa.yaml`.

## Log-Format

coraza-spoa schreibt **zerolog JSON Lines** in das shared Volume `/var/log/coraza/coraza.log`.
Der Backend-Tailer verarbeitet nur Zeilen mit einem `match`-Feld:

```json
{
  "level": "error",
  "match": {
    "client": "1.2.3.4",
    "rule_id": 941100,
    "msg": "XSS Attack Detected via libinjection",
    "severity": "CRITICAL",
    "disruptive": true,
    "uri": "/search?q=<script>alert(1)</script>",
    "phase": "request-body",
    "tags": ["attack-xss", "OWASP_CRS"]
  },
  "time": "2024-01-01T00:00:00Z"
}
```

`disruptive: true` = Block, `disruptive: false` = Detection

## OWASP CRS Ruleset anpassen

Alle Änderungen in `services/coraza/coraza-spoa.yaml` unter `directives:`.
Nach jeder Änderung: `docker compose restart coraza-spoa` — kein Rebuild nötig.

### Paranoia Level erhöhen

Das CRS hat 4 Paranoia-Level (PL1 = Standard, PL4 = sehr streng).
Höheres Level = mehr Rules aktiv = mehr False Positives möglich.

```yaml
directives: |
  Include @coraza.conf-recommended
  Include @crs-setup.conf.example

  # Paranoia Level 2 aktivieren (Standard ist 1)
  SecAction \
    "id:900000,\
    phase:1,\
    nolog,\
    pass,\
    t:none,\
    setvar:tx.blocking_paranoia_level=2"

  Include @owasp_crs/*.conf
  SecRuleEngine On
```

### Anomaly Score Threshold anpassen

CRS arbeitet mit Anomaly Scoring — erst wenn der Score einen Schwellwert überschreitet, wird geblockt.
Default: Inbound = 5, Outbound = 4.

```yaml
directives: |
  Include @coraza.conf-recommended
  Include @crs-setup.conf.example

  # Threshold erhöhen = weniger Blocks (toleranter)
  SecAction \
    "id:900110,\
    phase:1,\
    nolog,\
    pass,\
    t:none,\
    setvar:tx.inbound_anomaly_score_threshold=10,\
    setvar:tx.outbound_anomaly_score_threshold=10"

  Include @owasp_crs/*.conf
  SecRuleEngine On
```

### Einzelne Rules deaktivieren

```yaml
directives: |
  Include @coraza.conf-recommended
  Include @crs-setup.conf.example
  Include @owasp_crs/*.conf
  SecRuleEngine On

  # Rule 920350 deaktivieren (Host Header mit IP)
  SecRuleRemoveById 920350

  # Alle Rules mit Tag "attack-sqli" deaktivieren
  SecRuleRemoveByTag "attack-sqli"
```

### Nur Detection-Modus (kein Blocking)

```yaml
directives: |
  Include @coraza.conf-recommended
  Include @crs-setup.conf.example
  Include @owasp_crs/*.conf

  # DetectionOnly = loggt, blockt aber nicht
  SecRuleEngine DetectionOnly
```

### Eigene Rules hinzufügen

```yaml
directives: |
  Include @coraza.conf-recommended
  Include @crs-setup.conf.example
  Include @owasp_crs/*.conf
  SecRuleEngine On

  # Eigene Rule: Block Requests mit bestimmtem User-Agent
  SecRule REQUEST_HEADERS:User-Agent "bad-bot" \
    "id:1000001,\
    phase:1,\
    deny,\
    status:403,\
    msg:'Bad Bot blocked'"
```

### Änderungen anwenden

```bash
# Nur coraza-spoa neu starten (kein Rebuild)
docker compose restart coraza-spoa

# Logs prüfen
docker compose logs -f coraza-spoa
```

## Projektstruktur

```
coraza-dashboard/
  docker-compose.yml       — Alle Services
  .env.example             — Konfigurationsvorlage
  scripts/
    traffic-gen.py         — Traffic-Generator für Tests
  services/
    backend/               — Go REST API + Log-Tailer
    frontend/              — Vue 3 Dashboard (Vite + Tailwind)
    haproxy/               — HAProxy + SPOE Config
    coraza/                — coraza-spoa Config (OWASP CRS v4)
  agents/                  — Subagent Task-Dateien (Build-History)
```
