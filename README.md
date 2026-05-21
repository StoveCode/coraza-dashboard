# coraza-dashboard

Logging & Monitoring Dashboard for **OWASP Coraza WAF** + **HAProxy**.

Displays block events, traffic statistics, and WAF activity in a dark-theme web dashboard — including live configuration of WAF rules without restart.

## Architecture

```
[Client]
   │
   ▼
[HAProxy :80]  ──── SPOE ────▶  [coraza-spoa :9000]
                                       │
                              JSON log lines (stdout)
                                       │
                               Docker fluentd log driver
                                       │
                                       ▼
                              [Fluent Bit :24224]
                              Filters & forwards WAF events
                                       │
                               POST /api/ingest
                                       │
                                       ▼
                              [Backend :8080] (Go)
                              REST API + Ingest Handler
                                       │
                              [PostgreSQL :5432]
                                       │
                                       ▼
                              [Frontend :3000] (Vue 3)
```

## Services

| Service       | Port  | Description                                                 |
|---------------|-------|-------------------------------------------------------------|
| `haproxy`     | 80    | Reverse Proxy with Coraza SPOE filter                       |
| `haproxy`     | 8404  | HAProxy Stats page (admin/changeme)                         |
| `coraza-spoa` | 9000  | OWASP Coraza WAF Agent (SPOE) — CRS v4.25.0                 |
| `fluentbit`   | 24224 | Log forwarder: receives coraza-spoa logs, pushes to backend |
| `backend`     | 8080  | Go REST API + Ingest Handler                                |
| `frontend`    | 3000  | Vue 3 Dashboard                                             |
| `httpbin`     | 8081  | Echo backend (for testing)                                  |
| `postgres`    | 5432  | PostgreSQL database                                         |

## Quick Start

```bash
git clone https://github.com/StoveCode/coraza-dashboard.git
cd coraza-dashboard
sudo bash scripts/install.sh
```

That's it. **No build required** — all images are pulled directly from Docker Hub.

The install script automatically handles:
- Installing Docker + Docker Compose (apt / dnf / pacman)
- Pulling `stove301/coraza-spoa:latest` from Docker Hub
- Creating `.env` from `.env.example`
- Starting the stack + health check

### Manual start (without install script)

```bash
git clone https://github.com/StoveCode/coraza-dashboard.git
cd coraza-dashboard
cp .env.example .env
docker compose up -d
```

Docker Compose pulls all images automatically from Docker Hub — no `docker build` required.

## Docker Hub Images

All images are publicly available on Docker Hub:

| Image                                   | Description                                      |
|-----------------------------------------|--------------------------------------------------|
| `stove301/coraza-dashboard-backend`     | Go REST API + Ingest Handler                     |
| `stove301/coraza-dashboard-frontend`    | Vue 3 Dashboard                                  |
| `stove301/coraza-spoa`                  | OWASP Coraza WAF SPOE Agent (CRS v4.25.0)        |

**Tags:**
- `latest` — current stable version
- `crs-v4.25.0` — coraza-spoa with OWASP CRS v4.25.0

Local builds are still possible — `build:` directives are preserved in `docker-compose.yml` and used via `docker compose build`.

## URLs

| URL                              | Description                    |
|----------------------------------|--------------------------------|
| http://localhost:3000            | WAF Dashboard                  |
| http://localhost:8080/api/health | Backend Health Check           |
| http://localhost:8080/api/stats  | Aggregated stats (JSON)        |
| http://localhost:8404/stats      | HAProxy Stats (admin/changeme) |
| http://localhost:8081            | httpbin Echo Backend           |

## Traffic Generator

Test script that generates random legitimate and malicious traffic:

```bash
# 1 request every 10 seconds, 100% attacks
python3 scripts/traffic-gen.py --rate 0.1 --ratio 1.0

# Options
python3 scripts/traffic-gen.py --help
#   --target  Target URL (default: http://localhost:80)
#   --rate    Requests per second (default: 1.0)
#   --ratio   Fraction of attack requests (default: 0.4)
```

