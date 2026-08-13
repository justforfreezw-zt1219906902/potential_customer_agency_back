# Database

## Technology and ownership

The project uses PostgreSQL. Liquibase is the source of truth for schema
evolution. Production schema changes must be represented by version-controlled
changesets; the Go application must not create or alter tables at runtime.

The current Go request flow does not persist leads through a repository. The
ABM read path uses `pgxpool` and an explicit SQL account repository. The
Liquibase schema contains the existing `leads` history and the ABM Intelligence
MVP entities.

## Migration location

```text
db/changelog/db.changelog-master.yaml
db/changelog/changes/
```

The master changelog currently includes:

- `001-create-leads-table.yaml`
- `002-update-leads-for-short-form.yaml`
- `003-create-abm-intelligence-schema.yaml`

See [the ABM schema guide](ABM_Intelligence_Database_Schema_Guide.md) for
entity meaning, field semantics, relationships, and value rules. This file
describes what the model means; this document describes how it is managed.

## Local workflow

Local PostgreSQL is started manually by developers and is not part of Railway
deployment:

```text
scripts/db-local-up.sh
        ↓
scripts/liquibase-migrate.sh
        ↓
scripts/seed-demo-data.sh (explicit local/demo step)
        ↓
go run ./cmd/api
```

`db-local-up.sh` starts the Docker Compose PostgreSQL service. The migration
script reads the database address and credentials from the environment and
runs the master changelog.

`seed-demo-data.sh` is separate from Liquibase, idempotent, and intended only
for explicit local/demo data. It is never run automatically by application or
Railway startup.

The explicit Demo Seed also includes deterministic `source_document` and
Signal rows so Account Signal source traceability and empty/inactive Signal
states can be verified. The physical schema is unchanged.

## Railway workflow

Railway runs `scripts/railway-start.sh`. That script runs Liquibase first and
starts the Go service only when migration succeeds. It does not create a local
database container.

## Schema change rule

```text
Design schema change
        ↓
Add Liquibase changeset
        ↓
Update detailed schema documentation when semantics change
        ↓
Run migration locally
        ↓
Deploy through the normal deployment process
```
