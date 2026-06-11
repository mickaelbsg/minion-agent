# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Common Commands

| Purpose | Command |
|---|---|
| Build the binary (Go) | `go build -o minion ./...` |
| Run unit / integration tests (Go) | `go test ./...` |
| Run a single test (Go) | `go test ./path/to/package -run TestName` |
| Lint / static analysis (Go) | `go vet ./...` |
| Format code (Go) | `gofmt -w .` |
| Run the agent (systemd) | `systemctl start minion` |
| Enable on boot (systemd) | `systemctl enable minion` |
| View logs (systemd) | `journalctl -u minion -f` |

> **Note:** The repository currently contains only specification documents. When source code is added, typical Go tooling (go.mod, source files under `cmd/` or `pkg/`) will be used. Adjust commands accordingly.

## High‑Level Architecture

The Minion agent is a Go service running as a `systemd` unit. The main logical components (as described in `SPEC.md`):

1. **Collector Engine** – gathers local data (users, logins, sudo events, Fail2Ban, Wazuh, systemd services, host info, etc.).
2. **Event Engine** – detects relevant changes and emits events for further processing.
3. **Storage Engine** – persists collected data in a local SQLite database (`/opt/minion/minion.db`).
4. **API Server** – HTTP API (`/api/v1/*`) providing authenticated access to collected data. Authentication uses an API key (Argon2id hash) and IP allow‑list defined in `/etc/minion/config.yaml`.
5. **System Integration** – runs as a `systemd` service, exposing health (`GET /api/v1/health`) and other endpoints.

```
+----------------------+   HTTP API   +-------------------+
|      Severino        |<------------|      Minion       |
+----------+-----------+             +---------+---------+
           |                               |
           |  Collectors (users, logins, …) |
           |                               |
+----------v-----------+   SQLite   +-----v-------+
|   Storage Engine    |<---------->|  Event Engine |
+----------------------+            +---------------+
```

## Configuration & Runtime Files

- **Configuration**: `/etc/minion/config.yaml` – defines allowed IPs, API keys, and enabled clients.
- **Database**: `/opt/minion/minion.db` – local SQLite store.
- **Log**: `/var/log/minion/minion.log` – written by the binary and visible via `journalctl`.
- **Binary location**: installed to `/usr/local/bin/minion` (or any location on `$PATH`).

## Development Workflow

1. **Add source files** – typical Go layout (`cmd/minion/main.go`, `internal/...`).
2. **Create `go.mod`** – run `go mod init github.com/yourorg/minion`.
3. **Run `go mod tidy`** to fetch dependencies.
4. **Write tests** alongside code (files ending with `_test.go`).
5. **Run lint / format** before committing.
6. **Commit** and optionally push; CI can run `go test` and `go vet`.

## Reference Documents

- **Specification** – `SPEC.md` – detailed functional spec, component list, API endpoints, config schema.
- **Product Requirements** – `PRD.md` (currently placeholder).
- **Architecture Decision Record** – `ADR.md` – decisions about base architecture.