Attack vectors: SQLi, XSS, LFI, Path Traversal, RCE, SSRF, Scanner UAs, Recon

## Screenshots

### Dashboard
![Dashboard](docs/screenshots/2.png)

### Events
![Events](docs/screenshots/1.png)

### Rules Management
![Rules Management](docs/screenshots/3.png)

---

## Dashboard Features

### Charts (compact + expandable)
All charts are displayed in compact mode by default. Click the **⤢ Expand** button to open a full-size modal.

| Chart               | Description                           |
|---------------------|---------------------------------------|
| Events / Hour       | Timeline of the last 24h              |
| Top Client IPs      | Most attacking IPs                    |
| Top Rules           | Most frequently triggered WAF rules   |
| Top Tags            | CRS attack categories (XSS, SQLi...)  |
| Phase Distribution  | Where in the request phase blocks occur |

### Rules Management (`/rules`)
Live WAF configuration without restart:

| Feature                  | Description                                                                                         |
|--------------------------|-----------------------------------------------------------------------------------------------------|
| **Engine Mode**          | `On` / `Detection Only` / `Off`                                                                     |
| **Paranoia Level**       | Level 1–4 (or disabled for manual rule selection)                                                   |
| **Anomaly Thresholds**   | Inbound (request) + outbound (response) score threshold                                             |
| **Response Check**       | Enable/disable Data Leakage Prevention                                                              |
| **CRS Categories**       | SQLi, XSS, RCE, LFI, SSRF, Scanner etc. via toggle                                                 |
| **Rule Picker**          | Disable individual CRS rules via catalog search (ID, name, tag, severity)                           |
| **Orphaned Rules**       | Disabled rules that no longer exist in CRS are marked and can be removed                            |
| **CRS Version**          | Currently loaded CRS version displayed, mismatch warning when dashboard ≠ coraza-spoa              |

Changes → **Save Changes** → backend writes new `coraza-spoa.yaml` → `docker compose restart coraza-spoa`

## API Reference (Backend)

| Method | Path                    | Description                                             |
|--------|-------------------------|---------------------------------------------------------|
| GET    | /api/health             | Health Check                                            |
| GET    | /api/events             | WAF events (paginated, filterable)                      |
| GET    | /api/stats              | Aggregated statistics                                   |
| GET    | /api/metrics            | Prometheus metrics (when CORAZA_METRICS_URL is set)     |
| POST   | /api/ingest             | Log ingest endpoint for Fluent Bit                      |
| GET    | /api/rules/config       | Current rules configuration                             |
| PUT    | /api/rules/config       | Save configuration                                      |
| GET    | /api/rules/categories   | List of CRS categories                                  |
| GET    | /api/rules/catalog      | Full CRS rule catalog (ID, name, tag, severity, PL)     |
| GET    | /api/system/versions    | CRS versions from backend + coraza-spoa                 |

### GET /api/events Parameters

| Parameter    | Type    | Description                         |
|--------------|---------|-------------------------------------|
| `limit`      | int     | Number of entries (default: 50)     |
| `offset`     | int     | Pagination offset                   |
| `from`       | RFC3339 | From timestamp                      |
| `to`         | RFC3339 | To timestamp                        |
| `disruptive` | bool    | true=blocks, false=detections       |
| `client_ip`  | string  | Filter by client IP                 |

### GET /api/stats Response

```json
{
  "total_blocks": 1234,
  "total_inbound_blocks": 1100,
  "total_outbound_blocks": 134,
  "total_detections": 56,
  "avg_anomaly_score": 7.3,
  "max_anomaly_score": 25,
  "score_distribution": [
    {"range": "0-5", "count": 120},
    {"range": "6-10", "count": 45}
  ],
  "top_ips": [{"label": "1.2.3.4", "count": 42}],
  "top_ips_blocked": [{"label": "1.2.3.4", "count": 38}],
  "top_rules": [{"rule_id": 941100, "msg": "XSS Attack...", "count": 30}],
  "top_tags": [{"label": "attack-xss", "count": 152}],
  "top_phases": [{"label": "request-body", "count": 535}],
  "events_per_hour": [{"hour": "2024-01-01T00:00:00Z", "count": 10}]
}
```

