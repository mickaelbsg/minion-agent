# Uzinha - Minion Control Panel

A "Uzinha" é um painel de controle web para monitorar e gerenciar múltiplos minions remotamente.

## Início Rápido

```bash
# 1. Configurar minions
vim uzinha/config.json

# 2. Compilar e rodar
cd uzinha
chmod +x run.sh
./run.sh

# 3. Abrir no navegador
open http://localhost:8080
```

## Configuração

Edite `uzinha/config.json`:

```json
{
  "minions": [
    {
      "name": "meu-servidor",
      "host": "https://192.168.1.100:9870",
      "api_key": "minion_sk_...",
      "insecure": true
    }
  ],
  "server": {
    "port": 8080
  }
}
```

### Campos

| Campo | Descrição |
|---|---|
| `name` | Nome identificador do minion |
| `host` | URL completa do minion (https://ip:porta) |
| `api_key` | API key para autenticação |
| `insecure` | Ignorar verificação TLS (true para self-signed) |

## Endpoints da API

A Uzinha expõe:

| Endpoint | Descrição |
|---|---|
| `GET /` | Dashboard principal |
| `GET /api/minions` | Lista todos os minions com dados básicos |
| `GET /api/minion/?name=X` | Dados completos de um minion |

## Dados Coletados

Para cada minion, a Uzinha coleta:

- **Agent**: agent_id, hostname, versão, uptime, capabilities
- **System**: OS, kernel, hostname
- **Memory**: Total, disponível, livre
- **Disk**: Uso de disco por partição
- **Users**: Usuários do sistema

## Funcionalidades

- Dashboard com cards de cada minion
- Indicador online/offline
- Auto-refresh a cada 30 segundos
- Detalhes completos ao clicar em um minion
- Visualização JSON raw
- Tema escuro

## Exemplo de Uso

```bash
# Rodar em background
nohup ./uzinha &

# Verificar minions
curl http://localhost:8080/api/minions

# Ver detalhes de um minion
curl "http://localhost:8080/api/minion/?name=minion-local"
```

## Solução de Problemas

### Minion aparece offline
- Verifique se o minion está rodando: `curl -k https://IP:9870/api/v1/health`
- Verifique a API key no config.json
- Verifique se o IP está correto

### Erro de conexão TLS
- Defina `"insecure": true` no config.json para TLS self-signed

### Dados não atualizam
- O dashboard faz refresh automático a cada 30 segundos
- Recarregue a página manualmente com F5
