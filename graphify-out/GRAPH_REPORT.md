# Graph Report - minion-agent  (2026-08-02)

## Corpus Check
- 105 files · ~42,906 words
- Verdict: corpus is large enough that graph structure adds value.

## Summary
- 806 nodes · 1213 edges · 78 communities (68 shown, 10 thin omitted)
- Extraction: 87% EXTRACTED · 13% INFERRED · 0% AMBIGUOUS · INFERRED: 161 edges (avg confidence: 0.8)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `ec8ef57d`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- closeWithError
- New
- Model
- NewService
- Server
- .idempotentAction
- Consume
- GetJournalLogs
- Global Constraints
- New
- rateLimiter
- info.go
- parseMemory
- parseUsers
- At
- ADR.md
- Minion Agent
- Repository Guidelines
- test-deb-lifecycle.sh
- GetLogins
- install.sh
- install_minion.sh
- .handleHeartbeat
- .handleIdempotencyInProgress
- opencode.json
- CPUInfo
- .RevokeClient
- graphify.js
- build_deb.sh
- Server
- Storage
- Storage
- graphify-nvidia.sh
- minion
- Demo rápida do Minion Agent
- CLAUDE.md
- SPEC.md
- API
- ADR-011 — Princípio do Privilégio Mínimo e Deny-by-Default
- Minion Agent Improvements Implementation Plan
- Coletores
- Rotação de API key
- Expiração opcional de clientes
- Política de Segurança
- Tecnologias
- Gerenciamento de Clientes
- Recuperação de upgrade do pacote Debian
- Idempotência no desbloqueio Fail2Ban
- Componentes
- Identidade e capacidades do agente
- Heartbeat do Minion
- Retenção de idempotência
- Rate limiting
- Modelo de Segurança
- Alternativas Consideradas
- Credencial bootstrap do pacote Debian
- Limites de segurança HTTP
- Diagnóstico de idempotências em andamento
- Requisitos Não Funcionais
- Alternativas Consideradas
- Alternativas Consideradas
- Alternativas Consideradas
- Alternativas Consideradas
- Consequências
- Consequências
- client-revocation.md
- Validação da instalação Debian

## God Nodes (most connected - your core abstractions)
1. `Model` - 41 edges
2. `Server` - 27 edges
3. `Minion Agent` - 23 edges
4. `New()` - 21 edges
5. `NewService()` - 18 edges
6. `Repository Guidelines` - 18 edges
7. `12. Endpoints` - 15 edges
8. `closeWithError()` - 14 edges
9. `Default()` - 14 edges
10. `Save()` - 14 edges

## Surprising Connections (you probably didn't know these)
- `handleBootstrapCommands()` --calls--> `NewService()`  [INFERRED]
  cmd/minion/commands.go → internal/admin/service.go
- `setup()` --calls--> `NewService()`  [INFERRED]
  cmd/minion/commands.go → internal/admin/service.go
- `handleClientCommands()` --calls--> `NewService()`  [INFERRED]
  cmd/minion/commands.go → internal/admin/service.go
- `run()` --calls--> `Load()`  [INFERRED]
  cmd/minion/runner.go → internal/config/config.go
- `main()` --calls--> `HashAPIKey()`  [INFERRED]
  cmd/check/main.go → internal/security/security.go

## Import Cycles
- None detected.

## Communities (78 total, 10 thin omitted)

### Community 0 - "closeWithError"
Cohesion: 0.06
Nodes (32): CommandRunner, ConfigUpdate, CreatedClient, errorCloser, execRunner, fakeCloser, SetupOptions, SetupResult (+24 more)

### Community 1 - "New"
Cohesion: 0.06
Nodes (37): DB, T, TestClientExpirationCanBeCleared(), TestClientExpirationDisablesExpiredClient(), TestClientSchemaMigratesExistingDatabase(), T, TestUpdateClientAllowedIPs(), TestUpdateClientAllowedIPsRejectsMissingClient() (+29 more)

### Community 2 - "Model"
Cohesion: 0.12
Nodes (15): Cmd, File, ensureInteractive(), filepathOrFallback(), formatBoolPath(), Client, Service, newInput() (+7 more)

### Community 3 - "NewService"
Cohesion: 0.09
Nodes (41): fakeRunner, Client, Config, RateLimitConfig, SecurityConfig, Service, T, newRevocationTestService() (+33 more)

