# API Documentation

Base URL:

```text
Local: http://localhost:8080
Production: your deployed backend URL
```

## Submit Lead

Submit a lead form. The backend will create or update the contact in HubSpot, ensure the company exists, associate the contact to the company, send an internal notification email, and return the HubSpot contact ID.

```http
POST /api/lead
Content-Type: application/json
```

### Request Body

```json
{
  "firstName": "John",
  "familyName": "Doe",
  "company": "Example Inc",
  "workEmail": "john@example.com",
  "owner": "90579791"
}
```

### Fields

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `firstName` | string | Yes | Lead first name. |
| `familyName` | string | Yes | Lead family name / last name. |
| `company` | string | Yes | Company name. |
| `workEmail` | string | Yes | Work email address. Must be a valid email. |
| `owner` | string | No | HubSpot owner ID. If provided, it is sent to HubSpot as `hubspot_owner_id`. |

### Validation Rules

- `firstName` cannot be empty.
- `familyName` cannot be empty.
- `company` cannot be empty.
- `workEmail` cannot be empty.
- `workEmail` must be a valid email address.

### Success Response

HTTP status:

```http
200 OK
```

Body:

```json
{
  "message": "lead submitted successfully",
  "hubspot_contact_id": "796484549850"
}
```

### Error Responses

Invalid JSON body:

```http
400 Bad Request
```

```json
{
  "error": "invalid JSON request body"
}
```

Validation error example:

```http
400 Bad Request
```

```json
{
  "error": "workEmail must be valid"
}
```

HubSpot failure example:

```http
502 Bad Gateway
```

```json
{
  "error": "failed to create contact in HubSpot"
}
```

Unexpected backend error:

```http
500 Internal Server Error
```

```json
{
  "error": "internal server error"
}
```

### Backend Processing Flow

```text
POST /api/lead
↓
Validate request body
↓
Extract company domain from workEmail
↓
Search HubSpot company by domain
↓
Create HubSpot company if not found
↓
Search HubSpot contact by workEmail
↓
Create or update HubSpot contact
↓
Associate contact to company as Primary Company
↓
Send internal notification email with Resend
↓
Return success response
```

### HubSpot Mapping

Contact properties:

| API Field | HubSpot Property |
| --- | --- |
| `firstName` | `firstname` |
| `familyName` | `lastname` |
| `workEmail` | `email` |
| `owner` | `hubspot_owner_id` |

Company behavior:

- The company domain is derived from `workEmail`.
- Example: `john@example.com` uses `example.com`.
- The backend searches for a HubSpot company with that domain.
- If no company exists, the backend creates one using `company` and the derived domain.

### Email Notification

After HubSpot succeeds, the backend sends an internal email through Resend.

If email sending fails:

- the error is logged
- the API request still returns success
- HubSpot changes are not rolled back

Notification email subject:

```text
New Lead Submitted - Example Inc
```

Notification email body contains:

```text
First Name
Family Name
Work Email
Company
HubSpot Contact ID
Submission Time
```

### Curl Example

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

### CORS

The backend handles browser preflight requests:

```http
OPTIONS /api/lead
```

Allowed origins are configured with:

```text
CORS_ALLOWED_ORIGINS
```

## List Accounts

Returns the current demo company's target-account read model for Account
Discovery. The company scope is server-side; clients cannot provide it.

```http
GET /api/accounts
```

The response is an object with an `items` array. There is no pagination or
server-side search/filtering in the current MVP.

```json
{
  "items": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "name": "Oracle",
      "industry": "Technology",
      "hq": "Austin, TX",
      "lifecycle": "SQL",
      "analysis": {
        "icpScore": 86.5,
        "icpFit": "High",
        "signalScore": 91,
        "resonanceScore": 89.2,
        "tier": "Focus Accounts",
        "nextBestAction": "Prepare signal-led outreach"
      },
      "activeSignalCount": 14
    }
  ]
}
```

`industry`, `hq`, and `analysis` may be `null`. `activeSignalCount` is derived
from active signals and is `0` when none exist. The default order is Focus
Accounts, Tier 1, Tier 2, Below ICP, then unanalyzed accounts; within a tier,
resonance score descends and name is alphabetical.

When `analysis` exists, only the `action` value from `next_best_action` is
returned as `nextBestAction`. Malformed persisted analysis data produces the
normal safe HTTP 500 response.

## Account Overview

Returns one scoped target-account overview, including stable account facts and
the latest account analysis. `accountId` must be a PostgreSQL UUID. The account
must belong to the server-side demo company scope.

```http
GET /api/accounts/{accountId}
```

Example response:

