# Graph Report - minion-agent  (2026-07-03)

## Corpus Check
- 42 files · ~20,066 words
- Verdict: corpus is large enough that graph structure adds value.

## Summary
- 336 nodes · 435 edges · 42 communities (32 shown, 10 thin omitted)
- Extraction: 91% EXTRACTED · 9% INFERRED · 0% AMBIGUOUS · INFERRED: 37 edges (avg confidence: 0.8)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `af859f51`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- [[_COMMUNITY_.auth|.auth]]
- [[_COMMUNITY_Server|Server]]
- [[_COMMUNITY_Storage|Storage]]
- [[_COMMUNITY_middleware.go|middleware.go]]
- [[_COMMUNITY_Config|Config]]
- [[_COMMUNITY_GetIPTablesRules|GetIPTablesRules]]
- [[_COMMUNITY_GetJournalLogs|GetJournalLogs]]
- [[_COMMUNITY_GetLogins|GetLogins]]
- [[_COMMUNITY_GetSudoEvents|GetSudoEvents]]
- [[_COMMUNITY_GetUsers|GetUsers]]
- [[_COMMUNITY_install.sh|install.sh]]
- [[_COMMUNITY_install_minion.sh|install_minion.sh]]
- [[_COMMUNITY_CPUInfo|CPUInfo]]
- [[_COMMUNITY_GetDiskUsage|GetDiskUsage]]
- [[_COMMUNITY_GetFail2BanEvents|GetFail2BanEvents]]
- [[_COMMUNITY_GetMemory|GetMemory]]
- [[_COMMUNITY_GetServices|GetServices]]
- [[_COMMUNITY_GetSystem|GetSystem]]
- [[_COMMUNITY_GetWazuhStatus|GetWazuhStatus]]
- [[_COMMUNITY_build_deb.sh|build_deb.sh]]
- [[_COMMUNITY_IsIPBlocked|IsIPBlocked]]
- [[_COMMUNITY_graphify-nvidia.sh|graphify-nvidia.sh]]
- [[_COMMUNITY_minion|minion]]
- [[_COMMUNITY_12. Endpoints|12. Endpoints]]
- [[_COMMUNITY_SPEC|SPEC.md]]
- [[_COMMUNITY_CLAUDE|CLAUDE.md]]
- [[_COMMUNITY_API|API]]
- [[_COMMUNITY_Repository Guidelines|Repository Guidelines]]
- [[_COMMUNITY_ADR-011 — Princípio do Privilégio Mínimo e Deny-by-Default|ADR-011 — Princípio do Privilégio Mínimo e Deny-by-Default]]
- [[_COMMUNITY_Minion Agent Improvements Implementation Plan|Minion Agent Improvements Implementation Plan]]
- [[_COMMUNITY_Coletores|Coletores]]
- [[_COMMUNITY_Tecnologias|Tecnologias]]
- [[_COMMUNITY_Componentes|Componentes]]
- [[_COMMUNITY_Gerenciamento de Clientes|Gerenciamento de Clientes]]
- [[_COMMUNITY_Modelo de Segurança|Modelo de Segurança]]
- [[_COMMUNITY_Alternativas Consideradas|Alternativas Consideradas]]
- [[_COMMUNITY_Requisitos Não Funcionais|Requisitos Não Funcionais]]
- [[_COMMUNITY_Alternativas Consideradas|Alternativas Consideradas]]
- [[_COMMUNITY_Alternativas Consideradas|Alternativas Consideradas]]
- [[_COMMUNITY_Alternativas Consideradas|Alternativas Consideradas]]
- [[_COMMUNITY_Consequências|Consequências]]

## God Nodes (most connected - your core abstractions)
1. `Server` - 23 edges
2. `Minion Agent` - 23 edges
3. `12. Endpoints` - 15 edges
4. `Storage` - 11 edges
5. `API` - 10 edges
6. `HashAPIKey()` - 8 edges
7. `Repository Guidelines` - 8 edges
8. `Coletores` - 8 edges
9. `ADR-011 — Princípio do Privilégio Mínimo e Deny-by-Default` - 8 edges
10. `Minion Agent Improvements Implementation Plan` - 8 edges

