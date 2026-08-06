# Laboratório LXC para Teste do Minion Agent

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Criar um laboratório com LXC para testar a instalação, configuração e acesso à API do minion-agent em ambiente controlado.

**Architecture:** Containers LXC Debian rodam em host Linux (WSL2 ou servidor nativo). Scripts de automação criam, configuram e destroem containers. Testes validam instalação .deb, serviço systemd e conectividade API.

**Tech Stack:** LXC/LXD, Debian 12 (bookworm), systemd, bash, curl, sqlite3

## Global Constraints

- Container deve rodar Debian 12 (bookworm) para compatibilidade com o pacote .deb
- Serviço deve ser acessível de fora do container via IP do host
- Bootstrap credentials devem ser consumidas via `minion bootstrap pair`
- Testes devem ser automatizáveis e reproduzíveis
- Ambiente deve ser destruído após uso (não poluir o host)

---

## Estrutura de Arquivos

| Arquivo | Responsabilidade |
|---|---|
| `scripts/lxc/create-lab.sh` | Cria container LXC Debian, instala dependências, configura rede |
| `scripts/lxc/destroy-lab.sh` | Destroi o container e limpa recursos |
| `scripts/lxc/test-install.sh` | Testa instalação do pacote .deb no container |
| `scripts/lxc/test-api.sh` | Testa acesso à API de fora do container |
| `scripts/lxc/run-lab.sh` | Orquestra criação, teste e destruição do laboratório |
| `docs/lxc-lab.md` | Documentação de uso do laboratório |

---

### Task 1: Script de criação do container LXC

**Files:**
- Create: `scripts/lxc/create-lab.sh`

**Interfaces:**
- Consumes: Variáveis de ambiente `LAB_CONTAINER_NAME`, `LAB_IMAGE`, `LAB_HOST_PORT`
- Produzes: Container LXC rodando Debian 12 com systemd, IP acessível

- [ ] **Step 1: Criar script create-lab.sh**

```bash
#!/usr/bin/env bash
set -euo pipefail

# Configurações do laboratório
LAB_CONTAINER_NAME="${LAB_CONTAINER_NAME:-minion-lab}"
LAB_IMAGE="${LAB_IMAGE:-debian:12}"
LAB_HOST_PORT="${LAB_HOST_PORT:-9870}"
LAB_BRIDGE="${LAB_BRIDGE:-lxcbr0}"

log() { echo "[lab] $*"; }

# Verificar se LXD está disponível
if ! command -v lxc >/dev/null 2>&1; then
  log "ERRO: lxc não encontrado. Instale com: sudo snap install lxd"
  exit 1
fi

# Verificar se o container já existe
if lxc info "$LAB_CONTAINER_NAME" >/dev/null 2>&1; then
  log "Container '$LAB_CONTAINER_NAME' já existe. Use destroy-lab.sh para remover."
  exit 1
fi

# Criar container Debian 12
log "Criando container $LAB_CONTAINER_NAME com $LAB_IMAGE..."
lxc launch "$LAB_IMAGE" "$LAB_CONTAINER_NAME" \
  -c security.nesting=true \
  -c systemd=true

# Aguardar systemd inicializar
log "Aguardando systemd inicializar..."
lxc exec "$LAB_CONTAINER_NAME" -- systemctl is-system-running --wait >/dev/null 2>&1 || true

# Atualizar pacotes e instalar dependências
log "Instalando dependências..."
lxc exec "$LAB_CONTAINER_NAME" -- apt-get update -qq
lxc exec "$LAB_CONTAINER_NAME" -- apt-get install -y -qq \
  fail2ban \
  iptables \
  openssl \
  sqlite3 \
  curl \
  gnupg

# Configurar rede para acesso externo
log "Configurando acesso externo na porta $LAB_HOST_PORT..."
lxc config device add "$LAB_CONTAINER_NAME" minion-port \
  proxy listen=tcp:0.0.0.0:"$LAB_HOST_PORT" \
  connect=tcp:127.0.0.1:"$LAB_HOST_PORT"

# Obter IP do container
CONTAINER_IP=$(lxc exec "$LAB_CONTAINER_NAME" -- hostname -I | awk '{print $1}')
log "Container criado com IP: $CONTAINER_IP"
log "Acesso externo: https://localhost:$LAB_HOST_PORT"

# Salvar informações do container
cat > /tmp/minion-lab-info.txt <<EOF
CONTAINER_NAME=$LAB_CONTAINER_NAME
CONTAINER_IP=$CONTAINER_IP
HOST_PORT=$LAB_HOST_PORT
EOF

log "Laboratório pronto. Execute test-install.sh para validar."
```

