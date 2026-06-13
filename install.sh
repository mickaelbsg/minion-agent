#!/usr/bin/env bash
set -euo pipefail

# ------------------------------------------------------------
# Minion Agent installer for a test machine
# ------------------------------------------------------------
# This script performs the full installation of the Minion Agent:
#   1. Installs required system packages (go, fail2ban, iptables, openssl).
#   2. Builds the Go binary.
#   3. Generates TLS certificates if they do not exist.
#   4. Installs the binary to /usr/local/bin/minion.
#   5. Installs the systemd unit file (runs as root).
#   6. Enables and starts the service.
# ------------------------------------------------------------

# Helper for logging
log(){ echo "[install] $*"; }

# Ensure we are root (or using sudo for privileged actions)
if [[ "$EUID" -ne 0 ]]; then
  log "This installer must be run as root (or via sudo)."
  exit 1
fi

# 1. Install dependencies
log "Installing required packages..."
apt-get update -qq
DEPS="golang-go fail2ban iptables openssl"
apt-get install -y $DEPS > /dev/null

# 2. Build the binary
log "Building Minion binary..."
cd /home/pc/projetos/minion-agent
# Ensure Go environment is clean
export GO111MODULE=on
go build -o minion ./cmd/minion

# 3. Prepare configuration directory
log "Setting up configuration directory..."
mkdir -p /etc/minion
# Copy example config if not present
if [[ ! -f /etc/minion/config.json ]]; then
  cp config.example.json /etc/minion/config.json
  chmod 644 /etc/minion/config.json
fi

# 4. Generate TLS certs (self‑signed) if missing
TLS_DIR="/etc/minion/tls"
mkdir -p "$TLS_DIR"
if [[ ! -f "$TLS_DIR/minion.crt" || ! -f "$TLS_DIR/minion.key" ]]; then
  log "Generating self‑signed TLS certificate..."
  openssl req -newkey rsa:4096 -nodes -keyout "$TLS_DIR/minion.key" \
    -x509 -days 3650 -out "$TLS_DIR/minion.crt" \
    -subj "/C=BR/ST=DF/L=Brasilia/O=Minion/OU=Test/CN=minion.local"
  chmod 600 "$TLS_DIR/minion.key"
  chmod 644 "$TLS_DIR/minion.crt"
fi

# 5. Install binary
log "Installing binary to /usr/local/bin/minion"
cp minion /usr/local/bin/minion
chmod 755 /usr/local/bin/minion

# 6. Install systemd unit (runs as root)
log "Installing systemd unit file..."
cat > /etc/systemd/system/minion.service <<'EOF_UNIT'
[Unit]
Description=Minion Agent Service
After=network.target

[Service]
ExecStart=/usr/local/bin/minion --config /etc/minion/config.json
Restart=on-failure
# No User= line – runs as root for full Fail2Ban access
Environment=HOME=/root

[Install]
WantedBy=multi-user.target
EOF_UNIT

# Reload systemd, enable and start service
log "Reloading systemd daemon..."
systemctl daemon-reload
log "Enabling and starting Minion service..."
systemctl enable --now minion.service

log "Installation complete!"
log "Check status with: systemctl status minion.service"
log "Health endpoint: https://localhost:9871/api/v1/health"
