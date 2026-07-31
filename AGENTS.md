# Repository Guidelines

## Project Structure & Module Organization

`cmd/` contains the entrypoints: `cmd/minion` for the service and operator UI, `cmd/check` for API key hashing, and `cmd/verify` for verification helpers. Core code lives in `internal/`: `admin/` for reusable setup/config/client/status operations, `ui/` for the terminal wizard, `server/` for HTTP handlers and middleware, `storage/` for SQLite persistence, `security/` for API key and IP checks, `config/` for configuration loading/persistence, and `collectors/` for host-level data gathering. Deployment assets are under `systemd/`; packaging and install scripts live at the repo root. Long-form docs and architecture notes live in `docs/`.

## Build, Test, and Development Commands

- `go build -o minion ./cmd/minion` builds the service binary.
- `go build ./...` compiles all packages.
- `go test ./...` runs the full test suite.
- `go test ./... -v` matches CI verbosity.
- `golangci-lint run` runs linting used by CI.
- `./build_deb.sh` builds the Debian package.

SQLite uses `go-sqlite3`, so keep CGO enabled for normal builds. Example: `CGO_ENABLED=1 go build -ldflags="-s -w" -o minion ./cmd/minion`.

## Coding Style & Naming Conventions

Use standard Go style: tabs, `gofmt`, and idiomatic package boundaries. Keep package names lowercase and concise. Exported identifiers use `CamelCase`; tests use descriptive `TestXxx` names such as `TestAuditMiddlewareCreatesEntry`. Shell scripts should prefer `set -euo pipefail` and clear logging.

## Testing Guidelines

Tests use Go’s built-in `testing` package and live beside the code they cover, for example `internal/storage/storage_test.go`. Favor focused unit tests for handlers, storage, config parsing, and security logic. Before submitting changes, run `go test ./...`, `go build ./cmd/minion`, and `golangci-lint run`.

## Commit & Pull Request Guidelines

Recent history follows short conventional subjects like `feat: bootstrap initial api key during setup`, `fix: use constant-time api key comparison`, and `docs: rewrite README as complete user manual`. Prefer `feat:`, `fix:`, or `docs:` with an imperative summary. Avoid low-signal messages like `ok`. PRs should describe behavior changes, mention service/config impact, and include test evidence.

## Agent-Specific Instructions

Before changing code, check `graphify-out/graph.json`. Use `graphify query "<question>"` first for architecture, dependency, and code-location questions; use `graphify-out/GRAPH_REPORT.md` only for broad context. Treat the graph as a map, not source of truth: confirm changes in the real files. After meaningful code changes, run `graphify update .` or `graphify extract .` for a full semantic rebuild.

## Regra obrigatória: Graphify antes de codar

Antes de modificar qualquer código, o agente deve consultar o grafo do projeto.

Fluxo obrigatório:
1. Verificar se existe `graphify-out/graph.json`.
2. Para qualquer dúvida de arquitetura, dependência, fluxo ou localização de código, usar primeiro:
   `graphify query "<pergunta objetiva>"`
3. Usar `graphify-out/GRAPH_REPORT.md` apenas para visão geral da arquitetura.
4. Só depois de consultar o Graphify, abrir os arquivos reais necessários.
5. Nunca editar código baseado apenas no grafo. O grafo serve como mapa inicial; a alteração deve ser confirmada lendo os arquivos fonte.
6. Depois de mudanças relevantes, atualizar o grafo com:
   `graphify .`

## Visão obrigatória do produto

O Minion é a mão local dentro de servidores Linux. Ele coleta informações, expõe capacidades explícitas por uma API autenticada e, futuramente, executará ações administrativas controladas pelo Automation/n8n.

O Automation/n8n é o plano de controle. O Minion é o agente local de observabilidade e execução controlada.

O Minion não é um shell remoto, não aceita comandos livres e não possui IA embarcada.

## Princípio central: facilidade para o cliente

O produto deve ser simples de instalar, simples de conectar e imediatamente útil. O cliente não deve precisar entender Go, systemd, SQLite, TLS, Argon2, a estrutura interna do repositório ou editar JSON manualmente para começar a usar o Minion.

A experiência padrão obrigatória deve ser próxima de:

```bash
sudo dpkg -i minion-agent_<versao>_amd64.deb
```

Ao final desse comando, o Minion deve estar instalado, configurado, iniciado e pronto para conexão com o Automation.

Qualquer proposta que adicione várias etapas manuais ao fluxo padrão deve ser tratada como regressão de experiência e precisa de justificativa explícita.