### Community 4 - "Server"
Cohesion: 0.13
Nodes (16): DiskInfo, Service, SystemInfo, WazuhStatus, GetDiskUsage(), GetServices(), GetSystem(), GetWazuhStatus() (+8 more)

### Community 5 - ".idempotentAction"
Cohesion: 0.08
Nodes (30): Buffer, Header, copyHeaders(), HandlerFunc, ResponseWriter, Server, newBufferedResponseWriter(), payloadDigest() (+22 more)

### Community 6 - "Consume"
Cohesion: 0.13
Nodes (20): failingWriter, dispatchCommand(), handleBootstrapCommands(), handleClientCommands(), isRoot(), setup(), main(), run() (+12 more)

### Community 7 - "GetJournalLogs"
Cohesion: 0.14
Nodes (16): Fail2BanEvent, IPTablesRule, JournalEntry, SudoEvent, runCommandCombinedOutput(), runCommandOutput(), GetFail2BanEvents(), UnbanFail2BanIP() (+8 more)

### Community 8 - "Global Constraints"
Cohesion: 0.25
Nodes (7): Completion Criteria, Debian Self-Contained Installation Implementation Plan, Global Constraints, Task 1: Define Package Dependency and Bootstrap Contract, Task 2: Make `postinst` Idempotent After Dependency Resolution, Task 3: Align Package, Unit, and Documentation Artifacts, Task 4: Validate Real WSL Installation and Upgrade Lifecycle

### Community 9 - "New"
Cohesion: 0.09
Nodes (25): Handler, T, TestAuthFallsBackToConfigClientsWhenDatabaseIsEmpty(), TestAuthPrefersDatabaseClientsWhenAvailable(), Request, Server, Server, newHTTPServer() (+17 more)

### Community 10 - "rateLimiter"
Cohesion: 0.25
Nodes (10): Duration, Time, newRateLimiter(), T, TestRateLimiterBlocksAndRecovers(), TestRateLimiterCleansInactiveEntries(), TestRateLimiterIsolatesKeys(), Mutex (+2 more)

### Community 11 - "info.go"
Cohesion: 0.28
Nodes (11): Info, buildVersion(), capabilities(), deriveAgentID(), Get(), readMachineID(), readUptime(), T (+3 more)

### Community 12 - "parseMemory"
Cohesion: 0.36
Nodes (8): MemoryInfo, GetMemory(), Reader, parseMemory(), T, TestParseMemory(), TestParseMemoryRejectsAvailableGreaterThanTotal(), TestParseMemoryRejectsInvalidNumber()

### Community 13 - "parseUsers"
Cohesion: 0.33
Nodes (8): User, GetUsers(), Reader, isHumanUser(), parseUsers(), T, TestIsHumanUserRejectsInvalidUID(), TestParseUsersFiltersHumanAccounts()

### Community 14 - "At"
Cohesion: 0.36
Nodes (7): Status, At(), Get(), Time, T, TestAtNeverReturnsNegativeProcessUptime(), TestAtReturnsOnlineHeartbeat()

### Community 15 - "ADR.md"
Cohesion: 0.04
Nodes (47): ADR-001 — Arquitetura Base do Minion, ADR-002 — Linguagem de Desenvolvimento, ADR-003 — Distribuição em Binário Único, ADR-004 — Banco de Dados Local, ADR-005 — Modelo de Comunicação, ADR-006 — Modelo de Segurança, ADR-006A — Interface Operacional Guiada na CLI, ADR-007 — Geração de API Keys (+39 more)

### Community 16 - "Minion Agent"
Cohesion: 0.05
Nodes (38): 10. Gerenciamento de clientes, 11. Autenticação, 12. Endpoints, 13. Auditoria, 14. Build local, 15. Como gerar um `.deb` local, 16. Atualização manual, 17. Segurança operacional (+30 more)

### Community 17 - "Repository Guidelines"
Cohesion: 0.11
Nodes (18): Agent-Specific Instructions, Autenticação, Build, Test, and Development Commands, Coding Style & Naming Conventions, Commit & Pull Request Guidelines, Contrato da instalação Debian, Definição de pronto, Experiência esperada após a instalação (+10 more)

### Community 18 - "test-deb-lifecycle.sh"
Cohesion: 0.43
Nodes (4): assert_dependency(), assert_mode(), fail(), test-deb-lifecycle.sh script

### Community 19 - "GetLogins"
Cohesion: 0.83
Nodes (3): LoginEvent, findISOTimestampIndex(), GetLogins()

### Community 20 - "install.sh"
Cohesion: 0.67
Nodes (3): GO111MODULE, log(), install.sh script

