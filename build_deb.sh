#!/usr/bin/env bash
# Build a Debian package for Minion Agent
set -euo pipefail

PKG_NAME="minion"
PKG_VER="${PKG_VER:-1.0.4}"
ARCH="amd64"
BUILD_ROOT="$(mktemp -d)"
DEB_ROOT="$BUILD_ROOT/${PKG_NAME}_${PKG_VER}_${ARCH}"
MINION_BINARY="${MINION_BINARY:-}"

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
Depends: libc6 (>= 2.28), openssl, sqlite3, curl
Recommends: iptables, fail2ban
Description: Minion Agent - lightweight Linux observability agent and API server.
 Minion gathers host information and exposes an authenticated HTTPS API.
EOF

echo "/etc/minion/config.json" > "$DEB_ROOT/DEBIAN/conffiles"

cat > "$DEB_ROOT/DEBIAN/preinst" <<'EOS'
#!/bin/sh
set -e

BACKUP_DIR="/var/lib/minion/upgrade-backup"

if [ "$1" = "upgrade" ]; then
  systemctl stop minion.service >/dev/null 2>&1 || true

  rm -rf "$BACKUP_DIR"
  install -d -o root -g root -m 700 \
    "$BACKUP_DIR/etc/minion" \
    "$BACKUP_DIR/opt/minion" \
    "$BACKUP_DIR/usr/local/bin" \
    "$BACKUP_DIR/lib/systemd/system"

  [ ! -f /etc/minion/config.json ] || cp -a /etc/minion/config.json "$BACKUP_DIR/etc/minion/config.json"
  [ ! -d /etc/minion/tls ] || cp -a /etc/minion/tls "$BACKUP_DIR/etc/minion/tls"
  [ ! -f /opt/minion/minion.db ] || cp -a /opt/minion/minion.db "$BACKUP_DIR/opt/minion/minion.db"
  [ ! -f /opt/minion/minion.db-wal ] || cp -a /opt/minion/minion.db-wal "$BACKUP_DIR/opt/minion/minion.db-wal"
  [ ! -f /opt/minion/minion.db-shm ] || cp -a /opt/minion/minion.db-shm "$BACKUP_DIR/opt/minion/minion.db-shm"
  [ ! -f /usr/local/bin/minion ] || cp -a /usr/local/bin/minion "$BACKUP_DIR/usr/local/bin/minion"
  [ ! -f /lib/systemd/system/minion.service ] || cp -a /lib/systemd/system/minion.service "$BACKUP_DIR/lib/systemd/system/minion.service"

  printf '%s\n' "$2" > "$BACKUP_DIR/from-version"
  chmod -R go-rwx "$BACKUP_DIR"
  touch "$BACKUP_DIR/ready"
fi

exit 0
EOS
chmod 755 "$DEB_ROOT/DEBIAN/preinst"

cat > "$DEB_ROOT/DEBIAN/postinst" <<'EOS'
#!/bin/sh
set -e

CONFIG="/etc/minion/config.json"
DATA_DIR="/opt/minion"
TLS_DIR="/etc/minion/tls"
STATE_DIR="/var/lib/minion"
BACKUP_DIR="$STATE_DIR/upgrade-backup"
BOOTSTRAP_FILE="$STATE_DIR/bootstrap-credentials.txt"

