#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat >&2 <<'EOF'
Uso: sudo bash ./install.sh <minion_<versao>_amd64.deb>

O pacote Debian é o único artefato oficial de instalação do Minion.
Este wrapper valida o pacote e usa o APT para instalar o arquivo local com suas dependências.
EOF
}

if [[ $# -ne 1 ]]; then
  usage
  exit 64
fi

package=$1

if [[ ! -f "$package" ]]; then
  echo "Erro: pacote não encontrado: $package" >&2
  exit 66
fi

if [[ "$package" != *.deb ]]; then
  echo "Erro: o arquivo informado não possui extensão .deb" >&2
  exit 65
fi

if ! command -v dpkg-deb >/dev/null 2>&1 || ! command -v apt-get >/dev/null 2>&1; then
  echo "Erro: dpkg-deb e apt-get são obrigatórios." >&2
  exit 69
fi

package_name=$(dpkg-deb -f "$package" Package 2>/dev/null || true)
architecture=$(dpkg-deb -f "$package" Architecture 2>/dev/null || true)

if [[ "$package_name" != "minion" ]]; then
  echo "Erro: o arquivo não é um pacote oficial do Minion (Package: minion)." >&2
  exit 65
fi

if [[ "$architecture" != "amd64" ]]; then
  echo "Erro: arquitetura não suportada pelo pacote atual: ${architecture:-desconhecida}. Esperado: amd64." >&2
  exit 65
fi

if [[ ${EUID:-$(id -u)} -ne 0 ]]; then
  echo "Erro: execute com sudo para instalar o pacote Debian." >&2
  exit 77
fi

package=$(readlink -f -- "$package")
echo "Instalando o pacote oficial do Minion e resolvendo dependências: $package"
exec apt-get install -y -- "$package"
