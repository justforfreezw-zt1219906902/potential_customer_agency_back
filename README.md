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

2. Copy the example environment file:

```bash
cp .env.example .env
```

3. Set your HubSpot private app access token in `.env`:

```bash
HUBSPOT_ACCESS_TOKEN=pat-na1-your-token-here
CORS_ALLOWED_ORIGINS=*
RESEND_API_KEY=re_xxxxxxxxx
NOTIFICATION_EMAILS=['sales@mi-goto.com','email2','email3']
```

4. Install dependencies:

```bash
go mod tidy
```

5. Start the API:

```bash
go run ./cmd/api
```

The server starts on `http://localhost:8080` by default.

## Deployment Notes

Set these environment variables on your deployment platform:

```bash
PORT=8080
HUBSPOT_ACCESS_TOKEN=pat-na1-your-token-here
CORS_ALLOWED_ORIGINS=https://your-frontend-domain.com
RESEND_API_KEY=re_xxxxxxxxx
NOTIFICATION_EMAILS=['sales@mi-goto.com','email2','email3']
```

For multiple frontend domains, separate them with commas:

```bash
CORS_ALLOWED_ORIGINS=https://app.example.com,https://www.example.com
```

The backend handles browser `OPTIONS` preflight requests in code, so it does not depend on platform-specific CORS settings.

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
  "name": "John Doe",
  "email": "john@example.com",
  "company": "Example Inc",
  "website": "https://example.com",
  "phoneNumber": "+1 555 0100",
  "owner": "123456"
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
  "error": "email must be valid"
}
```

## Test With curl

```bash
curl -X POST http://localhost:8080/api/lead \
  -H "Content-Type: application/json" \
  -d '{
    "name": "John Doe",
    "email": "john@example.com",
  "company": "Example Inc",
  "website": "https://example.com",
  "phoneNumber": "+1 555 0100",
  "owner": "123456"
  }'
```

The `website` domain is used to find an existing HubSpot company. If no company exists for that domain, the API creates one first, then creates or updates the contact and associates it to the company through the legacy association API.

## Future Placeholders

The `internal/placeholder` package contains simple interfaces for future features:

- PostgreSQL
- Background jobs
- AI analysis service
- Report generation
- Authentication
- Email notifications

These are intentionally not implemented yet.
