# Identidade e capacidades do agente

O endpoint abaixo permite que o Automation/n8n identifique o Minion instalado em cada servidor e descubra as capacidades disponíveis antes de executar um workflow.

## Consultar o agente

```http
GET /api/v1/agent
Authorization: Bearer <API_KEY>
```

A rota usa a mesma autenticação por API key, cliente ativo e IP/CIDR permitido aplicada aos demais endpoints protegidos.

Exemplo de resposta:

```json
{
  "agent_id": "minion_7df85b88b1cf4bf4a6db63558f6f41cd",
  "hostname": "srv-linux-001",
  "version": "v1.2.0",
  "os": "linux",
  "architecture": "amd64",
  "uptime_seconds": 86400.42,
  "observed_at": "2026-07-31T21:45:00Z",
  "capabilities": [
    "agent.read",
    "disk.read",
    "fail2ban.read",
    "fail2ban.unban",
    "firewall.iptables.read",
    "ipblock.read",
    "journal.read",
    "logins.read",
    "memory.read",
    "privilege-events.read",
    "services.read",
    "system.read",
    "users.read",
    "wazuh.read"
  ]
}
```

## Campos

- `agent_id`: identificador estável derivado por SHA-256 do machine-id local. O valor original de `/etc/machine-id` não é exposto.
- `hostname`: nome atual da máquina.
- `version`: versão registrada no binário Go; builds locais sem versão retornam `devel`.
- `os` e `architecture`: plataforma usada para compilar e executar o agente.
- `uptime_seconds`: tempo de atividade do sistema lido de `/proc/uptime`.
- `observed_at`: horário UTC em que a resposta foi montada.
- `capabilities`: catálogo ordenado de leituras e ações explícitas implementadas pelo Minion.

## Uso pelo Automation

O Automation deve consultar esta rota durante cadastro, inventário ou heartbeat. Ele pode usar `agent_id` como identidade técnica e verificar `capabilities` antes de chamar outro endpoint.

O catálogo não autoriza comandos livres. Cada capacidade corresponde a um endpoint explícito implementado, validado e auditado pelo Minion.
