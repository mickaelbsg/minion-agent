# Troubleshooting

## Instalação

O fluxo suportado é instalar um `.deb` local com `apt`:

```bash
sudo apt install ./minion_<versao>_amd64.deb
```

Não use `dpkg -i` como fluxo normal: ele não resolve as dependências declaradas pelo pacote. O projeto não publica atualmente um repositório APT remoto, portanto `apt install minion` não é um comando válido para este produto.

O pacote instala automaticamente as dependências, cria configuração, TLS, SQLite, bootstrap e serviço `systemd`. Se a instalação falhar, preserve a saída do `apt` e consulte:

```bash
sudo systemctl status minion.service --no-pager -l
sudo journalctl -u minion.service -n 100 --no-pager
sudo dpkg-query -W -f='${Status} ${Version}\n' minion
```

## Instalação sem OpenSSL

Sintoma: uma versão antiga do instalador ou do comando `minion setup` falha com `openssl: command not found`, mesmo que configuração, banco e diretórios possam ser criados.

Causa: o fluxo legado delegava a criação do certificado bootstrap ao executável externo OpenSSL. O pacote atual gera e valida o par TLS dentro do próprio binário usando a biblioteca criptográfica do Go; `openssl` não deve aparecer em `Depends`, no `postinst` nem como requisito de recuperação.

Diagnóstico:

```bash
command -v openssl || true
dpkg-deb -f ./minion_<versao>_amd64.deb Depends
mkdir -p /tmp/minion-control
dpkg-deb --control ./minion_<versao>_amd64.deb /tmp/minion-control
grep -nE 'openssl|require_command' /tmp/minion-control/postinst || true
```

Correção: instale uma versão que utilize `minion package ensure-tls` no `postinst` e geração TLS nativa no `minion setup`. Não crie certificado ou chave manualmente para contornar estado parcial. Se apenas um dos arquivos existir, preserve-o para análise e corrija a origem antes de repetir a instalação.

Verificação esperada em host descartável:

```bash
sudo dpkg -i ./minion_<versao>_amd64.deb
sudo systemctl is-active minion.service
sudo stat -c '%a %U:%G %n' /etc/minion/tls /etc/minion/tls/minion.crt /etc/minion/tls/minion.key
sudo minion package ready --config /etc/minion/config.json
```

O resultado esperado é serviço `active`, diretório TLS `0700`, chave privada `0600`, certificado regular válido e readiness concluído sem executar OpenSSL. Upgrades e reinstalações devem preservar o par TLS existente byte a byte.

## Serviço ativo, API indisponível

O instalador valida `/api/v1/health` antes de concluir. Depois da instalação, confirme:

```bash
sudo systemctl is-active minion.service
curl --silent --show-error --fail --insecure https://127.0.0.1:9870/api/v1/health
```

Se o serviço estiver ativo mas a API não responder, verifique o journal, a porta e os arquivos TLS:

```bash
sudo ss -ltnp | grep ':9870'
sudo stat -c '%a %U:%G %n' /etc/minion/tls/minion.crt /etc/minion/tls/minion.key
sudo journalctl -u minion.service -n 100 --no-pager
```

TLS permanece obrigatório por padrão. `api.allow_insecure_http=true` é somente um fallback explícito para desenvolvimento.

## Unit systemd antiga

Uma instalação manual antiga em `/etc/systemd/system/minion.service` pode ter precedência sobre a unit empacotada. O `postinst` atual detecta essa unit regular, arquiva-a em:

```text
/var/lib/minion/legacy-systemd-minion.service
```

e ativa a unit oficial do pacote. O cliente não deve remover a unit manualmente. Para confirmar a unit efetivamente carregada:

```bash
systemctl show -p FragmentPath --value minion.service
readlink -f /lib/systemd/system/minion.service
```

Os caminhos resolvidos devem coincidir.

## Bootstrap e autenticação

Após uma instalação nova, a credencial fica em arquivo root-only e pode ser consumida uma vez:

```bash
sudo minion bootstrap pair --ips <AUTOMATION_IP/32>
```

Se o arquivo já não existir, ele foi consumido ou já havia clientes persistidos. Crie um cliente explicitamente:

```bash
sudo minion client create --name automation --ips <AUTOMATION_IP/32>
```

Clientes autenticados precisam de API key válida, cliente ativo e IP/CIDR autorizado. Não procure a API key em logs: somente o hash é persistido.

## Falha ao criar, rotacionar ou revogar cliente por falta de entropia

Sintoma: comandos administrativos de credenciais falham com erro semelhante a `failed to hash API key`, `failed to hash revocation secret` ou `generate Argon2id salt`. Nenhuma API key nova é criada e o cliente existente permanece inalterado.

Causa: o sistema operacional não conseguiu fornecer aleatoriedade criptograficamente segura para gerar o salt Argon2id. O Minion falha fechado e não usa salt previsível ou fallback determinístico.

Diagnóstico:

```bash
cat /proc/sys/kernel/random/entropy_avail
sudo dmesg --level=err,warn | tail -n 50
sudo journalctl -u minion.service -n 100 --no-pager
```

Em sistemas virtuais recém-inicializados, confirme também se o kernel dispõe de uma fonte de entropia adequada e se não há falhas no dispositivo aleatório:

```bash
ls -l /dev/random /dev/urandom
sudo systemctl status systemd-random-seed.service --no-pager -l
```

Correção: resolva a indisponibilidade de entropia no host e repita o comando. Não altere o banco SQLite manualmente e não tente substituir o Argon2id por hash estático.

Verificação esperada:

```bash
sudo minion client list
```

Para criação ou rotação, repita a operação e confirme que uma nova API key foi retornada apenas uma vez. Para revogação, confirme que o cliente aparece revogado. Em caso de falha, o cliente anterior deve continuar com o mesmo estado e credencial.

## Upgrade, rollback e remoção

Use o harness completo em um host Debian/Ubuntu com `systemd` como PID 1:

```bash
bash scripts/test-deb-lifecycle.sh <install.deb> <upgrade.deb> [broken.deb]
```

Ele valida instalação, permissões, bootstrap, API, upgrade, rollback e remoção. Upgrades preservam configuração, banco, TLS e clientes; remoção não apaga esses dados por padrão.

## WSL

O teste de lifecycle requer WSL com `systemd` habilitado e executando como PID 1, além de `sudo`, `apt`, `dpkg`, `sqlite3` e `curl`:

```bash
ps -p 1 -o comm=
systemctl is-system-running
```

O resultado esperado do primeiro comando é `systemd`.

## Scripts Bash falhando no WSL

Se `bash -n build_deb.sh` falhar em um `elif` ou `fi` no WSL, verifique os finais de linha:

```bash
file build_deb.sh scripts/test-deb-lifecycle.sh
git ls-files --eol build_deb.sh scripts/test-deb-lifecycle.sh
```

Os scripts devem usar `LF`. O repositório define isso em `.gitattributes`; depois de atualizar o checkout, confirme novamente a saída e execute:

```bash
bash -n build_deb.sh scripts/test-deb-lifecycle.sh
```
