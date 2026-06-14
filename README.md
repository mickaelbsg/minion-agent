# Minion Agent

Minion é um agente Linux leve escrito em Go.

Ele roda como serviço local, coleta dados operacionais do sistema e expõe uma API HTTP autenticada para sistemas externos, como o Severino.

O Minion não possui IA embarcada. Ele não interpreta, não decide e não executa ações administrativas na V1. Seu papel é coletar, organizar e disponibilizar dados locais.

## Objetivo

Reduzir acessos SSH recorrentes usados para coleta de dados em servidores Linux.

Em vez de vários sistemas conectarem via SSH para consultar logs, usuários, serviços e eventos de segurança, o Minion fica instalado localmente e responde através de uma API controlada por IP e API Key.

## Status

MVP inicial.

Implementado nesta primeira base:

- API HTTP REST.
- Autenticação por API Key.
- Restrição por IP/CIDR.
- Geração de API Key via CLI.
- Configuração em JSON.
- Coletores básicos de sistema, usuários, serviços, Fail2Ban, Wazuh e logins.
- Estrutura inicial para auditoria em SQLite.

## Estrutura

```
cmd/minion/               Entrada principal do binário
internal/config/          Leitura da configuração
internal/security/        API Key, hash e validação de IP
internal/server/          Servidor HTTP e middleware de autenticação
internal/collectors/      Coletores Linux
internal/storage/         Persistência local e auditoria
systemd/                  Unit file do serviço
config.example.json       Exemplo de configuração
```

## Endpoints

```
GET /api/v1/health
GET /api/v1/system
GET /api/v1/users
GET /api/v1/services
GET /api/v1/fail2ban
GET /api/v1/fail2ban/unban
GET /api/v1/ipblock?ip=<IP>
GET /api/v1/wazuh
GET /api/v1/logins
```

`/api/v1/health` não exige autenticação. Os demais endpoints exigem API Key e IP autorizado.

## Gerar chave de cliente

```bash
minion --create-client --name api_severino --ips 192.168.56.2/32
```
## Comandos de configuração simplificados

### `minion setup`

Executa todas as etapas de preparação de uma instalação nova:
- Cria o diretório `/etc/minion/tls` (se ainda não existir).
- Gera um certificado TLS auto‑assinado (caso ainda não exista).
- Copia o `config.example.json` para `/etc/minion/config.json` se o arquivo ainda não existir.
- Habilita e inicia o serviço `minion.service` via `systemctl`.

### `minion add client`

Atalho para criar um cliente API (equivalente a `--create-client`).
Exemplo:
```bash
sudo minion add client --name myclient --ips 0.0.0.0/0,::/0
```

Esses novos sub‑comandos foram implementados em `cmd/minion/main.go` e já estão disponíveis nos releases atuais.


O comando imprime:

```
Client: api_severino
Allowed IPs: 192.168.56.2/32
API Key: minion_sk_...
API Key Hash: ...
```

A API Key deve ser configurada no cliente consumidor. O hash deve ser salvo no `config.json` do Minion.

## Exemplo de configuração

```json
{
  "api": {
    "bind": "0.0.0.0:9871"
  },
  "db_path": "/var/lib/minion/minion.db",
  "clients": [
    {
      "name": "api_severino",
      "allowed_ips": ["192.168.56.2/32","127.0.0.1/32","::1/128"],
      "api_key_hash": "REPLACE_WITH_HASH",
      "enabled": true
    }
  ]
}
```

## Rodar

```bash
go build -o minion ./cmd/minion
./minion --config ./config.example.json
```

## Teste de API

```bash
curl http://localhost:9871/api/v1/health
```

Com autenticação:

```bash
curl \
  -H "Authorization: Bearer <API_KEY>" \
  http://localhost:9871/api/v1/system
```

## Observações

Esta base é um MVP. A próxima etapa recomendada é fortalecer a auditoria, persistir clientes no SQLite e trocar o hash simples por Argon2id.

---

## Auditoria

- Cada requisição HTTP autenticada (e a saúde do endpoint) é registrada na tabela `audit` do SQLite (`/opt/minion/minion.db`). Os campos são: cliente, IP, método, caminho e código de status.
- A tabela pode ser consultada diretamente via SQL, por exemplo:

