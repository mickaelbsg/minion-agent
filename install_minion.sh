#!/usr/bin/env bash
set -euo pipefail

# ------------------------------------------------------------
# Minion Agent minimal installer (test machine)
# ------------------------------------------------------------
# This script assumes the host already has all required dependencies:
#   - Go toolchain (go >= 1.22)
#   - fail2ban, iptables, openssl (or equivalent tools)
#   - sudo (for iptables access if needed)
# It will only:
#   1. Build the Minion binary from the source repository.
#   2. Install the binary to /usr/local/bin/minion.
#   3. Set up the configuration directory (/etc/minion) and copy an example
#      config if one does not already exist.
#   4. Generate a self‑signed TLS certificate (only if missing).
#   5. Install the systemd service unit (runs as root to allow Fail2Ban access).
#   6. Enable and start the service.
# ------------------------------------------------------------

log(){ echo "[install-minion] $*"; }

# -----------------------------------------------------------------
# 0. Verify we are running as root (required for installing files and
#    writing to /etc/systemd/system).
# -----------------------------------------------------------------
if [[ "$EUID" -ne 0 ]]; then
  log "ERROR: This installer must be run as root (or via sudo)."
  exit 1
fi

# -----------------------------------------------------------------
# 1. Build the binary
# -----------------------------------------------------------------
log "Building Minion binary..."
cd /home/pc/projetos/minion-agent
export GO111MODULE=on
# Clean any previous build artifacts
go clean -modcache
# Build the binary
go build -o minion ./cmd/minion
log "Binary built at $(pwd)/minion"

# -----------------------------------------------------------------
# 2. Install the binary
# -----------------------------------------------------------------
log "Installing binary to /usr/local/bin/minion"
cp minion /usr/local/bin/minion
chmod 755 /usr/local/bin/minion

# -----------------------------------------------------------------
# 3. Configuration directory
# -----------------------------------------------------------------
log "Creating configuration directory /etc/minion"
mkdir -p /etc/minion
# If a config does not already exist, copy the example config
if [[ ! -f /etc/minion/config.json ]]; then
  if [[ -f config.example.json ]]; then
    cp config.example.json /etc/minion/config.json
    chmod 644 /etc/minion/config.json
    log "Copied example config to /etc/minion/config.json"
  else
    log "WARNING: No config.example.json found – you must create /etc/minion/config.json manually."
  fi
else
  log "Config file already exists – leaving untouched."
fi

# -----------------------------------------------------------------
# 4. TLS certificates (self‑signed, if missing)
# -----------------------------------------------------------------
TLS_DIR="/etc/minion/tls"
mkdir -p "$TLS_DIR"
if [[ ! -f "$TLS_DIR/minion.crt" || ! -f "$TLS_DIR/minion.key" ]]; then
  log "Generating self‑signed TLS certificate (valid for 10 years)..."
  openssl req -newkey rsa:4096 -nodes -keyout "$TLS_DIR/minion.key" \
    -x509 -days 3650 -out "$TLS_DIR/minion.crt" \
    -subj "/C=BR/ST=DF/L=Brasilia/O=Minion/OU=Test/CN=minion.local"
  chmod 600 "$TLS_DIR/minion.key"
  chmod 644 "$TLS_DIR/minion.crt"
  log "TLS certificate created at $TLS_DIR/minion.{crt,key}"
else
  log "TLS certificate already present – skipping generation."
fi

# -----------------------------------------------------------------
# 5. Systemd unit file (runs as root for full Fail2Ban access)
# -----------------------------------------------------------------
UNIT_PATH="/etc/systemd/system/minion.service"
log "Installing systemd unit file to $UNIT_PATH"
cat > "$UNIT_PATH" <<'EOF_UNIT'
[Unit]
Description=Minion Agent Service
After=network.target

[Service]
ExecStart=/usr/local/bin/minion --config /etc/minion/config.json
Restart=on-failure
# No User= line – runs as root to allow fail2ban-client access
Environment=HOME=/root

[Install]
WantedBy=multi-user.target
EOF_UNIT

# -----------------------------------------------------------------
# 6. Reload systemd, enable and start the service
# -----------------------------------------------------------------
log "Reloading systemd daemon..."
systemctl daemon-reload
log "Enabling and starting Minion service..."
systemctl enable --now minion.service

log "Installation complete!"
log "Check service status:   systemctl status minion.service"
log "Health endpoint (no auth needed):   curl -k https://localhost:9871/api/v1/health"
