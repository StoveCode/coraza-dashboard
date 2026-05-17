#!/usr/bin/env bash
set -euo pipefail

# =============================================================================
# coraza-dashboard install script
# =============================================================================

SCRIPT_DIR=$(cd "$(dirname "$0")" && pwd)
cd "$SCRIPT_DIR/.."
PROJECT_ROOT="$(pwd)"

# ---------- Colors ------------------------------------------------------------
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
BOLD='\033[1m'
NC='\033[0m'

ok()   { echo -e "${GREEN}✓${NC} $*"; }
warn() { echo -e "${YELLOW}⚠${NC}  $*"; }
err()  { echo -e "${RED}✗${NC} $*"; }

section() {
  echo
  echo -e "${BOLD}${CYAN}=== $* ===${NC}"
  echo
}

# =============================================================================
# 1. Root check
# =============================================================================
section "Root check"
if [[ $EUID -ne 0 ]]; then
  err "This script must be run as root or with sudo."
  exit 1
fi
ok "Running as root"

# =============================================================================
# 2. OS / package-manager detection
# =============================================================================
section "OS Detection"

PKG_MANAGER=""
if command -v apt-get &>/dev/null; then
  PKG_MANAGER="apt"
  ok "Detected apt (Debian/Ubuntu)"
elif command -v dnf &>/dev/null; then
  PKG_MANAGER="dnf"
  ok "Detected dnf (Fedora/RHEL)"
elif command -v pacman &>/dev/null; then
  PKG_MANAGER="pacman"
  ok "Detected pacman (Arch)"
else
  err "Unsupported package manager. Please install dependencies manually."
  exit 1
fi

pkg_install() {
  case "$PKG_MANAGER" in
    apt)    apt-get install -y "$@" ;;
    dnf)    dnf install -y "$@" ;;
    pacman) pacman -S --noconfirm "$@" ;;
  esac
}

pkg_update() {
  case "$PKG_MANAGER" in
    apt)    apt-get update -y ;;
    dnf)    dnf check-update -y || true ;;
    pacman) pacman -Sy --noconfirm ;;
  esac
}

# =============================================================================
# 3. Dependencies
# =============================================================================
section "Dependencies"

# ---- Docker -----------------------------------------------------------------
if command -v docker &>/dev/null; then
  ok "Docker already installed: $(docker --version)"
else
  warn "Docker not found — installing..."
  pkg_update
  case "$PKG_MANAGER" in
    apt)
      pkg_install ca-certificates curl gnupg lsb-release
      install -m 0755 -d /etc/apt/keyrings
      curl -fsSL https://download.docker.com/linux/$(. /etc/os-release && echo "$ID")/gpg \
        | gpg --dearmor -o /etc/apt/keyrings/docker.gpg
      chmod a+r /etc/apt/keyrings/docker.gpg
      echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] \
        https://download.docker.com/linux/$(. /etc/os-release && echo "$ID") \
        $(. /etc/os-release && echo "$VERSION_CODENAME") stable" \
        > /etc/apt/sources.list.d/docker.list
      apt-get update -y
      pkg_install docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin
      ;;
    dnf)
      dnf config-manager --add-repo https://download.docker.com/linux/fedora/docker-ce.repo
      pkg_install docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin
      ;;
    pacman)
      pkg_install docker docker-compose
      ;;
  esac
  systemctl enable --now docker
  ok "Docker installed: $(docker --version)"
fi

# ---- Docker Compose plugin --------------------------------------------------
if docker compose version &>/dev/null; then
  ok "Docker Compose plugin: $(docker compose version)"
else
  warn "Docker Compose plugin not found — installing..."
  case "$PKG_MANAGER" in
    apt)    pkg_install docker-compose-plugin ;;
    dnf)    pkg_install docker-compose-plugin ;;
    pacman) warn "Install docker-compose manually on Arch if missing." ;;
  esac
  docker compose version &>/dev/null && ok "Docker Compose plugin installed" \
    || { err "Docker Compose plugin still not available."; exit 1; }
fi

# ---- git --------------------------------------------------------------------
if command -v git &>/dev/null; then
  ok "git: $(git --version)"
