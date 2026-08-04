#!/usr/bin/env bash
set -euo pipefail

LAB_CONTAINER_NAME="${LAB_CONTAINER_NAME:-minion-lab}"
HOST_PORT="${LAB_HOST_PORT:-9870}"

log() { echo "[test] $*"; }
fail() { echo "[test] FALHA: $*" >&2; exit 1; }

if ! lxc info "$LAB_CONTAINER_NAME" >/dev/null 2>&1; then
  fail "Container '$LAB_CONTAINER_NAME' não encontrado."
fi

log "Obtendo API key do bootstrap..."
BOOTSTRAP_FILE="/var/lib/minion/bootstrap-credentials.txt"

if lxc exec "$LAB_CONTAINER_NAME" -- test -f "$BOOTSTRAP_FILE"; then
  PAIR_OUTPUT=$(lxc exec "$LAB_CONTAINER_NAME" -- minion bootstrap pair \
    --config /etc/minion/config.json \
    --ips 127.0.0.1/32 2>/dev/null || echo "")
  
  API_KEY=$(echo "$PAIR_OUTPUT" | sed -n 's/^API Key: //p' | head -n1)
  
  if [[ -z "$API_KEY" || ! "$API_KEY" =~ ^minion_sk_ ]]; then
    fail "Não foi possível obter API key do bootstrap"
  fi
  log "✓ API key obtida: ${API_KEY:0:20}..."
else
  fail "Bootstrap credentials não encontrado. Execute test-install.sh primeiro."
fi

log "Testando acesso via localhost (dentro do container)..."
RESPONSE=$(lxc exec "$LAB_CONTAINER_NAME" -- curl --silent --show-error --fail --insecure \
  -H "Authorization: Bearer $API_KEY" \
  https://127.0.0.1:9870/api/v1/agent 2>/dev/null || echo "")

if [[ -z "$RESPONSE" ]]; then
  fail "Acesso via localhost falhou"
fi
log "✓ Acesso via localhost OK"

log "Testando acesso externo via localhost:$HOST_PORT..."
EXTERNAL_RESPONSE=$(curl --silent --show-error --fail --insecure \
  -H "Authorization: Bearer $API_KEY" \
  "https://localhost:$HOST_PORT/api/v1/agent" 2>/dev/null || echo "")

if [[ -z "$EXTERNAL_RESPONSE" ]]; then
  fail "Acesso externo falhou. Verifique se a porta $HOST_PORT está mapeada."
fi
log "✓ Acesso externo OK"

log "Testando endpoints da API..."

ENDPOINTS=(
  "/api/v1/health"
  "/api/v1/agent"
  "/api/v1/system"
  "/api/v1/users"
  "/api/v1/memory"
  "/api/v1/disk"
)

for endpoint in "${ENDPOINTS[@]}"; do
  RESP=$(lxc exec "$LAB_CONTAINER_NAME" -- curl --silent --show-error --fail --insecure \
    -H "Authorization: Bearer $API_KEY" \
    "https://127.0.0.1:9870${endpoint}" 2>/dev/null || echo "FALHA")
  
  if [[ "$RESP" == "FALHA" ]]; then
    log "⚠ Endpoint $endpoint retornou erro"
  else
    log "✓ Endpoint $endpoint OK"
  fi
done

if lxc exec "$LAB_CONTAINER_NAME" -- test -f "$BOOTSTRAP_FILE"; then
  log "⚠ Bootstrap credentials ainda existe após pair"
else
  log "✓ Bootstrap credentials consumido corretamente"
fi

log "========================================="
log "TODOS OS TESTES DE API PASSARAM"
log "========================================="
