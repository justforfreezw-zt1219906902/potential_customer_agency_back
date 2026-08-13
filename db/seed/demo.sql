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

INSERT INTO signal (id, signal_key, account_id, type, title, strength, evidence_status, is_active)
VALUES
('00000000-0000-0000-0000-000000002001', 'focus-active-one', '00000000-0000-0000-0000-000000000101', 'growth', 'Expansion signal', 'high', 'DERIVED', TRUE),
('00000000-0000-0000-0000-000000002002', 'focus-active-two', '00000000-0000-0000-0000-000000000101', 'hiring', 'Hiring signal', 'medium', 'SOURCE_BACKED', TRUE),
('00000000-0000-0000-0000-000000002003', 'focus-inactive', '00000000-0000-0000-0000-000000000101', 'funding', 'Historical funding signal', 'low', 'SOURCE_BACKED', FALSE),
('00000000-0000-0000-0000-000000002004', 'tier1-active', '00000000-0000-0000-0000-000000000102', 'hiring', 'Hiring signal', 'medium', 'SOURCE_BACKED', TRUE)
ON CONFLICT (id) DO NOTHING;
