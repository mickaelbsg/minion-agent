# Validação da instalação Debian

## Fluxo oficial

O fluxo oficial é `apt` instalando um `.deb` local. O projeto ainda não publica um repositório APT remoto; o arquivo deve estar disponível na máquina ou ser gerado antes:

```bash
PKG_VER=1.0.5 ./build_deb.sh
```

Instale com `apt`, não com `dpkg -i` diretamente:

```bash
sudo apt install ./minion_1.0.5_amd64.deb
```

O `apt` resolve as dependências declaradas antes de executar o `postinst`. O usuário não precisa instalar `fail2ban`, `iptables`, `openssl` ou `sqlite3` separadamente.

## Resultado esperado

Após a instalação, o `postinst` deve:

- preservar ou criar `/etc/minion/config.json`;
- criar TLS em `/etc/minion/tls`;
- inicializar `/opt/minion/minion.db`;
- criar a identidade determinística do agente a partir do `machine-id`;
- criar o cliente bootstrap somente quando o banco ainda não possui clientes;
- salvar a credencial em `/var/lib/minion/bootstrap-credentials.txt` com modo `0600`;
- habilitar e iniciar `minion.service`;
- substituir automaticamente uma unit antiga em `/etc/systemd/system/minion.service`, arquivando-a em `/var/lib/minion/legacy-systemd-minion.service`;
- validar que o serviço está ativo e que `/api/v1/health` aceita conexões antes de concluir o `postinst`;
- informar endereço HTTPS, `agent_id` e caminho da credencial bootstrap sem imprimir a API key.

## Comandos de validação

```bash
dpkg-query -W -f='${Status} ${Version}\n' minion
sudo systemctl is-active minion.service
sudo stat -c '%a %U:%G %n' /etc/minion/config.json /etc/minion/tls/minion.key /opt/minion/minion.db /var/lib/minion/bootstrap-credentials.txt
sudo sqlite3 /opt/minion/minion.db '.tables'
curl --silent --show-error --fail --insecure https://127.0.0.1:9870/api/v1/health
```

O pareamento com o Automation consome a credencial uma única vez:

```bash
sudo minion bootstrap pair --ips <AUTOMATION_IP/32>
```

Depois do pareamento, a API key deve ser armazenada no cofre do Automation e o arquivo bootstrap deve deixar de existir.

## Reinstalação e upgrade

Executar novamente `apt install ./minion_<versao>_amd64.deb` não deve recriar configuração, banco, TLS, identidade ou clientes existentes. Um upgrade com falha deve restaurar o estado operacional anterior e manter o snapshot root-only em `/var/lib/minion/upgrade-backup` para diagnóstico.

## Limite do comando `dpkg -i`

`dpkg -i` não resolve dependências. Se for usado diretamente em uma máquina sem os pacotes requeridos, ele deixa o pacote desembrulhado, mas não configurado. Por isso, `apt install ./package.deb` é o comando oficial e testado. `apt install minion` sem o caminho `./` não faz parte do fluxo atual, pois não há repositório APT remoto publicado.
