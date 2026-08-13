# ABM Intelligence Database Schema Guide

**Audience:** AI agents, backend developers, migration authors, and reviewers  
**Purpose:** Explain the MVP database model, table relationships, field semantics, value ranges, and normalization rules so an agent can generate safe SQL without guessing.

> This document describes the current MVP schema only. Authentication, users, roles, permissions, sessions, campaigns, notifications, and CRM-sync persistence are intentionally excluded.

---

# 1. System Context

The MVP supports the following flow:

```text
Frontend
  → Backend
  → Database
  → AI Service
  → Database
  → Backend
  → Frontend
```

The business flow is:

```text
Company Profile
    ↓
Target Account
    ↓
Source Documents
    ↓
Signals
    ↓
Account Analysis + Communication DNA
    ↓
Generated Assets
```

The central business entity is **`target_account`**.

A `target_account` is a company that the seller is researching, scoring, monitoring, and creating account-specific content for.

---

# 2. Entity Relationship Overview

```text
company_profile
    1 ───── N target_account
    1 ───── N icp_profile

target_account
    1 ───── N source_document
    1 ───── N signal
    1 ───── N ai_run
    1 ───── N account_analysis
    1 ───── N communication_dna
    1 ───── N generated_asset

source_document
    1 ───── N signal
    via signal.source_document_id

ai_run
    1 ───── N signal
    1 ───── N account_analysis
    1 ───── N communication_dna
    1 ───── N generated_asset

signal
    0 ───── 1 signal
    via superseded_by_signal_id

signal
    1 ───── N generated_asset
    via generated_asset.anchor_signal_id
```

## Relationship meanings

- **`company_profile`** = the seller/company using the system.
- **`target_account`** = a company being researched.
- **`icp_profile`** = the seller's Ideal Customer Profile configuration.
- **`source_document`** = raw evidence collected from public sources.
- **`signal`** = structured intelligence extracted from evidence.
- **`account_analysis`** = scores and reasoning derived from the account, ICP, and signals.
- **`communication_dna`** = AI-derived language and messaging characteristics of a target account.
- **`ai_run`** = execution trace for AI jobs.
- **`generated_asset`** = generated Content Studio output.

---

# 3. Global Modeling Rules

## 3.1 Primary keys

All business tables use:

```text
UUID PRIMARY KEY
```

The exact UUID generation mechanism must follow the repository's existing migration/application convention.

Do not introduce a new UUID strategy only for these tables.

---

## 3.2 Timestamps

Use `TIMESTAMPTZ` unless the repository already has a different established convention.

Typical fields:

```text
created_at
updated_at
started_at
finished_at
fetched_at
last_verified_at
```

---

## 3.3 JSONB

Use `JSONB` for structured data that:

- is generated or enriched by AI,
- is expected to evolve,
- contains nested arrays/objects,
- does not currently require relational filtering at high frequency.

Examples:

```text
products
value_propositions
communication_dna
icp config
icp_breakdown
next_best_action
generated content
traceability
```

Do not normalize these into many child tables during the MVP unless a concrete query requirement demands it.

---

## 3.4 Do not overuse ENUMs

Use `VARCHAR + CHECK` only for **stable value sets explicitly confirmed by the product**.

Use unrestricted `VARCHAR` for evolving taxonomies such as:

```text
signal.type
source_document.source_type
industry
ai_run.run_type
ai_run.run_status
ai_run.model
```

---

# 4. Table: `company_profile`

## Purpose

Represents the company using the ABM system.

In the current prototype this is conceptually the seller company, such as TechSmith.

It is **not** a user/login entity.

## Relationship

```text
company_profile
    1 ───── N target_account
    1 ───── N icp_profile
```

## Fields

