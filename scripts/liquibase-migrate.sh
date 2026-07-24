#!/usr/bin/env bash
set -euo pipefail

# This script runs database migrations with Liquibase.
# It works locally and on Railway.
# Railway should call this script through scripts/railway-start.sh only.

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

CHANGELOG_FILE="db/changelog/db.changelog-master.yaml"

load_env_file() {
  local env_file="$1"

  while IFS= read -r line || [[ -n "$line" ]]; do
    line="${line#"${line%%[![:space:]]*}"}"
    line="${line%"${line##*[![:space:]]}"}"

    if [[ -z "$line" || "$line" == \#* || "$line" != *=* ]]; then
      continue
    fi

    local key="${line%%=*}"
    local value="${line#*=}"

    key="${key#"${key%%[![:space:]]*}"}"
    key="${key%"${key##*[![:space:]]}"}"
    value="${value#"${value%%[![:space:]]*}"}"
    value="${value%"${value##*[![:space:]]}"}"
    value="${value%\"}"
    value="${value#\"}"
    value="${value%\'}"
    value="${value#\'}"

    case "$key" in
      DATABASE_URL|POSTGRES_DB|POSTGRES_USER|POSTGRES_PASSWORD|POSTGRES_SSLMODE|PGDATABASE|PGUSER|PGPASSWORD)
        if [[ -z "${!key:-}" ]]; then
          export "$key=$value"
        fi
        ;;
    esac
  done < "$env_file"
}

if [[ -f ".env.local" ]]; then
  echo "Loading local environment from .env.local"
  load_env_file ".env.local"
elif [[ -f ".env" ]]; then
  echo "Loading local environment from .env"
  load_env_file ".env"
fi

if [[ ! -f "$CHANGELOG_FILE" ]]; then
  echo "Missing Liquibase changelog: $CHANGELOG_FILE"
  exit 1
fi

if ! command -v liquibase >/dev/null 2>&1; then
  echo "Liquibase CLI is required but was not found."
  echo "Install Liquibase locally, or make sure it is available in the Railway build/runtime image."
  exit 1
fi

DB_USERNAME=""
DB_PASSWORD=""
JDBC_URL=""

: "${DATABASE_URL:?DATABASE_URL is required}"

DB_USERNAME="${POSTGRES_USER:-${PGUSER:-}}"
DB_PASSWORD="${POSTGRES_PASSWORD:-${PGPASSWORD:-}}"
DB_NAME="${POSTGRES_DB:-${PGDATABASE:-}}"

if [[ -z "$DB_USERNAME" ]]; then
  echo "POSTGRES_USER or PGUSER is required"
  exit 1
fi

if [[ -z "$DB_PASSWORD" ]]; then
  echo "POSTGRES_PASSWORD or PGPASSWORD is required"
  exit 1
fi

echo "Using DATABASE_URL for database address"
echo "Using POSTGRES_USER/PGUSER and POSTGRES_PASSWORD/PGPASSWORD for credentials"
echo "Using POSTGRES_DB/PGDATABASE for database name when provided"

if [[ "$DATABASE_URL" =~ ^jdbc:postgresql://([^:/?]+):?([0-9]*)(/([^?]+))?(\?(.*))?$ ]]; then
  DB_HOST="${BASH_REMATCH[1]}"
  DB_PORT="${BASH_REMATCH[2]:-5432}"
  URL_DB_NAME="${BASH_REMATCH[4]:-}"
  DB_QUERY="${BASH_REMATCH[6]:-}"
elif [[ "$DATABASE_URL" =~ ^postgres(ql)?://([^@/]+@)?([^:/?]+):?([0-9]*)(/([^?]+))?(\?(.*))?$ ]]; then
  DB_HOST="${BASH_REMATCH[3]}"
  DB_PORT="${BASH_REMATCH[4]:-5432}"
  URL_DB_NAME="${BASH_REMATCH[6]:-}"
  DB_QUERY="${BASH_REMATCH[8]:-}"
else
  echo "DATABASE_URL must be a database address, for example:"
  echo "  postgres://localhost:5432?sslmode=disable"
  echo "  postgres://localhost:5432/potential_customer?sslmode=disable"
  echo "  jdbc:postgresql://localhost:5432?sslmode=disable"
  exit 1
fi

DB_NAME="${DB_NAME:-$URL_DB_NAME}"
if [[ -z "$DB_NAME" ]]; then
  echo "POSTGRES_DB or PGDATABASE is required when DATABASE_URL does not include a database name"
  exit 1
fi

JDBC_URL="jdbc:postgresql://${DB_HOST}:${DB_PORT}/${DB_NAME}"
if [[ -n "$DB_QUERY" ]]; then
  JDBC_URL="${JDBC_URL}?${DB_QUERY}"
fi

echo "Running Liquibase migration"
echo "Changelog: $CHANGELOG_FILE"
echo "JDBC URL: $JDBC_URL"

liquibase \
  --changeLogFile="$CHANGELOG_FILE" \
  --url="$JDBC_URL" \
  --username="$DB_USERNAME" \
  --password="$DB_PASSWORD" \
  update

echo "Liquibase migration completed successfully"
