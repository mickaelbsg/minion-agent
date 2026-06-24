# ADR-011 — Princípio do Privilégio Mínimo e Deny-by-Default

**Status:** Aprovado
**Data:** 24/06/2026
**Projeto:** Minion

---

## Contexto

O Minion foi criado para reduzir riscos causados por automações externas com privilégios elevados nos servidores.

A motivação original veio do fato de que bots, pipelines, integrações, scripts e agentes externos poderiam exigir permissões administrativas diretamente nos hosts, criando superfície de ataque, ruído de segurança e dificuldade de auditoria.

Por isso, o Minion deve atuar como uma fronteira local controlada entre sistemas externos e o sistema operacional.

---

## Decisão

O Minion seguirá o princípio de privilégio mínimo.

A regra oficial é:

```text
Tudo é proibido, exceto o que foi explicitamente permitido.
```

Isso significa:

* Nenhuma ação existe por padrão.
* Nenhuma integração externa pode solicitar execução genérica.
* Nenhum cliente externo pode enviar instruções livres para o sistema operacional.
* Nenhum LLM, bot, dashboard ou orquestrador pode transformar texto livre em ação privilegiada dentro do host.
* Apenas capacidades previamente definidas, implementadas, validadas, documentadas e auditadas podem existir.

---

## Modelo Permitido

O Minion só deve expor capacidades explícitas por API.

O cliente solicita uma capacidade de negócio, e não uma instrução operacional livre.

O Minion deve:

1. Validar autenticação.
2. Validar IP permitido.
3. Validar status do cliente.
4. Validar payload.
5. Verificar se a capacidade está explicitamente permitida.
6. Executar apenas a operação interna prevista para aquela capacidade.
7. Registrar auditoria.
8. Retornar resultado previsível.

---

## Modelo Proibido

O Minion nunca deve implementar shell remoto, execução genérica ou ação arbitrária recebida de fora.

É proibido criar qualquer endpoint, função ou fluxo que permita ao cliente enviar instruções livres para o sistema operacional.

Também é proibido criar mecanismos indiretos que produzam o mesmo efeito, como prompts, webhooks ou payloads genéricos que resultem em ação privilegiada sem uma capacidade previamente definida.

---

## Regra de Implementação

Qualquer nova capacidade administrativa deve ser implementada como código explícito do Minion.

Cada capacidade deve possuir:

* endpoint próprio
* handler próprio
* validação própria
* allowlist interna quando aplicável
* limites claros de entrada
* auditoria obrigatória
* documentação
* testes automatizados

Se uma operação não estiver definida dessa forma, ela deve ser considerada proibida.

---

## Justificativa

Esse modelo reduz a superfície de ataque e impede que o Minion vire uma porta de execução privilegiada no servidor.

Mesmo que uma API Key vaze, um cliente seja comprometido ou um LLM gere uma instrução incorreta, o dano fica limitado às capacidades explícitas que o Minion expõe.

Essa decisão preserva o motivo original do projeto: eliminar elevação de privilégios distribuída e substituir automações privilegiadas soltas por uma API local controlada, previsível e auditável.

---

## Consequências

O desenvolvimento de novas ações será mais lento, pois cada capacidade precisará ser modelada e implementada explicitamente.

Em troca, o Minion terá uma postura de segurança mais forte, mais auditável e mais defensável perante times de segurança.

O Minion não será um executor remoto genérico.

O Minion será uma API local de capacidades permitidas.
