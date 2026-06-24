# Minion - Especificação Funcional v1.0

## Visão Geral

Minion é um agente Linux desenvolvido em Go, distribuído como binário único e executado como serviço systemd.

Seu objetivo é coletar informações operacionais e de segurança localmente, disponibilizando esses dados através de uma API HTTP autenticada para consumo por sistemas externos, como o Severino.

O Minion existe para reduzir a necessidade de SSH recorrente e evitar que automações externas precisem realizar elevação de privilégios diretamente nos servidores.

O Minion não possui inteligência artificial embarcada.

O Minion não realiza análises.

O Minion não toma decisões.

O Minion atua como camada local de coleta, armazenamento, exposição segura de informações e, em versões futuras, execução de capacidades administrativas explícitas via API.

---

# Objetivos

* Eliminar acessos SSH recorrentes para coleta de dados.
* Reduzir ruído em logs de autenticação.
* Centralizar coleta local de informações.
* Disponibilizar API padronizada.
* Permitir integração com Severino e outros sistemas.
* Operar com baixo consumo de recursos.
* Funcionar mesmo sem conectividade externa.
* Evitar credenciais privilegiadas distribuídas entre automações externas.
* Impedir que bots, pipelines, LLMs ou integrações externas executem comandos shell arbitrários no host.

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
           | HTTP API autenticada
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

Sistemas externos não acessam o host diretamente para operações privilegiadas. Eles consomem a API do Minion.

---

# Componentes

## Collector Engine

Responsável pela coleta local.

Coletores suportados:

* Usuários
* Logins
* Eventos de privilégio
* Fail2Ban
* Wazuh Agent
* Serviços Systemd
* Logs do Sistema
* Informações do Host
* Memória
* Disco
* Regras de firewall

---

## API Server

Responsável por:

* autenticação
* autorização
* exposição de endpoints
* separação entre consultas e capacidades administrativas futuras

---

## Event Engine

Detecta alterações relevantes localmente.

Exemplos:

* novo usuário
* login
* banimento fail2ban
* evento de privilégio
* alteração de serviço

---

## Storage Engine

Persistência local.

SQLite.

## Audit Logging

Todas as requisições HTTP são armazenadas na tabela `audit` do SQLite com campos de cliente, IP, método, path e status.

---

# Modelo de Segurança

## Conceito

O Minion controla quem pode acessá-lo.

Cada cliente autorizado possui:

* Nome
* IP permitido
* API Key
* Status

O Minion deve rodar com as permissões necessárias via systemd. O runtime do Minion não deve chamar `sudo` internamente.

É permitido que um administrador humano use `sudo` para instalar, configurar, iniciar, parar ou consultar logs do serviço. Essa permissão administrativa externa não deve ser confundida com o comportamento interno do Minion.

---

## Cliente

Exemplo:

```json
{
  "name": "api_severino",
  "allowed_ips": ["192.168.56.2/32"],
  "api_key_hash": "$argon2id$...",
  "enabled": true
}
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
minion client create --name api_severino --ips 192.168.56.2/32
```

Retorno:

```text
Client: api_severino
API Key: minion_sk_xxxxxxxxxxxxxx
API Key Hash: xxxxxxxxxxxxxx
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

Cliente autenticado possui acesso total à API disponível na V1.

Não existe RBAC.

Não existem permissões por rota.

RBAC e permissões por capacidade devem ser reavaliados antes da inclusão de ações administrativas amplas.

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

## Eventos de Privilégio

Monitoramento de eventos relacionados a elevação de privilégio e comandos administrativos executados por usuários humanos ou processos locais.

Fontes e eventos possíveis:

```text
sudo
su
pkexec
```

A presença da palavra `sudo` aqui se refere ao evento auditado no sistema operacional. O Minion não deve chamar `sudo` para executar seus próprios coletores.

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

## Eventos de Privilégio

Endpoint atual:

```http
GET /api/v1/sudo
```

Nome recomendado para evolução:

```http
GET /api/v1/privilege-events
```

O endpoint coleta eventos relacionados a elevação de privilégio. Ele não executa comandos administrativos.

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

# Ações Administrativas Futuras

A V1 é predominantemente observacional.

Futuramente, o Minion poderá executar ações administrativas, como:

* criar usuários
* excluir usuários
* bloquear IPs
* desbloquear IPs
* reiniciar serviços permitidos
* outras operações recorrentes e bem definidas

Essas ações não serão implementadas como shell remoto.

Não deve existir endpoint genérico de execução de comandos.

Proibido:

```http
POST /api/v1/execute
```

Proibido receber payload como:

```json
{
  "command": "useradd joao"
}
```

Modelo correto:

```http
POST /api/v1/users
```

```json
{
  "username": "joao",
  "shell": "/bin/bash"
}
```

Neste modelo, o cliente solicita uma capacidade. O Minion valida os parâmetros e executa internamente apenas a operação permitida por aquela capacidade.

Cada capacidade administrativa futura deve possuir:

* endpoint próprio
* validação própria
* whitelist interna de operação
* auditoria
* documentação
* comportamento previsível

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
* Sem necessidade de SSH recorrente para coleta
* Sem chamada interna a `sudo` no runtime
* Sem shell remoto
* Sem endpoint genérico para comandos arbitrários
* Ações futuras somente por capacidades explícitas e auditáveis da API

---

# Fora do Escopo da V1

* RBAC
* OAuth2
* OpenID Connect
* mTLS
* Interface Web
* Cluster
* Multi-tenancy
* Shell remoto
* Endpoint genérico de execução de comandos
* Ações administrativas amplas
* Inteligência Artificial
* Integração direta com LLM
* Atualização automática

---

# Filosofia do Produto

Minion não interpreta.

Minion não decide.

Minion não analisa.

Minion não é shell remoto.

Minion coleta, organiza e disponibiliza dados locais de forma segura e eficiente para consumo por sistemas externos.

Em versões futuras, Minion poderá executar capacidades administrativas específicas, mas apenas por API própria, whitelist interna e auditoria.