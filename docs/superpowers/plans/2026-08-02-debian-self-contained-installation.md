# Debian Self-Contained Installation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make `sudo apt install ./minion_<version>_amd64.deb` install, configure, start, and validate Minion without manual dependency installation, setup commands, or JSON editing.

**Architecture:** The Debian package will declare all required host dependencies in its Debian metadata; `apt install ./package.deb` resolves them before running the package maintainer scripts. The `postinst` then initializes persistent directories/configuration/TLS/SQLite/bootstrap credentials, enables the packaged systemd unit, and validates the running health endpoint. Existing config, database, TLS, identity, and credentials will be preserved on reinstall and upgrade; failures will leave a diagnostic state and invoke the existing rollback path.

**Tech Stack:** Debian `dpkg` maintainer scripts, Bash/POSIX shell, Go binary, systemd, OpenSSL, SQLite, WSL Ubuntu integration tests.

## Global Constraints

- The official customer path is one command: `sudo apt install ./minion_<version>_amd64.deb`.
- The package must not require `minion setup`, manual dependency installation, config editing, or source compilation on the customer host.
- TLS remains mandatory by default.
- API keys are never written to public logs or normal service output; the original bootstrap key is exposed only through a root-readable file/mechanism.
- Upgrades and reinstalls preserve configuration, database, certificates, agent identity, and existing credentials.
- The runtime must not provide shell execution or accept arbitrary commands.
- Validation must cover a clean WSL installation, reinstall/upgrade preservation, service state, TLS, SQLite initialization, and bootstrap usability.

---

### Task 1: Define Package Dependency and Bootstrap Contract

**Files:**
- Modify: `build_deb.sh:22-32` and generated maintainer scripts
- Review: `README.md:158-220`, `docs/bootstrap-credentials.md`, `internal/admin/service.go`
- Test: `scripts/test-deb-lifecycle.sh`

**Interfaces:**
- Produces a package whose `postinst` owns host dependency installation and first-boot initialization.
- The script must expose only a root-readable bootstrap credential path and a sanitized final status summary.

- [ ] **Step 1: Write failing lifecycle assertions**

Add assertions to the Debian lifecycle test for: missing host tools on a clean image, successful `apt install ./package.deb`, generated `/etc/minion/config.json`, generated TLS key/certificate, initialized `/opt/minion/minion.db`, active `minion.service`, and root-only bootstrap credential storage.

- [ ] **Step 2: Run the lifecycle test to capture the current failure**

Run:

```bash
sudo bash scripts/test-deb-lifecycle.sh
```

Expected current failure: the legacy `1.0.0` artifact leaves `minion` unconfigured because its dependency/bootstrap contract is stale.

- [ ] **Step 3: Specify the package contract in the build script comments/output**

Make the generated package metadata and post-install output explicitly describe that `postinst` installs host tools and completes initialization. Do not claim that SQLite CLI is required if the Go SQLite driver is already embedded in the binary.

- [ ] **Step 4: Re-run the focused assertions after implementation tasks**

The complete lifecycle test becomes the gate for this contract and must pass before declaring the package usable.

### Task 2: Make `postinst` Idempotent After Dependency Resolution

**Files:**
- Modify: `build_deb.sh:69-167`
- Review: `internal/admin/service.go`, `internal/bootstrap/credentials.go`, `internal/storage/storage.go`
- Test: generated package maintainer scripts through `scripts/test-deb-lifecycle.sh`

**Interfaces:**
- `postinst configure` runs after `apt` has installed required host commands, then runs `/usr/local/bin/minion setup --config /etc/minion/config.json` only when initialization is needed.
- Existing files and credentials are never recreated during reinstall or upgrade.

- [ ] **Step 1: Add a failing test for dependency resolution**

Run the package in a clean WSL snapshot or isolated root and assert that `apt install ./package.deb` installs `fail2ban`, `iptables`, `openssl`, and `sqlite3` before configuration.

- [ ] **Step 2: Remove non-resolvable host tools from mandatory `Depends`**

Keep the glibc floor plus `iptables`, `fail2ban`, `openssl`, and `sqlite3` in `Depends`; `apt` resolves them before `postinst` runs. Do not invoke `apt-get` from a maintainer script while `dpkg` holds the package database lock.

- [ ] **Step 3: Add an idempotent dependency installer**

In the generated `postinst`, verify that required commands are available, fail with an actionable message if package resolution was incomplete, and never print credentials or include them in command arguments.

- [ ] **Step 4: Preserve initialization state**

