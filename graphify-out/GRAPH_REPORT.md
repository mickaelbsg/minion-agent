# Graph Report - .  (2026-07-03)

## Corpus Check
- cluster-only mode — file stats not available

## Summary
- 133 nodes · 233 edges · 23 communities (12 shown, 11 thin omitted)
- Extraction: 86% EXTRACTED · 14% INFERRED · 0% AMBIGUOUS · INFERRED: 32 edges (avg confidence: 0.8)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `e4a9651f`
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

## God Nodes (most connected - your core abstractions)
1. `Server` - 22 edges
2. `Storage` - 10 edges
3. `HashAPIKey()` - 6 edges
4. `setup()` - 5 edges
5. `handleClientCommands()` - 5 edges
6. `Config` - 5 edges
7. `Load()` - 5 edges
8. `VerifyAPIKey()` - 5 edges
9. `New()` - 5 edges
10. `main()` - 4 edges

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

## Communities (23 total, 11 thin omitted)

### Community 0 - ".auth"
Cohesion: 0.17
Nodes (14): main(), handleClientCommands(), main(), setup(), main(), Load(), GenerateAPIKey(), HashAPIKey() (+6 more)

### Community 1 - "Server"
Cohesion: 0.43
Nodes (3): Request, ResponseWriter, Server

### Community 2 - "Storage"
Cohesion: 0.19
Nodes (7): DB, initSchema(), New(), T, TestStorageCreatesDBDir(), Client, Storage

### Community 3 - "middleware.go"
Cohesion: 0.18
Nodes (10): clientNameFromContext(), HandlerFunc, Request, ResponseWriter, Server, setClientNameOnWriter(), withClientName(), clientNameSetter (+2 more)

### Community 4 - "Config"
Cohesion: 0.22
Nodes (6): Client, Config, T, TestAuditMiddlewareCreatesEntry(), fileExists(), New()

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

## Knowledge Gaps
- **8 isolated node(s):** `build_deb.sh script`, `minion`, `GO111MODULE`, `GO111MODULE`, `contextKey` (+3 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **11 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `Server` connect `Server` to `.auth`, `Storage`, `Config`?**
  _High betweenness centrality (0.411) - this node is a cross-community bridge._
- **Why does `Storage` connect `Storage` to `Server`, `Config`?**
  _High betweenness centrality (0.161) - this node is a cross-community bridge._
- **Are the 5 inferred relationships involving `HashAPIKey()` (e.g. with `main()` and `handleClientCommands()`) actually correct?**
  _`HashAPIKey()` has 5 INFERRED edges - model-reasoned connections that need verification._
- **What connects `build_deb.sh script`, `minion`, `GO111MODULE` to the rest of the system?**
  _8 weakly-connected nodes found - possible documentation gaps or missing edges._