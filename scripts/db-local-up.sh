#!/usr/bin/env bash
set -euo pipefail

# This script is only for local development.
# Railway must not run this script, because Railway provides PostgreSQL as a managed service.

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

ENV_FILE=".env"
if [[ -f ".env.local" ]]; then
  ENV_FILE=".env.local"
elif [[ ! -f ".env" ]]; then
  echo "Missing .env.local or .env"
  echo "Create one from .env.example before starting local PostgreSQL."
  exit 1
fi

# Load the file so log messages can show the selected local port.
set -a
source "$ENV_FILE"
set +a

echo "Starting local PostgreSQL with docker compose"
echo "Using environment file: $ENV_FILE"

# docker-compose.yml uses POSTGRES_DB, POSTGRES_USER, POSTGRES_PASSWORD, and POSTGRES_PORT.
docker compose --env-file "$ENV_FILE" up -d postgres

echo "Local PostgreSQL is starting on port ${POSTGRES_PORT:-5432}"
