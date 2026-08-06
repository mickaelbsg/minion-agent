# Decisões Registradas

- O `.deb` local instalado com `sudo apt install ./minion_<versao>_amd64.deb` é o fluxo oficial. Não há repositório APT remoto publicado.
- O `postinst` é responsável por preparar configuração, TLS, SQLite, bootstrap e serviço; ele também aguarda `/api/v1/health` antes de declarar a instalação concluída.
- Units antigas regulares em `/etc/systemd/system/minion.service` são arquivadas em `/var/lib/minion/legacy-systemd-minion.service`, permitindo que a unit empacotada seja usada sem intervenção manual.
- `docs/troubleshooting.md` é a base operacional obrigatória para problemas reproduzíveis; correções devem registrar sintoma, causa, diagnóstico, correção e validação.
- Antes de iniciar uma tarefa, agentes devem verificar skills aplicáveis e seguir a skill correspondente quando existir.
- `main` é a linha principal de distribuição; branches remotas só devem permanecer quando vinculadas a PR aberto ou trabalho explicitamente em andamento.
- A release `v1.1.4` foi publicada no GitHub com `minion_1.1.4_amd64.deb` como asset.
- Uzinha é o painel de controle local para gerenciar minions e containers Incus — não é um serviço de produção.
- Incus (não LXC) é o orquestrador de containers no WSL2 — scripts devem usar `incus` em vez de `lxc`.
- O deploy do minion no container é feito via `incus file push` + `apt-get install` + postinst — não há download remoto.
- `runCommand` sempre passa por `wsl -e bash -c "..."` — todos os comandos Incus rodam dentro do WSL2.
- Respostas de erro JSON devem usar `json.Marshal` (via `jsonError()`) para evitar injeção de caracteres especiais.