### Community 21 - "install_minion.sh"
Cohesion: 0.67
Nodes (3): GO111MODULE, log(), install_minion.sh script

### Community 22 - ".handleHeartbeat"
Cohesion: 0.50
Nodes (3): Request, ResponseWriter, Server

### Community 23 - ".handleIdempotencyInProgress"
Cohesion: 0.50
Nodes (3): Request, ResponseWriter, Server

### Community 24 - "opencode.json"
Cohesion: 0.50
Nodes (3): plugin, $schema, .opencode/plugins/graphify.js

### Community 44 - "Demo rápida do Minion Agent"
Cohesion: 0.17
Nodes (11): Cenário demonstrado, Demo rápida do Minion Agent, Frase curta para LinkedIn, Objetivo da demo, Passo 1: build local, Passo 2: setup e criação de cliente, Passo 3: subir o serviço em ambiente controlado, Passo 4: health check (+3 more)

### Community 45 - "CLAUDE.md"
Cohesion: 0.18
Nodes (9): API Surface, Client Management, Common Commands, Configuration & Runtime Files, High-Level Architecture, Maintenance Notes, Mandatory Product Direction, Packaging / Installation (+1 more)

### Community 46 - "SPEC.md"
Cohesion: 0.18
Nodes (10): Arquitetura, Auditoria, Ações Administrativas Futuras, Estrutura de Diretórios, Filosofia do Produto, Fora do Escopo da V1, Minion - Especificação Funcional v1.0, Modelo de Permissões (+2 more)

### Community 47 - "API"
Cohesion: 0.20
Nodes (10): API, Eventos de Privilégio, Fail2Ban, Health, Journal, Logins, Serviços, Sistema (+2 more)

### Community 48 - "ADR-011 — Princípio do Privilégio Mínimo e Deny-by-Default"
Cohesion: 0.22
Nodes (8): ADR-011 — Princípio do Privilégio Mínimo e Deny-by-Default, Consequências, Contexto, Decisão, Justificativa, Modelo Permitido, Modelo Proibido, Regra de Implementação

### Community 49 - "Minion Agent Improvements Implementation Plan"
Cohesion: 0.22
Nodes (8): Minion Agent Improvements Implementation Plan, Self‑Review Checklist, Task 1: Garantir diretório do SQLite exista, Task 2: Trocar sal estático por sal aleatório nas chaves API, Task 3: Validar entrada no endpoint Fail2Ban unban, Task 4: Adicionar middleware de auditoria, Task 5: Linting & CI integration, Task 6: Systemd unit file

### Community 50 - "Coletores"
Cohesion: 0.25
Nodes (8): Coletores, Eventos de Privilégio, Fail2Ban, Host, Logins, Serviços, Usuários, Wazuh Agent

### Community 51 - "Rotação de API key"
Cohesion: 0.29
Nodes (6): Comando, Comportamento preservado, Procedimento seguro, Quando usar, Recuperação, Rotação de API key

### Community 52 - "Expiração opcional de clientes"
Cohesion: 0.29
Nodes (6): Compatibilidade, Consultar, Definir expiração, Expiração opcional de clientes, Operação com Automation/n8n, Remover expiração

### Community 53 - "Política de Segurança"
Cohesion: 0.29
Nodes (6): Como reportar uma vulnerabilidade, Divulgação responsável, Escopo de segurança, Política de Segurança, Processo de triagem, Versões suportadas

### Community 54 - "Tecnologias"
Cohesion: 0.29
Nodes (7): Banco Local, Configuração, Distribuição, Execução, Linguagem, Porta padrão, Tecnologias

### Community 55 - "Gerenciamento de Clientes"
Cohesion: 0.29
Nodes (7): Criar Cliente, Desabilitar Cliente, Gerenciamento de Clientes, Habilitar Cliente, Listar Clientes, Remover Cliente, UI guiada de operação

### Community 56 - "Recuperação de upgrade do pacote Debian"
Cohesion: 0.33
Nodes (5): Comportamento normal, Falha durante o upgrade, Garantias e limites, Recuperação de upgrade do pacote Debian, Snapshot mantido após falha

### Community 57 - "Idempotência no desbloqueio Fail2Ban"
Cohesion: 0.33
Nodes (5): Comportamento, Idempotência no desbloqueio Fail2Ban, Persistência e segurança, Requisição, Uso no n8n

