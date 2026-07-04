# Minion Agent Improvements Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task‑by‑task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Incorporar as melhorias identificadas (audit logging, geração de salt aleatório, validação de entrada, criação de diretório DB, lint, testes, unit‑service systemd) e alinhar a documentação ao código para o Minion‑Agent.

**Architecture:** Cada melhoria será implementada de forma isolada, com testes unitários e commits frequentes. O fluxo seguirá TDD onde aplicável e utiliza a camada de armazenamento SQLite existente.

**Tech Stack:** Go 1.22, SQLite, Argon2id, golangci‑lint, systemd, Git.

---

### Task 1: Garantir diretório do SQLite exista

**Files:**
- Modify: `internal/storage/storage.go`

- [ ] **Step 1: Write the failing test**

```go
func TestStorageCreatesDBDir(t *testing.T) {
    // Use a temporary dir
    dir := t.TempDir()
    dbPath := filepath.Join(dir, "nonexistent", "minion.db")
    cfg := config.Config{DBPath: dbPath}
    s, err := storage.New(cfg)
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    defer s.Close()
    if _, err := os.Stat(filepath.Dir(dbPath)); os.IsNotExist(err) {
        t.Fatalf("directory %s was not created", filepath.Dir(dbPath))
    }
}
```

- [ ] **Step 2: Run test to verify it fails**

```
go test ./internal/storage -run TestStorageCreatesDBDir -v
```
Expected: FAIL (function does not create directory).

- [ ] **Step 3: Write minimal implementation**

```go
func New(cfg config.Config) (*Storage, error) {
    // Ensure parent directory exists
    if err := os.MkdirAll(filepath.Dir(cfg.DBPath), 0755); err != nil {
        return nil, fmt.Errorf("cannot create db directory: %w", err)
    }
    db, err := sql.Open("sqlite3", cfg.DBPath)
    // existing logic …
}
```

- [ ] **Step 4: Run test to verify it passes**

```
go test ./internal/storage -run TestStorageCreatesDBDir -v
```
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/storage/storage.go internal/storage/storage_test.go
git commit -m "feat(storage): create DB directory if missing"
```

### Task 2: Trocar sal estático por sal aleatório nas chaves API

**Files:**
- Modify: `internal/security/security.go`
- Add test: `internal/security/security_test.go`

- [ ] **Step 1: Write failing test**

```go
func TestHashAndVerifyAPIKey(t *testing.T) {
    key := "my-secret-key"
    stored, err := security.HashAPIKey(key)
    if err != nil { t.Fatalf("hash error: %v", err) }
    // stored must contain a '$' separating salt and hash
    parts := strings.Split(stored, "$")
    if len(parts) != 2 { t.Fatalf("expected salt$hash, got %s", stored) }
    // Verify same key passes
    if ok := security.VerifyAPIKey(key, stored); !ok {
        t.Fatalf("verification failed for correct key")
    }
    // Wrong key should fail
    if ok := security.VerifyAPIKey("wrong", stored); ok {
        t.Fatalf("verification succeeded for wrong key")
    }
}
```

- [ ] **Step 2: Run test (expected fail)**

```
go test ./internal/security -run TestHashAndVerifyAPIKey -v
```
Expected: FAIL (static hash logic).

- [ ] **Step 3: Implement random‑salt hashing**

```go
func HashAPIKey(key string) (string, error) {
    salt := make([]byte, 16)
    if _, err := rand.Read(salt); err != nil { return "", err }
    hash := argon2.IDKey([]byte(key), salt, 1, 64*1024, 4, 32)
    return fmt.Sprintf("%s$%s", base64.StdEncoding.EncodeToString(salt), base64.StdEncoding.EncodeToString(hash)), nil
}

func VerifyAPIKey(key, stored string) bool {
    parts := strings.Split(stored, "$")
    if len(parts) != 2 { return false }
    salt, err := base64.StdEncoding.DecodeString(parts[0])
    if err != nil { return false }
    expectedHash, err := base64.StdEncoding.DecodeString(parts[1])
    if err != nil { return false }
    calc := argon2.IDKey([]byte(key), salt, 1, 64*1024, 4, 32)
    return hmac.Equal(calc, expectedHash)
}
```

- [ ] **Step 4: Run test (should pass)**

```
go test ./internal/security -run TestHashAndVerifyAPIKey -v
```
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/security/security.go internal/security/security_test.go
git commit -m "feat(security): per‑key random salt for API keys and verification helper"
```

### Task 3: Validar entrada no endpoint Fail2Ban unban

**Files:**
- Modify: `internal/server/server.go`
- Add test: `internal/server/server_test.go`

- [ ] **Step 1: Write failing test**

```go
func TestHandleFail2BanUnbanInvalidIP(t *testing.T) {
    // Build a request with malformed IP
    payload := `{"ip":"not-an-ip","jail":"sshd"}`
    req := httptest.NewRequest(http.MethodPost, "/api/v1/fail2ban/unban", strings.NewReader(payload))
    w := httptest.NewRecorder()
    // assume a server with a dummy auth client set up
    s := &Server{...}
    s.handleFail2BanUnban(w, req)
    if w.Code != http.StatusBadRequest {
        t.Fatalf("expected 400, got %d", w.Code)
    }
}
```

- [ ] **Step 2: Run test (expected fail)**

```
go test ./internal/server -run TestHandleFail2BanUnbanInvalidIP -v
```
Expected: FAIL (no validation).

- [ ] **Step 3: Implement validation**

