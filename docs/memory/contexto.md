# Contexto do Projeto

- O Minion é um agente local Linux em Go, executado como serviço `systemd`, que expõe uma API HTTPS autenticada para observabilidade e capacidades administrativas explícitas.
- O ponto de entrada principal é `cmd/minion`; a API fica em `internal/server`, os coletores em `internal/collectors`, a autenticação em `internal/security`, a configuração em `internal/config`, a persistência SQLite em `internal/storage` e operações de operador em `internal/admin` e `internal/ui`.
- O fluxo de runtime carrega `/etc/minion/config.json`, abre o SQLite em `/opt/minion/minion.db`, registra rotas `/api/v1/*`, autentica por API key mais IP/CIDR permitido e audita as requisições.
- O pacote `.deb` é o caminho oficial de instalação. `build_deb.sh` cria o pacote, os scripts maintainer inicializam configuração/TLS/banco/bootstrap, habilitam o serviço e preservam estado durante upgrade e remoção.
- A instalação deve funcionar com `sudo apt install ./minion_<versao>_amd64.deb`, sem compilação, cópia manual, edição de JSON ou execução separada de setup pelo cliente.
- O projeto usa `go-sqlite3`; builds e testes que dependem dele exigem CGO habilitado e toolchain C disponível.
- O CI executa `golangci-lint`, `go test ./... -v`, `go build ./cmd/minion`, dois builds de pacote e o teste completo de lifecycle em `scripts/test-deb-lifecycle.sh`.
- O teste de lifecycle Debian exige Linux com `systemd` como PID 1, `sudo`, `apt`, `dpkg`, `sqlite3`, `curl` e as dependências declaradas no pacote.
- TLS é obrigatório por padrão; HTTP inseguro só pode ser habilitado explicitamente para desenvolvimento.
- O produto não é shell remoto: não adicionar execução genérica, comandos livres, scripts arbitrários ou endpoints que convertam texto livre em ação privilegiada.
