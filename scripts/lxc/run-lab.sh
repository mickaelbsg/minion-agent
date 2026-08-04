#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
LAB_CONTAINER_NAME="${LAB_CONTAINER_NAME:-minion-lab}"
LAB_HOST_PORT="${LAB_HOST_PORT:-9870}"
DEB_PACKAGE="${1:-}"

log() { echo "[lab] $*"; }
fail() { echo "[lab] FALHA: $*" >&2; exit 1; }

cleanup() {
  log "Limpando laboratório..."
  "$SCRIPT_DIR/destroy-lab.sh" || true
}
trap cleanup EXIT

log "Verificando pré-requisitos..."
command -v lxc >/dev/null 2>&1 || fail "lxc não encontrado. Instale com: sudo snap install lxd"
command -v curl >/dev/null 2>&1 || fail "curl não encontrado"

if [[ -z "$DEB_PACKAGE" ]]; then
  DEB_PACKAGE=$(ls -t minion_*_amd64.deb 2>/dev/null | head -n1)
  if [[ -z "$DEB_PACKAGE" ]]; then
    log "Nenhum pacote .deb encontrado. Executando build_deb.sh..."
    PKG_VER=1.0.5 ./build_deb.sh
    DEB_PACKAGE=$(ls -t minion_*_amd64.deb | head -n1)
  fi
fi

if [[ ! -f "$DEB_PACKAGE" ]]; then
  fail "Pacote não encontrado: $DEB_PACKAGE"
fi

log "========================================="
log "LABORATÓRIO LXC - MINION AGENT"
log "========================================="
log "Container: $LAB_CONTAINER_NAME"
log "Pacote: $DEB_PACKAGE"
log "Porta: $LAB_HOST_PORT"
log "========================================="

log "Etapa 1/4: Criando container LXC..."
"$SCRIPT_DIR/create-lab.sh"

log "Etapa 2/4: Testando instalação..."
"$SCRIPT_DIR/test-install.sh" "$DEB_PACKAGE"

log "Etapa 3/4: Testando acesso à API..."
"$SCRIPT_DIR/test-api.sh"

log "========================================="
log "LABORATÓRIO CONCLUÍDO COM SUCESSO"
log "========================================="
log ""
log "Para testar manualmente:"
log "  lxc exec $LAB_CONTAINER_NAME -- bash"
log ""
log "Para acessar a API de fora:"
log "  curl -k -H 'Authorization: Bearer <API_KEY>' https://localhost:$LAB_HOST_PORT/api/v1/health"
log ""
log "Para destruir o laboratório:"
log "  $SCRIPT_DIR/destroy-lab.sh"
log "========================================="

trap - EXIT
