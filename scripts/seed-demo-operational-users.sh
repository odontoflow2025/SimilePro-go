#!/usr/bin/env bash
set -euo pipefail

if [[ "${EUID:-$(id -u)}" -ne 0 ]]; then
  echo 'Abra um shell administrativo primeiro com: sudo -s' >&2
  exit 1
fi
if [[ ! -f .env ]]; then
  echo 'Execute este script da raiz de SimilePro-go, onde está o .env do backend.' >&2
  exit 1
fi
if ! docker network inspect similepro_global_net >/dev/null 2>&1; then
  echo 'A rede similepro_global_net não existe. Inicie primeiro o ambiente Docker Simile Pro.' >&2
  exit 1
fi
if ! docker ps --filter name=similepro_postgres --filter status=running --format '{{.Names}}' | grep -qx similepro_postgres; then
  echo 'O container similepro_postgres não está em execução.' >&2
  exit 1
fi

read -r -s -p 'Senha para atendente.demo@similepro.invalid: ' SEED_DEMO_RECEPTIONIST_PASSWORD; printf '\n'
read -r -s -p 'Senha para tecnico.demo@similepro.invalid: ' SEED_DEMO_TECHNICIAN_PASSWORD; printf '\n'
read -r -s -p 'Senha para dentista.operacional@similepro.invalid: ' SEED_DEMO_DENTIST_PASSWORD; printf '\n'
export SEED_DEMO_RECEPTIONIST_PASSWORD SEED_DEMO_TECHNICIAN_PASSWORD SEED_DEMO_DENTIST_PASSWORD
trap 'unset SEED_DEMO_RECEPTIONIST_PASSWORD SEED_DEMO_TECHNICIAN_PASSWORD SEED_DEMO_DENTIST_PASSWORD' EXIT

docker run --rm \
  --network similepro_global_net \
  -v "$PWD:/app" \
  -w /app \
  -e SEED_DEMO_RECEPTIONIST_PASSWORD \
  -e SEED_DEMO_TECHNICIAN_PASSWORD \
  -e SEED_DEMO_DENTIST_PASSWORD \
  golang:1.25-alpine \
  sh -c 'apk add --no-cache git >/dev/null && go run ./cmd/seed-similepro-demo -operational-users-only -apply -confirm SIMILE-PRO-DEMO'
