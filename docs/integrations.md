# Integrations

## HubSpot

HubSpot is the lead system of record for the current submission flow. The
client authenticates with `HUBSPOT_ACCESS_TOKEN` and uses the HubSpot CRM APIs.
Contact ownership is configured by the backend through the required
`HUBSPOT_OWNER_ID` value.

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
| Backend `HUBSPOT_OWNER_ID` | `hubspot_owner_id` |

The configured owner is applied during both contact creation and contact
update. Public lead requests cannot select or override it. Optional lead
`context` is deliberately not mapped to a guessed HubSpot custom property.

HubSpot failures are returned as an external-service error, normally HTTP 502.
The contact/company operation is not rolled back by this application.

## Resend

Resend sends an internal notification after HubSpot succeeds. It uses
`RESEND_API_KEY`, `RESEND_FROM_EMAIL`, and `NOTIFICATION_EMAILS`.

The subject is `New Lead Submitted - <company>`. The body includes the lead
fields, optional context, and submission time. Empty context
is rendered as `(not provided)`. `RESEND_FROM_EMAIL` must be a valid sender for
the Resend account: use the Resend testing sender for local sandbox testing, or
a sender on a verified domain in production.

If Resend fails, the error is logged and the API still returns HTTP 200 because
the HubSpot operation already succeeded.

## Outreach Generation Boundary

The Content Studio outreach-email generation application contract is provider
neutral. The current adapter is Google Gemini API behind `outreach.Generator`.
It receives structured persisted seller, account, signal, analysis, and optional
Communication DNA context. Gemini does not use Google Search or external
grounding, and the endpoint does not send email or persist generated assets.
The adapter can be replaced behind the same interface later.

The backend-owned outreach prompt is versioned as `outreach-email-v1`. It uses
a fixed system instruction and a fixed user template whose context slots are
deterministic JSON. Seller company facts and seller messaging DNA are separate
slots, and target Communication DNA remains separate from both. The prompt
version is internal traceability metadata and does not change the public API.

Gemini inherits the request context created by the HTTP handler, including its
end-to-end deadline. The adapter does not add a shorter provider-only timeout,
retry requests, or enable tools, search, URL context, or external grounding.

## Future Integrations

No other external integration is currently implemented.
