# Repository Instructions

## Product Constraints

- Minion is a local Linux agent, not a remote shell. Never add free-form command execution, arbitrary scripts, generic execution endpoints, or embedded AI/LLM decision logic.
- The official customer flow is `sudo apt install ./minion_<version>_amd64.deb`; do not make compilation, manual file copying, JSON editing, or a separate setup command part of the normal installation.
- Prioritize, in order: Debian installation, authentication/client lifecycle, operational security, observability, Automation/n8n integration, then explicit administrative actions.
- TLS is mandatory by default. Insecure HTTP is only an explicit development fallback.
- New administrative capabilities must be explicit endpoints with strict validation, allowlists where applicable, authentication, audit logging, timeouts, payload limits, and idempotency when applicable.
- Never expose API keys or other secrets in logs, responses, commits, issues, workflows, or build artifacts. Store client credentials as hashes only.

## Architecture

- `cmd/minion` is the service and operator CLI entrypoint. `cmd/check` and `cmd/verify` are helper/debug binaries, not production entrypoints.
- `internal/server` owns HTTP routes, authentication, rate limiting, idempotency, and audit middleware.
- `internal/collectors` reads host state; `internal/security` handles API keys and IP/CIDR checks; `internal/storage` persists SQLite clients/audit/idempotency data; `internal/config` owns JSON configuration.
- `internal/admin` is the reusable setup/client/status service layer; `internal/ui` is the Bubble Tea terminal wizard.
- Runtime defaults are `/etc/minion/config.json`, `/etc/minion/tls/minion.{crt,key}`, `/opt/minion/minion.db`, `/var/lib/minion/bootstrap-credentials.txt`, and `/usr/local/bin/minion`.
- The service runs with root privileges because collectors inspect privileged host state. Preserve the hardening and writable paths in `systemd/minion.service` when changing packaging.
- SQLite-backed clients are the primary authentication source; config-file clients are only the fallback when SQLite has no persisted clients.

## Commands

- `go test ./...` runs the full unit suite; focus a test with `go test ./internal/<package> -run TestName`.
- `go test ./... -v` matches CI. CI runs lint, then verbose tests, then `go build ./cmd/minion`.
- `go build ./...` compiles all packages; `go build -o minion ./cmd/minion` builds the operator binary.
- `golangci-lint run` is configured by `.golangci.yml` with `govet`, `staticcheck`, `gofmt`, and `goimports`.
- `CGO_ENABLED=1 go build -ldflags="-s -w" -o minion ./cmd/minion` is the production-style build. `go-sqlite3` requires CGO and a C compiler.
- `PKG_VER=1.0.5 ./build_deb.sh` builds an amd64 package. The script accepts `MINION_BINARY` for lifecycle rollback tests.
- `bash scripts/test-deb-lifecycle.sh <install.deb> <upgrade.deb> [broken.deb]` validates fresh install, bootstrap, permissions, upgrade preservation, rollback, and removal. It needs Linux with systemd as PID 1, root/sudo, apt, dpkg, sqlite3, and curl.

## Packaging Rules

- `.deb` maintainer scripts must create secure config/data/TLS/state directories, initialize the service, create bootstrap credentials once, validate service health, preserve state across upgrades, and leave data on removal.
- Verify package changes with both `dpkg-deb` inspection and `scripts/test-deb-lifecycle.sh`; do not rely only on a local binary build.
- `install.sh` and `install_minion.sh` are development/manual workflows and must not become competing customer installation paths.

## Graphify

- Before architecture or code-location changes, check `graphify-out/graph.json` and query the graph first when available; confirm every result in source files because the graph is only a map.
- If the `graphify` command is unavailable, use the repository wrapper `./scripts/graphify-nvidia.sh` with the configured local environment, or inspect the real files directly rather than inventing graph results.
- After meaningful code changes, rebuild/update the graph with the repository's configured Graphify workflow; do not treat generated `graphify-out/` artifacts as source of truth.

## Verification

- Before declaring work complete, run the focused tests, `go test ./...`, `golangci-lint run`, and the relevant build. For packaging changes, also run package inspection and the Debian lifecycle harness when the environment supports systemd.
- Keep operational documentation aligned with executable scripts and CI. When they conflict, trust the script/configuration and update stale prose rather than encoding the conflict here.
