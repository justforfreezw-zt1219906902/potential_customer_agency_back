# Repository Guidance

This repository is a small Go/Gin backend for lead submission. Keep changes
focused and preserve the existing layered structure.

## Documentation navigation

- API behavior: read `docs/api.md`.
- Database schema or models: read `docs/database.md`,
  `docs/ABM_Intelligence_Database_Schema_Guide.md`, and `db/changelog/`.
- HubSpot, Resend, or other external systems: read `docs/integrations.md`.
- Docker, Railway, Liquibase, scripts, or environment configuration: read
  `docs/deployment.md`.
- Package boundaries or execution flow: read `docs/architecture.md`.
- Durable or non-obvious architectural behavior: read `docs/decisions.md`.

## Working rules

1. Read only the documentation relevant to the task, then inspect the actual
   implementation before editing documentation or code.
2. Make the smallest change necessary and preserve useful human-written text.
3. Do not rewrite all documentation for a narrow change.
4. Do not silently resolve meaningful conflicts between code and documentation;
   report them.
5. Clearly distinguish current behavior from future or planned functionality.
6. Database schema changes must use version-controlled Liquibase changesets.
7. Go application code must not create or alter production database schema.
8. Do not add authentication, jobs, AI, reports, notifications, or persistence
   behavior unless explicitly requested and implemented.
9. Do not add secrets to source code or documentation.

## Validation

For applicable coding tasks, use `go test ./...` and `go build ./...` as the
normal lightweight validation baseline. Database changes should also validate
the relevant Liquibase changelog and migration path.

## Documentation Impact Check

Every coding task must finish with:

```text
Documentation Impact:
- API: Yes / No
- Database: Yes / No
- Integration: Yes / No
- Deployment: Yes / No
- Architecture: Yes / No
- Long-term Decision: Yes / No
```

Documentation updates are required when a public contract, schema, external
integration, deployment workflow, package boundary, or durable decision changes.
Routine refactoring, formatting, private renaming, logging-only changes, and
implementation-only bug fixes do not require documentation changes.
