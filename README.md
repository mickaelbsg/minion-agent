# Minion Agent

![Go](https://img.shields.io/badge/Go-1.25%2B-00ADD8?style=for-the-badge&logo=go)
![SQLite](https://img.shields.io/badge/SQLite-003B57?style=for-the-badge&logo=sqlite)
![REST API](https://img.shields.io/badge/API-REST-007ACC?style=for-the-badge)
![Systemd](https://img.shields.io/badge/Systemd-Linux-E44D26?style=for-the-badge&logo=linux)

## 1. Visão geral

O **Minion Agent** é um agente local para servidores Linux, desenvolvido em Go e executado como serviço `systemd`.

Ele foi criado para resolver um problema específico de infraestrutura e segurança: reduzir a necessidade de automações externas acessarem servidores via SSH, com credenciais privilegiadas ou elevação de privilégios distribuída.

Em vez de cada automação, bot, pipeline, dashboard ou agente de IA acessar diretamente o host, o Minion fica instalado localmente no servidor e expõe uma API HTTP autenticada para consultas e, futuramente, capacidades administrativas explícitas.

O Minion não é um shell remoto.

O Minion não recebe comandos livres.

O Minion não possui IA embarcada.

O Minion não decide, não interpreta e não executa planos de LLM.

Ele coleta, organiza e disponibiliza dados locais de forma segura e auditável.

## 2. Por que o Minion existe

O Minion nasceu de uma preocupação de segurança: automações externas exigindo privilégios elevados diretamente nos servidores.

O modelo antigo tende a gerar:

- SSH recorrente para coleta de dados.
- Credenciais administrativas espalhadas.
- Uso excessivo de elevação de privilégio em automações.
- Ruído em logs de segurança.
- Dificuldade de auditoria.
- Risco de um bot, pipeline ou LLM virar uma ponte para ação privilegiada no servidor.

O modelo proposto é:

```text
Sistema externo
    ↓
API autenticada
    ↓
Minion local
    ↓
Sistema operacional
```

Sistemas como Severino, n8n, dashboards ou outros consumidores consultam a API do Minion. Eles não precisam executar comandos diretamente no host.

## 3. Princípio de segurança

O Minion segue o princípio de privilégio mínimo:

```text
Tudo é proibido, exceto o que foi explicitamente permitido.
```

Na prática:

- Nenhuma ação existe por padrão.
- Nenhum endpoint genérico de execução deve existir.
- Nenhum cliente externo deve enviar comandos livres para o sistema operacional.
- Nenhum LLM, bot, webhook ou orquestrador deve transformar texto livre em ação privilegiada dentro do host.
- Toda capacidade deve ser explícita, implementada no código, validada, documentada e auditada.

O uso de `sudo` é aceitável para o administrador humano instalar pacotes, configurar o serviço ou consultar logs. O runtime do Minion não deve chamar `sudo` internamente.

## 4. O que o Minion faz hoje

A versão atual é majoritariamente observacional.

Ela expõe dados como:

- Informações do sistema.
- Usuário atual do processo.
- Uso de memória.
- Uso de disco.
- Regras de iptables.
- Eventos relacionados a privilégio.
- Journal do sistema.
- Status do Fail2Ban.
- Verificação se um IP está bloqueado.
- Status base do Wazuh.
- Auditoria das requisições.

Existe também uma capacidade controlada para `fail2ban/unban`, validando IP e jail permitida. Essa rota deve ser tratada como ação administrativa explícita.

## 5. O que o Minion não deve fazer

O Minion não deve implementar:

- Shell remoto.
- Execução genérica de comandos.
- Campo livre para comando.
- Campo livre para script.
- Execução direta de plano de LLM.
- Webhook genérico que resulte em ação privilegiada.
- Agente autônomo com permissão de alterar o host sem capacidade explícita.

Futuras ações, como criar usuários, excluir usuários, bloquear IPs, desbloquear IPs ou reiniciar serviços, devem ser implementadas somente como endpoints próprios, com validação, whitelist interna quando aplicável e auditoria.

## 6. Arquitetura

```text
+----------------------+
|      Severino        |
| n8n / Dashboard / IA |
+----------+-----------+
           |
           | HTTP API autenticada
           |
+----------+-----------+
|        Minion        |
|  systemd + SQLite    |
+----------+-----------+
           |
           | Coletores locais
           |
+----------+-----------+
| Servidor Linux       |
+----------------------+
```

Componentes principais:

- `cmd/minion`: ponto de entrada do binário.
- `internal/server`: servidor HTTP, autenticação, rotas e handlers.
- `internal/collectors`: coletores locais.
- `internal/security`: geração e validação de API keys e allowlist de IP.
- `internal/storage`: SQLite, clientes autorizados e auditoria.
- `internal/config`: carregamento de configuração.

## 7. Requisitos

Ambiente recomendado:

- Linux Debian/Ubuntu.
- systemd.
- Go compatível com o `go.mod` do projeto.
- gcc/build-essential para compilar dependências CGO.
- SQLite.
- `last`/`wtmp` disponível no sistema para o endpoint `/api/v1/logins`.
- Fail2Ban, se for usar endpoints relacionados.
- iptables, se for usar coletores de firewall.

Instalação dos pacotes básicos em Debian/Ubuntu:

```bash
sudo apt update
sudo apt install -y golang build-essential gcc sqlite3 dpkg-dev
```

## 8. Instalação via pacote `.deb`

Baixe ou gere o pacote `.deb`.

Instale com:

```bash
sudo dpkg -i minion_1.0.0_amd64.deb
```

Se houver dependência pendente:

```bash
sudo apt-get install -f
```

Verifique o serviço:

```bash
sudo systemctl status minion
```

A unit carregada deve vir de `/lib/systemd/system/minion.service` quando instalada via pacote. Se existir uma unit antiga em `/etc/systemd/system/minion.service`, ela tem prioridade sobre a unit do pacote e pode causar falhas como `status=217/USER`.

Correção para unit antiga:

```bash
sudo systemctl stop minion || true
sudo rm -f /etc/systemd/system/minion.service
sudo systemctl daemon-reload
sudo systemctl reset-failed minion
sudo systemctl restart minion
```

Acompanhe logs:

```bash
sudo journalctl -u minion -f
```

## 9. Configuração

Arquivo principal:

```text
/etc/minion/config.json
```

Exemplo:

```json
{
  "api": {
    "bind": "0.0.0.0:9870"
  },
  "db_path": "/opt/minion/minion.db",
  "clients": []
}
```

O banco SQLite armazena clientes autorizados e auditoria.

Caminho padrão:

```text
/opt/minion/minion.db
```

## Graphify via NVIDIA

Para rodar o `graphify` deste projeto usando o endpoint compatível com OpenAI da NVIDIA, use o wrapper local:

```bash
./scripts/graphify-nvidia.sh extract . --no-cluster
./scripts/graphify-nvidia.sh label .
./scripts/graphify-nvidia.sh query "How does Server connect the project?"
```

O wrapper carrega variáveis de `.graphify.env.local` e define por padrão:

```text
OPENAI_BASE_URL=https://integrate.api.nvidia.com/v1
OPENAI_MODEL=openai/gpt-oss-20b
```

O provider customizado do `graphify` fica em `.graphify/providers.json` e desabilita thinking com `chat_template_kwargs.enable_thinking=false`, para a NVIDIA retornar texto em `message.content`, que é o formato esperado pelo `graphify`.

O arquivo versionado é `.graphify.env.example`. O arquivo `.graphify.env.local` fica fora do Git.

## 10. Gerenciamento de clientes

O Minion controla acesso por cliente.

Cada cliente possui:

- Nome.
- IPs ou CIDRs permitidos.
- API key.
- Hash da API key.
- Status ativo/inativo.

Criar cliente:

```bash
sudo minion add client --name severino --ips 192.168.56.2/32
```

Saída esperada:

```text
Client: severino
API Key: minion_sk_xxxxxxxxxxxxxxxxx
API Key Hash: xxxxxxxxxxxxxxxxx
```

A API key deve ser copiada no momento da criação. O Minion armazena apenas o hash.

Listar clientes:

```bash
sudo minion client list
```

Desabilitar cliente:

```bash
sudo minion client disable severino
```

Habilitar cliente:

```bash
sudo minion client enable severino
```

Remover cliente:

```bash
sudo minion client delete severino
```

## 11. Autenticação

Todos os endpoints, exceto health, exigem:

- IP de origem permitido.
- Header Authorization com API key válida.
- Cliente ativo.

Formato:

```http
Authorization: Bearer <API_KEY>
```

Exemplo:

```bash
curl -k -H "Authorization: Bearer <API_KEY>" https://localhost:9870/api/v1/system
```

## 12. Endpoints

### Health

```http
GET /api/v1/health
```

Não exige autenticação.

### Sistema

```http
GET /api/v1/system
```

Retorna informações básicas do host.

### Usuários

```http
GET /api/v1/users
```

Retorna usuários humanos cadastrados no sistema a partir de `/etc/passwd`.

Critério atual:

- inclui `root` (`uid=0`);
- inclui usuários com `uid >= 1000`;
- ignora contas com shell terminando em `/nologin` ou `/false`.

Exemplo de resposta:

```json
[
  {
    "username": "root",
    "uid": "0",
    "gid": "0",
    "home": "/root"
  },
  {
    "username": "automation",
    "uid": "1004",
    "gid": "1004",
    "home": "/home/automation"
  }
]
```

### Serviços

```http
GET /api/v1/services
```

Endpoint reservado para listagem de serviços.

### Fail2Ban

```http
GET /api/v1/fail2ban
```

Lista jails e IPs banidos detectados pelo Fail2Ban.

### Fail2Ban Unban

```http
POST /api/v1/fail2ban/unban
```

Ação administrativa explícita para remover um IP de uma jail permitida.

Payload:

```json
{
  "ip": "192.0.2.10",
  "jail": "sshd"
}
```

Regras atuais:

- IP precisa ser válido.
- Jail precisa estar na allowlist interna do código.
- Requisição exige autenticação.
- Requisição é auditada.

### IP Block

```http
GET /api/v1/ipblock?ip=<IP>
```

Verifica se um IP está bloqueado.

Exemplo:

```bash
curl -k -H "Authorization: Bearer <API_KEY>" "https://localhost:9870/api/v1/ipblock?ip=192.0.2.10"
```

Validações:

- `ip` é obrigatório.
- `ip` precisa ser um IP válido.

### Wazuh

```http
GET /api/v1/wazuh
```

Retorna estrutura base de status do Wazuh.

### Logins

```http
GET /api/v1/logins
```

Retorna histórico recente de logins bem-sucedidos usando o comando `last`, que lê o banco `wtmp` do sistema.

Comportamento atual:

- consulta até 50 entradas recentes;
- ignora eventos `reboot`, `shutdown` e `runlevel`;
- preserva nomes completos com `last -w`;
- marca `success=true` porque `last` registra sessões autenticadas;
- usa `ip="local"` quando a sessão não possui IP remoto.

Exemplo de resposta:

```json
[
  {
    "user": "mickaelgomes",
    "ip": "192.168.22.115",
    "success": true,
    "timestamp": "2026-06-24T10:03:42-03:00"
  },
  {
    "user": "automation",
    "ip": "192.168.22.115",
    "success": true,
    "timestamp": "2026-06-23T14:22:04-03:00"
  }
]
```

Observação: tentativas de login malsucedidas não aparecem neste endpoint; elas devem ser obtidas por logs de autenticação ou journal.

### Memória

```http
GET /api/v1/memory
```

Retorna informações de memória baseadas em `/proc/meminfo`.

### IPTables

```http
GET /api/v1/iptables
```

Retorna regras do iptables em formato estruturado.

### Disco

```http
GET /api/v1/disk
```

Retorna uso de disco.

### Eventos de privilégio

```http
GET /api/v1/sudo
```

Retorna eventos relacionados a uso de privilégio registrados no journal.

Observação: o nome atual do endpoint é `/sudo`, mas ele coleta eventos; ele não executa sudo. Uma evolução recomendada é renomear para `/api/v1/privilege-events` mantendo compatibilidade temporária.

### Journal

```http
GET /api/v1/journal?limit=100&level=err
```

Retorna logs do journal.

Parâmetros:

- `limit`: quantidade de linhas. Valor padrão: 100. Limite máximo atual: 1000.
- `level`: nível do journal.

## 13. Auditoria

Todas as requisições passam pelo middleware de auditoria.

São registrados:

- Cliente.
- IP de origem.
- Método HTTP.
- Path.
- Status HTTP.
- Timestamp.

Tabela SQLite:

```text
audit
```

A auditoria é parte central do modelo de segurança do Minion.

## 14. Build local

Clone o repositório:

```bash
git clone https://github.com/mickaelbsg/minion-agent.git
cd minion-agent
```

Baixe dependências:

```bash
go mod tidy
```

ou:

```bash
go mod download
```

Compile:

```bash
go build -o minion ./cmd/minion
```

Verifique o binário:

```bash
./minion --help
```

Execute testes:

```bash
go test ./...
```

Build de produção:

```bash
CGO_ENABLED=1 go build -ldflags="-s -w" -o minion ./cmd/minion
```

Observação: o projeto usa SQLite via `go-sqlite3`, portanto CGO deve estar habilitado.

## 15. Como gerar um `.deb` local

Crie a estrutura de empacotamento:

```bash
mkdir -p packaging/minion/DEBIAN
mkdir -p packaging/minion/usr/local/bin
mkdir -p packaging/minion/etc/minion
mkdir -p packaging/minion/lib/systemd/system
```

Compile o binário:

```bash
CGO_ENABLED=1 go build -ldflags="-s -w" -o minion ./cmd/minion
```

Copie o binário:

```bash
cp minion packaging/minion/usr/local/bin/minion
chmod 0755 packaging/minion/usr/local/bin/minion
```

Crie o arquivo `packaging/minion/DEBIAN/control`:

```text
Package: minion
Version: 1.0.0
Section: admin
Priority: optional
Architecture: amd64
Maintainer: Minion Team
Description: Local Linux infrastructure agent with authenticated API
```

Crie o arquivo `packaging/minion/etc/minion/config.json`:

```json
{
  "api": {
    "bind": "0.0.0.0:9870"
  },
  "db_path": "/opt/minion/minion.db",
  "clients": []
}
```

Crie o arquivo `packaging/minion/lib/systemd/system/minion.service`:

```ini
[Unit]
Description=Minion Agent
After=network.target

[Service]
Type=simple
ExecStart=/usr/local/bin/minion --config /etc/minion/config.json
Restart=always
RestartSec=5
User=root
Group=root

[Install]
WantedBy=multi-user.target
```

Ajuste permissões:

```bash
chmod 0644 packaging/minion/DEBIAN/control
chmod 0644 packaging/minion/etc/minion/config.json
chmod 0644 packaging/minion/lib/systemd/system/minion.service
```

Gere o pacote:

```bash
dpkg-deb --build packaging/minion
```

Resultado esperado:

```text
packaging/minion.deb
```

Renomeie com versão e arquitetura, se desejar:

```bash
mv packaging/minion.deb minion_1.0.0_amd64.deb
```

Instale localmente:

```bash
sudo dpkg -i minion_1.0.0_amd64.deb
```

Recarregue systemd:

```bash
sudo systemctl daemon-reload
sudo systemctl enable --now minion
```

Verifique:

```bash
sudo systemctl status minion
```

## 16. Atualização manual

Compile uma nova versão:

```bash
CGO_ENABLED=1 go build -ldflags="-s -w" -o minion ./cmd/minion
```

Substitua no pacote e gere novo `.deb`:

```bash
cp minion packaging/minion/usr/local/bin/minion
dpkg-deb --build packaging/minion
```

Instale por cima:

```bash
sudo dpkg -i packaging/minion.deb
sudo systemctl restart minion
```

## 17. Segurança operacional

Recomendações:

- Restringir `allowed_ips` ao IP exato do consumidor.
- Evitar `0.0.0.0/0` em produção.
- Guardar a API key em cofre ou variável segura.
- Rotacionar API keys periodicamente.
- Monitorar a tabela de auditoria.
- Expor a API apenas em rede interna ou túnel controlado.
- Usar firewall local para limitar acesso à porta 9870.
- Revisar qualquer nova capacidade administrativa antes de liberar.

Exemplo ruim para produção:

```text
allowed_ips: 0.0.0.0/0
```

Exemplo recomendado:

```text
allowed_ips: 192.168.56.2/32
```

## 18. Troubleshooting

Verificar serviço:

```bash
sudo systemctl status minion
```

Ver logs:

```bash
sudo journalctl -u minion -f
```

Testar health:

```bash
curl -k https://localhost:9870/api/v1/health
```

Testar autenticação:

```bash
curl -k -H "Authorization: Bearer <API_KEY>" https://localhost:9870/api/v1/system
```

Erro `missing authorization header`:

- Header `Authorization` não foi enviado.

Erro `invalid credentials`:

- API key incorreta.
- IP de origem não está permitido.
- Cliente está desabilitado.

Erro ao compilar SQLite:

- Verifique se `gcc` e `build-essential` estão instalados.
- Verifique se `CGO_ENABLED=1` está ativo.

## 19. Roadmap

Próximos passos recomendados:

- CI com `go test ./...`, `go vet ./...` e `gofmt`.
- Geração automática de `.deb` via GitHub Actions.
- Endpoint de consulta de auditoria.
- Renomear `/api/v1/sudo` para `/api/v1/privilege-events`.
- Testes automatizados para autenticação, allowlist de IP e parsers.
- RBAC antes de ampliar ações administrativas.
- mTLS em versão futura.
- Novas capacidades administrativas explícitas, sempre com privilégio mínimo e deny-by-default.

## 20. Documentos relacionados

- `ADR.md`
- `SPEC.md`
- `docs/adr/ADR-011-principio-privilegio-minimo.md`

## 21. Filosofia do projeto

O Minion não é um executor remoto genérico.

O Minion é uma API local de capacidades permitidas.

A regra central permanece:

```text
Tudo é proibido, exceto o que foi explicitamente permitido.
```