```json
{
  "id": "00000000-0000-0000-0000-000000000101",
  "name": "Focus Account",
  "domain": "focus.example",
  "webUrl": "https://focus.example",
  "industry": "Technology",
  "hq": "Austin, TX",
  "employees": 1200,
  "revenue": { "amountM": 450.5, "currency": "USD" },
  "founded": 2012,
  "description": "Synthetic demo account for overview verification.",
  "lifecycle": "SQL",
  "analysis": {
    "icpScore": 86.5,
    "icpFit": "High",
    "signalScore": 91,
    "resonanceScore": 89.2,
    "tier": "Focus Accounts",
    "whyThisAccount": "Strong synthetic enterprise fit.",
    "whyNow": "Multiple current demo buying signals.",
    "nextBestAction": {
      "action": "Prepare signal-led outreach",
      "rationale": "Current demo evidence supports immediate outreach.",
      "timeWindow": "Now",
      "priority": "Critical"
    }
  }
}
```

Nullable account facts remain `null`, including `industry`, `hq`, `employees`,
`revenue`, `founded`, and `description`. An account without analysis still
returns HTTP 200 with `analysis: null`. Revenue uses the stored `revenue_m`
value as `amountM` without display conversion. The latest analysis is selected
by `created_at DESC, id DESC`.

Errors:

- Invalid UUID: `400 Bad Request`, `{"error":"accountId must be a valid UUID"}`
- Missing or out-of-scope account: `404 Not Found`, `{"error":"account not found"}`
- Database or malformed persisted analysis data: safe `500 Internal Server Error`

The existing `GET /api/accounts` remains the compact list read model; its
`analysis.nextBestAction` remains a string or `null`, while this detail endpoint
returns the complete supported object.

## Communication DNA

Returns the latest persisted Communication DNA for a scoped Account. DNA is
versioned and selected by `created_at DESC, id DESC`.

```http
GET /api/accounts/{accountId}/communication-dna
```

The response is always wrapped as `{ "data": ... }`. An existing account with
no DNA returns `200` and `{ "data": null }`. A missing or out-of-scope account
returns `404` with `{"error":"account not found"}`. Invalid UUIDs return `400`.
Malformed JSONB shapes, invalid evidence statuses, or source integrity errors
return the safe `500` response.

DNA uses typed camelCase fields including `tone`, `vocabulary`,
`valuePropositions`, `problemFraming`, `proofStyle`, `ctaPatterns`,
`recurringPhrases`, `doRules`, and `dontRules`. Evidence statuses are exactly
`SOURCE_BACKED`, `DERIVED`, or `INSUFFICIENT_DATA`. `buyingSignalSources` is
derived from same-account Signal to Source Document links, deduplicated by URL;
source and internal database IDs are not exposed.

## Account Signals

Returns all persisted Signals for a scoped Account, including inactive Signals.
Source Documents are embedded only as traceability data; there is no public
generic Source Document endpoint.

```http
GET /api/accounts/{accountId}/signals
```

The `accountId` path parameter must be a PostgreSQL UUID and is scoped by the
server-side demo company. Example response:

```json
{
  "summary": {
    "total": 3,
    "active": 2,
    "byType": { "Job Posting": 1, "News & Events": 1, "Company Data": 1 }
  },
  "items": [
    {
      "id": "00000000-0000-0000-0000-000000002001",
      "type": "Job Posting",
      "title": "Hiring enterprise solutions engineers",
      "body": "The account is expanding its enterprise solutions team.",
      "strength": "high",
      "relevance": "High",
      "signalDate": "2026-08-01",
      "signalDateRaw": null,
      "freshnessLabel": "Recent",
      "evidenceStatus": "SOURCE_BACKED",
      "verified": true,
      "isActive": true,
      "scoreEligible": true,
      "source": {
        "name": "Demo Careers",
        "type": "Career Page",
        "url": "https://focus.example/careers"
      }
    }
  ]
}
```

`strength` preserves `high`, `medium`, or `low`. `evidenceStatus` preserves
`SOURCE_BACKED`, `DERIVED`, or `INSUFFICIENT_DATA`. Optional signal fields and
`source` remain `null`; `verified` belongs to the Signal, not its source.
`signalDate` is date-only (`YYYY-MM-DD`).

An existing account with no Signals returns HTTP 200 with `items: []`,
`summary.total: 0`, `summary.active: 0`, and `summary.byType: {}`. Invalid UUID
returns HTTP 400. Missing or out-of-scope accounts return HTTP 404 with
`{"error":"account not found"}`. Database or cross-account source integrity
errors return the normal safe HTTP 500 response.
