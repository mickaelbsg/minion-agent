# Estado de clientes no lifecycle Debian

## Invariante

No fluxo de instalação, upgrade e reinstalação do pacote Debian, a consulta interna pelo nome `bootstrap` não significa que um cliente com esse nome precisa continuar existindo. Ela responde se o estado de autenticação SQLite já foi inicializado.

- Banco sem clientes: o pacote deve criar o bootstrap inicial e publicar a credencial uma única vez em arquivo root-only.
- Banco com qualquer cliente persistido: o pacote deve preservar o estado e não gerar nova API key, mesmo que o cliente `bootstrap` tenha sido removido.
- Consultas por qualquer nome diferente de `bootstrap`: a correspondência continua exata.
- Erro ao ler configuração ou SQLite: o pacote deve falhar fechado e não assumir banco vazio.

## Motivo

Depois do pareamento com o Automation/n8n, o operador pode remover o cliente bootstrap e manter apenas clientes operacionais. Exigir permanentemente o nome `bootstrap` faria upgrades e reinstalações falharem ou incentivaria recriação insegura de credenciais.

## Proteções esperadas

O lifecycle Debian deve comprovar que configuração, TLS, banco, clientes e API keys existentes permanecem inalterados; o arquivo de bootstrap não reaparece; e nenhuma credencial é enviada para stdout, journal ou artefatos.
