#!/usr/bin/env bash
set -euo pipefail

LAB_CONTAINER_NAME="${LAB_CONTAINER_NAME:-minion-lab}"
DEB_PACKAGE="${1:-}"

log() { echo "[test] $*"; }
fail() { echo "[test] FALHA: $*" >&2; exit 1; }

if ! lxc info "$LAB_CONTAINER_NAME" >/dev/null 2>&1; then
  fail "Container '$LAB_CONTAINER_NAME' não encontrado. Execute create-lab.sh primeiro."
fi

if [[ -z "$DEB_PACKAGE" ]]; then
  DEB_PACKAGE=$(ls -t minion_*_amd64.deb 2>/dev/null | head -n1)
  if [[ -z "$DEB_PACKAGE" ]]; then
    fail "Nenhum pacote .deb encontrado. Execute build_deb.sh primeiro."
  fi
fi

if [[ ! -f "$DEB_PACKAGE" ]]; then
  fail "Pacote não encontrado: $DEB_PACKAGE"
fi

log "Usando pacote: $DEB_PACKAGE"

log "Copiando pacote para o container..."
lxc file push "$DEB_PACKAGE" "$LAB_CONTAINER_NAME/tmp/minion.deb"

log "Instalando pacote no container..."
lxc exec "$LAB_CONTAINER_NAME" -- bash -c \
  "DEBIAN_FRONTEND=noninteractive apt-get install -y -qq /tmp/minion.deb"

log "Validando instalação..."

INSTALLED_VERSION=$(lxc exec "$LAB_CONTAINER_NAME" -- dpkg-query -W -f='${Version}' minion 2>/dev/null || echo "")
if [[ -z "$INSTALLED_VERSION" ]]; then
  fail "Pacote não está instalado"
fi
log "✓ Pacote instalado: $INSTALLED_VERSION"

if ! lxc exec "$LAB_CONTAINER_NAME" -- systemctl is-active --quiet minion.service; then
  fail "Serviço não está ativo"
fi
log "✓ Serviço está ativo"

for path in /etc/minion/config.json /etc/minion/tls/minion.crt /etc/minion/tls/minion.key /opt/minion/minion.db; do
  if ! lxc exec "$LAB_CONTAINER_NAME" -- test -f "$path"; then
    fail "Arquivo não encontrado: $path"
  fi
  log "✓ Arquivo existe: $path"
done

for path in /etc/minion/config.json /etc/minion/tls/minion.key /opt/minion/minion.db; do
  MODE=$(lxc exec "$LAB_CONTAINER_NAME" -- stat -c '%a' "$path")
  if [[ "$MODE" != "600" ]]; then
    fail "Permissão incorreta em $path: $MODE (esperado: 600)"
  fi
  log "✓ Permissão correta em $path: $MODE"
done

if lxc exec "$LAB_CONTAINER_NAME" -- test -f /var/lib/minion/bootstrap-credentials.txt; then
  BOOTSTRAP_MODE=$(lxc exec "$LAB_CONTAINER_NAME" -- stat -c '%a' /var/lib/minion/bootstrap-credentials.txt)
  log "✓ Bootstrap credentials criado (modo: $BOOTSTRAP_MODE)"
else
  log "⚠ Bootstrap credentials não encontrado (pode ter sido consumido)"
fi

log "Testando health check via API..."
HEALTH_RESPONSE=$(lxc exec "$LAB_CONTAINER_NAME" -- curl --silent --show-error --fail --insecure \
  https://127.0.0.1:9870/api/v1/health 2>/dev/null || echo "")

if [[ -z "$HEALTH_RESPONSE" ]]; then
  fail "Health check falhou"
fi
log "✓ Health check OK: $HEALTH_RESPONSE"

log "========================================="
log "TODOS OS TESTES DE INSTALAÇÃO PASSARAM"
log "========================================="