| Field | PostgreSQL Type | Required | Range / Allowed Values | Description |
|---|---|---:|---|---|
| `id` | UUID | Yes | UUID | Primary key |
| `name` | VARCHAR(255) | Yes | Free text | Seller/company name |
| `tagline` | VARCHAR(500) | No | Free text | Company tagline |
| `website` | TEXT | No | URL/domain text | Company website |
| `hq` | VARCHAR(255) | No | Free text | Headquarters |
| `founded` | SMALLINT | No | Valid year if known | Founding year |
| `description` | TEXT | No | Free text | Company description |
| `products` | JSONB | Yes | JSON array | Products offered by the seller |
| `value_propositions` | JSONB | Yes | JSON array | Seller value propositions |
| `target_verticals` | JSONB | Yes | JSON array | Industries/verticals the seller targets |
| `buyer_personas` | JSONB | Yes | JSON array | Buyer personas the seller targets |
| `communication_dna` | JSONB | Yes | JSON object | Seller tone, proof style, CTA style, vocabulary |
| `signal_catalog` | JSONB | Yes | JSON array | Buying-signal concepts that the system should scan for |
| `created_at` | TIMESTAMPTZ | Yes | Timestamp | Record creation time |
| `updated_at` | TIMESTAMPTZ | Yes | Timestamp | Last modification time |

## Recommended JSON defaults

```text
products             = []
value_propositions   = []
target_verticals     = []
buyer_personas       = []
communication_dna    = {}
signal_catalog       = []
```

## Example `products`

```json
[
  {
    "name": "Camtasia Editor",
    "desc": "Screen recording + professional video editing"
  },
  {
    "name": "Snagit",
    "desc": "Screen capture, markup & auto step-by-step guides"
  }
]
```

## Example `communication_dna`

```json
{
  "tone": "Practical, warm, human",
  "proof": "Enterprise logos + research",
  "cta": "Consultative for enterprise",
  "vocab": "knowledge sharing, training videos"
}
```

---

# 5. Table: `target_account`

## Purpose

Represents a company being researched by the ABM system.

Examples:

```text
Oracle
Salesforce
IBM
ServiceNow
SecureVault GmbH
```

The table stores **company facts**, not AI scoring results.

## Relationship

```text
company_profile
    1 ───── N target_account
```

A seller can research many target accounts.

## Fields

| Field | PostgreSQL Type | Required | Range / Allowed Values | Description |
|---|---|---:|---|---|
| `id` | UUID | Yes | UUID | Primary key |
| `company_profile_id` | UUID | Yes | FK | Owning seller profile |
| `name` | VARCHAR(255) | Yes | Free text | Company name |
| `domain` | VARCHAR(255) | Yes | Domain | Canonical company domain |
| `web_url` | TEXT | Yes | URL | Main website URL |
| `impressum_url` | TEXT | No | URL | Impressum/legal page |
| `industry` | VARCHAR(255) | No | Open taxonomy | Industry classification |
| `hq` | VARCHAR(255) | No | Free text | Headquarters |
| `employees` | INTEGER | No | `>= 0` | Employee count |
| `revenue_m` | NUMERIC(18,2) | No | `>= 0` when known | Revenue normalized to millions |
| `revenue_currency` | CHAR(3) | No | ISO-like currency code | Example: EUR, USD |
| `founded` | SMALLINT | No | Valid year if known | Founding year |
| `description` | TEXT | No | Free text | Account description |
| `notes` | TEXT | No | Free text | Manually supplied context |
| `lifecycle` | VARCHAR(30) | Yes | See below | CRM/account lifecycle |
| `created_at` | TIMESTAMPTZ | Yes | Timestamp | Created time |
| `updated_at` | TIMESTAMPTZ | Yes | Timestamp | Updated time |

## Confirmed `lifecycle` values

Only these values are currently confirmed:

```text
Lead
MQL
SQL
Opportunity
Deal
Customer
```

Recommended constraint:

```sql
CHECK (
  lifecycle IN (
    'Lead',
    'MQL',
    'SQL',
    'Opportunity',
    'Deal',
    'Customer'
  )
)
```

## Important rules

### Industry

Do **not** restrict `industry` with an ENUM.

The prototype already contains multiple industries and allows future expansion.

### Domain uniqueness

Recommended logical uniqueness:

```text
(company_profile_id, lower(domain))
```

This means the same target domain should not be duplicated for the same seller profile.

### Revenue normalization

Examples:

```text
€18M   → revenue_m = 18
$57.4B → revenue_m = 57400
```

Currency is stored separately.

---

# 6. Table: `icp_profile`

## Purpose

Stores the seller's configurable Ideal Customer Profile.

The ICP determines account-fit scoring and tier thresholds.

