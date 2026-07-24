# Potential Customer Agency Backend

A simple Go backend starter for a future B2B SaaS application.

The current service accepts lead submissions, ensures the company exists in HubSpot, creates or updates the contact, associates the contact with the company, and sends an internal notification email.

## Stack

- Go 1.24+
- Gin
- Layered project structure
- Environment-based configuration
- Centralized JSON error handling
- Basic request logging

## Project Structure

```text
.
├── cmd
│   └── api
│       └── main.go
├── config
│   └── config.go
├── internal
│   ├── errors
│   ├── handler
│   ├── hubspot
│   ├── middleware
│   ├── models
│   ├── placeholder
│   └── service
├── routes
│   └── routes.go
├── .env.example
├── go.mod
└── README.md
```

## Local Setup

1. Install Go 1.24 or newer.

2. Install Docker and Liquibase.

3. Copy the example environment file:

```bash
cp .env.example .env
```

4. Set your environment variables in `.env`:

```bash
PORT=8080
HUBSPOT_ACCESS_TOKEN=pat-na1-your-token-here
CORS_ALLOWED_ORIGINS=*
RESEND_API_KEY=re_xxxxxxxxx
RESEND_FROM_EMAIL="Mi Goto <onboarding@resend.dev>"
NOTIFICATION_EMAILS=['sales@mi-goto.com','email2','email3']
POSTGRES_DB=potential_customer
POSTGRES_USER=postgres
POSTGRES_PASSWORD=postgres
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_SSLMODE=disable
DATABASE_URL=postgres://localhost:5432?sslmode=disable
```

`DATABASE_URL` should contain only the database address. The Liquibase script reads credentials from `POSTGRES_USER` / `PGUSER`, reads the password from `POSTGRES_PASSWORD` / `PGPASSWORD`, and reads the database name from `POSTGRES_DB` / `PGDATABASE`.

5. Install Go dependencies:

```bash
go mod tidy
```

6. Start local PostgreSQL manually:

```bash
./scripts/db-local-up.sh
```

7. Run Liquibase migrations:

```bash
./scripts/liquibase-migrate.sh
```

8. Start the API:

```bash
go run ./cmd/api
```

The server starts on `http://localhost:8080` by default.

## Deployment Notes

Set these environment variables on Railway:

```bash
PORT=8080
HUBSPOT_ACCESS_TOKEN=pat-na1-your-token-here
CORS_ALLOWED_ORIGINS=https://your-frontend-domain.com
RESEND_API_KEY=re_xxxxxxxxx
RESEND_FROM_EMAIL="Mi Goto <hello@mi-goto.com>"
NOTIFICATION_EMAILS=['sales@mi-goto.com','email2','email3']
POSTGRES_USER=postgres
POSTGRES_PASSWORD=your-railway-password
POSTGRES_DB=your-railway-database
DATABASE_URL=postgres://your-railway-host:5432?sslmode=require
```

For multiple frontend domains, separate them with commas:

```bash
CORS_ALLOWED_ORIGINS=https://app.example.com,https://www.example.com
```

The backend handles browser `OPTIONS` preflight requests in code, so it does not depend on platform-specific CORS settings.

## PostgreSQL And Liquibase

Schema changes are handled by Liquibase only. The Go application should not create or alter database tables directly.

Changelog files live under:

```text
db/changelog/db.changelog-master.yaml
db/changelog/changes/001-create-leads-table.yaml
db/changelog/changes/002-update-leads-for-short-form.yaml
```

### Local Development

Local PostgreSQL is started manually by developers:

```bash
./scripts/db-local-up.sh
```

This script uses Docker Compose and reads database settings from `.env.local` if present, otherwise from `.env`.

Run migrations:

```bash
./scripts/liquibase-migrate.sh
```

Start the backend:

```bash
go run ./cmd/api
```

Expected local workflow:

```text
./scripts/db-local-up.sh
./scripts/liquibase-migrate.sh
go run ./cmd/api
```

### Railway Deployment

Railway should use the project `Dockerfile` instead of default Nixpacks.

The Docker image installs:

- Go runtime
- Liquibase CLI
- PostgreSQL JDBC driver
- application source code
- `db/changelog`
- `scripts`

The Dockerfile final command is:

```bash
./scripts/railway-start.sh
```

The Railway script does this:

```text
Run Liquibase migration
↓
If migration succeeds, start Go backend
↓
If migration fails, stop deployment
```

Do not run `scripts/db-local-up.sh` in Railway. Railway PostgreSQL should provide `DATABASE_URL`.

## Lead Flow

```text
POST /api/lead
↓
HubSpot contact created or updated
↓
Internal notification email sent with Resend
↓
Success response returned
```

If the internal email fails, the error is logged but the API still returns success after HubSpot succeeds.

## API

### Submit Lead

```http
POST /api/lead
Content-Type: application/json
```

Request body:

```json
{
  "firstName": "John",
  "familyName": "Doe",
  "company": "Example Inc",
  "workEmail": "john@example.com",
  "owner": "90579791"
}
```

Success response:

```json
{
  "message": "lead submitted successfully",
  "hubspot_contact_id": "123456789"
}
```

Error response:

```json
{
  "error": "workEmail must be valid"
}
```

## Test With curl

```bash
curl -X POST http://localhost:8080/api/lead \
  -H "Content-Type: application/json" \
  -d '{
    "firstName": "John",
    "familyName": "Doe",
    "company": "Example Inc",
    "workEmail": "john@example.com",
    "owner": "90579791"
  }'
```

The `workEmail` domain is used to find an existing HubSpot company. If no company exists for that domain, the API creates one first, then creates or updates the contact and associates it to the company through the legacy association API.

## Future Placeholders

The `internal/placeholder` package contains simple interfaces for future features:

- PostgreSQL
- Background jobs
- AI analysis service
- Report generation
- Authentication
- Email notifications

These are intentionally not implemented yet.
