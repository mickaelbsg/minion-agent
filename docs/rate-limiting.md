# Rate limiting

O Minion aplica dois limites em memória aos endpoints autenticados:

1. Um bucket por IP é consultado antes da validação da API key. Isso reduz tentativas de força bruta e protege o custo da verificação Argon2id.
2. Após autenticação, um bucket independente é aplicado ao nome do cliente. Clientes distintos não compartilham esse limite.

Quando um limite é excedido, a API retorna `429 Too Many Requests` e o cabeçalho `Retry-After` com o número mínimo de segundos antes de uma nova tentativa.

## Configuração

Os valores abaixo são os padrões usados quando os campos não existem em configurações antigas:

```json
{
  "security": {
    "rate_limit": {
      "ip_burst": 30,
      "ip_refill_per_second": 5,
      "client_burst": 60,
      "client_refill_per_second": 10
    }
  }
}
```

`burst` define quantas requisições podem ocorrer de uma vez. `refill_per_second` define a reposição contínua de tokens.

O endpoint público `/api/v1/health` não passa pelo rate limit de autenticação. Os demais endpoints autenticados consomem o bucket do IP e, após autenticação, o bucket do cliente.

## Operação com Automation/n8n

Os padrões permitem polling frequente, mas rajadas maiores devem usar retry com backoff e respeitar `Retry-After`. Não aumente os limites para compensar loops incorretos no workflow.

Em ambientes com vários clientes atrás do mesmo NAT, todos compartilham o bucket de IP antes da autenticação. O limite por cliente continua isolado após a credencial ser validada.

## Segurança e auditoria

Rejeições registram somente o escopo controlado (`ip` ou `client`), o IP já usado pela auditoria, a rota, o método e o status. API keys, cabeçalhos de autorização e payloads não são persistidos.

O estado fica apenas na memória do processo. Reiniciar o serviço limpa os buckets. Isso é adequado ao modelo atual de um agente por servidor e não exige Redis ou outro serviço externo.