## Relationship

```text
company_profile
    1 ───── N icp_profile
```

Versioning is supported so previous configurations can remain reproducible.

## Fields

| Field | PostgreSQL Type | Required | Range / Allowed Values | Description |
|---|---|---:|---|---|
| `id` | UUID | Yes | UUID | Primary key |
| `company_profile_id` | UUID | Yes | FK | Seller/company profile |
| `version` | INTEGER | Yes | `> 0` | Configuration version |
| `config` | JSONB | Yes | JSON object | Complete ICP configuration |
| `is_active` | BOOLEAN | Yes | `true/false` | Whether this is the active ICP |
| `created_at` | TIMESTAMPTZ | Yes | Timestamp | Creation time |

Recommended uniqueness:

```text
(company_profile_id, version)
```

## Current confirmed ICP dimensions

```text
industry
size
revenue
techstack
geography
recency
techadopt
growth
```

Current names:

```text
Industry
Company Size
Revenue
Tech Stack
Geography
Buying Signal Recency
Technology Adoption
Growth Trajectory
```

## Current default configuration

```text
Each dimension weight = 12.5
```

Tier thresholds:

```text
Focus Accounts:
    ICP >= 80
    Signal >= 70

Tier 1:
    ICP >= 70
    High Signals >= 1

Tier 2:
    ICP >= 50
```

Resonance:

```text
ICP weight    = 0.4
Signal weight = 0.6
```

Signal-score configuration:

```text
Relevance = 10–50
ICP Match = 0–30
Freshness = 0–20
```

Keep these settings inside `config` rather than creating many configuration tables.

---

# 7. Table: `source_document`

## Purpose

Stores raw evidence collected by research/crawling.

This is the traceability foundation.

Conceptually:

```text
Source Document
    ↓
Signal
    ↓
Account Analysis
    ↓
Generated Asset
```

## Relationship

```text
target_account
    1 ───── N source_document
```

and:

```text
source_document
    1 ───── N signal
```

For the MVP, `signal.source_document_id` identifies the primary structured source.

## Fields

| Field | PostgreSQL Type | Required | Range / Allowed Values | Description |
|---|---|---:|---|---|
| `id` | UUID | Yes | UUID | Primary key |
| `account_id` | UUID | Yes | FK | Account this source belongs to |
| `source_name` | VARCHAR(500) | No | Free text | Human-readable source/domain |
| `source_type` | VARCHAR(255) | No | Open taxonomy | Type of source |
| `url` | TEXT | Yes | URL | Source URL |
| `content` | TEXT | No | Free text | Extracted/raw page content |
| `content_hash` | VARCHAR(64) | No | Hash | Used to detect content changes |
| `fetched_at` | TIMESTAMPTZ | No | Timestamp | When the source was fetched |
| `created_at` | TIMESTAMPTZ | Yes | Timestamp | Record creation time |

## Observed `source_type` examples

These are examples, **not an exhaustive allowed-value list**:

```text
Career Page
Website
Press
Event
Register
Job Portal
Press Release
Investor
Interpretation
SEC 10-K
Press Release / Blog
Partner Blog / Release Notes Summary
Career Page (aggregiert)
```

Do not create a restrictive CHECK or ENUM for this field.

---

# 8. Table: `ai_run`

## Purpose

Tracks calls from the backend to an AI service.

This table exists primarily for:

- debugging,
- reproducibility,
- auditability,
- model comparison,
- failure diagnosis.

## Relationship

```text
target_account
    1 ───── N ai_run
```

An `ai_run` may produce:

```text
signal
account_analysis
communication_dna
generated_asset
```

## Fields

| Field | PostgreSQL Type | Required | Range / Allowed Values | Description |
|---|---|---:|---|---|
| `id` | UUID | Yes | UUID | Primary key |
| `account_id` | UUID | No | FK | Account context, if account-specific |
| `run_type` | VARCHAR(100) | Yes | Open | Type of AI operation |
| `model` | VARCHAR(100) | No | Open | Model identifier |
| `run_status` | VARCHAR(50) | Yes | Open | Execution status |
| `request_payload` | JSONB | No | JSON | Payload sent to AI |
| `response_payload` | JSONB | No | JSON | Raw AI response |
| `error_message` | TEXT | No | Free text | Failure details |
| `started_at` | TIMESTAMPTZ | Yes | Timestamp | Start time |
| `finished_at` | TIMESTAMPTZ | No | Timestamp | Completion time |

