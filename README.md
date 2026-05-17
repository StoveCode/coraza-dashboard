# coraza-dashboard

WAF Logging & Monitoring Dashboard for OWASP Coraza + HAProxy.

## Architecture

```
Client → HAProxy → Coraza SPOE → Audit Logs → Backend (Go) → PostgreSQL
                                                      ↑
                                            Frontend (React/Next) ← (coming soon)
```

## Quick Start

```bash
cp .env.example .env
# adjust passwords in .env
docker compose up --build
```

Backend API available at: `http://localhost:8080`

## Services

| Service   | Port | Description                          |
|-----------|------|--------------------------------------|
| backend   | 8080 | Go REST API + Log Collector          |
| postgres  | 5432 | PostgreSQL database                  |
| haproxy   | 80   | Reverse proxy with Coraza SPOE       |

## API Endpoints

| Method | Path          | Description                       |
|--------|---------------|-----------------------------------|
| GET    | /api/health   | Health check                      |
| GET    | /api/events   | List WAF events (paginated)       |
| GET    | /api/stats    | Aggregated stats                  |
| POST   | /api/log      | Receive Coraza audit log          |

See [docs/](docs/) for full API documentation.
