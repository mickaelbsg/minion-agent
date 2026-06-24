# Minion Agent

![Go](https://img.shields.io/badge/Go-1.20%2B-00ADD8?style=for-the-badge&logo=go)
![SQLite](https://img.shields.io/badge/SQLite-003B57?style=for-the-badge&logo=sqlite)
![REST API](https://img.shields.io/badge/API-REST-007ACC?style=for-the-badge&logo=rest)
![Systemd](https://img.shields.io/badge/Systemd-E44D26?style=for-the-badge&logo=linux)

## 🚀 Visão Geral do Projeto

O **Minion Agent** é um agente de sistema leve e eficiente, desenvolvido em Go, projetado para operar em servidores Linux (especialmente Ubuntu) com baixo consumo de recursos. Seu principal objetivo é **coletar informações detalhadas do servidor** e **expor esses dados através de uma API REST segura**, eliminando a necessidade de acesso SSH direto para monitoramento e gerenciamento de informações.

Este projeto visa fornecer uma plataforma robusta para:

*   **Coleta de Dados:** Obter métricas e informações cruciais sobre CPU, memória, usuários, serviços, logs de autenticação, regras de firewall (iptables), e status de serviços como Fail2Ban e Wazuh.
*   **Acesso Remoto Seguro:** Oferecer uma API controlada por chaves de acesso (API Keys) e restrição de IPs, permitindo que sistemas externos consultem o estado do servidor de forma programática.
*   **Automação Futura:** Servir como base para futuras funcionalidades de escrita, onde o agente poderá executar comandos e realizar alterações no servidor de forma controlada, sem a necessidade de intervenção manual via SSH.

O Minion Agent é distribuído como um pacote `.deb`, facilitando a instalação e o gerenciamento como um serviço `systemd`.

## 💡 Arquitetura e Design

A arquitetura do Minion Agent é guiada por princípios de simplicidade, eficiência e segurança. As principais decisões de design incluem:

*   **Linguagem de Desenvolvimento (Go):** Escolhida por sua capacidade de gerar binários estáticos, excelente desempenho, baixo consumo de recursos, portabilidade e concorrência nativa. Isso garante que o agente seja leve e funcione bem em diversos ambientes Linux [1].
*   **Banco de Dados Local (SQLite):** Utilizado para persistir configurações, clientes autorizados, eventos e auditoria. O SQLite é embarcado, não requer um servidor de banco de dados separado e oferece bom desempenho para uso local [1].
*   **Modelo de Comunicação (API REST):** A comunicação com sistemas externos é feita exclusivamente via API REST, promovendo facilidade de integração e compatibilidade universal [1].
*   **Segurança:** Implementa um modelo de segurança baseado em **API Keys** e **restrição de IPs** para controlar o acesso à API. As API Keys são geradas pelo próprio agente e armazenadas como hashes seguros (Argon2id) [1].
*   **Agente Observacional:** Na sua versão inicial, o Minion Agent atua estritamente como um coletor de dados, sem funcionalidades de execução remota de comandos ou inteligência artificial. Essas responsabilidades são delegadas a outros componentes do ecossistema (como o Severino), garantindo uma clara separação de responsabilidades [1].

## 🛠️ Instalação e Configuração

### Pré-requisitos

*   Sistema operacional Linux (preferencialmente Ubuntu).
*   `sudo` ou acesso root.

### 1. Baixar o Pacote `.deb`

Você pode baixar a versão mais recente do pacote `.deb` diretamente do repositório GitHub:

```bash
wget https://github.com/mickaelbsg/minion-agent/releases/download/v1.0.0/minion_1.0.0_amd64.deb
```

### 2. Instalar o Agente

Após baixar o pacote, instale-o usando `dpkg`:

```bash
sudo dpkg -i minion_1.0.0_amd64.deb
```

Este comando irá:
*   Instalar o binário `minion` em `/usr/local/bin/`.
*   Copiar o arquivo de serviço `minion.service` para `/lib/systemd/system/`.
*   Criar o diretório de configuração `/etc/minion/` e copiar um `config.json` de exemplo.
*   Habilitar e iniciar o serviço `minion` via `systemd`.

### 3. Configuração Inicial

O arquivo de configuração principal está localizado em `/etc/minion/config.json`. Você precisará editá-lo para definir as API Keys e IPs permitidos.

Exemplo de `config.json`:

```json
{
  "api": {
    "bind": "0.0.0.0:9870"
  },
  "db_path": "/var/lib/minion/minion.db",
  "clients": [
    {
      "name": "meu_cliente_api",
      "allowed_ips": ["192.168.1.0/24", "127.0.0.1/32"],
      "api_key_hash": "minion_sk_...", // Hash da API Key gerada
      "enabled": true
    }
  ]
}
```

**Gerando uma API Key:**

Você pode gerar uma nova API Key e seu hash correspondente usando o próprio binário `minion`:

```bash
sudo minion add client --name meu_novo_cliente --ips 0.0.0.0/0
```

Copie o `API Key Hash` gerado e cole-o no seu `config.json`.

### 4. Reiniciar o Serviço

Após qualquer alteração no `config.json`, reinicie o serviço para que as mudanças sejam aplicadas:

```bash
sudo systemctl restart minion.service
```

Para verificar o status e os logs do serviço:

```bash
sudo systemctl status minion.service
sudo journalctl -u minion -f
```

## 🔌 Endpoints da API

O Minion Agent expõe uma API RESTful para acessar os dados coletados. Todos os endpoints, exceto `/health`, exigem autenticação via `Authorization: Bearer <API_KEY>` e um IP permitido.

| Método | Endpoint               | Descrição                                         | Autenticação |
| :----- | :--------------------- | :------------------------------------------------ | :----------- |
| `GET`  | `/api/v1/health`       | Verifica o status do agente.                      | Não          |
| `GET`  | `/api/v1/system`       | Coleta informações do sistema (hostname, OS, kernel, arquitetura, uptime). | Sim          |
| `GET`  | `/api/v1/users`        | Lista os usuários do sistema.                     | Sim          |
| `GET`  | `/api/v1/services`     | Lista os serviços em execução.                    | Sim          |
| `GET`  | `/api/v1/logins`       | Coleta informações de logins recentes.            | Sim          |
| `GET`  | `/api/v1/memory`       | Coleta informações de uso de memória.             | Sim          |
| `GET`  | `/api/v1/iptables`     | Lista as regras de iptables.                      | Sim          |
| `GET`  | `/api/v1/fail2ban`     | Coleta o status do Fail2Ban.                      | Sim          |
| `POST` | `/api/v1/fail2ban/unban` | Desbloqueia um IP do Fail2Ban. Requer JSON: `{"ip": "<IP>", "jail": "<JAIL>"}`. | Sim          |
| `GET`  | `/api/v1/ipblock?ip=<IP>` | Verifica se um IP está bloqueado.                 | Sim          |
| `GET`  | `/api/v1/wazuh`        | Coleta o status do Wazuh.                         | Sim          |

### Exemplo de Uso com `curl`

**Verificar Saúde (sem autenticação):**

```bash
curl -k https://localhost:9870/api/v1/health
# Saída: {"status":"ok"}
```

**Coletar Informações do Sistema (com autenticação):**

```bash
curl -k -H "Authorization: Bearer <SUA_API_KEY>" https://localhost:9870/api/v1/system
```

## 🛣️ Roadmap Futuro

O projeto Minion Agent está em constante evolução. As próximas etapas incluem:

*   **Fortalecimento da Segurança:** Implementação de mTLS e rotação automática de API Keys.
*   **Observabilidade Avançada:** Integração com Prometheus (`/metrics`) e dashboards Grafana.
*   **Funcionalidades de Escrita:** Adição de endpoints para executar comandos controlados no servidor.
*   **Extensibilidade:** Sistema de plugins para adicionar novos coletores facilmente.
*   **CI/CD:** Automação completa de build, teste e deploy de pacotes `.deb`.

## 📚 Referências

[1] [ADR.md](https://github.com/mickaelbsg/minion-agent/blob/main/ADR.md)


