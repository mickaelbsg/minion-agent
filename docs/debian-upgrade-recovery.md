# Recuperação de upgrade do pacote Debian

O pacote Debian do Minion protege automaticamente o estado operacional existente antes de um upgrade.

## Comportamento normal

O operador continua usando somente:

```bash
sudo apt install ./minion_<versao>_amd64.deb
```

Quando o `dpkg` identifica um upgrade, o `preinst` interrompe temporariamente o serviço e cria um snapshot root-only em:

```text
/var/lib/minion/upgrade-backup
```

O snapshot pode conter:

- `/etc/minion/config.json`;
- certificado e chave TLS;
- banco SQLite e arquivos WAL/SHM;
- binário atualmente instalado;
- unit systemd atualmente instalada;
- versão anterior do pacote.

O diretório recebe modo `0700`. O conteúdo não é impresso no terminal, no journal ou nos logs do pacote.

Depois que o novo pacote inicializa o banco, valida o bootstrap, reinicia o serviço e confirma que `minion.service` está ativo, o snapshot temporário é removido automaticamente.

## Falha durante o upgrade

Se o bootstrap, o novo binário ou a inicialização do serviço falhar, o `postinst` restaura automaticamente:

- configuração;
- TLS;
- SQLite;
- binário anterior;
- unit systemd anterior.

Em seguida, o pacote tenta reiniciar o serviço anterior. O objetivo é manter o agente operacional e preservar a autenticação do Automation mesmo quando o novo pacote não puder ser configurado.

O `dpkg` continuará marcando a configuração do pacote novo como falha. Isso é intencional: o estado operacional é restaurado, mas o administrador ainda precisa investigar a causa antes de repetir o upgrade.

Use:

```bash
sudo systemctl status minion.service
sudo journalctl -u minion.service -n 100 --no-pager
sudo dpkg --audit
```

Depois de corrigir ou substituir o pacote defeituoso, execute novamente:

```bash
sudo apt install ./minion_<versao_corrigida>_amd64.deb
```

## Snapshot mantido após falha

Após um rollback, o snapshot é mantido em `/var/lib/minion/upgrade-backup` para inspeção e recuperação manual. Não copie esse diretório para locais públicos: ele pode conter a chave TLS privada e o banco de clientes, embora as API keys sejam armazenadas apenas como hash.

O snapshot será substituído automaticamente na próxima tentativa de upgrade. A remoção manual só deve ocorrer depois de confirmar que o serviço está saudável e que a recuperação não é mais necessária:

```bash
sudo rm -rf /var/lib/minion/upgrade-backup
```

## Garantias e limites

O rollback restaura o último estado operacional dos arquivos e tenta reativar o serviço anterior. Ele não altera silenciosamente o banco de dados do `dpkg` para fingir que o pacote novo foi configurado com sucesso. Também não substitui uma política externa de backup do servidor.

Mudanças futuras de schema devem permanecer compatíveis com rollback ou utilizar migrações transacionais e versionadas antes de serem incorporadas ao pacote.
