#!/usr/bin/env bash
# ML-KEM Web App — quick deployment script
# Run as: bash deploy.sh
# Requires: Docker Engine 24+ with Compose plugin

set -e

RED='\033[0;31m'; GREEN='\033[0;32m'; YELLOW='\033[1;33m'; NC='\033[0m'

info()  { echo -e "${GREEN}[INFO]${NC}  $1"; }
warn()  { echo -e "${YELLOW}[WARN]${NC}  $1"; }
error() { echo -e "${RED}[ERROR]${NC} $1"; exit 1; }

# ── Check Docker ──────────────────────────────────────────────────────────────
command -v docker >/dev/null 2>&1 || error "Docker is not installed. See DEPLOYMENT.md."
docker compose version >/dev/null 2>&1 || error "Docker Compose plugin not found. See DEPLOYMENT.md."

# ── Check Docker daemon ───────────────────────────────────────────────────────
docker info >/dev/null 2>&1 || {
  warn "Docker daemon not running — starting it..."
  sudo systemctl start docker
}

# ── Build and start ───────────────────────────────────────────────────────────
info "Building and starting ML-KEM containers..."
sudo docker compose up --build -d

# ── Wait for containers ───────────────────────────────────────────────────────
info "Waiting for services to be ready..."
sleep 4

# ── Status ────────────────────────────────────────────────────────────────────
sudo docker compose ps

# ── Print access URL ─────────────────────────────────────────────────────────
LOCAL_IP=$(ip route get 1.1.1.1 2>/dev/null | awk '{print $7; exit}')
echo ""
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}  ML-KEM app is running!${NC}"
echo -e "${GREEN}  Local:   http://localhost:3000${NC}"
if [ -n "$LOCAL_IP" ]; then
  echo -e "${GREEN}  Network: http://${LOCAL_IP}:3000${NC}"
fi
echo -e "${GREEN}========================================${NC}"
echo ""
info "To stop:    sudo docker compose down"
info "To view logs: sudo docker compose logs -f"
