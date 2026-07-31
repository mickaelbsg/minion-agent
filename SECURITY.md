# Política de Segurança

## Versões suportadas

O Minion Agent está em evolução ativa. Correções de segurança são aplicadas à versão mais recente disponível na branch `main` e, quando existirem releases publicadas, à release estável mais recente.

Versões antigas, forks e binários modificados por terceiros não possuem garantia de correção.

## Escopo de segurança

São considerados relatos de segurança, entre outros:

- bypass de autenticação ou da allowlist de IP;
- exposição de API keys, hashes, dados de auditoria ou informações sensíveis do host;
- execução de comandos fora das capacidades explícitas do Minion;
- elevação de privilégio causada pelo serviço, instalador ou pacote;
- validação insuficiente em endpoints administrativos;
- falhas de autorização entre clientes;
- criação de arquivos, permissões ou serviços de forma insegura;
- vulnerabilidades em dependências com impacto comprovável no projeto.

Problemas gerais de funcionamento, documentação ou sugestões sem impacto de segurança devem ser registrados como issues comuns.

## Como reportar uma vulnerabilidade

Não publique vulnerabilidades em issues, pull requests, discussões ou comentários públicos.

Use o recurso **Report a vulnerability** da aba **Security** do repositório, que cria um GitHub Security Advisory privado. Inclua, quando possível:

- descrição técnica e impacto;
- versão, commit ou release afetada;
- passos mínimos para reprodução;
- pré-condições necessárias;
- evidências sanitizadas;
- sugestão de correção, caso exista.

Nunca inclua API keys reais, senhas, tokens, certificados privados, endereços internos completos, dumps de banco ou logs com dados pessoais. Substitua esses valores por exemplos fictícios.

## Processo de triagem

O reporte será avaliado quanto a reprodutibilidade, severidade e impacto. Durante a análise, podem ser solicitadas informações adicionais por meio do advisory privado.

Quando a falha for confirmada, o objetivo será:

1. limitar a exposição pública dos detalhes;
2. preparar uma correção e testes de regressão;
3. validar possíveis impactos de compatibilidade e implantação;
4. publicar a correção antes ou junto da divulgação técnica;
5. creditar o pesquisador, quando autorizado.

Relatos duplicados, sem impacto demonstrável ou fora do escopo podem ser encerrados com justificativa.

## Divulgação responsável

Evite divulgar detalhes técnicos antes da disponibilização de uma correção. A data e o nível de detalhe da divulgação devem ser coordenados no advisory privado.

A existência desta política não autoriza testes destrutivos, acesso a sistemas de terceiros, indisponibilidade deliberada, engenharia social ou coleta de dados sem permissão.