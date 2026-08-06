#!/usr/bin/env bash
set -euo pipefail

LAB_CONTAINER_NAME="${LAB_CONTAINER_NAME:-minion-lab}"
LAB_IMAGE="${LAB_IMAGE:-debian:12}"
LAB_HOST_PORT="${LAB_HOST_PORT:-9870}"

log() { echo "[lab] $*"; }

if ! command -v lxc >/dev/null 2>&1; then
  log "ERRO: lxc não encontrado. Instale com: sudo snap install lxd"
  exit 1
fi

if lxc info "$LAB_CONTAINER_NAME" >/dev/null 2>&1; then
  log "Container '$LAB_CONTAINER_NAME' já existe. Use destroy-lab.sh para remover."
  exit 1
fi

log "Criando container $LAB_CONTAINER_NAME com $LAB_IMAGE..."
lxc launch "$LAB_IMAGE" "$LAB_CONTAINER_NAME" \
  -c security.nesting=true \
  -c systemd=true

log "Aguardando systemd inicializar..."
lxc exec "$LAB_CONTAINER_NAME" -- systemctl is-system-running --wait >/dev/null 2>&1 || true

log "Instalando dependências..."
lxc exec "$LAB_CONTAINER_NAME" -- apt-get update -qq
lxc exec "$LAB_CONTAINER_NAME" -- apt-get install -y -qq \
  fail2ban \
  iptables \
  openssl \
  sqlite3 \
  curl \
  gnupg

log "Configurando acesso externo na porta $LAB_HOST_PORT..."
lxc config device add "$LAB_CONTAINER_NAME" minion-port \
  proxy listen=tcp:0.0.0.0:"$LAB_HOST_PORT" \
  connect=tcp:127.0.0.1:"$LAB_HOST_PORT"

CONTAINER_IP=$(lxc exec "$LAB_CONTAINER_NAME" -- hostname -I | awk '{print $1}')
log "Container criado com IP: $CONTAINER_IP"
log "Acesso externo: https://localhost:$LAB_HOST_PORT"

cat > /tmp/minion-lab-info.txt <<EOF
CONTAINER_NAME=$LAB_CONTAINER_NAME
CONTAINER_IP=$CONTAINER_IP
HOST_PORT=$LAB_HOST_PORT
EOF

log "Laboratório pronto. Execute test-install.sh para validar."