Create `/etc/minion`, `/etc/minion/tls`, `/opt/minion`, and `/var/lib/minion` with root ownership and mode `0700`. Invoke setup only when config/TLS/database/identity/bootstrap state is absent. Existing configuration and credentials must remain untouched.

- [ ] **Step 5: Validate bootstrap creation and output handling**

Capture setup output in a mode `0600` root-owned temporary file, move it atomically to the documented bootstrap credential file only when a new credential was generated, remove temporary files on every path, and ensure the journal receives no API key.

- [ ] **Step 6: Run the focused lifecycle test**

Run:

```bash
sudo bash scripts/test-deb-lifecycle.sh
```

Expected: package configuration completes and all first-install assertions pass.

### Task 3: Align Package, Unit, and Documentation Artifacts

**Files:**
- Modify: `build_deb.sh:180-237`
- Modify: `README.md:158-220`
- Modify: `docs/bootstrap-credentials.md`
- Modify: `scripts/test-deb-lifecycle.sh`
- Review: `install.sh`, `install_minion.sh`, `systemd/minion.service`

**Interfaces:**
- The generated `.deb` is the only documented production installer.
- Development installers cannot silently install a competing systemd unit or contradict the package flow.

- [ ] **Step 1: Add a failing version consistency check**

Assert that the package version, generated filename, README example, and reported installed version use the same value supplied by `PKG_VER`.

- [ ] **Step 2: Generate the package from the current binary and current unit**

Ensure the build script uses the current `systemd/minion.service`, current bootstrap implementation, and current version metadata rather than leaving a stale checked-in `1.0.0` artifact as the apparent release package.

- [ ] **Step 3: Make manual installers explicitly development-only**

Update their messages and documentation so they do not compete with the `.deb` path or create an older unit under `/etc/systemd/system` that shadows `/lib/systemd/system/minion.service`.

- [ ] **Step 4: Document the one-command result**

Document the exact final output: service status, HTTPS address, agent ID, root-only bootstrap credential location, and Automation registration next step. Include upgrade, reinstall, removal-with-data-preservation, and failure recovery behavior.

- [ ] **Step 5: Run documentation and package consistency checks**

Run:

```bash
grep -R "minion_1.0.0\|minion_1.0.4" README.md docs build_deb.sh scripts
```

Expected: no stale version examples remain outside intentional historical notes.

### Task 4: Validate Real WSL Installation and Upgrade Lifecycle

**Files:**
- Modify: `scripts/test-deb-lifecycle.sh`
- Create: `docs/debian-install-validation.md`

**Interfaces:**
- Provides a repeatable WSL validation procedure and records evidence, not just exit codes.

- [ ] **Step 1: Build the package in WSL**

Build the Linux binary/package using the repository’s current source and record package metadata with `dpkg-deb -I` and file contents with `dpkg-deb -c`.

- [ ] **Step 2: Run first install from a clean state**

Execute only:

```bash
sudo apt install ./minion_<version>_amd64.deb
```

Then verify package status is `install ok installed`, service state is active, TLS files exist with safe permissions, SQLite schema exists, and the bootstrap credential is readable only by root.

- [ ] **Step 3: Verify API behavior**

Call the health endpoint over HTTPS, verify an authenticated endpoint using the bootstrap credential, verify invalid credentials fail, and verify no secret appears in `journalctl -u minion`.

- [ ] **Step 4: Verify reinstall preservation**

Record hashes/contents of config, database, TLS, agent identity, and credential state; run the same `apt install ./package.deb` again; assert the existing state is preserved and no replacement bootstrap key is generated.

- [ ] **Step 5: Verify upgrade rollback and final reporting**

Install the next package version, force a controlled post-install failure in a disposable test, verify prior operational files are restored, then run a successful upgrade and record evidence in `docs/debian-install-validation.md`.

- [ ] **Step 6: Run repository verification**

Run:

```bash
go test ./...
go build ./...
golangci-lint run
```

Run the WSL lifecycle script separately. If a tool is unavailable in the Windows host, run the equivalent command inside WSL and record that fact.

## Completion Criteria

- A clean WSL Ubuntu installation succeeds with only `sudo apt install ./package.deb`.
- No manual `apt-get`, `minion setup`, config editing, or certificate generation is required.
- Service starts with TLS enabled and health endpoint responds.
- Bootstrap credential is generated once, root-only, and never logged.
- Reinstall and upgrade preserve operational state.
- Failed package configuration restores the previous operational state.
- Package version, artifact contents, unit file, tests, and documentation agree.
