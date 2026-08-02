# Retenção de idempotência

O Minion mantém no SQLite os resultados concluídos de ações idempotentes para que retries do Automation/n8n não executem novamente a mesma operação.

Por padrão, registros concluídos são mantidos por 168 horas (7 dias). A limpeza ocorre na inicialização do serviço e remove somente registros com estado `completed` cujo `updated_at` seja anterior à janela configurada.

Registros `in_progress` nunca são removidos automaticamente. Eles podem representar uma ação executada cujo resultado não chegou a ser persistido; repeti-la automaticamente criaria risco operacional.

## Configuração

```json
{
  "security": {
    "idempotency_retention_hours": 168
  }
}
```

Configurações antigas que não possuem o campo recebem o valor padrão automaticamente em memória. O arquivo existente não precisa ser reescrito durante upgrade.

Use uma janela maior que o período máximo em que o Automation/n8n pode repetir uma execução. Valores muito baixos permitem que um request ID antigo seja tratado como novo depois que o registro concluído expirar.

## Logs

Quando há remoções, o Minion registra somente a quantidade de registros excluídos. Payloads, respostas, API keys e cabeçalhos de autorização não são incluídos.

Uma falha na limpeza é registrada e não impede o serviço de iniciar. A idempotência continua funcionando; apenas os registros antigos permanecem no banco até a próxima inicialização bem-sucedida.

## Limitação conhecida

Esta retenção controla crescimento dos registros concluídos. Recuperação de registros `in_progress` exige uma decisão operacional explícita e não faz parte desta etapa.
