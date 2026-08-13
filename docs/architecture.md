# Architecture

## Current system

```text
Frontend
   ↓
Gin HTTP API
   ↓
Handler
   ↓
Lead service
   ├── HubSpot client
   └── Resend email service
```

The current request path is `POST /api/lead`. The service validates the lead,
ensures the HubSpot company exists, creates or updates the contact, associates
the contact with the company, and sends an internal notification.

## Responsibilities

| Area | Responsibility |
| --- | --- |
| `cmd/api` | Loads configuration, wires dependencies, and starts the HTTP server. |
| `config` | Reads environment-backed runtime configuration. |
| `routes` | Registers HTTP routes. |
| `internal/handler` | Handles JSON binding, validation errors, and HTTP responses. |
| `internal/service` | Orchestrates lead business behavior and error policy. |
| `internal/hubspot` | Owns HubSpot HTTP requests and response decoding. |
| `internal/email` | Owns Resend notification requests behind `EmailService`. |
| `internal/middleware` | Provides CORS and centralized error handling. |
| `internal/models` | Defines request and response contracts. |
| `db/changelog` | Defines database schema changes through Liquibase. |
| `scripts` | Provides local database, migration, and Railway startup workflows. |

## Boundaries

Handlers should deal with HTTP concerns. Services should coordinate business
operations. Integration packages should own external API details. Liquibase is
the source of truth for database schema evolution.

Lead submissions are not persisted to PostgreSQL. The ABM read path uses an
AccountRepository and supports both list and detail reads:

```text
GET /api/accounts
GET /api/accounts/{accountId}
GET /api/accounts/{accountId}/signals
GET /api/accounts/{accountId}/communication-dna
   ↓
Account handler → Account service → Account repository → PostgreSQL
```

The Signals read model composes optional `source_document` traceability data
through the Account repository. There is no generic Signal or Source Document
CRUD API.

Communication DNA reads the latest `communication_dna` row and composes
Buying Signal source attribution from the existing Signal/Source Document data.
No AI generation service is implemented by this read path.

## Future / Not Yet Implemented

`internal/placeholder` contains interfaces reserved for future persistence,
background jobs, AI analysis, report generation, authentication, and other
capabilities. These are placeholders only.
