#!/usr/bin/env bash
# Build a Debian package for Minion Agent
set -euo pipefail

PKG_NAME="minion"
PKG_VER="1.0.0"
ARCH="amd64"
BUILD_ROOT="$(mktemp -d)"
DEB_ROOT="$BUILD_ROOT/${PKG_NAME}_${PKG_VER}_${ARCH}"

echo "Creating package directory $DEB_ROOT"
mkdir -p "$DEB_ROOT/DEBIAN"
mkdir -p "$DEB_ROOT/usr/local/bin"
mkdir -p "$DEB_ROOT/etc/minion"
mkdir -p "$DEB_ROOT/etc/minion/tls"
mkdir -p "$DEB_ROOT/lib/systemd/system"

# ---- Control file ----
cat > "$DEB_ROOT/DEBIAN/control" <<EOF
Package: $PKG_NAME
Version: $PKG_VER
Section: utils
Priority: optional
Architecture: $ARCH
Maintainer: Mickael Bergson <mickael@example.com>
Depends: libc6 (>= 2.28), iptables, fail2ban, openssl, sqlite3
Description: Minion Agent – lightweight Linux data collector and API server.
 Minion gathers system information, users, services, Fail2Ban bans, and exposes a secure HTTPS API.
EOF

# ---- postinst ----
cat > "$DEB_ROOT/DEBIAN/postinst" <<'EOS'
#!/bin/sh
set -e

# If a legacy/local unit exists in /etc, systemd gives it precedence over the
# packaged unit in /lib. Keep /etc in sync to avoid stale User=/Group= settings.
if [ -f /lib/systemd/system/minion.service ]; then
  cp /lib/systemd/system/minion.service /etc/systemd/system/minion.service
fi

systemctl daemon-reload
systemctl reset-failed minion.service || true
systemctl enable minion.service
systemctl restart minion.service
exit 0
EOS
chmod 755 "$DEB_ROOT/DEBIAN/postinst"

# ---- prerm ----
cat > "$DEB_ROOT/DEBIAN/prerm" <<'EOS'
#!/bin/sh
set -e
# Stop and disable the service before removal
systemctl stop minion.service || true
systemctl disable minion.service || true
exit 0
EOS
chmod 755 "$DEB_ROOT/DEBIAN/prerm"

# ---- Binary ----
if [[ ! -f "$(pwd)/minion" ]]; then
  echo "Compiling minion binary..."
  go build -o minion ./cmd/minion
fi
cp "$(pwd)/minion" "$DEB_ROOT/usr/local/bin/minion"
chmod 755 "$DEB_ROOT/usr/local/bin/minion"

# ---- Config ----
if [[ -f "config.example.json" ]]; then
  cp "config.example.json" "$DEB_ROOT/etc/minion/config.json"
else
  echo "{\"api\": {\"bind\": \"0.0.0.0:9870\", \"allow_insecure_http\": false}, \"security\": {\"allowed_fail2ban_jails\": [\"sshd\", \"apache-auth\", \"recidive\"]}, \"db_path\": \"/opt/minion/minion.db\", \"clients\": []}" > "$DEB_ROOT/etc/minion/config.json"
fi
chmod 600 "$DEB_ROOT/etc/minion/config.json"

# ---- systemd service ----
if [[ -f "systemd/minion.service" ]]; then
  cp "systemd/minion.service" "$DEB_ROOT/lib/systemd/system/minion.service"
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

# Build the .deb package
dpkg-deb --build "$DEB_ROOT" "${PKG_NAME}_${PKG_VER}_${ARCH}.deb"
echo "Package built: ${PKG_NAME}_${PKG_VER}_${ARCH}.deb"

# Clean up
rm -rf "$BUILD_ROOT"
