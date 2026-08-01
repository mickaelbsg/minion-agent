# Revogação definitiva de clientes

Use a revogação quando uma API key for comprometida, um integrador for desativado ou o acesso precisar ser encerrado de forma permanente sem apagar o histórico do cliente.

```bash
sudo minion client revoke automation
```

A operação exige root, desabilita o cliente, substitui o hash da chave atual por um novo hash Argon2id de um segredo aleatório descartado e registra `revoked_at` em UTC. A API key anterior deixa de funcionar imediatamente. Nenhuma chave ou hash é impresso pelo comando.

O registro continua visível em:

```bash
sudo minion client list
```

Clientes revogados aparecem com status `revoked` e não podem ser habilitados, rotacionados ou ter sua expiração alterada. Essa restrição evita reativação acidental de uma identidade comprometida.

Para restabelecer uma integração, crie uma nova identidade com outro nome e armazene a nova chave no gerenciador de credenciais do Automation/n8n:

```bash
sudo minion client create --name automation-2026 --ips 192.0.2.10/32
```

Depois de confirmar que a auditoria antiga não é mais necessária, o registro revogado pode ser removido explicitamente com `sudo minion client delete <nome>`. A exclusão não é necessária para bloquear o acesso.
