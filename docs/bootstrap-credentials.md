# Credencial bootstrap do pacote Debian

Depois da instalação do pacote:

```bash
sudo dpkg -i minion_1.0.3_amd64.deb
```

O Minion cria automaticamente o primeiro cliente local, inicializa TLS e SQLite e inicia o serviço. Para visualizar a API key inicial, execute:

```bash
sudo minion bootstrap show
```

O comando:

- exige root;
- recusa symlinks e arquivos com permissões inseguras;
- mostra a credencial somente uma vez;
- remove o arquivo de credencial após a exibição bem-sucedida;
- preserva o arquivo caso a saída falhe;
- não registra a API key no journal.

Armazene a API key imediatamente no cofre de credenciais do Automation/n8n. Não salve a chave em workflows, Code nodes, logs, tickets ou documentação.

Se a credencial já tiver sido consumida, crie um cliente novo explicitamente:

```bash
sudo minion client create --name automation --ips 192.0.2.10/32
```

Substitua o endereço de exemplo pelo IP ou CIDR real do plano de controle. A chave gerada é exibida uma única vez; somente o hash Argon2id permanece no Minion.

## Upgrade e reinstalação

Upgrades e reinstalações não recriam a credencial bootstrap quando já existe um cliente no banco. A configuração, os certificados TLS, o banco SQLite e os clientes persistidos são preservados.

## Recuperação

Verifique o estado do serviço:

```bash
sudo systemctl status minion.service
sudo journalctl -u minion.service -n 100
```

A ausência da credencial bootstrap não impede o serviço de funcionar. Ela indica que a chave já foi consumida ou que clientes já existiam antes da instalação atual.