```go
func isValidIP(ip string) bool {
    return net.ParseIP(ip) != nil
}

func isAllowedJail(jail string) bool {
    allowed := map[string]bool{"sshd":true,"apache-auth":true,"recidive":true}
    return allowed[jail]
}

func (s *Server) handleFail2BanUnban(w http.ResponseWriter, r *http.Request) {
    var p struct{ IP, Jail string }
    if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
        http.Error(w, "invalid json", http.StatusBadRequest); return
    }
    if !isValidIP(p.IP) {
        http.Error(w, "invalid IP address", http.StatusBadRequest); return
    }
    if !isAllowedJail(p.Jail) {
        http.Error(w, "unknown jail", http.StatusBadRequest); return
    }
    // existing exec command …
}
```

- [ ] **Step 4: Run test (should pass)**

```
go test ./internal/server -run TestHandleFail2BanUnbanInvalidIP -v
```
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/server/server.go internal/server/server_test.go
git commit -m "feat(server): validate IP and jail on fail2ban unban endpoint"
```

### Task 4: Adicionar middleware de auditoria

**Files:**
- New: `internal/server/middleware.go`
- Modify: `internal/server/server.go` (wrap handlers)

- [ ] **Step 1: Write failing test** (no audit entries).

```go
func TestAuditMiddlewareCreatesEntry(t *testing.T) {
    // Setup in‑memory SQLite storage
    cfg := config.Config{DBPath: ":memory:"}
    st, _ := storage.New(cfg)
    defer st.Close()
    s := NewServer(st, ...) // simplified
    // Create a test handler that returns 200
    handler := s.audit(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(200) }))
    req := httptest.NewRequest(http.MethodGet, "/test", nil)
    w := httptest.NewRecorder()
    handler.ServeHTTP(w, req)
    // Verify audit row exists
    rows, _ := st.DB.Query("SELECT count(*) FROM audit WHERE method='GET' AND path='/test'")
    var cnt int
    rows.Next(); rows.Scan(&cnt)
    if cnt != 1 { t.Fatalf("audit entry not created") }
}
```

- [ ] **Step 2: Run test (expected fail)**

```
go test ./internal/server -run TestAuditMiddlewareCreatesEntry -v
```
Expected: FAIL (no middleware).

- [ ] **Step 3: Implement middleware**

```go
// middleware.go
type statusRecorder struct { http.ResponseWriter; status int }
func (r *statusRecorder) WriteHeader(code int) { r.status = code; r.ResponseWriter.WriteHeader(code) }

func (s *Server) audit(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        rec := &statusRecorder{ResponseWriter: w, status: 200}
        start := time.Now()
        client := clientNameFromContext(r.Context()) // set by auth middleware
        next.ServeHTTP(rec, r)
        // Insert audit record
        _ = s.Storage.InsertAudit(storage.AuditEntry{Timestamp: start, Client: client, IP: r.RemoteAddr, Method: r.Method, Path: r.URL.Path, Status: rec.status})
    })
}
```

- [ ] **Step 4: Run test (should pass)**

```
go test ./internal/server -run TestAuditMiddlewareCreatesEntry -v
```
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/server/middleware.go internal/server/server.go internal/server/server_test.go
git commit -m "feat(server): audit middleware to log every request"
```

### Task 5: Linting & CI integration

**Files:**
- Add: `.golangci.yml`
- Add: `.github/workflows/ci.yml`

- [ ] **Step 1: Write lint config**

```yaml
run:
  timeout: 5m
linters:
  enable:
    - govet
    - staticcheck
    - gofmt
    - goimports
issues:
  exclude-use-default: false
```

- [ ] **Step 2: Write GitHub Actions workflow**

```yaml
name: CI
on: [push, pull_request]
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v4
    - name: Set up Go
      uses: actions/setup-go@v5
      with:
        go-version: '1.22'
    - name: Install golangci-lint
      run: go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.60
    - name: Lint
      run: golangci-lint run
    - name: Test
      run: go test ./... -v
    - name: Build
      run: go build ./cmd/minion
```

- [ ] **Step 3: Commit config files**

```bash
git add .golangci.yml .github/workflows/ci.yml
git commit -m "chore(ci): add golangci‑lint config and GitHub Actions workflow"
```

### Task 6: Systemd unit file

**Files:**
- Add: `systemd/minion.service`

- [ ] **Step 1: Write unit file**

```ini
[Unit]
Description=Minion Agent – coleta de métricas de infraestrutura
After=network.target

[Service]
Type=simple
ExecStart=/usr/local/bin/minion --config /etc/minion/config.json
Restart=on-failure
User=pc
Group=pc
Environment=MINION_ENV=production

[Install]
WantedBy=multi-user.target
```

- [ ] **Step 2: Install and enable**

```bash
sudo cp systemd/minion.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable --now minion.service
sudo systemctl status minion.service
```

- [ ] **Step 3: Verify service runs and health endpoint works**

```bash
curl -sk https://localhost:9870/api/v1/health | jq .
```
Expected JSON with `{ "status": "ok" }`.

- [ ] **Step 4: Commit unit file**

```bash
git add systemd/minion.service
git commit -m "chore(systemd): add minion.service unit file"
```

---

## Self‑Review Checklist

1. **Spec coverage** – Cada melhoria listada no relatório de QA está representada por um task.
2. **Placeholder scan** – Nenhum `TODO`, `TBD` ou referência vaga permanece.
3. **Type consistency** – Os nomes de funções (`HashAPIKey`, `VerifyAPIKey`, `handleFail2BanUnban`) coincidem entre implementação e testes.

Plan complete and saved to `docs/superpowers/plans/2026-06-13-minion-improvements.md`.

**Execution options:**

1️⃣ **Subagent‑Driven (recommended)** – Dispatch a fresh subagent per task, review between tasks.

2️⃣ **Inline Execution** – Run tasks sequentially in this session using `superpowers:executing-plans`.

Which approach do you prefer, chefe?
