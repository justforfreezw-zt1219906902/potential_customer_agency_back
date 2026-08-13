# Integrations

## HubSpot

HubSpot is the lead system of record for the current submission flow. The
client authenticates with `HUBSPOT_ACCESS_TOKEN` and uses the HubSpot CRM APIs.

For each request, the service:

1. derives a company domain from `workEmail`;
2. searches companies by domain and creates the company when absent;
3. searches contacts by email;
4. creates or updates the contact;
5. associates the contact with the company as the Primary Company.

The request fields map as follows:

| API field | HubSpot property |
| --- | --- |
| `firstName` | `firstname` |
| `familyName` | `lastname` |
| `workEmail` | `email` |
| `owner` | `hubspot_owner_id` |

HubSpot failures are returned as an external-service error, normally HTTP 502.
The contact/company operation is not rolled back by this application.

## Resend

Resend sends an internal notification after HubSpot succeeds. It uses
`RESEND_API_KEY`, `RESEND_FROM_EMAIL`, and `NOTIFICATION_EMAILS`.

The subject is `New Lead Submitted - <company>`. The body includes the lead
fields, HubSpot contact ID, and submission time. `RESEND_FROM_EMAIL` must be a
valid sender for the Resend account: use the Resend testing sender for local
sandbox testing, or a sender on a verified domain in production.

If Resend fails, the error is logged and the API still returns HTTP 200 because
the HubSpot operation already succeeded.

## Future Integrations

No other external integration is currently implemented.
