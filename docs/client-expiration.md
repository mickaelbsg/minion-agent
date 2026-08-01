# Expiração opcional de clientes

Clientes do Minion não expiram por padrão. Isso preserva o comportamento das instalações existentes e evita interrupções inesperadas na integração com Automation/n8n.

Use expiração para acessos temporários, homologação, migrações ou integrações que precisam deixar de funcionar automaticamente em uma data definida.

## Definir expiração

Execute como root e informe a data no formato RFC3339:

```bash
sudo minion client expire automation 2026-08-31T23:59:59Z
```

O instante é normalizado e armazenado em UTC. Após esse momento, a API key continua armazenada somente como hash, mas deixa de autenticar. Não é necessário reiniciar o serviço.

## Consultar

```bash
sudo minion client list
```

A coluna `EXPIRES AT` mostra a data configurada ou `never`.

## Remover expiração

```bash
sudo minion client expire automation never
```

Isso não troca a API key e não altera os IPs/CIDRs permitidos. Caso o cliente esteja manualmente desabilitado, ele continua desabilitado.

## Operação com Automation/n8n

Antes de definir uma validade, confirme que existe outro cliente administrativo funcional ou um procedimento de recuperação local. Para renovar o acesso sem trocar a credencial, defina uma nova data futura. Para trocar também a credencial, use:

```bash
sudo minion client rotate automation
```

Atualize a credencial protegida no Automation/n8n imediatamente após a rotação. Nunca armazene a API key em workflow, Code node, URL, issue ou log.

## Compatibilidade

Durante a abertura do banco, o Minion adiciona automaticamente a coluna nullable `expires_at` em bancos criados por versões anteriores. Clientes existentes recebem validade indefinida e continuam funcionando normalmente.
