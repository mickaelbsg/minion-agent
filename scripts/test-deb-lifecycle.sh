#!/usr/bin/env bash
set -euo pipefail

PACKAGE="${1:-./minion_1.0.4_amd64.deb}"
SERVICE="minion.service"
CONFIG="/etc/minion/config.json"
DB="/opt/minion/minion.db"
CERT="/etc/minion/tls/minion.crt"
KEY="/etc/minion/tls/minion.key"
BOOTSTRAP="/var/lib/minion/bootstrap-credentials.txt"

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

cleanup() {
  sudo dpkg --remove minion >/dev/null 2>&1 || true
}
trap cleanup EXIT

[[ -f "$PACKAGE" ]] || fail "package not found: $PACKAGE"
command -v systemctl >/dev/null || fail "systemctl is unavailable"
[[ "$(ps -p 1 -o comm=)" == "systemd" ]] || fail "test host is not running systemd as PID 1"

sudo apt-get update -qq
sudo DEBIAN_FRONTEND=noninteractive apt-get install -y -qq iptables fail2ban openssl sqlite3 curl >/dev/null

# Ensure a previous failed run cannot influence the fresh-install assertions.
sudo dpkg --remove minion >/dev/null 2>&1 || true
sudo rm -rf /etc/minion /opt/minion /var/lib/minion

sudo dpkg -i "$PACKAGE"
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
[[ -n "$clients_before" ]] || fail "no persisted API client found before reinstall"

sudo dpkg -i "$PACKAGE"
sudo systemctl is-active --quiet "$SERVICE" || fail "service is not active after reinstall"
sudo test ! -e "$BOOTSTRAP" || fail "reinstall recreated bootstrap credentials"

[[ "$(sudo sha256sum "$CONFIG" | awk '{print $1}')" == "$config_hash" ]] || fail "reinstall changed configuration"
[[ "$(sudo sha256sum "$CERT" | awk '{print $1}')" == "$cert_hash" ]] || fail "reinstall changed TLS certificate"
[[ "$(sudo sha256sum "$KEY" | awk '{print $1}')" == "$key_hash" ]] || fail "reinstall changed TLS private key"
[[ "$(client_fingerprint)" == "$clients_before" ]] || fail "reinstall changed persisted API clients"
unset clients_before

curl --silent --show-error --fail --insecure \
  -H "Authorization: Bearer $api_key" \
  https://127.0.0.1:9870/api/v1/heartbeat >/dev/null

sudo dpkg --remove minion
sudo test -f "$CONFIG" || fail "package removal deleted configuration"
sudo test -f "$DB" || fail "package removal deleted database"
sudo test -f "$CERT" || fail "package removal deleted TLS certificate"
sudo test -f "$KEY" || fail "package removal deleted TLS private key"

trap - EXIT
echo "Debian package lifecycle validated successfully."