else
  warn "git not found — installing..."
  pkg_install git
  ok "git installed: $(git --version)"
fi

# ---- python3 + pip3 + requests ----------------------------------------------
if command -v python3 &>/dev/null; then
  ok "python3: $(python3 --version)"
else
  warn "python3 not found — installing..."
  case "$PKG_MANAGER" in
    apt)    pkg_install python3 python3-pip ;;
    dnf)    pkg_install python3 python3-pip ;;
    pacman) pkg_install python python-pip ;;
  esac
  ok "python3 installed: $(python3 --version)"
fi

if command -v pip3 &>/dev/null || python3 -m pip --version &>/dev/null 2>&1; then
  ok "pip3 available"
else
  warn "pip3 not found — installing..."
  case "$PKG_MANAGER" in
    apt)    pkg_install python3-pip ;;
    dnf)    pkg_install python3-pip ;;
    pacman) pkg_install python-pip ;;
  esac
fi

if python3 -c "import requests" &>/dev/null 2>&1; then
  ok "Python 'requests' package already installed"
else
  warn "Installing Python 'requests' package..."
  pip3 install --quiet requests 2>/dev/null \
    || python3 -m pip install --quiet requests
  ok "Python 'requests' installed"
fi

# =============================================================================
# 4. Build coraza-spoa:local image
# =============================================================================
section "coraza-spoa:local image"

if docker image inspect coraza-spoa:local &>/dev/null; then
  warn "Image 'coraza-spoa:local' already exists — skipping build"
else
  SPOA_SRC="/tmp/coraza-spoa-source"
  if [[ -d "$SPOA_SRC/.git" ]]; then
    ok "coraza-spoa source already cloned — pulling latest..."
    git -C "$SPOA_SRC" pull --ff-only
  else
    ok "Cloning coraza-spoa source..."
    git clone https://github.com/corazawaf/coraza-spoa.git "$SPOA_SRC"
  fi

  ok "Building coraza-spoa:local (this may take a few minutes)..."
  docker build \
    -f "$SPOA_SRC/ftw/Dockerfile.coraza_spoa" \
    -t coraza-spoa:local \
    "$SPOA_SRC"
  ok "Image 'coraza-spoa:local' built successfully"
fi

# =============================================================================
# 5. .env setup
# =============================================================================
section ".env Setup"

if [[ -f "$PROJECT_ROOT/.env" ]]; then
  ok ".env already exists — skipping"
else
  if [[ -f "$PROJECT_ROOT/.env.example" ]]; then
    cp "$PROJECT_ROOT/.env.example" "$PROJECT_ROOT/.env"
    ok ".env created from .env.example"
    warn "Don't forget to change passwords and secrets in .env before going to production!"
  else
    warn ".env.example not found — skipping .env creation (create it manually if needed)"
  fi
fi

# =============================================================================
# 6. Start stack
# =============================================================================
section "Starting Docker Compose Stack"

docker compose up -d --build
ok "Stack started"

# =============================================================================
# 7. Health check
# =============================================================================
section "Health Check"

echo -n "Waiting for backend to become ready (max 30s)..."
TIMEOUT=30
ELAPSED=0
until curl -sf http://localhost:8080/api/health &>/dev/null; do
  if [[ $ELAPSED -ge $TIMEOUT ]]; then
    echo
    err "Backend did not respond within ${TIMEOUT}s."
    warn "Check logs with: docker compose logs backend"
    exit 1
  fi
  sleep 2
  ELAPSED=$((ELAPSED + 2))
  echo -n "."
done
echo
ok "Backend is healthy"

# =============================================================================
# 8. Done
# =============================================================================
section "Installation Complete"

echo -e "${GREEN}${BOLD}coraza-dashboard is up and running!${NC}"
echo
echo -e "  ${BOLD}Dashboard:${NC}      http://localhost:3000"
echo -e "  ${BOLD}Backend API:${NC}    http://localhost:8080"
echo -e "  ${BOLD}HAProxy Stats:${NC}  http://localhost:8404/stats"
echo
ok "All done 🦀"
