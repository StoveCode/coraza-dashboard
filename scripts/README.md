# Traffic-Generator — coraza-dashboard

Test-Script das zufälligen HTTP-Traffic durch HAProxy/Coraza schickt.
Nützlich zum Befüllen des Dashboards mit Test-Daten.

## Usage

```bash
python3 traffic-gen.py [--target URL] [--rate N] [--ratio N]
```

## Optionen

| Option     | Default              | Beschreibung                          |
|------------|----------------------|---------------------------------------|
| `--target` | `http://localhost:80`| Ziel-URL (HAProxy)                    |
| `--rate`   | `1.0`               | Requests pro Sekunde                  |
| `--ratio`  | `0.4`               | Anteil Angriffs-Requests (0.0 - 1.0)  |

## Beispiele

```bash
# 1 req/s, 50% Angriffe (Standard-Test)
python3 traffic-gen.py --rate 1 --ratio 0.5

# Stresstest: 5 req/s, 80% Angriffe
python3 traffic-gen.py --rate 5 --ratio 0.8

# Nur legitimer Traffic
python3 traffic-gen.py --ratio 0.0

# Gegen anderen Host
python3 traffic-gen.py --target http://192.168.1.210:80
```

## Angriffsvektoren

| Kategorie        | Beispiele                                    |
|------------------|----------------------------------------------|
| SQL Injection    | `OR '1'='1`, `UNION SELECT`, `DROP TABLE`    |
| XSS              | `<script>`, `onerror=`, `javascript:`        |
| LFI              | `../../../etc/passwd`, `../../../../etc/shadow` |
| RCE              | `; cat /etc/passwd`, `; id`, `$(whoami)`     |
| SSRF             | `http://169.254.169.254/`, `file:///etc/`    |
| Scanner-UAs      | `nikto/2.1.6`, `sqlmap/1.7`                  |
| Recon            | `/.env`, `/wp-admin/`, `/.git/config`        |

## Output

```
🚦 WAF Traffic Generator
   Target : http://localhost:80
   Rate   : 1.0 req/s
   Attacks: 50% of traffic

[14:23:07] 🔴 BLOCKED 403 GET  /download?file=../../../etc/passwd   [LFI-passwd]
[14:23:08] 🔴 BLOCKED 403 GET  /page?title=<img src=x onerror=...>  [XSS-img]
[14:23:09] ✅ ALLOWED 200 GET  /api/products?page=1                  [legit]
```

## Dependencies

```bash
pip install requests
```