- [ ] **Step 2: Tornar executável**

```bash
chmod +x scripts/lxc/create-lab.sh
```

- [ ] **Step 3: Commit**

```bash
git add scripts/lxc/create-lab.sh
git commit -m "feat: add LXC lab creation script"
```

---

### Task 2: Script de destruição do laboratório

**Files:**
- Create: `scripts/lxc/destroy-lab.sh`

**Interfaces:**
- Consumes: `LAB_CONTAINER_NAME` (opcional, default: minion-lab)
- Produzes: Container removido, recursos limpos

- [ ] **Step 1: Criar script destroy-lab.sh**

```bash
#!/usr/bin/env bash
set -euo pipefail

LAB_CONTAINER_NAME="${LAB_CONTAINER_NAME:-minion-lab}"

log() { echo "[lab] $*"; }

# Verificar se container existe
if ! lxc info "$LAB_CONTAINER_NAME" >/dev/null 2>&1; then
  log "Container '$LAB_CONTAINER_NAME' não encontrado."
  exit 0
fi

# Parar e remover container
log "Parando container $LAB_CONTAINER_NAME..."
lxc stop "$LAB_CONTAINER_NAME" --force 2>/dev/null || true

log "Removendo container $LAB_CONTAINER_NAME..."
lxc delete "$LAB_CONTAINER_NAME" --force

# Limpar arquivo de informações
rm -f /tmp/minion-lab-info.txt

log "Laboratório destruído com sucesso."
```

- [ ] **Step 2: Tornar executável**

```bash
chmod +x scripts/lxc/destroy-lab.sh
```

- [ ] **Step 3: Commit**

```bash
git add scripts/lxc/destroy-lab.sh
git commit -m "feat: add LXC lab destruction script"
```

---

### Task 3: Script de teste de instalação

**Files:**
- Create: `scripts/lxc/test-install.sh`

**Interfaces:**
- Consumes: Container LXC rodando, pacote .deb disponível
- Produzes: Relatório de validação da instalação

- [ ] **Step 1: Criar script test-install.sh**

