# Backend — coraza-dashboard

Go REST API + Log-Tailer für das coraza-dashboard.

## Funktion

- **Log-Tailer:** Folgt `/var/log/coraza/coraza.log` live via `nxadm/tail` (Follow + ReOpen)
- **REST API:** Liefert Events und Stats an das Frontend
- **Prometheus Scraper:** Optional — scrapt Metriken von coraza-spoa

## Endpoints

| Method | Pfad          | Beschreibung                                     |
|--------|---------------|--------------------------------------------------|
| GET    | /api/health   | Health Check → `{"status":"ok"}`                 |
| GET    | /api/events   | WAF-Events (paginated, filterbar)                |
| GET    | /api/stats    | Aggregierte Statistiken                          |
| GET    | /api/metrics  | Prometheus-Metriken (wenn CORAZA_METRICS_URL gesetzt) |

## WAFEvent Datenmodell

```go
type WAFEvent struct {
    ID         string    // UUID
    Timestamp  time.Time
    Client     string    // Client-IP
    Server     string    // Server-IP
    URI        string
    RuleID     int
    RuleMsg    string
    RuleFile   string
    Severity   string    // CRITICAL, ERROR, WARNING, NOTICE, INFO, DEBUG
    SeverityID int
    Phase      string    // request-headers, request-body, response-body, ...
    PhaseID    int
    Disruptive bool      // true=Block, false=Detection
    Tags       []string  // OWASP CRS Tags (attack-xss, attack-sqli, ...)
    Data       string    // Matched data
    UniqueID   string    // HAProxy Transaction ID
}
```

## Umgebungsvariablen

| Variable               | Default           | Beschreibung                        |
|------------------------|-------------------|-------------------------------------|
| `DB_HOST`              | `postgres`        | PostgreSQL Host                     |
| `DB_PORT`              | `5432`            | PostgreSQL Port                     |
| `DB_NAME`              | `coraza_dashboard`| Datenbank Name                      |
| `DB_USER`              | `coraza`          | Datenbank User                      |
| `DB_PASSWORD`          | `changeme`        | Datenbank Passwort                  |
| `SERVER_PORT`          | `8080`            | HTTP Server Port                    |
| `CORS_ALLOWED_ORIGINS` | `*`               | Erlaubte CORS Origins (kommasep.)   |
| `LOG_FILE`             | `/var/log/coraza/coraza.log` | Pfad zum Coraza Log   |
| `CORAZA_METRICS_URL`   | (leer)            | Prometheus URL von coraza-spoa      |
| `LOG_LEVEL`            | `info`            | Log-Level (debug/info/warn/error)   |

## Lokal starten

```bash
cd services/backend
go run .
```

## Docker

```bash
docker build -t coraza-dashboard-backend .
```
