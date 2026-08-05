# Integrações opcionais no pacote Debian

A instalação base do Minion não depende de Fail2Ban nem de iptables. O pacote mantém essas ferramentas como recomendações porque elas habilitam capacidades específicas, mas TLS, configuração, SQLite, bootstrap, systemd e o endpoint de health funcionam sem elas.

## Instalação base

```bash
sudo dpkg -i minion_<versao>_amd64.deb
```

Após a instalação, valide:

```bash
sudo systemctl is-active minion
curl -k https://127.0.0.1:9870/api/v1/health
sudo stat -c '%U:%G %a %n' /var/lib/minion/bootstrap-credentials.txt
```

O arquivo bootstrap deve pertencer a `root:root` e possuir modo `0600`. A API key não é exibida pelo `dpkg` nem pelo `postinst`.

## Capacidades opcionais

Sem `fail2ban`, os endpoints de status, verificação e desbloqueio relacionados ao Fail2Ban não conseguem consultar ou alterar jails. Instale a integração somente nos hosts que realmente utilizam essa capacidade:

```bash
sudo apt-get install fail2ban
```

Sem `iptables`, o coletor de regras de iptables não possui uma fonte local para consultar. Instale-o apenas quando esse backend de firewall for usado:

```bash
sudo apt-get install iptables
```

A ausência dessas ferramentas não habilita fallback, shell remoto nem execução livre de comandos. As capacidades dependentes devem retornar erro controlado e permanecer indisponíveis até o pacote correspondente ser instalado.

## Instalação com recomendações

Para instalar o pacote local e também resolver automaticamente as recomendações disponíveis:

```bash
sudo apt install ./minion_<versao>_amd64.deb
```

Esse fluxo é opcional. O contrato da instalação principal continua sendo que `dpkg -i` deixa o núcleo do agente configurado, ativo e pronto para pareamento quando as dependências obrigatórias declaradas já estão presentes.

## Upgrade e remoção

Upgrades e reinstalações não recriam a credencial bootstrap já consumida e não sobrescrevem configuração, TLS ou SQLite existentes. A remoção simples preserva os dados:

```bash
sudo dpkg --remove minion
```

A remoção completa e destrutiva exige purge explícito:

```bash
sudo dpkg --purge minion
```
