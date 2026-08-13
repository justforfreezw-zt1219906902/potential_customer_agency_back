# Decisions

## DEC-001 — Liquibase owns schema changes

Status: Accepted

Decision: Database schema changes are version-controlled Liquibase changesets.

Reason: Migrations need a reviewable, repeatable history across local and
Railway environments.

Impact: Do not add runtime `CREATE TABLE` or `ALTER TABLE` behavior to the Go
application.

## DEC-002 — Local PostgreSQL startup is manual

Status: Accepted

Decision: Developers start the local PostgreSQL Docker service explicitly with
`scripts/db-local-up.sh`.

Reason: Railway provides managed PostgreSQL and must not create a local
container during deployment.

Impact: Local database startup and Railway startup remain separate workflows.

## DEC-003 — Railway migrates before starting Go

Status: Accepted

Decision: `scripts/railway-start.sh` runs Liquibase before starting the API.

Reason: The application should start only against the expected schema.

Impact: A failed migration stops startup and prevents the backend from running.

## DEC-004 — Notification failure does not undo HubSpot success

Status: Accepted

Decision: Resend notification errors are logged but do not fail a lead request
after HubSpot succeeds.

Reason: Internal notification is secondary to the lead operation and there is
no queue or retry worker in the current scope.

Impact: A successful response does not guarantee that the notification email
was delivered; logs are the current diagnostic path.

## DEC-005 — Explicit SQL through pgxpool

Status: Accepted

Decision: Runtime PostgreSQL access uses `pgxpool` and explicit SQL in focused
repositories.

Reason: The current read slice needs predictable queries and does not justify
an ORM, query generator, or generic persistence abstraction.

Impact: Repository queries own SQL and scanning; services do not access SQL.

## DEC-006 — Demo data is separate from migrations

Status: Accepted

Decision: Demonstration rows are created only by the explicit seed script.

Reason: Schema history and optional development data have different lifecycles.

Impact: Liquibase and Railway startup never seed demo records automatically.
