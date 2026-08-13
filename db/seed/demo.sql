-- Explicit development/demo data only. This file is never run by Liquibase or Railway startup.
-- The company UUID must match DEMO_COMPANY_PROFILE_ID in the runtime environment.

INSERT INTO company_profile (id, name, products, value_propositions, target_verticals, buyer_personas, communication_dna, signal_catalog)
VALUES ('00000000-0000-0000-0000-000000000001', 'Demo Company', '[]', '[]', '[]', '[]', '{}', '[]')
ON CONFLICT (id) DO NOTHING;

INSERT INTO target_account (id, company_profile_id, name, domain, web_url, industry, hq, employees, revenue_m, revenue_currency, founded, description, lifecycle)
VALUES
('00000000-0000-0000-0000-000000000101', '00000000-0000-0000-0000-000000000001', 'Focus Account', 'focus.example', 'https://focus.example', 'Technology', 'Austin, TX', 1200, 450.5, 'USD', 2012, 'Synthetic demo account for overview verification.', 'SQL'),
('00000000-0000-0000-0000-000000000102', '00000000-0000-0000-0000-000000000001', 'Tier One Account', 'tier1.example', 'https://tier1.example', 'Software', 'Berlin', 800, 210, 'EUR', 2015, 'Synthetic demo account.', 'MQL'),
('00000000-0000-0000-0000-000000000103', '00000000-0000-0000-0000-000000000001', 'Tier Two Account', 'tier2.example', 'https://tier2.example', 'Finance', NULL, NULL, NULL, NULL, NULL, NULL, 'Lead'),
('00000000-0000-0000-0000-000000000104', '00000000-0000-0000-0000-000000000001', 'Below ICP Account', 'below.example', 'https://below.example', NULL, 'Munich', NULL, NULL, NULL, NULL, NULL, 'Lead'),
('00000000-0000-0000-0000-000000000105', '00000000-0000-0000-0000-000000000001', 'Unanalyzed Account', 'unanalyzed.example', 'https://unanalyzed.example', 'Retail', 'Hamburg', NULL, NULL, NULL, NULL, NULL, 'Lead')
ON CONFLICT (id) DO UPDATE SET
  company_profile_id = EXCLUDED.company_profile_id, name = EXCLUDED.name, domain = EXCLUDED.domain,
  web_url = EXCLUDED.web_url, industry = EXCLUDED.industry, hq = EXCLUDED.hq,
  employees = EXCLUDED.employees, revenue_m = EXCLUDED.revenue_m,
  revenue_currency = EXCLUDED.revenue_currency, founded = EXCLUDED.founded,
  description = EXCLUDED.description, lifecycle = EXCLUDED.lifecycle;

INSERT INTO source_document (id, account_id, source_name, source_type, url, content)
VALUES
('00000000-0000-0000-0000-000000003001', '00000000-0000-0000-0000-000000000101', 'Demo Careers', 'Career Page', 'https://focus.example/careers', 'Synthetic demo careers source.'),
('00000000-0000-0000-0000-000000003002', '00000000-0000-0000-0000-000000000101', 'Demo Newsroom', 'News', 'https://focus.example/news', 'Synthetic demo newsroom source.'),
('00000000-0000-0000-0000-000000003003', '00000000-0000-0000-0000-000000000102', NULL, NULL, 'https://tier1.example/company', NULL)
ON CONFLICT (id) DO UPDATE SET
  account_id = EXCLUDED.account_id, source_name = EXCLUDED.source_name,
  source_type = EXCLUDED.source_type, url = EXCLUDED.url, content = EXCLUDED.content;

