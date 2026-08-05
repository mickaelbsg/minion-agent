#!/usr/bin/env bash
set -euo pipefail

PACKAGE="${1:-./minion_1.0.5_amd64.deb}"
SERVICE="minion.service"
BOOTSTRAP="/var/lib/minion/bootstrap-credentials.txt"

fail() {
  echo "minimal Debian install test failed: $*" >&2
  exit 1
}

cleanup() {
  sudo dpkg --purge minion >/dev/null 2>&1 || true
  sudo rm -rf /etc/minion /opt/minion /var/lib/minion
}
trap cleanup EXIT

[[ -f "$PACKAGE" ]] || fail "package not found: $PACKAGE"
[[ "$(ps -p 1 -o comm=)" == "systemd" ]] || fail "test host is not running systemd as PID 1"

# The base agent must not depend on observability integrations.
depends="$(dpkg-deb -f "$PACKAGE" Depends)"
recommends="$(dpkg-deb -f "$PACKAGE" Recommends)"
printf '%s\n' "$depends" | grep -Eq '(^|,)[[:space:]]*(iptables|fail2ban)([[:space:](,]|$)' && \
  fail "iptables or fail2ban is still a hard dependency"
printf '%s\n' "$recommends" | grep -Eq '(^|,)[[:space:]]*iptables([[:space:](,]|$)' || \
  fail "iptables is not declared as recommended"
printf '%s\n' "$recommends" | grep -Eq '(^|,)[[:space:]]*fail2ban([[:space:](,]|$)' || \
  fail "fail2ban is not declared as recommended"

control_dir="$(mktemp -d)"
dpkg-deb --control "$PACKAGE" "$control_dir"
if grep -Eq 'require_command[[:space:]]+(iptables|fail2ban-client)' "$control_dir/postinst"; then
  rm -rf "$control_dir"
  fail "postinst still blocks installation on optional observability commands"
fi
rm -rf "$control_dir"

cleanup

# Remove optional packages so the test cannot pass because of the runner image.
sudo apt-get remove -y fail2ban iptables >/dev/null 2>&1 || true
command -v fail2ban-client >/dev/null 2>&1 && fail "fail2ban-client is still available"
command -v iptables >/dev/null 2>&1 && fail "iptables is still available"

install_output="$(mktemp)"
if ! sudo DEBIAN_FRONTEND=noninteractive dpkg -i "$PACKAGE" >"$install_output" 2>&1; then
  rm -f "$install_output"
  fail "dpkg -i failed without optional observability packages; output suppressed"
fi
if grep -Eq '(^|[[:space:]])API Key:|minion_sk_' "$install_output"; then
  rm -f "$install_output"
  fail "dpkg -i exposed bootstrap credentials; captured output was discarded"
fi
rm -f "$install_output"

sudo systemctl is-active --quiet "$SERVICE" || fail "service is not active"
sudo test -f /etc/minion/config.json || fail "configuration was not created"
sudo test -f /etc/minion/tls/minion.crt || fail "TLS certificate was not created"
sudo test -f /etc/minion/tls/minion.key || fail "TLS private key was not created"
sudo test -f /opt/minion/minion.db || fail "SQLite database was not created"
sudo test -f "$BOOTSTRAP" || fail "bootstrap credential file was not created"
[[ "$(sudo stat -c '%u:%g:%a' "$BOOTSTRAP")" == "0:0:600" ]] || \
  fail "bootstrap credential file is not root:root mode 0600"

curl --silent --show-error --fail --insecure --max-time 3 \
  https://127.0.0.1:9870/api/v1/health >/dev/null || fail "health endpoint is unavailable"

trap - EXIT
cleanup
echo "Minimal Debian installation without iptables and Fail2Ban validated successfully."
