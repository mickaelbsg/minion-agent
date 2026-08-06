# Troubleshooting

## Instalação Debian e API

- Sintoma confirmado: `systemd` reportava o serviço ativo, mas uma conexão imediata à API podia ocorrer antes de o socket estar pronto.
- Causa confirmada: o `postinst` verificava apenas `systemctl is-active`.
- Correção confirmada: o pacote declara `curl` e o `postinst` aguarda `https://127.0.0.1:9870/api/v1/health` antes de concluir.
- Sintoma confirmado: uma unit antiga em `/etc/systemd/system/minion.service` mascarava a unit empacotada em `/lib/systemd/system/minion.service`.
- Causa confirmada: a precedência de units do `systemd`.
- Correção confirmada: a unit regular antiga é arquivada em `/var/lib/minion/legacy-systemd-minion.service` e a unit do pacote passa a ser usada automaticamente.
- Verificação reproduzível: `bash scripts/test-deb-lifecycle.sh <install.deb> <upgrade.deb> [broken.deb]` passou no Ubuntu WSL com instalação, API, upgrade, rollback e remoção.
- Dead end evitado: remover manualmente a unit antiga ou usar `dpkg -i`; o primeiro cria uma etapa de operação para o cliente e o segundo não resolve dependências.

## Scripts no WSL

- Sintoma confirmado: `bash -n build_deb.sh` falhava em `elif` no checkout Windows.
- Causa confirmada: finais de linha `CRLF` no working tree montado pelo WSL.
- Correção confirmada: `.gitattributes` fixa `LF` para `*.sh`; os scripts foram normalizados.
- Verificação: `git ls-files --eol` mostra `i/lf w/lf` e `bash -n build_deb.sh scripts/test-deb-lifecycle.sh` passa.

## Uzinha + Incus no WSL2

### Containers sem acesso à internet (WSL2)
- Sintoma: `apt-get update` falha dentro dos containers Incus com "Network is unreachable".
- Causa: WSL2 não configura NAT automaticamente para a bridge `incusbr0` (10.162.89.0/24).
- Correção: Adicionar regras de iptables no WSL2:
  ```bash
  sudo iptables -t nat -A POSTROUTING -s 10.162.89.0/24 ! -o incusbr0 -j MASQUERADE
  sudo iptables -A FORWARD -i incusbr0 -j ACCEPT
  sudo iptables -A FORWARD -o incusbr0 -m state --state RELATED,ESTABLISHED -j ACCEPT
  ```
- A Uzinha agora executa `ensureNAT()` automaticamente antes de criar containers.

### `-c systemd=true` não existe no Incus
- Sintoma: `incus launch` falha com "Unknown configuration key: systemd".
- Causa: Incus (sucessor do LXC/LXD) não suporta a chave `systemd=true`.
- Correção: Removida a flag. O systemd funciona normalmente como PID 1 no Debian 12 sem essa config.

### JSON com control characters (`\r`) no frontend
- Sintoma: "Bad control character in string literal in JSON at position 95" no JavaScript do frontend.
- Causa: `incus launch`, `incus exec` e `incus file push` escrevem progresso com `\r` (carriage return) no output. Quando o comando falha, o `ExitError.Stderr` incluía esses caracteres, que eram injetados diretamente no JSON de erro via `fmt.Sprintf("{"error":"%s"}", err.Error())`.
- Correção: (1) `runCommand` agora usa `stripControl()` para remover `\r` e outros caracteres de controle < 32. (2) Respostas de erro agora usam `json.Marshal` via `jsonError()` em vez de `fmt.Sprintf`.
- Verificação: todos os endpoints retornam JSON limpo sem caracteres de controle.

### Glob pegava .deb mais antigo
- Sintoma: Deploy instalava minion 1.0.0 em vez de 1.1.4.
- Causa: `filepath.Glob("../../minion_*_amd64.deb")` retornava arquivos em ordem alfabética; `minion_1.0.0` vinha antes de `minion_1.1.4`.
- Correção: Adicionada `latestDeb()` para selecionar a versão mais recente por comparação de strings.

### Caminho relativo errado para .deb
- Sintoma: `filepath.Glob("../../minion_*_amd64.deb")` não encontrava arquivos.
- Causa: Uzinha roda em `uzinha/`, mas `../../` ia dois níveis acima do project root.
- Correção: Corrigido para `../minion_*_amd64.deb`.

### `incus file push` com path Windows
- Sintoma: `incus file push` falhava porque o path Windows não é acessível pelo WSL.
- Causa: O path `C:\Users\...\minion.deb` precisa ser convertido para `/mnt/c/Users/...`.
- Correção: Adicionada `windowsToWSLPath()` que converte paths Windows para WSL.

### `CombinedOutput()` misturava stdout e stderr
- Sintoma: JSON parsing falhava intermitentemente.
- Causa: `cmd.CombinedOutput()` misturava output de sucesso com mensagens de erro.
- Correção: Trocado para `cmd.Output()` com captura separada de stderr via `ExitError.Stderr`.