rollback_upgrade() {
  [ -f "$BACKUP_DIR/ready" ] || return 0

  echo "Minion upgrade failed; restoring the previous operational state." >&2
  systemctl stop minion.service >/dev/null 2>&1 || true

  if [ -f "$BACKUP_DIR/etc/minion/config.json" ]; then
    install -d -o root -g root -m 700 /etc/minion
    cp -a "$BACKUP_DIR/etc/minion/config.json" /etc/minion/config.json
  fi
  if [ -d "$BACKUP_DIR/etc/minion/tls" ]; then
    rm -rf /etc/minion/tls
    cp -a "$BACKUP_DIR/etc/minion/tls" /etc/minion/tls
  fi

  rm -f /opt/minion/minion.db /opt/minion/minion.db-wal /opt/minion/minion.db-shm
  for file in minion.db minion.db-wal minion.db-shm; do
    [ ! -f "$BACKUP_DIR/opt/minion/$file" ] || cp -a "$BACKUP_DIR/opt/minion/$file" "/opt/minion/$file"
  done

  [ ! -f "$BACKUP_DIR/usr/local/bin/minion" ] || cp -a "$BACKUP_DIR/usr/local/bin/minion" /usr/local/bin/minion
  [ ! -f "$BACKUP_DIR/lib/systemd/system/minion.service" ] || cp -a "$BACKUP_DIR/lib/systemd/system/minion.service" /lib/systemd/system/minion.service

  systemctl daemon-reload >/dev/null 2>&1 || true
  systemctl reset-failed minion.service >/dev/null 2>&1 || true
  if systemctl restart minion.service >/dev/null 2>&1 && systemctl is-active --quiet minion.service; then
    echo "Previous Minion service restored and running. Package configuration remains failed; inspect dpkg status before retrying the upgrade." >&2
  else
    echo "Previous files were restored, but the Minion service could not be restarted. Run: journalctl -u minion.service -n 100" >&2
  fi
}

on_exit() {
  rc=$?
  if [ "$rc" -ne 0 ]; then
    rollback_upgrade
  fi
  exit "$rc"
}
trap on_exit EXIT

install -d -o root -g root -m 700 /etc/minion "$TLS_DIR" "$DATA_DIR" "$STATE_DIR"
if [ ! -f "$CONFIG" ]; then
  cat > "$CONFIG" <<'EOF_CONFIG'
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
EOF_CONFIG
fi
chmod 600 "$CONFIG"

require_command() {
  command -v "$1" >/dev/null 2>&1 || {
    echo "Minion installation is missing required command: $1" >&2
    echo "Install the missing runtime dependency and retry 'sudo dpkg --configure minion'." >&2
    exit 1
  }
}

require_command systemctl
require_command openssl
require_command sqlite3
require_command curl
require_command stat

PACKAGED_UNIT="/lib/systemd/system/minion.service"
LEGACY_UNIT="/etc/systemd/system/minion.service"
LEGACY_UNIT_BACKUP="$STATE_DIR/legacy-systemd-minion.service"
if [ -f "$PACKAGED_UNIT" ] && [ -f "$LEGACY_UNIT" ] && [ ! -L "$LEGACY_UNIT" ]; then
  cp -a "$LEGACY_UNIT" "$LEGACY_UNIT_BACKUP"
  rm -f "$LEGACY_UNIT"
  echo "Replaced legacy /etc/systemd/system/minion.service with the packaged unit."
fi

systemctl daemon-reload
systemctl reset-failed minion.service >/dev/null 2>&1 || true

bootstrap_client_existed=false
if [ -f "$DATA_DIR/minion.db" ] && \
   [ "$(sqlite3 "$DATA_DIR/minion.db" "SELECT COUNT(*) FROM clients WHERE name = 'bootstrap';" 2>/dev/null || printf '0')" -gt 0 ]; then
  bootstrap_client_existed=true
fi

umask 077
if ! /usr/local/bin/minion setup --config "$CONFIG" --name bootstrap --ips 127.0.0.1/32; then
  echo "Minion bootstrap failed; package configuration was not completed." >&2
  exit 1
fi

if [ -e "$BOOTSTRAP_FILE" ]; then
  if [ ! -f "$BOOTSTRAP_FILE" ] || [ -L "$BOOTSTRAP_FILE" ]; then
    echo "Minion bootstrap credential was not published as a regular root-only file." >&2
    exit 1
  fi
  if [ "$(stat -c '%u:%g:%a' "$BOOTSTRAP_FILE")" != "0:0:600" ]; then
    echo "Minion bootstrap credential has unsafe ownership or permissions." >&2
    exit 1
  fi
  echo "Minion bootstrap credential created securely."
elif [ "$bootstrap_client_existed" != true ]; then
  echo "Minion bootstrap credential was not published for the newly created bootstrap client." >&2
  exit 1
