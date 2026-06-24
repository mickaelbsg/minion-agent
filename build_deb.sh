#!/usr/bin/env bash
# Build a Debian package for Minion Agent
set -e

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
# Reload systemd, enable and start the service
systemctl daemon-reload
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
  go build -o minion ./cmd/minion/main.go
fi
cp "$(pwd)/minion" "$DEB_ROOT/usr/local/bin/minion"
chmod 755 "$DEB_ROOT/usr/local/bin/minion"

# ---- Config ----
if [[ -f "config.example.json" ]]; then
  cp "config.example.json" "$DEB_ROOT/etc/minion/config.json"
else
  echo "{\"api\": {\"bind\": \"0.0.0.0:9870\"}, \"db_path\": \"/etc/minion/minion.db\"}" > "$DEB_ROOT/etc/minion/config.json"
fi
chmod 644 "$DEB_ROOT/etc/minion/config.json"

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
Restart=always
User=root
Group=root
Environment=PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin

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
