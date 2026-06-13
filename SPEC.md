# Minion - Especificação Funcional v1.0

## Visão Geral

Minion é um agente Linux desenvolvido em Go, distribuído como binário único e executado como serviço systemd.

Seu objetivo é coletar informações operacionais e de segurança localmente, disponibilizando esses dados através de uma API HTTP autenticada para consumo por sistemas externos, como o Severino.

O Minion não possui inteligência artificial embarcada.

O Minion não realiza análises.

O Minion não toma decisões.

O Minion atua exclusivamente como camada de coleta, armazenamento local e exposição segura de informações.

---

# Objetivos

* Eliminar acessos SSH recorrentes para coleta de dados.
* Reduzir ruído em logs de autenticação.
* Centralizar coleta local de informações.
* Disponibilizar API padronizada.
* Permitir integração com Severino e outros sistemas.
* Operar com baixo consumo de recursos.
* Funcionar mesmo sem conectividade externa.

---

# Tecnologias

## Linguagem

Go

## Distribuição

Binário único

```text
minion
```

## Execução

systemd

```bash
systemctl enable minion
systemctl start minion
```

## Banco Local

SQLite

Arquivo:

```text
/opt/minion/minion.db
```

## Configuração

```text
/etc/minion/config.json
```

## Porta padrão

```text
9870/tcp
```

---

# Arquitetura

```text
+----------------------+
|      Severino        |
+----------+-----------+
           |
           | HTTP API
           |
+----------+-----------+
|        Minion        |
+----------+-----------+
           |
           |
+----------+-----------+
| Linux Operating Sys. |
+----------------------+
```

---

# Componentes

## Collector Engine

Responsável pela coleta local.

Coletores suportados:

* Usuários
* Logins
* Eventos sudo
* Fail2Ban
* Wazuh Agent
* Serviços Systemd
* Logs do Sistema
* Informações do Host

---

## API Server

Responsável por:

* autenticação
* autorização
* exposição de endpoints

---

## Event Engine

Detecta alterações relevantes localmente.

Exemplos:

* novo usuário
* login
* banimento fail2ban
* evento sudo
* alteração de serviço

---

## Storage Engine

Persistência local.

SQLite.

## Audit Logging

All HTTP requests are stored in the `audit` SQLite table with fields client, ip, method, path, status.

---

# Modelo de Segurança

## Conceito

O Minion controla quem pode acessá-lo.

Cada cliente autorizado possui:

* Nome
* IP permitido
* API Key
* Status

---

## Cliente

Exemplo:

```yaml
clients:
  - name: api_severino

    allowed_ips:
      - 192.168.56.2/32

    api_key_hash: "$argon2id$..."

    enabled: true
```

---

## Fluxo de Autenticação

```text
Request
   ↓
IP autorizado?
   ↓
API Key válida?
   ↓
Cliente ativo?
   ↓
Acesso liberado
```

---

## API Key

Gerada pelo próprio Minion.

Exemplo:

```text
minion_sk_ZA8f91LxK...
```

A chave é exibida apenas uma vez.

O Minion armazena somente o hash.

Algoritmo:

```text
Argon2id
```

---

# Gerenciamento de Clientes

## Criar Cliente

```bash
minion client create api_severino \
  --ip 192.168.56.2/32
```

Retorno:

```text
Cliente: api_severino

API Key:
minion_sk_xxxxxxxxxxxxxx
```

---

## Listar Clientes

```bash
minion client list
```

---

## Desabilitar Cliente

```bash
minion client disable api_severino
```

---

## Habilitar Cliente

```bash
minion client enable api_severino
```

---

## Remover Cliente

```bash
minion client delete api_severino
```

---

# Modelo de Permissões

Versão 1:

```text
ALL OR NOTHING
```

Cliente autenticado possui acesso total à API.

Não existe RBAC.

Não existem permissões por rota.

---

# Coletores

## Host

Coleta:

* hostname
* fqdn
* sistema operacional
* kernel
* uptime
* cpu
* memória
* disco

---

## Usuários

Coleta:

```text
/etc/passwd
```

Retorna:

* usuário
* uid
* shell
* home

---

## Logins

Coleta:

* SSH
* Login local

Fontes:

```text
journalctl
auth.log
```

---

## Elevação de Privilégio

Monitoramento:

```text
sudo
su
pkexec
```

---

## Fail2Ban

Coleta:

* jails
* bans
* unbans

---

## Wazuh Agent

Coleta:

* status
* versão
* conectividade
* eventos

---

## Serviços

Monitoramento:

```text
systemctl
```

Retorna:

* nome
* estado
* uptime

---

# API

## Health

```http
GET /api/v1/health
```

---

## Sistema

```http
GET /api/v1/system
```

---

## Usuários

```http
GET /api/v1/users
```

---

## Logins

```http
GET /api/v1/logins
```

---

## Eventos Sudo

```http
GET /api/v1/sudo
```

---

## Fail2Ban

```http
GET /api/v1/fail2ban
```

---

## Wazuh

```http
GET /api/v1/wazuh
```

---

## Serviços

```http
GET /api/v1/services
```

---

## Journal

```http
GET /api/v1/journal
```

Parâmetros:

```http
?limit=100
?level=error
```

---

# Auditoria

Toda requisição deve ser registrada.

Formato:

```text
TIMESTAMP
CLIENT
IP
METHOD
ENDPOINT
STATUS
```

Exemplo:

```text
2026-06-11T15:42:10Z
CLIENT=api_severino
IP=192.168.56.2
METHOD=GET
PATH=/api/v1/users
STATUS=200
```

---

# Estrutura de Diretórios

```text
/etc/minion/
├── config.json

/opt/minion/
├── minion.db

/var/log/minion/
├── minion.log

/usr/local/bin/
├── minion
```

---

# Requisitos Não Funcionais

## Consumo

Memória alvo:

```text
< 100 MB
```

CPU:

```text
< 2%
```

em operação normal.

---

## Disponibilidade

O Minion deve continuar operando mesmo que:

* Severino esteja offline
* rede esteja indisponível
* Wazuh esteja desconectado

---

## Segurança

* API Key obrigatória
* Allow List por IP
* Hash Argon2id
* Sem credenciais em texto puro
* Sem necessidade de SSH para coleta

---

# Fora do Escopo da V1

* RBAC
* OAuth2
* OpenID Connect
* mTLS
* Interface Web
* Cluster
* Multi-tenancy
* Execução remota de comandos
* Inteligência Artificial
* Integração direta com LLM
* Atualização automática

---

# Filosofia do Produto

Minion não interpreta.

Minion não decide.

Minion não analisa.

Minion coleta, organiza e disponibiliza dados locais de forma segura e eficiente para consumo por sistemas externos.