else
  echo "Existing Minion bootstrap client preserved; no new credential was generated."
fi

chmod 700 /etc/minion "$TLS_DIR" "$DATA_DIR" "$STATE_DIR"
[ ! -f "$TLS_DIR/minion.key" ] || chmod 600 "$TLS_DIR/minion.key"
[ ! -f "$TLS_DIR/minion.crt" ] || chmod 644 "$TLS_DIR/minion.crt"
[ ! -f "$DATA_DIR/minion.db" ] || chmod 600 "$DATA_DIR/minion.db"

systemctl enable minion.service >/dev/null
if ! systemctl is-active --quiet minion.service; then
  systemctl start minion.service
fi

if ! systemctl is-active --quiet minion.service; then
  echo "Minion service failed to start. Run: journalctl -u minion.service -n 100" >&2
  exit 1
fi

ready=false
for _ in $(seq 1 30); do
  if curl --silent --show-error --fail --insecure --max-time 2 \
    https://127.0.0.1:9870/api/v1/health >/dev/null; then
    ready=true
    break
  fi
  sleep 1
done
if [ "$ready" != true ]; then
  echo "Minion API failed readiness validation. Run: journalctl -u minion.service -n 100" >&2
  exit 1
fi

bind_address=$(sed -n 's/^[[:space:]]*"bind"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' "$CONFIG" | head -n 1)
bind_host=${bind_address%:*}
bind_port=${bind_address##*:}
case "$bind_host" in
  ""|"0.0.0.0"|"::"|"[::]")
    display_host=$(hostname -I 2>/dev/null | awk '{print $1}')
    [ -n "$display_host" ] || display_host="127.0.0.1"
    ;;
  *)
    display_host="$bind_host"
    ;;
esac
[ -n "$bind_port" ] || bind_port="9870"

case "$display_host" in
  \[*\]) display_url_host="$display_host" ;;
  *:*) display_url_host="[$display_host]" ;;
  *) display_url_host="$display_host" ;;
esac

machine_id=$(cat /etc/machine-id 2>/dev/null || true)
if [ -z "$machine_id" ]; then
  machine_id=$(hostname 2>/dev/null || printf 'unknown')
fi
agent_id="minion_$(printf 'minion-agent:%s' "$machine_id" | sha256sum | cut -c1-32)"

rm -rf "$BACKUP_DIR"
trap - EXIT

echo "Minion installed and running."
echo "Service: active (minion.service)"
echo "Address: https://${display_url_host}:${bind_port}"
echo "Agent ID: $agent_id"
if [ -f "$BOOTSTRAP_FILE" ]; then
  echo "Bootstrap credential (root-only): $BOOTSTRAP_FILE"
  echo "Next step: run 'sudo minion bootstrap pair --ips <AUTOMATION_IP/32>' to authorize Automation and display the initial API key once."
else
  echo "Bootstrap credential: already consumed; existing client credentials were preserved."
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

cat > "$DEB_ROOT/DEBIAN/postrm" <<'EOS'
#!/bin/sh
set -e

if [ "$1" = "purge" ]; then
  systemctl stop minion.service >/dev/null 2>&1 || true
  systemctl disable minion.service >/dev/null 2>&1 || true

  rm -rf -- /etc/minion /opt/minion /var/lib/minion

  systemctl daemon-reload >/dev/null 2>&1 || true
  systemctl reset-failed minion.service >/dev/null 2>&1 || true
fi

exit 0
EOS
chmod 755 "$DEB_ROOT/DEBIAN/postrm"

if [[ -z "$MINION_BINARY" ]]; then
  MINION_BINARY="$BUILD_ROOT/minion"
  echo "Compiling minion binary..."
  go build -o "$MINION_BINARY" ./cmd/minion
elif [[ ! -f "$MINION_BINARY" ]]; then
  echo "Compiling minion binary..."
  go build -o "$MINION_BINARY" ./cmd/minion
fi
install -m 755 "$MINION_BINARY" "$DEB_ROOT/usr/local/bin/minion"

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