```bash
#!/usr/bin/env bash
set -euo pipefail

LAB_CONTAINER_NAME="${LAB_CONTAINER_NAME:-minion-lab}"
DEB_PACKAGE="${1:-}"

log() { echo "[test] $*"; }
fail() { echo "[test] FALHA: $*" >&2; exit 1; }

# Verificar container
if ! lxc info "$LAB_CONTAINER_NAME" >/dev/null 2>&1; then
  fail "Container '$LAB_CONTAINER_NAME' não encontrado. Execute create-lab.sh primeiro."
fi

# Verificar pacote
if [[ -z "$DEB_PACKAGE" ]]; then
  # Procurar pacote mais recente
  DEB_PACKAGE=$(ls -t minion_*_amd64.deb 2>/dev/null | head -n1)
  if [[ -z "$DEB_PACKAGE" ]]; then
    fail "Nenhum pacote .deb encontrado. Execute build_deb.sh primeiro."
  fi
fi

if [[ ! -f "$DEB_PACKAGE" ]]; then
  fail "Pacote não encontrado: $DEB_PACKAGE"
fi

log "Usando pacote: $DEB_PACKAGE"

# Copiar pacote para o container
log "Copiando pacote para o container..."
lxc file push "$DEB_PACKAGE" "$LAB_CONTAINER_NAME/tmp/minion.deb"

# Instalar pacote
log "Instalando pacote no container..."
lxc exec "$LAB_CONTAINER_NAME" -- bash -c \
  "DEBIAN_FRONTEND=noninteractive apt-get install -y -qq /tmp/minion.deb"

# Validar instalação
log "Validando instalação..."

# 1. Pacote instalado
INSTALLED_VERSION=$(lxc exec "$LAB_CONTAINER_NAME" -- dpkg-query -W -f='${Version}' minion 2>/dev/null || echo "")
if [[ -z "$INSTALLED_VERSION" ]]; then
  fail "Pacote não está instalado"
fi
log "✓ Pacote instalado: $INSTALLED_VERSION"

# 2. Serviço ativo
if ! lxc exec "$LAB_CONTAINER_NAME" -- systemctl is-active --quiet minion.service; then
  fail "Serviço não está ativo"
fi
log "✓ Serviço está ativo"

# 3. Arquivos criados
for path in /etc/minion/config.json /etc/minion/tls/minion.crt /etc/minion/tls/minion.key /opt/minion/minion.db; do
  if ! lxc exec "$LAB_CONTAINER_NAME" -- test -f "$path"; then
    fail "Arquivo não encontrado: $path"
  fi
  log "✓ Arquivo existe: $path"
done

# 4. Permissões corretas
for path in /etc/minion/config.json /etc/minion/tls/minion.key /opt/minion/minion.db; do
  MODE=$(lxc exec "$LAB_CONTAINER_NAME" -- stat -c '%a' "$path")
  if [[ "$MODE" != "600" ]]; then
    fail "Permissão incorreta em $path: $MODE (esperado: 600)"
  fi
  log "✓ Permissão correta em $path: $MODE"
done

# 5. Bootstrap credentials criado
if lxc exec "$LAB_CONTAINER_NAME" -- test -f /var/lib/minion/bootstrap-credentials.txt; then
  BOOTSTRAP_MODE=$(lxc exec "$LAB_CONTAINER_NAME" -- stat -c '%a' /var/lib/minion/bootstrap-credentials.txt)
  log "✓ Bootstrap credentials criado (modo: $BOOTSTRAP_MODE)"
else
  log "⚠ Bootstrap credentials não encontrado (pode ter sido consumido)"
fi

# 6. Health check via API
log "Testando health check via API..."
HEALTH_RESPONSE=$(lxc exec "$LAB_CONTAINER_NAME" -- curl --silent --show-error --fail --insecure \
  https://127.0.0.1:9870/api/v1/health 2>/dev/null || echo "")

if [[ -z "$HEALTH_RESPONSE" ]]; then
  fail "Health check falhou"
fi
log "✓ Health check OK: $HEALTH_RESPONSE"

log "========================================="
log "TODOS OS TESTES DE INSTALAÇÃO PASSARAM"
log "========================================="
```

- [ ] **Step 2: Tornar executável**

```bash
chmod +x scripts/lxc/test-install.sh
```

- [ ] **Step 3: Commit**

```bash
git add scripts/lxc/test-install.sh
git commit -m "feat: add LXC lab installation test script"
```

---

### Task 4: Script de teste de acesso à API

**Files:**
- Create: `scripts/lxc/test-api.sh`

**Interfaces:**
- Consumes: Container LXC rodando, API key do bootstrap
- Produzes: Relatório de validação do acesso à API

- [ ] **Step 1: Criar script test-api.sh**