INSERT INTO account_analysis (id, account_id, icp_score, icp_fit, signal_score, resonance_score, tier, why_this_account, why_now, next_best_action, created_at)
VALUES
('00000000-0000-0000-0000-000000001001', '00000000-0000-0000-0000-000000000101', 80, 'High', 75, 70, 'Focus Accounts', 'Strong synthetic enterprise fit.', 'Historical demo timing.', '{"action":"Review account signals"}', '2026-07-01T10:00:00Z'),
('00000000-0000-0000-0000-000000001002', '00000000-0000-0000-0000-000000000101', 86.5, 'High', 91, 89.2, 'Focus Accounts', 'Strong synthetic enterprise fit.', 'Multiple current demo buying signals.', '{"action":"Prepare signal-led outreach","rationale":"Current demo evidence supports immediate outreach.","timeWindow":"Now","priority":"Critical"}', '2026-08-01T10:00:00Z'),
('00000000-0000-0000-0000-000000001003', '00000000-0000-0000-0000-000000000102', 78, 'High', 82, 80, 'Tier 1', 'Synthetic software account fit.', 'Recent demo hiring signal.', '{"action":"Review buying committee"}', '2026-08-02T10:00:00Z'),
('00000000-0000-0000-0000-000000001004', '00000000-0000-0000-0000-000000000103', 65, 'Medium', 60, 55, 'Tier 2', NULL, NULL, '{"action":"Monitor account"}', '2026-08-03T10:00:00Z'),
('00000000-0000-0000-0000-000000001005', '00000000-0000-0000-0000-000000000104', 35, 'Low', 30, 25, 'Below ICP', NULL, NULL, '{"action":"Revisit ICP fit"}', '2026-08-04T10:00:00Z')
ON CONFLICT (id) DO UPDATE SET
  account_id = EXCLUDED.account_id, icp_score = EXCLUDED.icp_score, icp_fit = EXCLUDED.icp_fit,
  signal_score = EXCLUDED.signal_score, resonance_score = EXCLUDED.resonance_score,
  tier = EXCLUDED.tier, why_this_account = EXCLUDED.why_this_account,
  why_now = EXCLUDED.why_now, next_best_action = EXCLUDED.next_best_action,
  created_at = EXCLUDED.created_at;

INSERT INTO signal (id, signal_key, account_id, source_document_id, type, title, body, strength, relevance, signal_date, signal_date_raw, freshness_label, evidence_status, verified, is_active, score_eligible)
VALUES
('00000000-0000-0000-0000-000000002001', 'focus-active-one', '00000000-0000-0000-0000-000000000101', '00000000-0000-0000-0000-000000003001', 'Job Posting', 'Hiring enterprise solutions engineers', 'The account is expanding its enterprise solutions team.', 'high', 'High', '2026-08-01', NULL, 'Recent', 'SOURCE_BACKED', TRUE, TRUE, TRUE),
('00000000-0000-0000-0000-000000002002', 'focus-active-two', '00000000-0000-0000-0000-000000000101', '00000000-0000-0000-0000-000000003002', 'News & Events', 'Expansion announcement', NULL, 'medium', 'Medium', '2026-07-20', 'July 2026', 'Recent', 'SOURCE_BACKED', TRUE, TRUE, TRUE),
('00000000-0000-0000-0000-000000002003', 'focus-inactive', '00000000-0000-0000-0000-000000000101', NULL, 'Company Data', 'Historical company data', NULL, 'low', NULL, NULL, NULL, NULL, 'DERIVED', FALSE, FALSE, FALSE),
('00000000-0000-0000-0000-000000002004', 'tier1-active', '00000000-0000-0000-0000-000000000102', '00000000-0000-0000-0000-000000003003', 'Company Data', 'Company profile signal', NULL, 'medium', NULL, NULL, NULL, NULL, 'INSUFFICIENT_DATA', FALSE, TRUE, FALSE)
ON CONFLICT (id) DO UPDATE SET
  account_id = EXCLUDED.account_id, source_document_id = EXCLUDED.source_document_id,
  type = EXCLUDED.type, title = EXCLUDED.title, body = EXCLUDED.body,
  strength = EXCLUDED.strength, relevance = EXCLUDED.relevance, signal_date = EXCLUDED.signal_date,
  signal_date_raw = EXCLUDED.signal_date_raw, freshness_label = EXCLUDED.freshness_label,
  evidence_status = EXCLUDED.evidence_status, verified = EXCLUDED.verified,
  is_active = EXCLUDED.is_active, score_eligible = EXCLUDED.score_eligible;
