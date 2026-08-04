# Informações
<!-- Informações técnicas estáveis do projeto -->

- Binário compilado do Uzinha: `uzinha/uzinha` (excluído do .gitignore)
- Uzinha roda em `localhost:8080` e usa `wsl -e bash -c "incus ..."` para gerenciar containers
- Incus 6.0.5 instalado no WSL2 (Ubuntu) — sucessor do LXC/LXD
- Imagem Debian 12 usada para containers: fingerprint `ea8c12769f00`
- Bridge Incus: `incusbr0` (subnet `10.162.89.0/24`)
- NAT necessário no WSL2: `scripts/lxc/setup-wsl2-nat.sh` ou `ensureNAT()` automático
- O `.deb` mais recente é selecionado automaticamente pelo Uzinha via `latestDeb()`
- O minion precisa de CGO compilado (`CGO_ENABLED=1`) por causa do `go-sqlite3`
