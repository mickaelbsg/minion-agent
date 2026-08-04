#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

if [[ ! -f uzinha ]]; then
  echo "Compilando Uzinha..."
  go build -o uzinha .
fi

echo "Iniciando Uzinha..."
exec ./uzinha
