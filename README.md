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

```text
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

```text
GET /api/v1/health
GET /api/v1/system
GET /api/v1/users
GET /api/v1/services
GET /api/v1/fail2ban
GET /api/v1/wazuh
GET /api/v1/logins
```

`/api/v1/health` não exige autenticação. Os demais endpoints exigem API Key e IP autorizado.

## Gerar chave de cliente

```bash
minion --create-client --name api_severino --ips 192.168.56.2/32
```

O comando imprime:

```text
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
    "bind": "0.0.0.0:9870"
  },
  "db_path": "/opt/minion/minion.db",
  "clients": [
    {
      "name": "api_severino",
      "allowed_ips": ["192.168.56.2/32"],
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
curl http://localhost:9870/api/v1/health
```

Com autenticação:

```bash
curl \
  -H "Authorization: Bearer minion_sk_EXEMPLO" \
  http://localhost:9870/api/v1/system
```

## Observações

Esta base é um MVP. A próxima etapa recomendada é fortalecer a auditoria, persistir clientes no SQLite e trocar o hash simples por Argon2id.