## Surprising Connections (you probably didn't know these)
- `main()` --calls--> `HashAPIKey()`  [INFERRED]
  cmd/check/main.go → internal/security/security.go
- `main()` --calls--> `VerifyAPIKey()`  [INFERRED]
  cmd/check/main.go → internal/security/security.go
- `main()` --calls--> `Load()`  [INFERRED]
  cmd/minion/main.go → internal/config/config.go
- `setup()` --calls--> `Load()`  [INFERRED]
  cmd/minion/main.go → internal/config/config.go
- `setup()` --calls--> `GenerateAPIKey()`  [INFERRED]
  cmd/minion/main.go → internal/security/security.go

## Import Cycles
- None detected.

## Communities (42 total, 10 thin omitted)

### Community 0 - ".auth"
Cohesion: 0.08
Nodes (27): Client, main(), handleClientCommands(), main(), setup(), storageFromConfig(), main(), Client (+19 more)

### Community 1 - "Server"
Cohesion: 0.36
Nodes (4): IsIPBlocked(), Request, ResponseWriter, Server

### Community 2 - "Storage"
Cohesion: 0.19
Nodes (7): DB, initSchema(), New(), T, TestStorageCreatesDBDir(), Client, Storage

### Community 3 - "middleware.go"
Cohesion: 0.18
Nodes (10): clientNameFromContext(), HandlerFunc, Request, ResponseWriter, Server, setClientNameOnWriter(), withClientName(), clientNameSetter (+2 more)

### Community 4 - "Config"
Cohesion: 0.05
Nodes (43): ADR-001 — Arquitetura Base do Minion, ADR-002 — Linguagem de Desenvolvimento, ADR-003 — Distribuição em Binário Único, ADR-004 — Banco de Dados Local, ADR-005 — Modelo de Comunicação, ADR-006 — Modelo de Segurança, ADR-007 — Geração de API Keys, ADR-008 — Modelo de Permissões (+35 more)

### Community 5 - "GetIPTablesRules"
Cohesion: 1.00
Nodes (3): IPTablesRule, GetIPTablesRules(), ParseIPTablesRules()

### Community 6 - "GetJournalLogs"
Cohesion: 1.00
Nodes (3): JournalEntry, GetJournalLogs(), ParseJournalLogs()

### Community 7 - "GetLogins"
Cohesion: 0.83
Nodes (3): LoginEvent, findISOTimestampIndex(), GetLogins()

### Community 8 - "GetSudoEvents"
Cohesion: 1.00
Nodes (3): SudoEvent, GetSudoEvents(), ParseSudoEvents()

### Community 9 - "GetUsers"
Cohesion: 0.83
Nodes (3): User, GetUsers(), isHumanUser()

### Community 10 - "install.sh"
Cohesion: 0.67
Nodes (3): GO111MODULE, log(), install.sh script

### Community 11 - "install_minion.sh"
Cohesion: 0.67
Nodes (3): GO111MODULE, log(), install_minion.sh script

### Community 20 - "IsIPBlocked"
Cohesion: 0.09
Nodes (22): 10. Gerenciamento de clientes, 11. Autenticação, 13. Auditoria, 14. Build local, 15. Como gerar um `.deb` local, 16. Atualização manual, 17. Segurança operacional, 18. Troubleshooting (+14 more)

### Community 23 - "12. Endpoints"
Cohesion: 0.13
Nodes (15): 12. Endpoints, Disco, Eventos de privilégio, Fail2Ban, Fail2Ban Unban, Health, IP Block, IPTables (+7 more)

### Community 24 - "SPEC.md"
Cohesion: 0.18
Nodes (10): Arquitetura, Auditoria, Ações Administrativas Futuras, Estrutura de Diretórios, Filosofia do Produto, Fora do Escopo da V1, Minion - Especificação Funcional v1.0, Modelo de Permissões (+2 more)

### Community 25 - "CLAUDE.md"
Cohesion: 0.20
Nodes (8): API Surface, Client Management, Common Commands, Configuration & Runtime Files, High-Level Architecture, Maintenance Notes, Packaging / Installation, Reference Documents

### Community 26 - "API"
Cohesion: 0.20
Nodes (10): API, Eventos de Privilégio, Fail2Ban, Health, Journal, Logins, Serviços, Sistema (+2 more)

