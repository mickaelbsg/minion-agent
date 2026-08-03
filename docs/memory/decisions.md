# Decisões Registradas

- O `.deb` local instalado com `sudo apt install ./minion_<versao>_amd64.deb` é o fluxo oficial. Não há repositório APT remoto publicado.
- O `postinst` é responsável por preparar configuração, TLS, SQLite, bootstrap e serviço; ele também aguarda `/api/v1/health` antes de declarar a instalação concluída.
- Units antigas regulares em `/etc/systemd/system/minion.service` são arquivadas em `/var/lib/minion/legacy-systemd-minion.service`, permitindo que a unit empacotada seja usada sem intervenção manual.
- `docs/troubleshooting.md` é a base operacional obrigatória para problemas reproduzíveis; correções devem registrar sintoma, causa, diagnóstico, correção e validação.
- Antes de iniciar uma tarefa, agentes devem verificar skills aplicáveis e seguir a skill correspondente quando existir.
- `main` é a linha principal de distribuição; branches remotas só devem permanecer quando vinculadas a PR aberto ou trabalho explicitamente em andamento.
- A release `v1.1.4` foi publicada no GitHub com `minion_1.1.4_amd64.deb` como asset.
