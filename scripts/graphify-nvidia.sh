#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
ENV_FILE="${GRAPHIFY_ENV_FILE:-$ROOT_DIR/.graphify.env.local}"

if [[ -f "$ENV_FILE" ]]; then
  set -a
  # shellcheck disable=SC1090
  . "$ENV_FILE"
  set +a
fi

: "${OPENAI_BASE_URL:=https://integrate.api.nvidia.com/v1}"
: "${OPENAI_MODEL:=openai/gpt-oss-20b}"
: "${GRAPHIFY_ALLOW_LOCAL_PROVIDERS:=1}"

export OPENAI_BASE_URL
export OPENAI_MODEL
export GRAPHIFY_ALLOW_LOCAL_PROVIDERS

if [[ -z "${OPENAI_API_KEY:-}" ]]; then
  echo "OPENAI_API_KEY is not set. Put it in $ENV_FILE or export it in your shell." >&2
  exit 1
fi

if [[ $# -eq 0 ]]; then
  exec graphify
fi

backend_flag_present=0
for arg in "$@"; do
  if [[ "$arg" == "--backend" || "$arg" == --backend=* ]]; then
    backend_flag_present=1
    break
  fi
done

case "${1:-}" in
  extract|label|cluster-only)
    if [[ $backend_flag_present -eq 0 ]]; then
      exec graphify "$@" --backend nvidia
    fi
    ;;
esac

exec graphify "$@"