## Configuration (.env)

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

# Log ingest via Fluent Bit (no file tailer)
LOG_INGEST_MODE=true

# Optional: Prometheus scraping from coraza-spoa
# CORAZA_METRICS_URL=http://coraza-spoa:9090/metrics
```

## Log Pipeline

coraza-spoa writes **zerolog JSON lines** to **stdout**. Fluent Bit receives the logs via Docker `fluentd` log driver, filters for WAF events (lines with a `"match"` field), and pushes them via HTTP to `/api/ingest`.

```json
{
  "level": "error",
  "match": {
    "client": "1.2.3.4",
    "rule_id": 941100,
    "msg": "XSS Attack Detected via libinjection",
    "severity": "CRITICAL",
    "disruptive": false,
    "uri": "/search?q=<script>alert(1)</script>",
    "phase": "request-body",
    "tags": ["attack-xss", "OWASP_CRS"]
  },
  "time": "2024-01-01T00:00:00Z"
}
```

`disruptive: true` = block, `disruptive: false` = detection (SecRuleEngine DetectionOnly)

## OWASP CRS v4

coraza-spoa automatically loads the **OWASP Core Rule Set v4** (currently: **v4.25.0**). Configuration in `services/coraza/coraza-spoa.yaml`.

### Paranoia Level

```yaml
directives: |
  Include @coraza.conf-recommended
  Include @crs-setup.conf.example

  # Enable PL2 (default: 1)
  SecAction "id:900000,phase:1,nolog,pass,t:none,setvar:tx.blocking_paranoia_level=2"

  Include @owasp_crs/*.conf
  SecRuleEngine On
```

### Anomaly Score Threshold

```yaml
directives: |
  Include @coraza.conf-recommended
  Include @crs-setup.conf.example

  # More lenient: higher threshold
  SecAction "id:900110,phase:1,nolog,pass,t:none,\
    setvar:tx.inbound_anomaly_score_threshold=10,\
    setvar:tx.outbound_anomaly_score_threshold=10"

  Include @owasp_crs/*.conf
  SecRuleEngine On
```

### Disabling Rules

```yaml
directives: |
  Include @coraza.conf-recommended
  Include @crs-setup.conf.example
  Include @owasp_crs/*.conf
  SecRuleEngine On

  SecRuleRemoveById 920350
  SecRuleRemoveByTag "attack-sqli"
```

### Applying Changes

```bash
docker compose restart coraza-spoa
docker compose logs -f coraza-spoa
```

## CRS Update Process

The dashboard shows a **mismatch warning** when the CRS version in the backend differs from the one in coraza-spoa.

Update process:
1. A new `stove301/coraza-spoa:crs-vX.Y.Z` image appears on Docker Hub
2. `docker compose pull && docker compose up -d`
3. The dashboard mismatch warning disappears automatically

## Known Limitations

- **Rules Management Reload**: `PUT /api/rules/config` writes the new config, but requires a manual `docker compose restart coraza-spoa` (coraza-spoa `-autoreload` uses fsnotify, which fails on hosts with many inotify instances).

## Host Requirements

On hosts with many running processes (k8s, many containers), you may need to increase the inotify limit:

```bash
echo "fs.inotify.max_user_instances=512" >> /etc/sysctl.conf
sysctl -p
```

## Project Structure

```
coraza-dashboard/
  docker-compose.yml         — All services
  .env.example               — Configuration template
  scripts/
    install.sh               — Full install script
    traffic-gen.py           — Traffic generator for testing
  services/
    backend/                 — Go REST API + Ingest Handler
    frontend/                — Vue 3 Dashboard (Vite + Tailwind)
    fluentbit/               — Fluent Bit config (log forwarder)
    haproxy/                 — HAProxy + SPOE config
    coraza/                  — coraza-spoa config (OWASP CRS v4)
  agents/                    — Subagent task files (build history)
```