```sql
SELECT timestamp, client_name, ip, method, path, status FROM audit ORDER BY timestamp DESC LIMIT 20;
```

## Systemd Unit

- O repositório inclui o arquivo de unidade `systemd/minion.service`.
- Instala‑lo copiando para `/etc/systemd/system/` e recarregue o daemon:

```bash
sudo cp systemd/minion.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable --now minion.service
```

## 📦 Ritmo detalhado do Minion Agent
*(Como funciona, por que existe e quais passos seguir para evoluir o projeto)*

### 1️⃣ Visão geral
| ✅ | **O que é** | Um agente leve escrito em Go que roda como *systemd service* em servidores Linux e expõe, via API HTTPS (mTLS), dados de observabilidade e controle remoto. |
|---|------------|----------------------------------------------------------------------------------------------|
| 🎯 | **Objetivo** | Eliminar a necessidade de SSH manual para auditoria, coleta de métricas e operações de segurança. Tudo pode ser feito com chamadas HTTP autenticadas. |
| 🔐 | **Segurança** | API protegida por **Bearer‑token** (hash Argon2id + IP whitelist). Comunicação criptografada com TLS auto‑signed (ou certs reais). O service roda como **root** via systemd para ter permissão de `fail2ban-client`. |
| 🛠️ | **Escopo MVP** | • Coleta de: system info, usuários, serviços, logs, Fail2Ban, Wazuh. <br>• Endpoints: `/health`, `/system`, `/users`, `/services`, `/fail2ban`, `/fail2ban/unban`, `/ipblock`. <br>• Persistência em SQLite. <br>• Deploy como .deb (single‑binary) + systemd unit. |

### 2️⃣ Arquitetura (código + infra)
```
+-------------------+      +-------------------+      +-------------------+
|   collectors/    | ---> |   storage/       | ---> |   HTTP server     |
| (fail2ban, ...)  |      | (SQLite DB)      |      | (mux, auth, json) |
+-------------------+      +-------------------+      +-------------------+
        ^                       ^                         ^
        |                       |                         |
        +--- internal/-----------+-------------------------+
                (core structs, utils)
```
* **Collectors** – funções puras que rodam `exec.Command` (iptables, fail2ban, `cat /etc/hosts.deny`, etc.) e retornam structs.
* **Storage** – camada mínima que grava eventos em `minion.db` usando `database/sql`.
* **Server** – *net/http* + mux; middleware `auth` verifica:
  1. Header `Authorization: Bearer ***`
  2. IP do cliente na whitelist (config.json)
  3. Hash da API‑key (ideal Argon2id).
* **Systemd unit** – `ExecStart=/usr/local/bin/minion --config /etc/minion/config.json` 
  *Rodando como root* → `fail2ban-client` funciona sem *sudo*.

### 3️⃣ Fluxo de uso típico
1. **Instalação** – `apt install ./minion_1.0.0.deb` (ou script `install_minion.sh`).
2. **Configuração** – editar `/etc/minion/config.json` → definir **port**, **db_path**, **clients** (nome, IPs permitidos, API‑key hash).
3. **Start** – `systemctl enable --now minion.service`; logs via `journalctl -u minion`.
4. **Consumo** – exemplos cURL:
   * Health: `curl -k https://localhost:9871/api/v1/health`
   * Listar IP bloqueado: `curl -k -H "Authorization: Bearer ***" https://localhost:9871/api/v1/ipblock?ip=203.0.113.12`
   * Desbloquear Fail2Ban: `POST /api/v1/fail2ban/unban` com JSON `{ "ip":"1.2.3.4", "jail":"ssh" }`
5. **Integração** – scripts, CI/CD ou UI centralizada podem chamar a API sem abrir shells.

