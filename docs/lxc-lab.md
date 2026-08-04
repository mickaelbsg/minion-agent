# Laboratório LXC para Teste do Minion Agent

## Visão Geral

O laboratório LXC permite testar a instalação, configuração e acesso à API do minion-agent em ambiente controlado usando containers Linux.

## Pré-requisitos

- Linux com suporte a LXC/LXD (WSL2 com Nested Containers ou servidor nativo)
- `lxc` instalado (`sudo snap install lxd`)
- Pacote `.deb` do minion compilado (`PKG_VER=1.0.5 ./build_deb.sh`)

## Uso Rápido

```bash
# Executar laboratório completo (cria, testa e mantém container)
./scripts/lxc/run-lab.sh

# Ou executar etapas individualmente
./scripts/lxc/create-lab.sh          # Cria container
./scripts/lxc/test-install.sh        # Testa instalação
./scripts/lxc/test-api.sh            # Testa API
./scripts/lxc/destroy-lab.sh         # Destroi container
```

## Estrutura

```
scripts/lxc/
├── create-lab.sh      # Cria container LXC Debian
├── destroy-lab.sh     # Destroi o container
├── test-install.sh    # Testa instalação do .deb
├── test-api.sh        # Testa acesso à API
└── run-lab.sh         # Orquestra tudo
```

## Configuração

Variáveis de ambiente opcionais:

| Variável | Default | Descrição |
|---|---|---|
| `LAB_CONTAINER_NAME` | `minion-lab` | Nome do container |
| `LAB_IMAGE` | `debian:12` | Imagem base do container |
| `LAB_HOST_PORT` | `9870` | Porta mapeada no host |

Exemplo:
```bash
LAB_CONTAINER_NAME=minion-test LAB_HOST_PORT=8080 ./scripts/lxc/run-lab.sh
```

## Testes Executados

### Instalação
- Pacote instalado corretamente
- Serviço systemd ativo
- Arquivos criados com permissões corretas
- Bootstrap credentials gerado
- Health check via API

### API
- Acesso via localhost (dentro do container)
- Acesso externo (do host)
- Endpoints principais: /health, /agent, /system, /users, /memory, /disk
- Consumo correto do bootstrap credentials

## Acesso Manual

```bash
# Entrar no container
lxc exec minion-lab -- bash

# Verificar serviço
systemctl status minion.service
journalctl -u minion.service -f

# Testar API
curl -k -H "Authorization: Bearer <API_KEY>" https://127.0.0.1:9870/api/v1/health

# Do host
curl -k -H "Authorization: Bearer <API_KEY>" https://localhost:9870/api/v1/health
```

## Solução de Problemas

### Container não cria
- Verifique se LXD está rodando: `sudo snap services lxd`
- Inicie se necessário: `sudo snap start lxd`

### Serviço não inicia
- Verifique logs: `lxc exec minion-lab -- journalctl -u minion.service -f`
- Verifique dependências: `lxc exec minion-lab -- systemctl status minion.service`

### API não acessível de fora
- Verifique se a porta está mapeada: `lxc config device list minion-lab`
- Teste dentro do container primeiro
- Verifique firewall no host

### Bootstrap não consome
- Verifique se o arquivo existe: `lxc exec minion-lab -- ls -la /var/lib/minion/`
- Execute pair manualmente: `lxc exec minion-lab -- minion bootstrap pair --config /etc/minion/config.json --ips 127.0.0.1/32`
