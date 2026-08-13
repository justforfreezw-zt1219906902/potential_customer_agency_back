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
