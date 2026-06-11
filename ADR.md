# ADR-001 — Arquitetura Base do Minion

**Status:** Aprovado
**Data:** 11/06/2026
**Projeto:** Minion
**Autor:** Equipe Minion

---

# Contexto

A organização possui múltiplos sistemas que necessitam coletar informações operacionais de servidores Linux.

Atualmente a coleta é realizada principalmente através de:

* SSH
* Execução remota de comandos
* Scripts distribuídos
* Consultas recorrentes em arquivos de log

Esse modelo gera:

* Grande volume de logs de autenticação
* Alertas constantes em ferramentas de segurança
* Necessidade de gerenciamento de credenciais
* Complexidade operacional
* Baixa padronização entre integrações

Foi identificada a necessidade de criar uma camada padronizada para coleta e exposição de informações locais dos servidores Linux.

---

# Decisão

Será desenvolvido um agente denominado **Minion**.

O Minion será executado localmente em servidores Linux e será responsável por:

* Coletar informações operacionais.
* Armazenar informações localmente.
* Expor uma API HTTP para consulta.
* Controlar o acesso aos seus dados.

O Minion não executará análises, correlações ou decisões automatizadas.

---

# Consequências

## Positivas

* Eliminação da necessidade de SSH recorrente.
* Redução de ruído em logs.
* Padronização de integrações.
* Menor dependência de credenciais administrativas.
* Arquitetura simples e previsível.

## Negativas

* Necessidade de instalação do agente.
* Necessidade de manutenção do binário.
* Necessidade de atualização dos agentes.

---

# ADR-002 — Linguagem de Desenvolvimento

**Status:** Aprovado

---

# Contexto

O agente deverá operar em diversos ambientes Linux com baixo consumo de recursos.

---

# Decisão

O Minion será desenvolvido em Go.

---

# Justificativa

* Binário único.
* Excelente portabilidade.
* Baixo consumo de memória.
* Facilidade de distribuição.
* Boa concorrência nativa.
* Sem dependência de runtime externo.

---

# Alternativas Consideradas

## Python

Rejeitado.

Motivos:

* Dependência de interpretador.
* Distribuição mais complexa.
* Maior consumo de memória.

## Node.js

Rejeitado.

Motivos:

* Necessidade de runtime.
* Maior complexidade operacional.

---

# ADR-003 — Distribuição em Binário Único

**Status:** Aprovado

---

# Contexto

O processo de instalação deve ser simples.

---

# Decisão

O Minion será distribuído como binário único.

---

# Justificativa

Instalação simplificada:

```bash
cp minion /usr/local/bin/
systemctl enable minion
```

Sem dependências externas.

---

# ADR-004 — Banco de Dados Local

**Status:** Aprovado

---

# Contexto

O agente necessita persistir:

* Configurações
* Clientes autorizados
* Eventos
* Auditoria

---

# Decisão

Utilizar SQLite.

---

# Justificativa

* Embutido.
* Sem servidor adicional.
* Excelente desempenho local.
* Backup simples.
* Compatível com binário único.

---

# Alternativas Consideradas

## PostgreSQL

Rejeitado.

Complexidade incompatível com o objetivo do produto.

## Arquivos JSON

Rejeitado.

Dificuldade para consultas futuras.

---

# ADR-005 — Modelo de Comunicação

**Status:** Aprovado

---

# Contexto

Era necessário definir como sistemas externos irão acessar o Minion.

---

# Decisão

O Minion disponibilizará API REST.

---

# Justificativa

* Simplicidade.
* Facilidade de integração.
* Compatibilidade universal.
* Curva de aprendizado reduzida.

---

# Alternativas Consideradas

## gRPC

Adiado para futuras versões.

## WebSocket

Adiado para futuras versões.

---

# ADR-006 — Modelo de Segurança

**Status:** Aprovado

---

# Contexto

O Minion deve permitir acesso apenas a clientes autorizados.

---

# Decisão

O acesso será protegido por:

* Restrição por IP.
* API Key.
* Status ativo/inativo.

---

# Fluxo

```text
Cliente
   ↓
Validação de IP
   ↓
Validação da API Key
   ↓
Validação do Status
   ↓
Acesso liberado
```

---

# Justificativa

Fornece proteção adequada para ambientes internos sem aumentar a complexidade da V1.

---

# Alternativas Consideradas

## OAuth2

Rejeitado.

Complexidade excessiva para o escopo atual.

## OpenID Connect

Rejeitado.

Dependência externa desnecessária.

## mTLS

Postergado para V2.

---

# ADR-007 — Geração de API Keys

**Status:** Aprovado

---

# Contexto

Era necessário definir quem controla as credenciais de acesso.

---

# Decisão

O próprio Minion gerará as API Keys.

---

# Justificativa

* Controle local.
* Independência de sistemas externos.
* Menor acoplamento.
* Instalação simplificada.

---

# Regras

* A chave é exibida apenas uma vez.
* Apenas o hash será armazenado.
* Utilizar Argon2id para armazenamento.

---

# ADR-008 — Modelo de Permissões

**Status:** Aprovado

---

# Contexto

Foi avaliada a implementação de RBAC.

---

# Decisão

A V1 utilizará modelo ALL OR NOTHING.

---

# Funcionamento

Cliente autenticado:

```text
Acesso total
```

Cliente não autenticado:

```text
Sem acesso
```

---

# Justificativa

* Menor complexidade.
* Menor esforço de manutenção.
* Menor volume de código.
* Atende ao caso de uso atual.

---

# ADR-009 — Ausência de Inteligência Artificial

**Status:** Aprovado

---

# Contexto

O ecossistema inclui o Severino, responsável por funções de IA.

---

# Decisão

O Minion não possuirá:

* LLM
* Embeddings
* Vetores
* RAG
* Agentes autônomos

---

# Justificativa

Separação clara de responsabilidades.

Minion:

```text
Coletar
Armazenar
Disponibilizar
```

Severino:

```text
Interpretar
Correlacionar
Analisar
Responder
```

---

# ADR-010 — Execução Remota de Comandos

**Status:** Aprovado

---

# Contexto

Foi avaliada a possibilidade de permitir execução remota.

---

# Decisão

A V1 não executará comandos remotos.

---

# Justificativa

* Redução da superfície de ataque.
* Simplificação da segurança.
* Foco na coleta de informações.

---

# Consequências

O Minion será exclusivamente um agente observacional na versão inicial.

Não realizará alterações no sistema operacional.

Não executará ações administrativas.

Não substituirá ferramentas de automação como Ansible, AWX ou SaltStack.