## Prioridades obrigatórias do projeto

Os agentes devem seguir esta ordem:

1. instalação Debian completa e simples;
2. autenticação e ciclo de vida dos clientes;
3. segurança operacional;
4. observabilidade completa;
5. integração avançada com Automation/n8n;
6. ações administrativas controladas.

Enquanto a instalação com `dpkg -i` não estiver funcional, segura, testada e documentada, não priorize funcionalidades que não contribuam diretamente para esse objetivo.

## Contrato da instalação Debian

O pacote `.deb` é o caminho oficial de instalação do produto. Ele deve:

- instalar o binário no caminho apropriado;
- instalar, habilitar e iniciar a unit systemd;
- criar os diretórios de configuração, dados e TLS com permissões seguras;
- gerar configuração inicial quando ela não existir;
- gerar certificado TLS quando ele não existir;
- inicializar o banco SQLite;
- gerar identidade persistente do agente;
- criar um cliente bootstrap seguro;
- gerar API key forte e armazenar somente seu hash;
- disponibilizar a API key original uma única vez por mecanismo legível apenas pelo root;
- nunca registrar a API key no journal, logs públicos, commits, issues ou artefatos;
- preservar configuração, banco, certificados e credenciais durante upgrades;
- não recriar credenciais em reinstalações;
- validar o serviço no pós-instalação;
- permitir remoção sem apagar dados por padrão;
- documentar instalação, upgrade, rollback, remoção e recuperação de falhas.

O cliente não deve precisar compilar o projeto, copiar arquivos, executar `minion setup` separadamente ou editar `config.json` para concluir uma instalação padrão.

`install.sh`, `install_minion.sh` e fluxos manuais são temporários ou destinados ao desenvolvimento. Eles não devem competir com o `.deb` como método oficial.

## Experiência esperada após a instalação

A saída final deve informar claramente:

- sucesso ou falha;
- estado do serviço;
- endereço da API;
- `agent_id`;
- localização root-only da credencial bootstrap;
- próximo passo para cadastrar o agente no Automation.

Exemplo conceitual:

```text
Minion instalado com sucesso.
Status: ativo
Endereço: https://192.168.1.50:9870
Agent ID: minion_a1b2c3d4
Credencial bootstrap: /root/minion-bootstrap.key
Próximo passo: cadastre este agente no Automation.
```

## Autenticação

O modelo atual é API key mais IP/CIDR autorizado. As evoluções devem priorizar:

- bootstrap seguro;
- criação, listagem, habilitação e desabilitação de clientes;
- rotação e revogação de chaves;
- armazenamento somente de hash Argon2id;
- expiração opcional;
- integração simples com Automation/n8n;
- ausência de segredos em logs e respostas.

Até que RBAC seja explicitamente planejado, clientes autenticados continuam com acesso completo às capacidades expostas.

## Segurança obrigatória

Nunca implementar:

- shell remoto;
- endpoint genérico de execução;
- comando livre enviado pelo Automation;
- scripts arbitrários;
- segredos no código, logs, issues, workflows ou artefatos.

Capacidades administrativas devem ser endpoints explícitos, com validação rigorosa, allowlist, autenticação, auditoria, timeout, limite de payload e idempotência quando aplicável.

TLS deve permanecer obrigatório por padrão. HTTP inseguro só pode existir como opção explícita de desenvolvimento.

## Valor perceptível por execução

Toda execução automatizada de manutenção deve entregar avanço funcional perceptível. Não use um ciclo inteiro apenas para lint, formatação, Dependabot, documentação isolada ou refatoração cosmética.

Uma correção interna pode ocupar uma execução somente quando bloquear diretamente a próxima funcionalidade ou corrigir risco de segurança relevante. A issue e o relatório devem explicar o bloqueio e a capacidade destravada.

## Definição de pronto

Uma entrega só está pronta quando:

- resolve um problema real do usuário ou operador;
- preserva a simplicidade de uso;
- possui testes adequados;
- possui documentação operacional;
- lint, testes e build passam;
- não expõe segredos;
- não cria shell remoto;
- não quebra instalação, upgrade ou rollback;
- possui risco e impacto claramente descritos.

## Pergunta obrigatória antes de implementar

Antes de qualquer mudança, o agente deve responder internamente:

> Isso deixa o Minion mais fácil de instalar, mais seguro ou mais útil para o cliente?

Se a resposta não for clara, a mudança não deve ser prioridade.