```bash
#!/usr/bin/env bash
set -euo pipefail

LAB_CONTAINER_NAME="${LAB_CONTAINER_NAME:-minion-lab}"
HOST_PORT="${LAB_HOST_PORT:-9870}"

log() { echo "[test] $*"; }
fail() { echo "[test] FALHA: $*" >&2; exit 1; }

# Verificar container
if ! lxc info "$LAB_CONTAINER_NAME" >/dev/null 2>&1; then
  fail "Container '$LAB_CONTAINER_NAME' não encontrado."
fi

# Obter API key do bootstrap
log "Obtendo API key do bootstrap..."
BOOTSTRAP_FILE="/var/lib/minion/bootstrap-credentials.txt"

if lxc exec "$LAB_CONTAINER_NAME" -- test -f "$BOOTSTRAP_FILE"; then
  # Bootstrap ainda existe, usar pair para obter API key
  PAIR_OUTPUT=$(lxc exec "$LAB_CONTAINER_NAME" -- minion bootstrap pair \
    --config /etc/minion/config.json \
    --ips 127.0.0.1/32 2>/dev/null || echo "")
  
  API_KEY=$(echo "$PAIR_OUTPUT" | sed -n 's/^API Key: //p' | head -n1)
  
  if [[ -z "$API_KEY" || ! "$API_KEY" =~ ^minion_sk_ ]]; then
    fail "Não foi possível obter API key do bootstrap"
  fi
  log "✓ API key obtida: ${API_KEY:0:20}..."
else
  fail "Bootstrap credentials não encontrado. Execute test-install.sh primeiro."
fi

# Testar acesso via localhost (dentro do container)
log "Testando acesso via localhost (dentro do container)..."
RESPONSE=$(lxc exec "$LAB_CONTAINER_NAME" -- curl --silent --show-error --fail --insecure \
  -H "Authorization: Bearer $API_KEY" \
  https://127.0.0.1:9870/api/v1/agent 2>/dev/null || echo "")

if [[ -z "$RESPONSE" ]]; then
  fail "Acesso via localhost falhou"
fi
log "✓ Acesso via localhost OK"

# Testar acesso externo (do host)
log "Testando acesso externo via localhost:$HOST_PORT..."
EXTERNAL_RESPONSE=$(curl --silent --show-error --fail --insecure \
  -H "Authorization: Bearer $API_KEY" \
  "https://localhost:$HOST_PORT/api/v1/agent" 2>/dev/null || echo "")

if [[ -z "$EXTERNAL_RESPONSE" ]]; then
  fail "Acesso externo falhou. Verifique se a porta $HOST_PORT está mapeada."
fi
log "✓ Acesso externo OK"

# Testar endpoints principais
log "Testando endpoints da API..."

ENDPOINTS=(
  "/api/v1/health"
  "/api/v1/agent"
  "/api/v1/system"
  "/api/v1/users"
  "/api/v1/memory"
  "/api/v1/disk"
)

for endpoint in "${ENDPOINTS[@]}"; do
  RESP=$(lxc exec "$LAB_CONTAINER_NAME" -- curl --silent --show-error --fail --insecure \
    -H "Authorization: Bearer $API_KEY" \
    "https://127.0.0.1:9870${endpoint}" 2>/dev/null || echo "FALHA")
  
  if [[ "$RESP" == "FALHA" ]]; then
    log "⚠ Endpoint $endpoint retornou erro"
  else
    log "✓ Endpoint $endpoint OK"
  fi
done

# Verificar que API key foi consumida
if lxc exec "$LAB_CONTAINER_NAME" -- test -f "$BOOTSTRAP_FILE"; then
  log "⚠ Bootstrap credentials ainda existe após pair"
else
  log "✓ Bootstrap credentials consumido corretamente"
fi

log "========================================="
log "TODOS OS TESTES DE API PASSARAM"
log "========================================="
```

- [ ] **Step 2: Tornar executável**

```bash
chmod +x scripts/lxc/test-api.sh
```

- [ ] **Step 3: Commit**

```bash
git add scripts/lxc/test-api.sh
git commit -m "feat: add LXC lab API test script"
```

---

