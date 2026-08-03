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