## Important rule

The prototype does **not** define the complete allowed values of:

```text
run_type
run_status
model
```

Therefore an agent must **not invent ENUMs or CHECK constraints** for these fields.

---

# 9. Table: `signal`

## Purpose

Stores structured buying/account signals.

A signal represents a meaningful piece of evidence or derived intelligence about a target account.

Examples:

```text
Job posting
Funding event
Personnel change
Communication gap
Product launch
Company data
```

## Relationship

```text
target_account
    1 ───── N signal
```

Optional provenance relationships:

```text
source_document
    1 ───── N signal

ai_run
    1 ───── N signal
```

Signals can also supersede older signals:

```text
old signal
    ↓ superseded_by_signal_id
new signal
```

## Fields

| Field | PostgreSQL Type | Required | Range / Allowed Values | Description |
|---|---|---:|---|---|
| `id` | UUID | Yes | UUID | Primary key |
| `signal_key` | VARCHAR(100) | No | Human-readable stable key | Example: `sc-ora-13` |
| `account_id` | UUID | Yes | FK | Target account |
| `ai_run_id` | UUID | No | FK | AI run that created/updated the signal |
| `source_document_id` | UUID | No | FK | Primary source document |
| `type` | VARCHAR(100) | Yes | Open taxonomy | Signal category |
| `title` | TEXT | Yes | Free text | Short signal title |
| `body` | TEXT | No | Free text | Detailed evidence/reasoning |
| `strength` | VARCHAR(10) | Yes | `high`, `medium`, `low` | Signal strength |
| `relevance` | VARCHAR(10) | No | `High`, `Medium`, `Low` | Relevance to seller/use case |
| `signal_date` | DATE | No | Valid date | Exact event/publication date when known |
| `signal_date_raw` | VARCHAR(50) | No | Free text | Partial/ambiguous source date |
| `freshness_label` | VARCHAR(255) | No | Free text | UI-friendly freshness string |
| `evidence_status` | VARCHAR(30) | Yes | See below | Evidence quality/classification |
| `verified` | BOOLEAN | Yes | `true/false` | Whether the signal was verified |
| `last_verified_at` | TIMESTAMPTZ | No | Timestamp | Most recent verification |
| `is_active` | BOOLEAN | Yes | `true/false` | Whether it is currently a live signal |
| `closed_observed_at` | DATE | No | Date | Date the system observed it was closed |
| `score_eligible` | BOOLEAN | Yes | `true/false` | Whether it contributes to current Signal Score |
| `exclusion_reason` | TEXT | No | Free text | Reason it is excluded |
| `superseded_by_signal_id` | UUID | No | Self-FK | New signal replacing this one |
| `created_at` | TIMESTAMPTZ | Yes | Timestamp | Created time |
| `updated_at` | TIMESTAMPTZ | Yes | Timestamp | Updated time |

---

## 9.1 Confirmed `strength` values

```text
high
medium
low
```

Recommended constraint:

```sql
CHECK (strength IN ('high', 'medium', 'low'))
```

---

## 9.2 Confirmed `relevance` values

Normalized values:

```text
High
Medium
Low
```

Recommended constraint:

```sql
CHECK (relevance IN ('High', 'Medium', 'Low'))
```

### Important normalization

A prototype string such as:

```text
High (historical, not counted)
```

must **not** be stored literally in `relevance`.

It should be normalized as:

```text
relevance      = High
score_eligible = false
```

---

## 9.3 Confirmed `evidence_status` values

```text
SOURCE_BACKED
DERIVED
INSUFFICIENT_DATA
```

Recommended constraint:

```sql
CHECK (
  evidence_status IN (
    'SOURCE_BACKED',
    'DERIVED',
    'INSUFFICIENT_DATA'
  )
)
```

### Meaning

**SOURCE_BACKED**

The claim is directly supported by an identified source.

**DERIVED**

The claim is inferred from one or more source-backed observations.

**INSUFFICIENT_DATA**

Current evidence is not sufficient to make the claim reliably.