### 4️⃣ Roadmap (etapas de evolução)
| Fase | Sprint (≈2 sem) | Principais entregas | Motivo |
|------|----------------|----------------------|--------|
| **0 – Preparação** | 1 | • Script `install_minion.sh` → .deb <br>• Docs *README* + *API spec* (OpenAPI) | Facilita adoção interna |
| **1 – MVP Completo** | 2‑3 | • Endpoints atuais testados em CI <br>• Argon2id para API‑key <br>• Persistir clientes no SQLite via API | Satisfaz “não usar SSH” e traz segurança real |
| **2 – Observabilidade avançada** | 4‑5 | • Exporter Prometheus (`/metrics`) <br>• Dashboard Grafana pré‑configurado <br>• Webhook opcional (Telegram/Slack) | Visualização em tempo real |
| **3 – Integração CI/CD** | 6‑7 | • GitHub Actions que compila, testa e publica .deb <br>• Deploy automático via Ansible/Puppet | Automatiza distribuição |
| **4 – Extensibilidade** | 8‑9 | • Plugin system (carregar collectors via Go plugins ou scripts Bash) <br>• Exemplo: collector de *auditd* | Permite squads criar coletores sem recompilar |
| **5 – Hardenização** | 10‑11 | • SELinux/AppArmor profiles <br>• Rotação automática de API‑keys <br>• Testes de penetração (OWASP) | Cumpre compliance de segurança |
| **6 – Multi‑tenant** | 12‑13 | • Suporte a *profiles* por cliente (DB isolado, whitelist) <br>• UI mínima (React + OpenAPI) | Caso o Minion seja SaaS interno |
| **7 – Release final** | 14 | • Tag v1.0.0 <br>• Publishes de binaries (amd64, arm64) <br>• Docs no GitHub Pages | Produto estável para produção |

> Cada sprint inclui **testes unitários**, **lint** (`golangci‑lint`) e **pipeline de CI** (`go test ./... && go vet && staticcheck`).

### 5️⃣ Como o Minion resolve o “problema do SSH”
| Problema | Como o Minion atua |
|----------|-------------------|
| Login manual em cada host | API HTTPS já escuta; basta chamar a URL a partir de um script/orquestrador. |
| Gerenciamento de credenciais espalhadas | Apenas um **Bearer‑token** por host (hash no config); rotação via `/rotate-key`. |
| Auditoria de quem fez o que | Todos os requests são logados no *systemd journal* (JSON) → ingestão em SIEM. |
| Bloqueio/Desbloqueio de IPs | `GET /ipblock`, `POST /fail2ban/unban` – sem abrir *shell*. |
| Visibilidade em tempo real | Exporter Prometheus → alertas automáticos quando banimentos aumentam. |

### 6️⃣ Guia rápido de “primeiros passos”
```bash
# 1️⃣ Baixe o .deb (ou use o script)
wget https://github.com/mickaelbsg/minion-agent/releases/download/v1.0.0/minion_1.0.0.deb
sudo dpkg -i minion_1.0.0.deb

# 2️⃣ Ajuste a configuração (exemplo mínimo)
sudo cp /etc/minion/config.example.json /etc/minion/config.json
sudo nano /etc/minion/config.json   # → coloque sua API‑key

# 3️⃣ (Opcional) Troque o certificado auto‑signed por um real
sudo cp ~/certs/minion.crt /etc/minion/tls/
sudo cp ~/certs/minion.key /etc/minion/tls/

# 4️⃣ Inicie o serviço
sudo systemctl enable --now minion.service
sudo journalctl -u minion -f   # acompanha logs

# 5️⃣ Teste o health endpoint
curl -k https://localhost:9871/api/v1/health
# → {"status":"ok"}

# 6️⃣ Consulte um IP bloqueado
curl -k -H "Authorization: Bearer ***" https://localhost:9871/api/v1/ipblock?ip=203.0.113.12
# → {"blocked":true}
```

### 7️⃣ Checklist de “pronto para produção”
- [ ] **Binário estático** (`CGO_ENABLED=0`).
- [ ] **Package .deb** publicado em repositório interno apt.
- [ ] **TLS** (auto‑signed ou CA) configurada e verificada.
- [ ] **API‑key** usando Argon2id (upgrade futuro).
- [ ] **Unit tests** ≥ 80 % coverage + CI pipeline.
- [ ] **Logs no journal** em JSON → ingestão SIEM.
- [ ] **Fail2Ban** unban funcional (testado).
- [ ] **Documentação** completa + exemplos de clientes (Bash, Python).
- [ ] **Rollback plan** (keep old .deb, `systemctl stop minion && dpkg -r minion`).

---

*Este README foi atualizado em 2026‑06‑13 para incluir o ritmo detalhado do projeto e as próximas etapas de desenvolvimento.*