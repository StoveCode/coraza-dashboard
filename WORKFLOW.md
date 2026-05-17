# WORKFLOW.md — coraza-dashboard

## Projekt-Überblick
Logging & Monitoring Dashboard für OWASP Coraza WAF + HAProxy.
Zeigt Block-Events, Traffic-Statistiken und WAF-Aktivitäten in einem Web-Dashboard.

## Architektur

```
[Client]
   │
   ▼
[HAProxy]  ←──── SPOE ────→  [coraza-spoa]  (WAF, Rule Evaluation)
   │                               │
   │                        Audit Log (HTTP Webhook / JSON)
   │                               │
   ▼                               ▼
[Backend API]  ←────────── Log-Collector (Go)
   │
   ├── PostgreSQL (Event Storage)
   │
   ▼
[Frontend Dashboard] (Vue 3)
```

## Services (Docker Compose)
| Service       | Beschreibung                                | Port   |
|---------------|---------------------------------------------|--------|
| `haproxy`     | Reverse Proxy / Load Balancer               | 80/443 |
| `coraza-spoa` | WAF SPOE Agent (Coraza)                     | 9000   |
| `backend`     | Go API (Log-Collector + REST API)           | 8080   |
| `frontend`    | Vue 3 Dashboard (Nginx)                     | 3000   |
| `postgres`    | Event & Statistik Datenbank                 | 5432   |

## Konfiguration
- **Alle Verbindungen via `.env`** — kein Hardcoding
- `.env.example` im Repo-Root als Referenz
- Docker Compose liest `.env` automatisch

## Rollen
```
Stefan (Product Owner)
    │
🦀 Rusty (PM + Orchestrator) — fasst KEINEN Code an
    │
    ├── ⚙️ backend-agent  (Go: API, Log-Collector, DB-Schema)
    └── 🎨 frontend-agent (Vue 3: Dashboard, Charts, Tabellen)
```

## Workflow-Schritte

### 1. Backend zuerst
- Go REST API: `/api/events`, `/api/stats`, `/api/health`
- Log-Collector: HTTP-Endpoint empfängt Coraza Audit Logs
- PostgreSQL Schema: Events, Stats
- Docker: Multi-stage Build, Env-Vars

### 2. Frontend danach
- Vue 3 + Vite + Pinia + Vue Router
- Dashboard: Block-Events Tabelle, Charts (Top-IPs, Rule-Hits, Timeline)
- Verbindet sich mit Backend API (URL via .env / Build-Arg)
- Nginx-Serving

### 3. Integration
- docker-compose.yml verbindet alle Services
- .env.example dokumentiert alle Variablen
- HAProxy + coraza-spoa als Demo-Config beigelegt

## Coding Standards
- **Backend:** Go, idiomatic, Fehler-Handling, structured logging
- **Frontend:** TypeScript, Composition API, kein Options API
- **Kein Hardcoding** — alle Hosts/Ports/Credentials via Env
- **Docker:** Multi-stage Builds, non-root User

## ADR-Pflicht
Bei jeder Architektur-Entscheidung → `docs/ADR-XXX.md`

## Agents Output-Format
```
STATUS: done | failed
COMMITS: -
FILES: geänderte Dateien
PROBLEMS: keine | Liste
NOTES: Kontext für nächsten Agent
FEEDBACK: Verbesserungsvorschläge
```
