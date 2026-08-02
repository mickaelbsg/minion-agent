# Diagnóstico de idempotências em andamento

O Minion preserva registros `in_progress` quando uma ação administrativa foi iniciada, mas seu resultado não pôde ser persistido com segurança. Esses registros não são liberados automaticamente porque repetir a ação pode causar impacto operacional.

## Consulta

```http
GET /api/v1/idempotency/in-progress
Authorization: Bearer <api-key>
```

A rota é autenticada, passa pelo rate limit e pela auditoria existentes e é estritamente somente leitura.

Parâmetros opcionais:

- `limit`: entre 1 e 100; padrão 50;
- `action`: atualmente aceita somente `fail2ban_unban`.

Exemplo:

```http
GET /api/v1/idempotency/in-progress?action=fail2ban_unban&limit=20
```

Resposta:

```json
{
  "count": 1,
  "records": [
    {
      "client_name": "automation",
      "action": "fail2ban_unban",
      "request_id": "n8n-unban-20260802-0001",
      "created_at": "2026-08-02T12:00:00Z",
      "updated_at": "2026-08-02T12:00:00Z"
    }
  ]
}
```

Os registros mais antigos aparecem primeiro. A resposta não contém API key, hash do payload, corpo enviado, corpo de resposta ou cabeçalhos de autorização.

## Procedimento operacional

Ao encontrar um registro antigo:

1. correlacione `request_id`, cliente e horário com o workflow do Automation/n8n;
2. confirme externamente se a ação produziu efeito no servidor;
3. não repita a requisição com outro request ID sem avaliar o risco;
4. trate qualquer liberação manual como uma operação administrativa separada e auditada.

Esta API não conclui, remove, libera ou repete registros. Recuperação automática de ações com resultado incerto permanece proibida.
