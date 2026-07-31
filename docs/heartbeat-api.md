# Heartbeat do Minion

O endpoint autenticado `GET /api/v1/heartbeat` foi criado para polling periódico pelo Automation/n8n.

## Requisição

```http
GET /api/v1/heartbeat
Authorization: Bearer <API_KEY>
```

## Resposta

```json
{
  "status": "online",
  "agent_id": "minion_7df85b88b1cf4bf4a6db63558f6f41cd",
  "hostname": "srv-linux-001",
  "version": "v1.2.0",
  "system_uptime_seconds": 86400.42,
  "process_uptime_seconds": 900.12,
  "process_started_at": "2026-07-31T22:00:00Z",
  "observed_at": "2026-07-31T22:15:00Z"
}
```

## Uso no Automation

Consulte o endpoint em intervalos regulares e registre `observed_at` como a última comunicação confirmada. Um intervalo inicial entre 1 e 5 minutos é adequado para inventário operacional; alertas de indisponibilidade devem considerar mais de uma falha consecutiva para evitar falso positivo de rede.

`system_uptime_seconds` representa o tempo do servidor Linux. `process_uptime_seconds` representa somente o processo Minion e permite detectar reinicializações do agente sem confundir com reinicialização do host.

O endpoint exige a mesma autenticação por API key, cliente ativo e IP/CIDR permitido usada nas demais rotas protegidas.