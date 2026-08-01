#!/usr/bin/env bash
set -euo pipefail

INSTALL_PACKAGE="${1:-./minion_1.0.4_amd64.deb}"
UPGRADE_PACKAGE="${2:-$INSTALL_PACKAGE}"
BROKEN_PACKAGE="${3:-}"
SERVICE="minion.service"
CONFIG="/etc/minion/config.json"
DB="/opt/minion/minion.db"
CERT="/etc/minion/tls/minion.crt"
KEY="/etc/minion/tls/minion.key"
BINARY="/usr/local/bin/minion"
BOOTSTRAP="/var/lib/minion/bootstrap-credentials.txt"
UPGRADE_BACKUP="/var/lib/minion/upgrade-backup"

fail() {
  echo "deb lifecycle test failed: $*" >&2
  exit 1
}

assert_mode() {
  local path="$1"
  local expected="$2"
  local actual
  actual="$(sudo stat -c '%a' "$path")"
  [[ "$actual" == "$expected" ]] || fail "$path mode is $actual, expected $expected"
}

client_fingerprint() {
  sudo sqlite3 "$DB" \
    "SELECT name || '|' || allowed_ips || '|' || api_key_hash || '|' || enabled FROM clients ORDER BY name;"
}

package_version() {
  dpkg-deb -f "$1" Version
}

cleanup() {
  sudo dpkg --remove minion >/dev/null 2>&1 || true
}
trap cleanup EXIT

[[ -f "$INSTALL_PACKAGE" ]] || fail "install package not found: $INSTALL_PACKAGE"
[[ -f "$UPGRADE_PACKAGE" ]] || fail "upgrade package not found: $UPGRADE_PACKAGE"
[[ -z "$BROKEN_PACKAGE" || -f "$BROKEN_PACKAGE" ]] || fail "broken package not found: $BROKEN_PACKAGE"
command -v systemctl >/dev/null || fail "systemctl is unavailable"
[[ "$(ps -p 1 -o comm=)" == "systemd" ]] || fail "test host is not running systemd as PID 1"

install_version="$(package_version "$INSTALL_PACKAGE")"
upgrade_version="$(package_version "$UPGRADE_PACKAGE")"
if [[ "$INSTALL_PACKAGE" != "$UPGRADE_PACKAGE" ]]; then
  dpkg --compare-versions "$upgrade_version" gt "$install_version" || \
    fail "upgrade package version $upgrade_version must be newer than $install_version"
fi
if [[ -n "$BROKEN_PACKAGE" ]]; then
  broken_version="$(package_version "$BROKEN_PACKAGE")"
  dpkg --compare-versions "$broken_version" gt "$upgrade_version" || \
    fail "broken package version $broken_version must be newer than $upgrade_version"
fi

sudo apt-get update -qq
sudo DEBIAN_FRONTEND=noninteractive apt-get install -y -qq iptables fail2ban openssl sqlite3 curl >/dev/null

# Ensure a previous failed run cannot influence the fresh-install assertions.
sudo dpkg --remove minion >/dev/null 2>&1 || true
sudo rm -rf /etc/minion /opt/minion /var/lib/minion

sudo dpkg -i "$INSTALL_PACKAGE"
[[ "$(dpkg-query -W -f='${Version}' minion)" == "$install_version" ]] || fail "unexpected installed package version"
sudo systemctl is-active --quiet "$SERVICE" || fail "service is not active after installation"

for path in "$CONFIG" "$DB" "$CERT" "$KEY" "$BOOTSTRAP"; do
  sudo test -f "$path" || fail "missing file after installation: $path"
done

assert_mode "$CONFIG" 600
assert_mode "$DB" 600
assert_mode "$KEY" 600
assert_mode "$BOOTSTRAP" 600

pair_output="$(sudo /usr/local/bin/minion bootstrap pair --config "$CONFIG" --ips 127.0.0.1/32)"
api_key="$(printf '%s\n' "$pair_output" | sed -n 's/^API Key: //p' | head -n 1)"
[[ "$api_key" == minion_sk_* ]] || fail "bootstrap pair did not return a valid API key"
echo "::add-mask::$api_key"
unset pair_output
sudo test ! -e "$BOOTSTRAP" || fail "bootstrap credential file was not removed after pairing"