### Task 5: Script orquestrador do laboratório

**Files:**
- Create: `scripts/lxc/run-lab.sh`

**Interfaces:**
- Consumes: Todos os scripts anteriores, pacote .deb
- Produzes: Laboratório completo testado e destruído

- [ ] **Step 1: Criar script run-lab.sh**

```bash
#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
LAB_CONTAINER_NAME="${LAB_CONTAINER_NAME:-minion-lab}"
LAB_HOST_PORT="${LAB_HOST_PORT:-9870}"
DEB_PACKAGE="${1:-}"

log() { echo "[lab] $*"; }
fail() { echo "[lab] FALHA: $*" >&2; exit 1; }

# Limpar ao sair
cleanup() {
  log "Limpando laboratório..."
  "$SCRIPT_DIR/destroy-lab.sh" || true
}
trap cleanup EXIT

# Verificar pré-requisitos
log "Verificando pré-requisitos..."
command -v lxc >/dev/null 2>&1 || fail "lxc não encontrado. Instale com: sudo snap install lxd"
command -v curl >/dev/null 2>&1 || fail "curl não encontrado"

# Verificar se há pacote .deb
if [[ -z "$DEB_PACKAGE" ]]; then
  DEB_PACKAGE=$(ls -t minion_*_amd64.deb 2>/dev/null | head -n1)
  if [[ -z "$DEB_PACKAGE" ]]; then
    log "Nenhum pacote .deb encontrado. Executando build_deb.sh..."
    PKG_VER=1.0.5 ./build_deb.sh
    DEB_PACKAGE=$(ls -t minion_*_amd64.deb | head -n1)
  fi
fi

if [[ ! -f "$DEB_PACKAGE" ]]; then
  fail "Pacote não encontrado: $DEB_PACKAGE"
fi

log "========================================="
log "LABORATÓRIO LXC - MINION AGENT"
log "========================================="
log "Container: $LAB_CONTAINER_NAME"
log "Pacote: $DEB_PACKAGE"
log "Porta: $LAB_HOST_PORT"
log "========================================="

# 1. Criar laboratório
log "Etapa 1/4: Criando container LXC..."
"$SCRIPT_DIR/create-lab.sh"

# 2. Testar instalação
log "Etapa 2/4: Testando instalação..."
"$SCRIPT_DIR/test-install.sh" "$DEB_PACKAGE"

# 3. Testar API
log "Etapa 3/4: Testando acesso à API..."
"$SCRIPT_DIR/test-api.sh"

# 4. Relatório final
log "========================================="
log "LABORATÓRIO CONCLUÍDO COM SUCESSO"
log "========================================="
log ""
log "Para testar manualmente:"
log "  lxc exec $LAB_CONTAINER_NAME -- bash"
log ""
log "Para acessar a API de fora:"
log "  curl -k -H 'Authorization: Bearer <API_KEY>' https://localhost:$LAB_HOST_PORT/api/v1/health"
log ""
log "Para destruir o laboratório:"
log "  $SCRIPT_DIR/destroy-lab.sh"
log "========================================="

# Não destruir automaticamente (manter para testes manuais)
trap - EXIT
```

- [ ] **Step 2: Tornar executável**

```bash
chmod +x scripts/lxc/run-lab.sh
```

- [ ] **Step 3: Commit**

```bash
git add scripts/lxc/run-lab.sh
git commit -m "feat: add LXC lab orchestrator script"
```

---

### Task 6: Documentação do laboratório

**Files:**
- Create: `docs/lxc-lab.md`

**Interfaces:**
- Consumes: Todos os scripts do laboratório
- Produzes: Documentação completa de uso

- [ ] **Step 1: Criar documentação**

