# Potential Customer Agency Backend

## Overview

A small Go/Gin backend for submitting leads to HubSpot and notifying an
internal team through Resend. It is structured as a foundation for a future
B2B SaaS application.

## Current Capabilities

- `POST /api/lead` accepts the short lead form.
- HubSpot company lookup/creation and contact creation/update.
- Primary Company association in HubSpot.
- Best-effort internal email notification through Resend.
- PostgreSQL-backed Account Intelligence read APIs.
- Account Discovery list, scoped Account Overview, and scoped Account Signals.
- Latest persisted Communication DNA with traceable buying-signal sources.
- Server-side demo-company scoping and an explicit demo seed workflow.
- PostgreSQL schema migrations through Liquibase.
- Local Docker PostgreSQL and Railway startup support.

## Technology Stack

Go 1.24+, Gin, PostgreSQL, Liquibase, Docker Compose, Railway, HubSpot CRM,
and Resend.

## Project Structure

```text
cmd/api        Application entry point
config         Environment-backed configuration
routes         HTTP route registration
internal       Handlers, services, integrations, models, middleware
db/changelog   Liquibase schema history
scripts        Local database, migration, and Railway startup scripts
docs           Contracts, boundaries, operations, and decisions
```

## Quick Start

```bash
cp .env.example .env
# Set HUBSPOT_ACCESS_TOKEN and local values in .env
./scripts/db-local-up.sh
./scripts/liquibase-migrate.sh
go run ./cmd/api
```

The local database container is started manually. The API listens on port
`8080` by default. See [Deployment](docs/deployment.md) for environment and
database details, and [API](docs/api.md) for the request contract.

## Production / Railway

Railway should use the project `Dockerfile` and run only
`./scripts/railway-start.sh`. The script runs Liquibase first and starts the Go
backend only after migration succeeds. Do not run the local Docker database
script on Railway.

## Documentation

- [Architecture](docs/architecture.md)
- [API](docs/api.md)
- [Database](docs/database.md)
- [Integrations](docs/integrations.md)
- [Deployment](docs/deployment.md)
- [Decisions](docs/decisions.md)
- [Detailed ABM schema guide](docs/ABM_Intelligence_Database_Schema_Guide.md)
