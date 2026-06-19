#!/usr/bin/env bash
set -euo pipefail

# This script runs database migrations with Liquibase.
# It works locally and on Railway.
# Railway should call this script through scripts/railway-start.sh only.

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

CHANGELOG_FILE="db/changelog/db.changelog-master.yaml"

if [[ -f ".env.local" ]]; then
  echo "Loading local environment from .env.local"
  set -a
  source ".env.local"
  set +a
elif [[ -f ".env" ]]; then
  echo "Loading local environment from .env"
  set -a
  source ".env"
  set +a
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
: "${POSTGRES_USER:?POSTGRES_USER is required}"
: "${POSTGRES_PASSWORD:?POSTGRES_PASSWORD is required}"

DB_USERNAME="$POSTGRES_USER"
DB_PASSWORD="$POSTGRES_PASSWORD"

echo "Using DATABASE_URL for database address"
echo "Using POSTGRES_USER and POSTGRES_PASSWORD for credentials"

if [[ "$DATABASE_URL" =~ ^jdbc:postgresql://(.+)$ ]]; then
  JDBC_URL="$DATABASE_URL"
elif [[ "$DATABASE_URL" =~ ^postgres(ql)?://([^@/]+@)?([^:/?]+):?([0-9]*)/([^?]+)(\?(.*))?$ ]]; then
  DB_HOST="${BASH_REMATCH[3]}"
  DB_PORT="${BASH_REMATCH[4]:-5432}"
  DB_NAME="${BASH_REMATCH[5]}"
  DB_QUERY="${BASH_REMATCH[7]:-}"

  JDBC_URL="jdbc:postgresql://${DB_HOST}:${DB_PORT}/${DB_NAME}"
  if [[ -n "$DB_QUERY" ]]; then
    JDBC_URL="${JDBC_URL}?${DB_QUERY}"
  fi
else
  echo "DATABASE_URL must be a database address, for example:"
  echo "  postgres://localhost:5432/potential_customer?sslmode=disable"
  echo "  jdbc:postgresql://localhost:5432/potential_customer?sslmode=disable"
  exit 1
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
