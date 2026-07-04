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
