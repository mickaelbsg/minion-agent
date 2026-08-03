# Validação da instalação Debian

## Fluxo oficial

O pacote deve ser instalado com `apt`, não com `dpkg -i` diretamente:

```bash
sudo apt install ./minion_<versao>_amd64.deb
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
- validar que o serviço está ativo;
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

`dpkg -i` não resolve dependências. Se for usado diretamente em uma máquina sem os pacotes requeridos, ele deixa o pacote desembrulhado, mas não configurado. Por isso, `apt install ./package.deb` é o comando oficial e testado.
