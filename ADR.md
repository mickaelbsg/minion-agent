# ADR-001 — Arquitetura Base do Minion

**Status:** Aprovado
**Data:** 11/06/2026
**Projeto:** Minion
**Autor:** Equipe Minion

---

# Contexto

A organização possui múltiplos sistemas que necessitam coletar informações operacionais de servidores Linux.

Historicamente, a coleta e automação eram realizadas principalmente através de:

* SSH
* Execução remota de comandos
* Scripts distribuídos
* Consultas recorrentes em arquivos de log
* Automações com necessidade de elevação de privilégios nos hosts

Esse modelo gera:

* Grande volume de logs de autenticação
* Alertas constantes em ferramentas de segurança
* Necessidade de gerenciamento de credenciais privilegiadas
* Complexidade operacional
* Baixa padronização entre integrações
* Risco de elevação de privilégios distribuída entre automações, bots, pipelines e integrações externas

A motivação principal do Minion nasceu de uma preocupação de segurança: automações externas estavam exigindo subida de privilégios nos servidores, o que aumentava a superfície de ataque e dificultava auditoria.

Foi identificada a necessidade de criar uma camada padronizada para coleta e exposição de informações locais dos servidores Linux, evitando que cada automação precise acessar o host diretamente com privilégios elevados.

---

# Decisão

Será desenvolvido um agente denominado **Minion**.

O Minion será executado localmente em servidores Linux como serviço systemd e será responsável por:

* Coletar informações operacionais.
* Armazenar informações localmente.
* Expor uma API HTTP para consulta.
* Controlar o acesso aos seus dados.
* Concentrar localmente as permissões necessárias para inspecionar o sistema operacional.

O Minion não executará análises, correlações ou decisões automatizadas.

Sistemas externos, como Severino, n8n, dashboards, LLMs e outras integrações, não devem acessar diretamente o host com credenciais privilegiadas. Eles devem consumir exclusivamente a API do Minion.

---

# Consequências

## Positivas

* Eliminação da necessidade de SSH recorrente.
* Redução de ruído em logs.
* Padronização de integrações.
* Menor dependência de credenciais administrativas distribuídas.
* Arquitetura simples e previsível.
* Redução da superfície de ataque causada por automações externas privilegiadas.
* Auditoria centralizada das consultas e futuras ações administrativas.

## Negativas

* Necessidade de instalação do agente.
* Necessidade de manutenção do binário.
* Necessidade de atualização dos agentes.
* O Minion passa a ser um componente sensível e deve ser protegido, versionado e auditado com rigor.

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

Além disso, o Minion existe para reduzir a necessidade de credenciais privilegiadas distribuídas em automações externas.

---

# Decisão

O acesso será protegido por:

* Restrição por IP.
* API Key.
* Status ativo/inativo.

O Minion será a fronteira local de segurança entre sistemas externos e o sistema operacional.

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

O modelo reduz a exposição de credenciais administrativas fora do host e permite que a segurança audite um ponto controlado em vez de múltiplas automações privilegiadas.

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
Executar capacidades locais explícitas quando aprovadas em versões futuras
```

Severino:

```text
Interpretar
Correlacionar
Analisar
Responder
Solicitar ações controladas quando autorizado
```

---

# ADR-010 — Ações Administrativas Controladas

**Status:** Aprovado

---

# Contexto

Foi avaliada a possibilidade de permitir ações administrativas nos hosts.

O Minion foi criado justamente para evitar que automações externas precisem elevar privilégios diretamente nos servidores. Portanto, o modelo de ação precisa preservar essa decisão de segurança.

---

# Decisão

A V1 será predominantemente observacional.

Ações administrativas amplas, como criar usuários, excluir usuários, bloquear IPs, desbloquear IPs e outras alterações no sistema operacional, não fazem parte do escopo inicial.

Versões futuras poderão incluir ações administrativas, desde que sejam expostas como capacidades explícitas, específicas e auditáveis da API.

Cada ação administrativa futura deverá ter:

* Endpoint próprio.
* Handler próprio.
* Validação própria de entrada.
* Registro de auditoria.
* Comportamento previsível e documentado.
* Mapeamento interno para uma operação permitida.

Exemplos aceitáveis para versões futuras:

```text
POST /api/v1/users
DELETE /api/v1/users/{username}
POST /api/v1/ipblock
DELETE /api/v1/ipblock/{ip}
POST /api/v1/services/{name}/restart
```

Não será permitido endpoint genérico para execução de comandos shell.

Não será permitido receber comando arbitrário por payload, string, prompt, JSON ou qualquer outro formato externo.

Exemplo proibido:

```text
POST /api/v1/execute
{
  "command": "useradd joao"
}
```

O modelo correto é API de capacidade, não shell remoto.

Quando uma capacidade precisar executar algo no sistema operacional, o comando real deverá estar implementado internamente no código do Minion ou em whitelist local controlada pelo próprio Minion, nunca recebido bruto de um cliente externo.

Exemplo correto:

```text
POST /api/v1/users
{
  "username": "joao",
  "shell": "/bin/bash"
}
```

Neste modelo, o cliente solicita a capacidade `criar usuário`. O Minion valida os parâmetros e executa internamente apenas a operação permitida para aquela capacidade.

---

# Justificativa

* Preserva a motivação original do projeto: eliminar elevação de privilégios distribuída em automações externas.
* Reduz superfície de ataque.
* Evita execução arbitrária por bots, LLMs, pipelines ou integrações externas.
* Impede que clientes externos transformem o Minion em shell remoto.
* Permite auditoria por ação de negócio, e não por comando bruto.
* Mantém separação clara entre inteligência/orquestração externa e capacidade local controlada.

---

# Consequências

O Minion poderá evoluir para executar alterações no sistema operacional, mas somente através de endpoints específicos, validados, registrados em auditoria e projetados com regras claras.

O Minion não substituirá ferramentas de automação como Ansible, AWX ou SaltStack para execução genérica. Seu papel será expor capacidades locais seguras para operações recorrentes e bem definidas.
