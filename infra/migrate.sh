#!/usr/bin/env bash
set -euo pipefail

# Usage:
#   DATABASE_URL="postgres://..." ./infra/migrate.sh up
#   DATABASE_URL="postgres://..." ./infra/migrate.sh down 1
#
# Notes:
# - Uses dockerized golang-migrate so you don't need it installed locally.
# - Works for local Postgres or Supabase Postgres.

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
MIGRATIONS_DIR="${ROOT_DIR}/api/migrations"

if [[ -z "${DATABASE_URL:-}" ]]; then
  echo "DATABASE_URL is required" >&2
  exit 1
fi

CMD="${1:-up}"
N="${2:-}"

IMAGE="migrate/migrate:v4.18.2"

if [[ "${CMD}" == "up" ]]; then
  docker run --rm -v "${MIGRATIONS_DIR}:/migrations" "${IMAGE}" -path=/migrations -database "${DATABASE_URL}" up
  exit 0
fi

if [[ "${CMD}" == "down" ]]; then
  if [[ -z "${N}" ]]; then
    echo "Usage: ./infra/migrate.sh down <n>" >&2
    exit 1
  fi
  docker run --rm -v "${MIGRATIONS_DIR}:/migrations" "${IMAGE}" -path=/migrations -database "${DATABASE_URL}" down "${N}"
  exit 0
fi

echo "Unknown command: ${CMD} (use up|down)" >&2
exit 1