---

## 9.4 Signal type

Observed types currently include:

```text
Job Posting
Funding & Expansion
Communication Gap
News & Events
Personnel Change
Company Data
```

This list is **not closed**.

Do not implement an ENUM/CHECK for `signal.type`.

---

## 9.5 `is_active` vs `evidence_status`

These fields answer different questions.

```text
is_active
→ Is this still a current/live signal?

evidence_status
→ How strong is the evidence supporting the signal?
```

A signal may be:

```text
evidence_status = SOURCE_BACKED
is_active       = false
```

This is valid.

Example:

A job posting was directly verified but later closed.

---

## 9.6 `closed_observed_at`

This is intentionally **not** named `closed_at`.

Meaning:

> The date on which our system first verified that the signal was already closed.

It does not claim to know the exact real-world closing timestamp.

---

## 9.7 `score_eligible`

`score_eligible` answers:

> Should this signal contribute to the account's current Signal Score?

Typical rule:

```text
live + relevant signal
    → score_eligible = true

historical closed signal
    → score_eligible = false
```

The historical signal may still remain valuable as evidence.

---

## 9.8 `superseded_by_signal_id`

Used when an old signal has been replaced by a newer/better signal.

Example:

```text
sc-ora-1
    ↓
superseded by
    ↓
sc-ora-4
```

This should be a self-referencing FK, not merely a text sentence in `body`.

---

## 9.9 Signal count is derived

Do **not** store `signal_count` on `target_account`.

Calculate counts from `signal`.

### Total signals

```sql
COUNT(*)
WHERE account_id = ?
```

### Active signals

```sql
COUNT(*)
WHERE account_id = ?
  AND is_active = true
```

Therefore:

```text
totalSignalCount
activeSignalCount
```

are API/query-derived values, not source-of-truth database columns.

---

# 10. Table: `account_analysis`

## Purpose

Stores derived scoring and account-level reasoning.

This table should contain AI/scoring results, while `target_account` contains stable company facts.

## Relationship

```text
target_account
    1 ───── N account_analysis
```

Historical analysis rows should be preserved.

A new analysis should usually create a new row rather than overwriting the previous analysis.

## Fields

| Field | PostgreSQL Type | Required | Range / Allowed Values | Description |
|---|---|---:|---|---|
| `id` | UUID | Yes | UUID | Primary key |
| `account_id` | UUID | Yes | FK | Target account |
| `icp_profile_id` | UUID | No | FK | ICP configuration used |
| `ai_run_id` | UUID | No | FK | AI execution that produced analysis |
| `icp_score` | NUMERIC(5,2) | Yes | `0..100` | ICP Fit score |
| `icp_fit` | VARCHAR(20) | Yes | `High`, `Medium`, `Low` | ICP classification |
| `icp_breakdown` | JSONB | Yes | JSON object | Per-dimension ICP explanation |
| `signal_score` | NUMERIC(5,2) | Yes | `0..100` | Current Signal Score |
| `signal_score_breakdown` | JSONB | No | JSON | Optional score details |
| `resonance_score` | NUMERIC(5,2) | Yes | `0..100` | Combined ICP + Signal score |
| `tier` | VARCHAR(30) | Yes | See below | Account priority tier |
| `why_this_account` | TEXT | No | Free text | Why the account matters |
| `why_now` | TEXT | No | Free text | Why action is timely |
| `next_best_action` | JSONB | No | JSON object | Recommended action |
| `created_at` | TIMESTAMPTZ | Yes | Timestamp | Analysis creation time |

## Score constraints

```sql
CHECK (icp_score BETWEEN 0 AND 100)
CHECK (signal_score BETWEEN 0 AND 100)
CHECK (resonance_score BETWEEN 0 AND 100)
```

---

## 10.1 `icp_fit`

Confirmed values:

```text
High
Medium
Low
```

---

## 10.2 `tier`

Confirmed values:

```text
Focus Accounts
Tier 1
Tier 2
Below ICP
```

`Below ICP` exists in scoring logic even if normal UI filters emphasize the first three.

Recommended constraint:

```sql
CHECK (
  tier IN (
    'Focus Accounts',
    'Tier 1',
    'Tier 2',
    'Below ICP'
  )
)
```