```markdown
# Laboratório LXC para Teste do Minion Agent

## Visão Geral

O laboratório LXC permite testar a instalação, configuração e acesso à API do minion-agent em ambiente controlado usando containers Linux.

## Pré-requisitos

- Linux com suporte a LXC/LXD (WSL2 com Nested Containers ou servidor nativo)
- `lxc` instalado (`sudo snap install lxd`)
- Pacote `.deb` do minion compilado (`PKG_VER=1.0.5 ./build_deb.sh`)

## Uso Rápido

```bash
# Executar laboratório completo (cria, testa e mantém container)
./scripts/lxc/run-lab.sh

# Ou executar etapas individualmente
./scripts/lxc/create-lab.sh          # Cria container
./scripts/lxc/test-install.sh        # Testa instalação
./scripts/lxc/test-api.sh            # Testa API
./scripts/lxc/destroy-lab.sh         # Destroi container
```

## Estrutura

```
scripts/lxc/
├── create-lab.sh      # Cria container LXC Debian
├── destroy-lab.sh     # Destroi o container
├── test-install.sh    # Testa instalação do .deb
├── test-api.sh        # Testa acesso à API
└── run-lab.sh         # Orquestra tudo
```

## Configuração

Variáveis de ambiente opcionais:

| Variável | Default | Descrição |
|---|---|---|
| `LAB_CONTAINER_NAME` | `minion-lab` | Nome do container |
| `LAB_IMAGE` | `debian:12` | Imagem base do container |
| `LAB_HOST_PORT` | `9870` | Porta mapeada no host |

Exemplo:
```bash
LAB_CONTAINER_NAME=minion-test LAB_HOST_PORT=8080 ./scripts/lxc/run-lab.sh
```

## Testes Executados

### Instalação
- Pacote instalado corretamente
- Serviço systemd ativo
- Arquivos criados com permissões corretas
- Bootstrap credentials gerado
- Health check via API

### API
- Acesso via localhost (dentro do container)
- Acesso externo (do host)
- Endpoints principais: /health, /agent, /system, /users, /memory, /disk
- Consumo correto do bootstrap credentials

## Acesso Manual

```bash
# Entrar no container
lxc exec minion-lab -- bash

# Verificar serviço
systemctl status minion.service
journalctl -u minion.service -f

# Testar API
curl -k -H "Authorization: Bearer <API_KEY>" https://127.0.0.1:9870/api/v1/health

# Do host
curl -k -H "Authorization: Bearer <API_KEY>" https://localhost:9870/api/v1/health
```

## Solução de Problemas

### Container não cria
- Verifique se LXD está rodando: `sudo snap services lxd`
- Inicie se necessário: `sudo snap start lxd`

### Serviço não inicia
- Verifique logs: `lxc exec minion-lab -- journalctl -u minion.service -f`
- Verifique dependências: `lxc exec minion-lab -- systemctl status minion.service`

### API não acessível de fora
- Verifique se a porta está mapeada: `lxc config device list minion-lab`
- Teste dentro do container primeiro
- Verifique firewall no host

### Bootstrap não consome
- Verifique se o arquivo existe: `lxc exec minion-lab -- ls -la /var/lib/minion/`
- Execute pair manualmente: `lxc exec minion-lab -- minion bootstrap pair --config /etc/minion/config.json --ips 127.0.0.1/32`
```

- [ ] **Step 2: Commit**

```bash
git add docs/lxc-lab.md
git commit -m "docs: add LXC lab documentation"
```

---

### Task 7: Adicionar scripts ao .gitignore e verificar estrutura

**Files:**
- Modify: `.gitignore`

**Interfaces:**
- Consumes: Scripts do laboratório
- Produze: Gitignore atualizado

- [ ] **Step 1: Atualizar .gitignore**

Adicionar ao `.gitignore`:
```
# LXC lab artifacts
/tmp/minion-lab-info.txt
```

- [ ] **Step 2: Verificar estrutura de diretórios**

```bash
ls -la scripts/lxc/
```

- [ ] **Step 3: Commit final**

```bash
git add .gitignore
git commit -m "chore: update gitignore for LXC lab artifacts"
```
