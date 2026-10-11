# Seed de demonstração Simile Pro

Cria uma clínica fictícia, a conta administrativa solicitada, uma dentista, 30
pacientes, agendamentos em dias úteis e receitas para atendimentos concluídos.
O intervalo padrão vai de `2020-01-01` até ontem, considerando a data local.

Execute a partir da raiz do backend, onde está o `.env` com as configurações do
banco e a chave de criptografia usada pelos modelos:

```bash
go run ./cmd/seed-similepro-demo
```

Esse modo só mostra a prévia e não conecta nem grava no banco. Para aplicar,
informe a senha administrativa sem colocá-la no histórico do shell:

```bash
read -rsp 'Senha administrativa: ' SEED_ADMIN_PASSWORD
printf '\n'
export SEED_ADMIN_PASSWORD
go run ./cmd/seed-similepro-demo -apply -confirm SIMILE-PRO-DEMO
unset SEED_ADMIN_PASSWORD
```

O comando é idempotente para os registros identificados pelo seed. Ele usa uma
transação e aborta se o e-mail já pertencer a outra clínica ou se o identificador
reservado da clínica estiver ocupado. Datas podem ser informadas com `-start`
e `-end` no formato `AAAA-MM-DD`. Ele não executa migrações de schema.
