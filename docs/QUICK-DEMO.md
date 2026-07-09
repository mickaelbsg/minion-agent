# Demo rápida do Minion Agent

Este roteiro serve para demonstrar o valor do Minion em poucos minutos, sem vender o projeto como shell remoto ou executor genérico.

## Objetivo da demo

Mostrar um agente local para servidores Linux que expõe dados operacionais por API autenticada, com allowlist de IP, auditoria e princípio de privilégio mínimo.

## Cenário demonstrado

Um sistema externo, como dashboard, orquestrador, n8n, agente de IA ou automação, precisa consultar informações do servidor sem abrir SSH recorrente nem espalhar credenciais privilegiadas.

Fluxo:

```text
Sistema externo
    -> API autenticada
    -> Minion local
    -> Coletores locais do Linux
```

## Passo 1: build local

```bash
go mod download
CGO_ENABLED=1 go build -ldflags="-s -w" -o minion ./cmd/minion
```

## Passo 2: setup e criação de cliente

```bash
sudo ./minion setup --name demo --ips 127.0.0.1/32
```

Guarde a API key gerada. Ela aparece uma única vez; o Minion armazena apenas o hash.

## Passo 3: subir o serviço em ambiente controlado

Para teste local, use configuração de desenvolvimento com HTTP inseguro apenas em ambiente isolado. Em produção, use TLS.

```bash
sudo ./minion --config /etc/minion/config.json
```

## Passo 4: health check

```bash
curl -k https://localhost:9870/api/v1/health
```

Resposta esperada:

```json
{"status":"ok"}
```

## Passo 5: consulta autenticada

```bash
curl -k \
  -H "Authorization: Bearer <API_KEY>" \
  https://localhost:9870/api/v1/system
```

## Passo 6: validação de segurança

Sem API key, a resposta esperada é erro de autenticação:

```bash
curl -k https://localhost:9870/api/v1/system
```

Resultado esperado:

```json
{"error":"missing authorization header"}
```

Com API key inválida ou IP fora da allowlist, a resposta esperada é:

```json
{"error":"invalid credentials"}
```

## Pontos para destacar em uma apresentação

- O Minion não é shell remoto.
- O Minion não aceita comando livre.
- O Minion não executa planos de LLM.
- O acesso é controlado por cliente, API key e allowlist de IP.
- A API key é armazenada como hash.
- As requisições passam por auditoria.
- A ideia é reduzir SSH recorrente e privilégio espalhado em automações.

## Frase curta para LinkedIn

Minion Agent é um agente local em Go para servidores Linux, criado para expor dados operacionais por API autenticada e auditável, reduzindo dependência de SSH recorrente em automações e integrações com IA.
