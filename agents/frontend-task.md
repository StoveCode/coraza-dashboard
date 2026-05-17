# Frontend Agent Task — coraza-dashboard

## Deine Rolle
Du bist Senior Frontend Engineer. Du schreibst Code, committest, und meldest dich strukturiert zurück.
Du sprichst NUR mit Rusty (dem PM). Kein direkter Kontakt zu Stefan.

## Projekt
Vue 3 Dashboard für OWASP Coraza WAF Logging & Monitoring.

## Tech Stack
- **Framework:** Vue 3 + TypeScript + Vite
- **State:** Pinia
- **Router:** Vue Router 4
- **Charts:** Chart.js + vue-chartjs
- **HTTP:** axios
- **Styling:** Tailwind CSS (oder UnoCSS) — dark-themed, professionell
- **Serving:** Nginx (Docker)

## Arbeitsverzeichnis
`/home/ubuntu/.openclaw/workspace/coraza-dashboard/services/frontend/`

## Backend API (Referenz)

Base URL: `http://localhost:8080` (konfigurierbar via `.env`)

```
GET  /api/health
GET  /api/events?limit=50&offset=0&from=<RFC3339>&to=<RFC3339>&action=block|detect&client_ip=<ip>
     → { total, limit, offset, events: [WAFEvent] }
GET  /api/stats
     → { total_blocks, total_detections, top_ips, top_rules, events_per_hour }
POST /api/log  (interner Collector, nicht im Frontend nötig)
```

WAFEvent:
```typescript
{
  id: string           // UUID
  timestamp: string    // RFC3339
  client_ip: string
  method: string
  uri: string
  rule_id: string
  rule_msg: string
  severity: string
  action: "block" | "detect"
  raw_log: object
}
```

## Was zu tun ist

### 1. Projektstruktur
```
services/frontend/
  index.html
  vite.config.ts
  tsconfig.json
  package.json
  .env.example        — VITE_API_BASE_URL=http://localhost:8080
  Dockerfile          — Multi-stage (node build → nginx:alpine)
  nginx.conf          — SPA fallback, proxy /api → backend
  src/
    main.ts
    App.vue
    router/index.ts
    stores/
      events.ts       — Pinia store für Events
      stats.ts        — Pinia store für Stats
    api/
      client.ts       — axios instance, base URL aus import.meta.env
      events.ts       — API calls für Events
      stats.ts        — API calls für Stats
    views/
      Dashboard.vue   — Haupt-Dashboard
      Events.vue      — Events-Tabelle mit Filterung
    components/
      StatCard.vue    — Zahl + Label Karte (Total Blocks, Detections etc.)
      EventsTable.vue — Tabelle mit Paginierung
      TimelineChart.vue  — Events pro Stunde (Line Chart)
      TopIPsChart.vue    — Top blockierte IPs (Bar Chart)
      TopRulesChart.vue  — Top Rule Hits (Horizontal Bar)
      FilterBar.vue   — Zeitraum, Action-Filter, IP-Suche
```

### 2. Dashboard View (Haupt-Seite)
Zeigt:
- **4 StatCards:** Total Blocks, Total Detections, Unique IPs, Most Hit Rule
- **TimelineChart:** Events der letzten 24h (stündlich)
- **TopIPsChart:** Top 10 blockierte IPs
- **TopRulesChart:** Top 10 WAF Rules (Rule ID + Msg)
- **Auto-Refresh:** alle 30 Sekunden (via setInterval)

### 3. Events View
- Tabelle mit allen Columns: Timestamp, IP, Method, URI, Rule ID, Rule Msg, Severity, Action
- **Action-Badge:** block = rot, detect = gelb
- **Filter:** Zeitraum (from/to DatePicker), Action (All/Block/Detect), IP-Suche
- **Paginierung:** 50 pro Seite
- **Klick auf Row:** Raw Log JSON expandierbar (Accordion)

### 4. Design
- **Dark Theme** — professionell, Security-Dashboard Feeling
- Farben: Grau-Töne + Rot für Blocks + Gelb für Detections + Grün für OK
- Kompakt und übersichtlich, kein Spielzeug-Look
- Navbar mit Links: Dashboard | Events

### 5. Env-Konfiguration
`.env.example`:
```
VITE_API_BASE_URL=http://localhost:8080
```
Nginx im Docker-Container soll `/api` Requests zum Backend proxyen (URL via Nginx-Template konfigurierbar).

### 6. Dockerfile
```dockerfile
# Stage 1: Build
FROM node:20-alpine AS builder
WORKDIR /app
COPY package*.json .
RUN npm ci
COPY . .
RUN npm run build

# Stage 2: Serve
FROM nginx:alpine
COPY --from=builder /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 3000
```

### 7. docker-compose.yml updaten
Im Root-`docker-compose.yml` den Frontend-Service auskommentieren/aktivieren:
```yaml
frontend:
  build: ./services/frontend
  ports:
    - "${FRONTEND_PORT:-3000}:3000"
  environment:
    - VITE_API_BASE_URL=${BACKEND_URL:-http://backend:8080}
  depends_on:
    - backend
```
Und in `.env.example` ergänzen:
```
FRONTEND_PORT=3000
```

### 8. Git
- Committe am Ende alles mit sinnvollen Messages
- Kein Push — Rusty macht das

## Output-Format (PFLICHT am Ende)
```
STATUS: done | failed
FILES: alle erstellten/geänderten Dateien
PROBLEMS: keine | Liste
NOTES: Besonderheiten, bekannte Einschränkungen
FEEDBACK: Verbesserungsvorschläge an Rusty
```
