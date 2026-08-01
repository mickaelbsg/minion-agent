#!/usr/bin/env bash
# Build a Debian package for Minion Agent
set -euo pipefail

PKG_NAME="minion"
PKG_VER="${PKG_VER:-1.0.4}"
ARCH="amd64"
BUILD_ROOT="$(mktemp -d)"
DEB_ROOT="$BUILD_ROOT/${PKG_NAME}_${PKG_VER}_${ARCH}"

trap 'rm -rf "$BUILD_ROOT"' EXIT

echo "Creating package directory $DEB_ROOT"
mkdir -p "$DEB_ROOT/DEBIAN"
mkdir -p "$DEB_ROOT/usr/local/bin"
mkdir -p "$DEB_ROOT/etc/minion/tls"
mkdir -p "$DEB_ROOT/opt/minion"
mkdir -p "$DEB_ROOT/var/lib/minion"
mkdir -p "$DEB_ROOT/lib/systemd/system"

cat > "$DEB_ROOT/DEBIAN/control" <<EOF
Package: $PKG_NAME
Version: $PKG_VER
Section: utils
Priority: optional
Architecture: $ARCH
Maintainer: Mickael Bergson <mickael@example.com>
Depends: libc6 (>= 2.28), iptables, fail2ban, openssl, sqlite3
Description: Minion Agent - lightweight Linux observability agent and API server.
 Minion gathers host information and exposes an authenticated HTTPS API.
EOF

echo "/etc/minion/config.json" > "$DEB_ROOT/DEBIAN/conffiles"

cat > "$DEB_ROOT/DEBIAN/postinst" <<'EOS'
#!/bin/sh
set -e

CONFIG="/etc/minion/config.json"
DATA_DIR="/opt/minion"
TLS_DIR="/etc/minion/tls"
STATE_DIR="/var/lib/minion"
BOOTSTRAP_FILE="$STATE_DIR/bootstrap-credentials.txt"
BOOTSTRAP_TMP="$STATE_DIR/.bootstrap-credentials.tmp"

install -d -o root -g root -m 700 /etc/minion "$TLS_DIR" "$DATA_DIR" "$STATE_DIR"
chmod 600 "$CONFIG"

systemctl daemon-reload
systemctl reset-failed minion.service >/dev/null 2>&1 || true

umask 077
if /usr/local/bin/minion setup --config "$CONFIG" --name bootstrap --ips 127.0.0.1/32 >"$BOOTSTRAP_TMP" 2>&1; then
  if grep -q '^API Key:' "$BOOTSTRAP_TMP"; then
    mv -f "$BOOTSTRAP_TMP" "$BOOTSTRAP_FILE"
    chmod 600 "$BOOTSTRAP_FILE"
    echo "Minion bootstrap credential created securely."
  else
    rm -f "$BOOTSTRAP_TMP"
  fi
else
  cat "$BOOTSTRAP_TMP" >&2
  rm -f "$BOOTSTRAP_TMP"
  echo "Minion bootstrap failed; package configuration was not completed." >&2
  exit 1
fi

chmod 700 /etc/minion "$TLS_DIR" "$DATA_DIR" "$STATE_DIR"
[ ! -f "$TLS_DIR/minion.key" ] || chmod 600 "$TLS_DIR/minion.key"
[ ! -f "$TLS_DIR/minion.crt" ] || chmod 644 "$TLS_DIR/minion.crt"
[ ! -f "$DATA_DIR/minion.db" ] || chmod 600 "$DATA_DIR/minion.db"

systemctl enable minion.service >/dev/null
systemctl restart minion.service

if ! systemctl is-active --quiet minion.service; then
  echo "Minion service failed to start. Run: journalctl -u minion.service -n 100" >&2
  exit 1
fi

echo "Minion installed and running."
echo "Status: systemctl status minion.service"
echo "Health: https://127.0.0.1:9870/api/v1/health"
if [ -f "$BOOTSTRAP_FILE" ]; then
  echo "Next step: run 'sudo minion bootstrap pair --ips <AUTOMATION_IP/32>' to authorize Automation and display the initial API key once."
fi
exit 0
EOS
chmod 755 "$DEB_ROOT/DEBIAN/postinst"

cat > "$DEB_ROOT/DEBIAN/prerm" <<'EOS'
#!/bin/sh
set -e
systemctl stop minion.service >/dev/null 2>&1 || true
if [ "$1" = "remove" ]; then
  systemctl disable minion.service >/dev/null 2>&1 || true
fi
exit 0
EOS
chmod 755 "$DEB_ROOT/DEBIAN/prerm"

if [[ ! -f "$(pwd)/minion" ]]; then
  echo "Compiling minion binary..."
  go build -o minion ./cmd/minion
fi
install -m 755 "$(pwd)/minion" "$DEB_ROOT/usr/local/bin/minion"

if [[ -f "config.example.json" ]]; then
  install -m 600 config.example.json "$DEB_ROOT/etc/minion/config.json"
else
  cat > "$DEB_ROOT/etc/minion/config.json" <<'EOS'
{
  "api": {
    "bind": "0.0.0.0:9870",
    "allow_insecure_http": false
  },
  "security": {
    "allowed_fail2ban_jails": ["sshd", "apache-auth", "recidive"]
  },
  "db_path": "/opt/minion/minion.db",
  "clients": []
}
EOS
  chmod 600 "$DEB_ROOT/etc/minion/config.json"
fi

if [[ -f "systemd/minion.service" ]]; then
  install -m 644 systemd/minion.service "$DEB_ROOT/lib/systemd/system/minion.service"
else
  cat > "$DEB_ROOT/lib/systemd/system/minion.service" <<'EOS'
[Unit]
Description=Minion Agent Service
After=network.target

[Service]
Type=simple
ExecStart=/usr/local/bin/minion --config /etc/minion/config.json
Restart=on-failure
RestartSec=5
Environment=HOME=/root
Environment=PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin
NoNewPrivileges=true
PrivateTmp=true
ProtectHome=true
ProtectSystem=strict
ReadWritePaths=/opt/minion /etc/minion
RestrictSUIDSGID=true
LockPersonality=true
MemoryDenyWriteExecute=true

[Install]
WantedBy=multi-user.target
EOS
fi
chmod 644 "$DEB_ROOT/lib/systemd/system/minion.service"
chmod 700 "$DEB_ROOT/etc/minion" "$DEB_ROOT/etc/minion/tls" "$DEB_ROOT/opt/minion" "$DEB_ROOT/var/lib/minion"

dpkg-deb --root-owner-group --build "$DEB_ROOT" "${PKG_NAME}_${PKG_VER}_${ARCH}.deb"
echo "Package built: ${PKG_NAME}_${PKG_VER}_${ARCH}.deb"