### Community 58 - "Componentes"
Cohesion: 0.33
Nodes (6): API Server, Audit Logging, Collector Engine, Componentes, Event Engine, Storage Engine

### Community 59 - "Identidade e capacidades do agente"
Cohesion: 0.40
Nodes (4): Campos, Consultar o agente, Identidade e capacidades do agente, Uso pelo Automation

### Community 60 - "Heartbeat do Minion"
Cohesion: 0.40
Nodes (4): Heartbeat do Minion, Requisição, Resposta, Uso no Automation

### Community 61 - "Retenção de idempotência"
Cohesion: 0.40
Nodes (4): Configuração, Limitação conhecida, Logs, Retenção de idempotência

### Community 62 - "Rate limiting"
Cohesion: 0.40
Nodes (4): Configuração, Operação com Automation/n8n, Rate limiting, Segurança e auditoria

### Community 63 - "Modelo de Segurança"
Cohesion: 0.40
Nodes (5): API Key, Cliente, Conceito, Fluxo de Autenticação, Modelo de Segurança

### Community 64 - "Alternativas Consideradas"
Cohesion: 0.50
Nodes (4): Alternativas Consideradas, mTLS, OAuth2, OpenID Connect

### Community 65 - "Credencial bootstrap do pacote Debian"
Cohesion: 0.50
Nodes (3): Credencial bootstrap do pacote Debian, Recuperação, Upgrade e reinstalação

### Community 66 - "Limites de segurança HTTP"
Cohesion: 0.50
Nodes (3): Impacto para Automation/n8n, Limites atuais, Limites de segurança HTTP

### Community 67 - "Diagnóstico de idempotências em andamento"
Cohesion: 0.50
Nodes (3): Consulta, Diagnóstico de idempotências em andamento, Procedimento operacional

### Community 68 - "Requisitos Não Funcionais"
Cohesion: 0.50
Nodes (4): Consumo, Disponibilidade, Requisitos Não Funcionais, Segurança

### Community 69 - "Alternativas Consideradas"
Cohesion: 0.67
Nodes (3): Alternativas Consideradas, Node.js, Python

### Community 70 - "Alternativas Consideradas"
Cohesion: 0.67
Nodes (3): Alternativas Consideradas, Arquivos JSON, PostgreSQL

### Community 71 - "Alternativas Consideradas"
Cohesion: 0.67
Nodes (3): Alternativas Consideradas, gRPC, WebSocket

### Community 72 - "Alternativas Consideradas"
Cohesion: 0.67
Nodes (3): Alternativas Consideradas, Apenas comandos por flags, Painel web local

### Community 73 - "Consequências"
Cohesion: 0.67
Nodes (3): Consequências, Negativas, Positivas

### Community 74 - "Consequências"
Cohesion: 0.67
Nodes (3): Consequências, Negativas, Positivas

### Community 77 - "Validação da instalação Debian"
Cohesion: 0.29
Nodes (6): Comandos de validação, Fluxo oficial, Limite do comando `dpkg -i`, Reinstalação e upgrade, Resultado esperado, Validação da instalação Debian

## Knowledge Gaps
- **271 isolated node(s):** `$schema`, `.opencode/plugins/graphify.js`, `build_deb.sh script`, `minion`, `GO111MODULE` (+266 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **10 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `Server` connect `Server` to `New`, `rateLimiter`, `NewService`?**
  _High betweenness centrality (0.087) - this node is a cross-community bridge._
- **Why does `Config` connect `NewService` to `New`, `Server`?**
  _High betweenness centrality (0.073) - this node is a cross-community bridge._
- **Why does `New()` connect `New` to `rateLimiter`, `NewService`, `Server`, `.idempotentAction`?**
  _High betweenness centrality (0.035) - this node is a cross-community bridge._
- **Are the 17 inferred relationships involving `New()` (e.g. with `TestClientExpirationCanBeCleared()` and `TestClientExpirationDisablesExpiredClient()`) actually correct?**
  _`New()` has 17 INFERRED edges - model-reasoned connections that need verification._
- **Are the 16 inferred relationships involving `NewService()` (e.g. with `handleBootstrapCommands()` and `handleClientCommands()`) actually correct?**
  _`NewService()` has 16 INFERRED edges - model-reasoned connections that need verification._
- **What connects `$schema`, `.opencode/plugins/graphify.js`, `build_deb.sh script` to the rest of the system?**
  _271 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `closeWithError` be split into smaller, more focused modules?**
  _Cohesion score 0.06485671191553545 - nodes in this community are weakly interconnected._