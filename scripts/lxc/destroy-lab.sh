#!/usr/bin/env bash
set -euo pipefail

LAB_CONTAINER_NAME="${LAB_CONTAINER_NAME:-minion-lab}"

log() { echo "[lab] $*"; }

if ! lxc info "$LAB_CONTAINER_NAME" >/dev/null 2>&1; then
  log "Container '$LAB_CONTAINER_NAME' não encontrado."
  exit 0
fi

log "Parando container $LAB_CONTAINER_NAME..."
lxc stop "$LAB_CONTAINER_NAME" --force 2>/dev/null || true

log "Removendo container $LAB_CONTAINER_NAME..."
lxc delete "$LAB_CONTAINER_NAME" --force

rm -f /tmp/minion-lab-info.txt

log "Laboratório destruído com sucesso."