response="$(curl --silent --show-error --fail --insecure \
  -H "Authorization: Bearer $api_key" \
  https://127.0.0.1:9870/api/v1/agent)"
printf '%s' "$response" | grep -q '"agent_id"' || fail "authenticated agent endpoint returned an unexpected response"

config_hash="$(sudo sha256sum "$CONFIG" | awk '{print $1}')"
cert_hash="$(sudo sha256sum "$CERT" | awk '{print $1}')"
key_hash="$(sudo sha256sum "$KEY" | awk '{print $1}')"
clients_before="$(client_fingerprint)"
[[ -n "$clients_before" ]] || fail "no persisted API client found before package transition"

sudo dpkg -i "$UPGRADE_PACKAGE"
[[ "$(dpkg-query -W -f='${Version}' minion)" == "$upgrade_version" ]] || fail "package was not upgraded to $upgrade_version"
sudo systemctl is-active --quiet "$SERVICE" || fail "service is not active after package upgrade"
sudo test ! -e "$BOOTSTRAP" || fail "package upgrade recreated bootstrap credentials"
sudo test ! -e "$UPGRADE_BACKUP" || fail "successful package upgrade left a temporary rollback snapshot"

[[ "$(sudo sha256sum "$CONFIG" | awk '{print $1}')" == "$config_hash" ]] || fail "package upgrade changed configuration"
[[ "$(sudo sha256sum "$CERT" | awk '{print $1}')" == "$cert_hash" ]] || fail "package upgrade changed TLS certificate"
[[ "$(sudo sha256sum "$KEY" | awk '{print $1}')" == "$key_hash" ]] || fail "package upgrade changed TLS private key"
[[ "$(client_fingerprint)" == "$clients_before" ]] || fail "package upgrade changed persisted API clients"

curl --silent --show-error --fail --insecure \
  -H "Authorization: Bearer $api_key" \
  https://127.0.0.1:9870/api/v1/heartbeat >/dev/null

if [[ -n "$BROKEN_PACKAGE" ]]; then
  binary_hash="$(sudo sha256sum "$BINARY" | awk '{print $1}')"

  if sudo dpkg -i "$BROKEN_PACKAGE"; then
    fail "intentionally broken package unexpectedly installed successfully"
  fi

  sudo systemctl is-active --quiet "$SERVICE" || fail "service is not active after automatic rollback"
  sudo test -f "$UPGRADE_BACKUP/ready" || fail "failed upgrade did not retain the root-only recovery snapshot"
  assert_mode "$UPGRADE_BACKUP" 700

  [[ "$(sudo sha256sum "$BINARY" | awk '{print $1}')" == "$binary_hash" ]] || fail "rollback did not restore the previous binary"
  [[ "$(sudo sha256sum "$CONFIG" | awk '{print $1}')" == "$config_hash" ]] || fail "rollback did not restore configuration"
  [[ "$(sudo sha256sum "$CERT" | awk '{print $1}')" == "$cert_hash" ]] || fail "rollback did not restore TLS certificate"
  [[ "$(sudo sha256sum "$KEY" | awk '{print $1}')" == "$key_hash" ]] || fail "rollback did not restore TLS private key"
  [[ "$(client_fingerprint)" == "$clients_before" ]] || fail "rollback did not restore persisted API clients"

  curl --silent --show-error --fail --insecure \
    -H "Authorization: Bearer $api_key" \
    https://127.0.0.1:9870/api/v1/heartbeat >/dev/null || \
    fail "previous API key no longer authenticates after rollback"
fi
unset clients_before

sudo dpkg --remove minion
sudo test -f "$CONFIG" || fail "package removal deleted configuration"
sudo test -f "$DB" || fail "package removal deleted database"
sudo test -f "$CERT" || fail "package removal deleted TLS certificate"
sudo test -f "$KEY" || fail "package removal deleted TLS private key"

trap - EXIT
echo "Debian package install, upgrade, rollback and removal lifecycle validated successfully."
