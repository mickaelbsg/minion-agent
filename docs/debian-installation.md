# Instalação Debian oficial

O pacote `.deb` é o único artefato oficial de instalação, upgrade e remoção do Minion. Não compile nem copie o binário manualmente em servidores gerenciados.

## Instalação recomendada em host limpo

```bash
sudo bash ./install.sh ./minion_<versao>_amd64.deb
```

O wrapper valida o nome e a arquitetura do pacote e usa o APT para instalar o arquivo local junto com as dependências declaradas pelo pacote. A chamada explícita por `bash` funciona mesmo quando o arquivo foi obtido sem preservar o bit executável. Ele não instala Go, não compila código, não cria certificados e não grava uma unit systemd própria.

## Instalação direta com dpkg

Em hosts que já possuem as dependências do pacote:

```bash
sudo dpkg -i minion_<versao>_amd64.deb
```

Caso o `dpkg` informe dependências ausentes, conclua a mesma instalação do pacote com:

```bash
sudo apt-get -f install
```

Não execute scripts de compilação nem copie arquivos manualmente. Tanto o wrapper quanto o comando direto instalam o mesmo `.deb` e executam os mesmos scripts maintainer.

Ao final da instalação, confirme:

```bash
sudo systemctl is-active minion.service
sudo systemctl status minion.service --no-pager
```

A credencial bootstrap, quando criada pela primeira instalação, fica em:

```text
/var/lib/minion/bootstrap-credentials.txt
```

O arquivo é root-only. Use o comando indicado pelo pós-instalação para parear o Automation e consumir a credencial. Upgrades e reinstalações não devem recriar a credencial nem sobrescrever configuração, TLS ou SQLite existentes.

## Upgrade

```bash
sudo bash ./install.sh ./minion_<nova-versao>_amd64.deb
```

Ou, quando as dependências já estiverem satisfeitas:

```bash
sudo dpkg -i minion_<nova-versao>_amd64.deb
```

O pacote preserva `/etc/minion/config.json`, TLS, banco SQLite e clientes. Antes do upgrade, cria um snapshot root-only e tenta restaurar o estado anterior caso o novo serviço não configure ou inicie corretamente.

## Remoção preservando dados

Use a remoção normal quando pretende reinstalar o agente ou interromper seu uso sem perder identidade, configuração, certificados e clientes:

```bash
sudo dpkg -r minion
```

Esse comando para e desabilita o serviço, remove o binário e a unit pertencentes ao pacote, mas preserva:

- `/etc/minion`;
- `/opt/minion`;
- `/var/lib/minion`.

Uma instalação posterior do pacote reutiliza esse estado e não deve recriar a credencial bootstrap nem invalidar as API keys existentes.

## Purge destrutivo

Use purge somente quando deseja remover deliberadamente todo o estado local do Minion:

```bash
sudo dpkg --purge minion
```

Esse comando apaga configuração, certificados TLS, banco SQLite, clientes, credencial bootstrap, snapshots de upgrade e demais arquivos em `/etc/minion`, `/opt/minion` e `/var/lib/minion`.

O purge é irreversível sem backup. Não o utilize para upgrade, reinstalação ou correção temporária do serviço.

## Diagnóstico de falha

```bash
sudo dpkg --audit
sudo systemctl status minion.service --no-pager
sudo journalctl -u minion.service -n 100 --no-pager
```

Não recorra ao instalador antigo nem copie uma unit para `/etc/systemd/system`. Essa unit teria prioridade sobre a unit oficial em `/lib/systemd/system` e poderia manter comportamento obsoleto após upgrades.