---

## 10.3 `next_best_action`

Current structure:

```json
{
  "action": "Primary recommended action",
  "rationale": "Why this action is appropriate",
  "timeWindow": "Now",
  "priority": "Critical"
}
```

Observed `priority` values:

```text
Critical
High
Medium
Low
```

Because `next_best_action` is JSONB in the MVP, no database CHECK is required inside the JSON.

---

# 11. Table: `communication_dna`

## Purpose

Stores the target account's communication style extracted/derived by AI.

## Relationship

```text
target_account
    1 ───── N communication_dna
```

A new AI analysis can create a new DNA version.

## Fields

| Field | PostgreSQL Type | Required | Description |
|---|---|---:|---|
| `id` | UUID | Yes | Primary key |
| `account_id` | UUID | Yes | Target account FK |
| `ai_run_id` | UUID | No | AI execution FK |
| `tone` | JSONB | Yes | Primary/secondary tone and evidence |
| `vocabulary` | JSONB | Yes | Common terms and contexts |
| `value_propositions` | JSONB | Yes | Account value propositions |
| `problem_framing` | JSONB | Yes | How the account frames problems |
| `proof_style` | JSONB | Yes | How the account proves claims |
| `cta_patterns` | JSONB | Yes | Typical calls to action |
| `recurring_phrases` | JSONB | Yes | Repeated phrases |
| `do_rules` | JSONB | Yes | Messaging recommendations |
| `dont_rules` | JSONB | Yes | Messaging warnings |
| `created_at` | TIMESTAMPTZ | Yes | Creation time |

## Recommended defaults

```text
tone                = {}
vocabulary          = {}
value_propositions  = []
problem_framing     = {}
proof_style         = {}
cta_patterns        = {}
recurring_phrases   = []
do_rules            = []
dont_rules          = []
```

## Example `tone`

```json
{
  "primary": "Technical-authoritative",
  "secondary": "Enterprise-confident",
  "description": "Precise, engineering-led, scale-oriented.",
  "status": "DERIVED",
  "sources": ["oracle.com"]
}
```

## Example `vocabulary`

```json
{
  "terms": [
    {
      "word": "cloud infrastructure",
      "frequency": "high",
      "context": "Core",
      "source": "oracle.com",
      "sourceType": "Homepage"
    }
  ],
  "status": "DERIVED"
}
```

Observed vocabulary `frequency` values currently include:

```text
high
medium
```

Do not invent additional database restrictions.

---

# 12. Table: `generated_asset`

## Purpose

Stores Content Studio output generated for an account.

Examples:

```text
LinkedIn Ad
Landing Page
Outreach Email
LinkedIn Outreach
Sales Talking Points
```

## Relationship

```text
target_account
    1 ───── N generated_asset
```

Optional references:

```text
ai_run
    1 ───── N generated_asset

signal
    1 ───── N generated_asset
    via anchor_signal_id
```

## Fields

| Field | PostgreSQL Type | Required | Range / Allowed Values | Description |
|---|---|---:|---|---|
| `id` | UUID | Yes | UUID | Primary key |
| `account_id` | UUID | Yes | FK | Target account |
| `ai_run_id` | UUID | No | FK | AI generation run |
| `anchor_signal_id` | UUID | No | FK | Signal used as the generation anchor |
| `persona` | VARCHAR(30) | No | See below | Target persona |
| `asset_type` | VARCHAR(30) | Yes | See below | Generated asset family |
| `asset_subtype` | VARCHAR(30) | No | See below | Asset subtype |
| `content` | JSONB | Yes | JSON object | Generated content |
| `traceability` | JSONB | Yes | JSON array | Source/evidence mapping |
| `created_at` | TIMESTAMPTZ | Yes | Timestamp | Creation time |
| `updated_at` | TIMESTAMPTZ | Yes | Timestamp | Last modification |

---

## 12.1 Confirmed `persona` values

```text
marketing
sales
exec
```

Mappings:

```text
marketing → Head of Marketing / CMO
sales     → Head of Sales / VP Sales
exec      → CEO / Founder
```

Recommended constraint:

```sql
CHECK (
  persona IS NULL
  OR persona IN ('marketing', 'sales', 'exec')
)
```

