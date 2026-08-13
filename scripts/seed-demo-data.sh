#!/usr/bin/env bash
set -euo pipefail

# Explicit local/demo seed. This is never part of Liquibase or Railway startup.
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

if [[ -f ".env.local" ]]; then
  ENV_FILE=".env.local"
elif [[ -f ".env" ]]; then
  ENV_FILE=".env"
else
  echo "Missing .env.local or .env"
  exit 1
fi

while IFS= read -r line || [[ -n "$line" ]]; do
  line="${line#${line%%[![:space:]]*}}"
  line="${line%${line##*[![:space:]]}}"
  [[ -z "$line" || "$line" == \#* || "$line" != *=* ]] && continue
  key="${line%%=*}"
  value="${line#*=}"
  value="${value#\"}"; value="${value%\"}"
  value="${value#\'}"; value="${value%\'}"
  case "$key" in
    DATABASE_URL|POSTGRES_DB|POSTGRES_USER|POSTGRES_PASSWORD|DEMO_COMPANY_PROFILE_ID)
      export "$key=$value"
      ;;
  esac
done < "$ENV_FILE"

: "${DATABASE_URL:?DATABASE_URL is required}"
: "${POSTGRES_USER:?POSTGRES_USER is required}"
: "${POSTGRES_PASSWORD:?POSTGRES_PASSWORD is required}"
: "${POSTGRES_DB:?POSTGRES_DB is required}"
: "${DEMO_COMPANY_PROFILE_ID:?DEMO_COMPANY_PROFILE_ID is required}"

if [[ "$DEMO_COMPANY_PROFILE_ID" != "00000000-0000-0000-0000-000000000001" ]]; then
  echo "DEMO_COMPANY_PROFILE_ID must be the fixed demo UUID for this seed"
  exit 1
fi

export PGUSER="$POSTGRES_USER"
export PGPASSWORD="$POSTGRES_PASSWORD"
export PGDATABASE="$POSTGRES_DB"

echo "Seeding explicit demo data into the configured PostgreSQL database"
psql "$DATABASE_URL" --set ON_ERROR_STOP=1 --file db/seed/demo.sql
echo "Demo seed completed"
