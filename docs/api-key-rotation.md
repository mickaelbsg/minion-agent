# Rotação de API key

A rotação substitui a credencial de um cliente existente sem alterar seu nome, IPs/CIDRs autorizados ou estado habilitado/desabilitado.

## Quando usar

- suspeita de vazamento da chave;
- troca periódica de credenciais;
- mudança do secret store utilizado pelo Automation/n8n;
- desligamento de uma integração que conhecia a chave anterior.

## Comando

Execute como root no servidor onde o Minion está instalado:

```bash
sudo minion client rotate automation
```

Saída esperada:

```text
Client: automation
New API Key: minion_sk_...
The previous API key is now invalid. Update the credential in Automation/n8n immediately; this key will not be shown again.
```

A nova chave é mostrada uma única vez. O Minion armazena somente o hash Argon2id no SQLite.

## Procedimento seguro

1. Prepare a credencial do Minion no secret store do Automation/n8n.
2. Execute a rotação no servidor.
3. Copie a nova chave diretamente para a credencial protegida do Automation/n8n.
4. Atualize ou substitua a credencial usada pela integração.
5. Faça uma chamada autenticada, por exemplo ao endpoint `/api/v1/heartbeat`.
6. Remova qualquer cópia temporária da chave.

A chave anterior deixa de funcionar imediatamente. Portanto, existe uma pequena interrupção entre a execução do comando e a atualização da credencial no Automation/n8n.

## Comportamento preservado

A rotação não altera:

- nome do cliente;
- IPs e CIDRs autorizados;
- estado habilitado ou desabilitado;
- certificado TLS;
- identidade do agente;
- registros de auditoria existentes.

Rotacionar um cliente desabilitado não o habilita automaticamente.

## Recuperação

Caso a nova chave seja perdida antes de ser cadastrada no Automation/n8n, execute novamente:

```bash
sudo minion client rotate automation
```

A chave perdida será invalidada e uma nova será emitida.

Não registre API keys em scripts, issues, logs, histórico de comandos ou arquivos compartilhados. Use uma credencial nativa ou secret store no Automation/n8n.