---

## 12.2 Confirmed `asset_type` values

```text
ad
landing
email
linkedin
talking
```

Mappings:

```text
ad        → LinkedIn Ad
landing   → Landing Page
email     → Outreach Email
linkedin  → LinkedIn Outreach
talking   → Sales Talking Points
```

Recommended constraint:

```sql
CHECK (
  asset_type IN (
    'ad',
    'landing',
    'email',
    'linkedin',
    'talking'
  )
)
```

---

## 12.3 Confirmed `asset_subtype` values

Currently confirmed for LinkedIn Ads:

```text
single
document
tla
```

Mappings:

```text
single   → Single Image Ad
document → Document Ad
tla      → Thought Leadership Ad
```

Recommended constraint:

```sql
CHECK (
  asset_subtype IS NULL
  OR asset_subtype IN ('single', 'document', 'tla')
)
```

Do not invent subtypes for other asset types.

---

## 12.4 `traceability`

Purpose:

> Explain why generated content contains each claim/style element and where it came from.

Example:

```json
[
  {
    "label": "Anchor signal",
    "value": "Hiring a new enablement leader",
    "status": "SOURCE_BACKED",
    "source": "careers.example.com",
    "url": "https://careers.example.com/job/123"
  },
  {
    "label": "Tone",
    "value": "Technical-authoritative",
    "status": "DERIVED",
    "source": "example.com"
  }
]
```

This is a key product capability and should not be discarded.

---

# 13. Derived Values — Do Not Store as Source-of-Truth Columns

The following values should normally be calculated by the backend/query layer.

## Active signal count

```sql
SELECT COUNT(*)
FROM signal
WHERE account_id = :account_id
  AND is_active = TRUE;
```

## Total signal count

```sql
SELECT COUNT(*)
FROM signal
WHERE account_id = :account_id;
```

## Signal count by type

```sql
SELECT type, COUNT(*)
FROM signal
WHERE account_id = :account_id
GROUP BY type;
```

These can be returned by an API as:

```json
{
  "totalSignalCount": 14,
  "activeSignalCount": 13,
  "signalTypeCounts": {
    "Job Posting": 7,
    "News & Events": 5,
    "Company Data": 2
  }
}
```

Do not maintain duplicate count columns unless a later performance requirement justifies a materialized aggregate.

---

# 14. Critical Normalization Example

A historical prototype signal may conceptually look like:

```text
status      = "SOURCE_BACKED (closed)"
relevance   = "High (historical, not counted)"
freshness   = "Closed as of 2026-08-10"
supersededBy = "sc-ora-4"
```

Do **not** reproduce those overloaded strings in the database.

Normalize them as:

```text
evidence_status          = SOURCE_BACKED
relevance                = High
is_active                = false
closed_observed_at       = 2026-08-10
score_eligible           = false
superseded_by_signal_id  = <UUID of sc-ora-4>
```

Why:

```text
SOURCE_BACKED
```

describes evidence quality.

```text
is_active = false
```

describes lifecycle/currentness.

```text
score_eligible = false
```

describes scoring behavior.

These are independent concepts.

---

# 15. Agent Rules

An AI agent working with this database must follow these rules.

## Rule 1 — Do not guess allowed values

If a field has a documented closed value set, use only documented values.

If a field is documented as an open taxonomy, do not invent a database CHECK constraint.

---

## Rule 2 — Separate facts from analysis

Store stable company facts in:

```text
target_account
```

Store derived scoring/reasoning in:

```text
account_analysis
```

Do not copy ICP Score, Signal Score, Why Now, etc. into `target_account`.

---

## Rule 3 — Preserve evidence

Signals must remain traceable to evidence when possible.

Preferred chain:

```text
source_document
    ↓
signal
    ↓
account_analysis / communication_dna
    ↓
generated_asset
```

---

## Rule 4 — Closed does not mean insufficient data

This is valid:

```text
evidence_status = SOURCE_BACKED
is_active       = false
```

A source can strongly prove that something is no longer active.

---

## Rule 5 — Verified does not mean source-backed

`verified` and `evidence_status` are separate.

For example:

```text
evidence_status = DERIVED
verified        = true
```

is valid when a derived conclusion has been reviewed/validated.

