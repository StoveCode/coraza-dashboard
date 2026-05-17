# Frontend — coraza-dashboard

Vue 3 Dark-Theme Dashboard für OWASP Coraza WAF Monitoring.

## Features

- **StatCards:** Total Blocks, Detections, Unique IPs, Top Rule
- **Timeline Chart:** Events pro Stunde (letzte 24h)
- **Top IPs Chart:** Meistblockierte IP-Adressen
- **Top Rules Chart:** Häufigste WAF-Rule-Hits
- **Top Tags Chart:** OWASP CRS Angriffskategorien (attack-xss, attack-sqli, ...)
- **Phase Donut:** Request-Headers / Request-Body / Response Verteilung
- **Events-Tabelle:** Alle Events filterbar nach Zeit, disruptive/detect, IP
- **Auto-Refresh:** alle 30 Sekunden

## Tech Stack

- Vue 3 + TypeScript + Vite
- Pinia (State Management)
- Vue Router 4
- Chart.js + vue-chartjs
- Tailwind CSS (Dark Theme)
- Axios

## Umgebungsvariablen

| Variable         | Default | Beschreibung                     |
|------------------|---------|----------------------------------|
| `VITE_API_URL`   | (leer)  | API Base URL (leer = relativ)    |

Leer lassen damit Nginx die Requests zu `/api/` proxied.
Nur setzen wenn Backend auf einer anderen Domain läuft.

## Lokal entwickeln

```bash
cd services/frontend
npm install
npm run dev
# http://localhost:5173
```

Backend muss auf `http://localhost:8080` laufen (oder `VITE_API_URL` setzen).

## Docker

```bash
docker build -t coraza-dashboard-frontend .
```

Nginx served auf Port 3000 und proxied `/api/` zum Backend-Container.
