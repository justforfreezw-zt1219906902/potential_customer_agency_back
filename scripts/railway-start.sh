#!/usr/bin/env bash
set -euo pipefail

# This is the only script Railway should execute.
# It runs Liquibase first. If migration fails, deployment stops.
# If migration succeeds, the Go backend starts.

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

echo "Railway start: running database migration"
"$ROOT_DIR/scripts/liquibase-migrate.sh"

echo "Railway start: starting Go backend"
go run ./cmd/api
