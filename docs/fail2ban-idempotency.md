# Idempotência no desbloqueio Fail2Ban

O endpoint `POST /api/v1/fail2ban/unban` exige o cabeçalho `X-Request-ID` para evitar que retries do Automation/n8n executem a mesma ação mais de uma vez.

## Requisição

```http
POST /api/v1/fail2ban/unban
Authorization: Bearer <api-key>
Content-Type: application/json
X-Request-ID: n8n-unban-20260802-0001

{"ip":"192.0.2.10","jail":"sshd"}
```

O identificador deve ter de 8 a 128 caracteres e pode conter letras, números, ponto, sublinhado, dois-pontos e hífen. Gere um identificador novo para cada ação lógica e reutilize exatamente o mesmo valor apenas ao repetir a mesma requisição.

## Comportamento

- A primeira requisição reserva atomicamente o identificador para o cliente autenticado e executa a ação.
- Uma repetição concluída retorna o mesmo status e a resposta persistida, sem executar o comando novamente.
- O mesmo identificador com payload diferente retorna `409 Conflict`.
- Uma repetição enquanto a primeira execução ainda está em andamento retorna `409 Conflict`.
- O agente devolve `X-Request-ID` na resposta.

A chave é isolada por cliente autenticado e ação. Dois clientes diferentes podem utilizar o mesmo identificador sem colisão.

## Persistência e segurança

O SQLite armazena somente:

- nome do cliente;
- ação;
- request ID validado;
- SHA-256 do corpo recebido;
- estado;
- status HTTP;
- resposta necessária ao replay;
- datas de criação e atualização.

A API key, o cabeçalho `Authorization` e o corpo bruto da requisição não são armazenados. Registros interrompidos permanecem como `in_progress` e falham de forma fechada; o Minion não repete automaticamente uma ação cujo resultado final não pôde ser confirmado.

## Uso no n8n

Crie o request ID antes do nó HTTP e preserve o mesmo valor durante as tentativas de retry. Não gere um novo ID a cada tentativa, pois isso representa uma nova ação para o agente.
