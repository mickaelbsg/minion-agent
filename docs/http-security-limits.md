# Limites de segurança HTTP

O Minion aplica limites fixos no servidor HTTP para reduzir consumo indevido de conexões, memória e tempo de processamento.

## Limites atuais

- cabeçalhos HTTP: até 5 segundos para leitura;
- requisição completa: até 15 segundos para leitura;
- resposta: até 30 segundos para escrita;
- conexão ociosa persistente: até 60 segundos;
- corpo de requisição: até 64 KiB.

Requisições cujo corpo ultrapasse 64 KiB recebem `413 Request Entity Too Large`. O limite também é aplicado quando o cliente usa transferência segmentada e não informa `Content-Length`.

Os endpoints de coleta atuais usam `GET` sem corpo. O limite protege principalmente ações administrativas explícitas, como o desbloqueio controlado do Fail2Ban, sem permitir comandos livres ou shell remoto.

## Impacto para Automation/n8n

As chamadas normais do Automation/n8n utilizam payloads pequenos e não exigem alteração. Um retorno `413` indica payload incorreto ou excessivo e não deve ser contornado aumentando o corpo enviado.

Os timeouts são limites do servidor, não tempos de repetição do Automation. O plano central deve usar timeout próprio menor ou igual ao limite de resposta e tratar falhas de rede de forma controlada.
