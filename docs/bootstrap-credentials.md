# Credencial bootstrap do pacote Debian

Depois da instalação do pacote:

```bash
sudo dpkg -i minion_1.0.4_amd64.deb
```

O Minion cria automaticamente o primeiro cliente, inicializa TLS e SQLite e inicia o serviço. Para autorizar o Automation/n8n remoto e visualizar a API key em uma única operação, execute:

```bash
sudo minion bootstrap pair --ips 192.0.2.10/32
```

Substitua o endereço de exemplo pelo IP ou CIDR real do Automation. O comando mantém o bootstrap restrito a localhost até o pareamento, valida o destino informado, atualiza a allowlist e mostra a credencial uma única vez.

A operação:

- exige root;
- aceita somente IP ou CIDR válido;
- recusa symlinks e arquivos de credencial com permissões inseguras;
- não abre acesso para `0.0.0.0/0` automaticamente;
- preserva a credencial se a atualização da allowlist falhar;
- remove o arquivo somente depois de exibir a credencial com sucesso;
- não registra a API key no journal.

Para uso exclusivamente local, o comando anterior continua disponível:

```bash
sudo minion bootstrap show
```

Armazene a API key imediatamente no cofre de credenciais do Automation/n8n. Não salve a chave em workflows, Code nodes, logs, tickets ou documentação.

Se a credencial já tiver sido consumida, crie um cliente novo explicitamente:

```bash
sudo minion client create --name automation --ips 192.0.2.10/32
```

A chave gerada é exibida uma única vez; somente o hash Argon2id permanece no Minion.

## Upgrade e reinstalação

Upgrades e reinstalações não recriam a credencial bootstrap quando já existe um cliente no banco. A configuração, os certificados TLS, o banco SQLite e os clientes persistidos são preservados.

## Recuperação

Verifique o estado do serviço:

```bash
sudo systemctl status minion.service
sudo journalctl -u minion.service -n 100
```

A ausência da credencial bootstrap não impede o serviço de funcionar. Ela indica que a chave já foi consumida ou que clientes já existiam antes da instalação atual.