---

## Rule 6 — Do not delete historical analysis by default

Prefer version/history preservation for:

```text
account_analysis
communication_dna
ai_run
generated_asset
signal
```

Historical rows are useful for debugging and explaining change over time.

---

## Rule 7 — Do not store UI-only strings as core semantics

Examples of UI labels:

```text
"This week"
"~7 weeks"
"Closed as of 2026-08-10"
"13 active"
```

These may be stored as display metadata such as `freshness_label`, but lifecycle/scoring logic must rely on normalized fields.

---

# 16. Recommended Indexes

Exact syntax should follow the repository's migration conventions.

Recommended indexes:

```text
target_account(company_profile_id)

UNIQUE:
target_account(company_profile_id, lower(domain))

icp_profile(company_profile_id)
UNIQUE:
icp_profile(company_profile_id, version)

source_document(account_id)
source_document(account_id, fetched_at)

ai_run(account_id)
ai_run(account_id, started_at)

signal(account_id)
signal(account_id, is_active)
signal(account_id, type)
signal(account_id, score_eligible)
signal(account_id, signal_date)

UNIQUE where appropriate:
signal(account_id, signal_key)

account_analysis(account_id, created_at)

communication_dna(account_id, created_at)

generated_asset(account_id, created_at)
generated_asset(anchor_signal_id)
```

---

# 17. Foreign-Key Summary

```text
target_account.company_profile_id
    → company_profile.id

icp_profile.company_profile_id
    → company_profile.id

source_document.account_id
    → target_account.id

ai_run.account_id
    → target_account.id

signal.account_id
    → target_account.id

signal.ai_run_id
    → ai_run.id

signal.source_document_id
    → source_document.id

signal.superseded_by_signal_id
    → signal.id

account_analysis.account_id
    → target_account.id

account_analysis.icp_profile_id
    → icp_profile.id

account_analysis.ai_run_id
    → ai_run.id

communication_dna.account_id
    → target_account.id

communication_dna.ai_run_id
    → ai_run.id

generated_asset.account_id
    → target_account.id

generated_asset.ai_run_id
    → ai_run.id

generated_asset.anchor_signal_id
    → signal.id
```

Delete behavior must follow the existing repository convention. Do not introduce broad `ON DELETE CASCADE` behavior without reviewing current migration practices.

---

# 18. Current MVP Tables

The schema currently contains these nine core tables:

```text
1. company_profile
2. target_account
3. icp_profile
4. source_document
5. ai_run
6. signal
7. account_analysis
8. communication_dna
9. generated_asset
```

Explicitly out of scope for the current MVP:

```text
user
role
permission
session
login_history
campaign
notification
crm_sync
agent_discovery_suggestion
```

---

# 19. Mental Model for an AI Agent

When an agent receives a task, it should think about the database in this order:

```text
Who is selling?
→ company_profile

Which company are we researching?
→ target_account

What evidence do we have?
→ source_document

What meaningful events/facts were extracted?
→ signal

What did the scoring/reasoning engine conclude?
→ account_analysis

How does the target communicate?
→ communication_dna

Which AI call produced the result?
→ ai_run

What content did we generate?
→ generated_asset
```

This model should be used before generating SQL INSERT, UPDATE, SELECT, migration, or data-quality scripts.

---

# 20. Data-Quality Checklist for Agents

Before generating SQL or inserting data, validate:

1. All FK references point to existing records.
2. Required fields are present.
3. Score fields are within `0..100`.
4. `strength` uses only `high / medium / low`.
5. `relevance` uses only `High / Medium / Low`.
6. `evidence_status` uses only `SOURCE_BACKED / DERIVED / INSUFFICIENT_DATA`.
7. `lifecycle` uses only documented lifecycle values.
8. `tier` uses only documented tier values.
9. Closed signals are not encoded inside `evidence_status`.
10. Historical/non-active signals do not accidentally contribute to current scoring.
11. `superseded_by_signal_id` references a real signal.
12. Do not store derived signal counts as account facts.
13. Do not invent new enum/check values because a UI string looks similar.
14. Preserve source URLs and traceability whenever available.
15. Use repository-native UUID, timestamp, migration, and FK conventions.

---

**End of schema guide.**