### Community 27 - "Repository Guidelines"
Cohesion: 0.22
Nodes (8): Agent-Specific Instructions, Build, Test, and Development Commands, Coding Style & Naming Conventions, Commit & Pull Request Guidelines, Project Structure & Module Organization, Regra obrigatória: Graphify antes de codar, Repository Guidelines, Testing Guidelines

### Community 28 - "ADR-011 — Princípio do Privilégio Mínimo e Deny-by-Default"
Cohesion: 0.22
Nodes (8): ADR-011 — Princípio do Privilégio Mínimo e Deny-by-Default, Consequências, Contexto, Decisão, Justificativa, Modelo Permitido, Modelo Proibido, Regra de Implementação

### Community 29 - "Minion Agent Improvements Implementation Plan"
Cohesion: 0.22
Nodes (8): Minion Agent Improvements Implementation Plan, Self‑Review Checklist, Task 1: Garantir diretório do SQLite exista, Task 2: Trocar sal estático por sal aleatório nas chaves API, Task 3: Validar entrada no endpoint Fail2Ban unban, Task 4: Adicionar middleware de auditoria, Task 5: Linting & CI integration, Task 6: Systemd unit file

### Community 30 - "Coletores"
Cohesion: 0.25
Nodes (8): Coletores, Eventos de Privilégio, Fail2Ban, Host, Logins, Serviços, Usuários, Wazuh Agent

### Community 31 - "Tecnologias"
Cohesion: 0.29
Nodes (7): Banco Local, Configuração, Distribuição, Execução, Linguagem, Porta padrão, Tecnologias

### Community 32 - "Componentes"
Cohesion: 0.33
Nodes (6): API Server, Audit Logging, Collector Engine, Componentes, Event Engine, Storage Engine

### Community 33 - "Gerenciamento de Clientes"
Cohesion: 0.33
Nodes (6): Criar Cliente, Desabilitar Cliente, Gerenciamento de Clientes, Habilitar Cliente, Listar Clientes, Remover Cliente

### Community 34 - "Modelo de Segurança"
Cohesion: 0.40
Nodes (5): API Key, Cliente, Conceito, Fluxo de Autenticação, Modelo de Segurança

### Community 35 - "Alternativas Consideradas"
Cohesion: 0.50
Nodes (4): Alternativas Consideradas, mTLS, OAuth2, OpenID Connect

### Community 36 - "Requisitos Não Funcionais"
Cohesion: 0.50
Nodes (4): Consumo, Disponibilidade, Requisitos Não Funcionais, Segurança

### Community 37 - "Alternativas Consideradas"
Cohesion: 0.67
Nodes (3): Alternativas Consideradas, Node.js, Python

### Community 38 - "Alternativas Consideradas"
Cohesion: 0.67
Nodes (3): Alternativas Consideradas, Arquivos JSON, PostgreSQL

### Community 39 - "Alternativas Consideradas"
Cohesion: 0.67
Nodes (3): Alternativas Consideradas, gRPC, WebSocket

### Community 40 - "Consequências"
Cohesion: 0.67
Nodes (3): Consequências, Negativas, Positivas

## Knowledge Gaps
- **174 isolated node(s):** `build_deb.sh script`, `minion`, `GO111MODULE`, `GO111MODULE`, `contextKey` (+169 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **10 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `Server` connect `Server` to `.auth`, `Storage`?**
  _High betweenness centrality (0.075) - this node is a cross-community bridge._
- **Why does `Storage` connect `Storage` to `.auth`, `Server`?**
  _High betweenness centrality (0.032) - this node is a cross-community bridge._
- **Why does `New()` connect `.auth` to `Server`, `Storage`?**
  _High betweenness centrality (0.017) - this node is a cross-community bridge._
- **What connects `build_deb.sh script`, `minion`, `GO111MODULE` to the rest of the system?**
  _174 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `.auth` be split into smaller, more focused modules?**
  _Cohesion score 0.07692307692307693 - nodes in this community are weakly interconnected._
- **Should `Config` be split into smaller, more focused modules?**
  _Cohesion score 0.045454545454545456 - nodes in this community are weakly interconnected._
- **Should `IsIPBlocked` be split into smaller, more focused modules?**
  _Cohesion score 0.08695652173913043 - nodes in this community are weakly interconnected._