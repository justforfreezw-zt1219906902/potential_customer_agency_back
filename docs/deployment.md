# Deployment

## Local development

Copy `.env.example` to `.env`, set credentials, then run:

```bash
./scripts/db-local-up.sh
./scripts/liquibase-migrate.sh
./scripts/seed-demo-data.sh
go run ./cmd/api
```

The first script manually starts PostgreSQL through `docker-compose.yml`.
It is for local development only and must not run on Railway. The migration
script uses the database address from `DATABASE_URL` and credentials/database
name from the PostgreSQL environment variables.

## Railway

Railway should build using the project `Dockerfile`. The image includes Go,
Java 17, Liquibase, the PostgreSQL JDBC driver, the changelog, and startup
scripts.

```text
Railway
   ↓
scripts/railway-start.sh
   ↓
Liquibase migration
   ↓
success? ── no → startup fails
   │
  yes
   ↓
go run ./cmd/api
```

Only `scripts/railway-start.sh` is the Railway start command. It never starts
the local Docker PostgreSQL container.

## Environment variables

| Variable | Purpose |
| --- | --- |
| `PORT` | HTTP listening port. |
| `HUBSPOT_ACCESS_TOKEN` | HubSpot API authentication. |
| `HUBSPOT_OWNER_ID` | Required HubSpot owner assigned to contacts submitted through `POST /api/lead`. |
| `CORS_ALLOWED_ORIGINS` | Allowed browser origins, comma-separated. |
| `RESEND_API_KEY` | Resend API authentication. |
| `RESEND_FROM_EMAIL` | Resend sender address. |
| `NOTIFICATION_EMAILS` | Comma-separated or bracketed recipient list. |
| `DATABASE_URL` | PostgreSQL address, including Railway host and optional database/query. |
| `POSTGRES_DB` / `PGDATABASE` | Database name when not present in the address. |
| `POSTGRES_USER` / `PGUSER` | Database username. |
| `POSTGRES_PASSWORD` / `PGPASSWORD` | Database password. |
| `POSTGRES_PORT` | Local Docker host port. |
| `POSTGRES_SSLMODE` | Local/remote PostgreSQL SSL mode when included in the address. |
| `DEMO_COMPANY_PROFILE_ID` | Required server-side UUID scope for the demo account list. |
| `GEMINI_API_KEY` | Optional Google Gemini API key for outreach draft generation. |
| `GEMINI_MODEL` | Gemini model name; defaults to `gemini-3.6-flash`. |
| `OUTREACH_REQUEST_TIMEOUT_SECONDS` | Positive whole-second limit for the complete outreach request; defaults to `30`. |

Use placeholders only in `.env.example`; never commit actual tokens or
passwords.

Railway does not run the demo seed. Set `DEMO_COMPANY_PROFILE_ID` only when
the corresponding company profile exists in the target database.

The outreach timeout covers the complete handler-to-database-to-Gemini flow.
Invalid, zero, or negative values stop configuration loading. The HTTP server
write timeout is fixed at 40 seconds so the default 30-second outreach deadline
can return a controlled API response before the connection-level ceiling.
