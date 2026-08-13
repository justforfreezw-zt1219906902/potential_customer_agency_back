package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/justforfreezw-zt1219906902/potential_customer_agency_back/internal/models"
)

type AccountRepository struct {
	pool *pgxpool.Pool
}

func (r *AccountRepository) GetByID(ctx context.Context, companyProfileID, accountID uuid.UUID) (*models.AccountOverview, bool, error) {
	const query = `
SELECT
    a.id, a.name, a.domain, a.web_url, a.industry, a.hq, a.employees,
    a.revenue_m, a.revenue_currency, a.founded, a.description, a.lifecycle,
    latest.icp_score, latest.icp_fit, latest.signal_score, latest.resonance_score,
    latest.tier, latest.why_this_account, latest.why_now, latest.next_best_action
FROM target_account AS a
LEFT JOIN LATERAL (
    SELECT icp_score, icp_fit, signal_score, resonance_score, tier,
           why_this_account, why_now, next_best_action
    FROM account_analysis
    WHERE account_id = a.id
    ORDER BY created_at DESC, id DESC
    LIMIT 1
) AS latest ON TRUE
WHERE a.id = $1 AND a.company_profile_id = $2`

	var account models.AccountOverview
	var industry, hq, description *string
	var employees *int32
	var revenueM, icpScore, signalScore, resonanceScore *string
	var revenueCurrency *string
	var founded *int16
	var icpFit, tier, whyThisAccount, whyNow *string
	var nextBestAction []byte

	err := r.pool.QueryRow(ctx, query, accountID, companyProfileID).Scan(
		&account.ID, &account.Name, &account.Domain, &account.WebURL, &industry, &hq, &employees,
		&revenueM, &revenueCurrency, &founded, &description, &account.Lifecycle,
		&icpScore, &icpFit, &signalScore, &resonanceScore, &tier, &whyThisAccount, &whyNow, &nextBestAction,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, false, nil
		}
		return nil, false, fmt.Errorf("query account overview: %w", err)
	}
	account.Industry, account.HQ, account.Employees = industry, hq, employees
	account.Founded, account.Description = founded, description
	account.Revenue, err = parseRevenue(revenueM, revenueCurrency)
	if err != nil {
		return nil, false, fmt.Errorf("parse account revenue %s: %w", account.ID, err)
	}
	if tier != nil {
		if icpScore == nil || icpFit == nil || signalScore == nil || resonanceScore == nil {
			return nil, false, fmt.Errorf("account %s has incomplete analysis data", account.ID)
		}
		analysis, err := parseOverviewAnalysis(*icpScore, *icpFit, *signalScore, *resonanceScore, *tier, whyThisAccount, whyNow, nextBestAction)
		if err != nil {
			return nil, false, fmt.Errorf("parse account overview %s: %w", account.ID, err)
		}
		account.Analysis = analysis
	}
	return &account, true, nil
}

func (r *AccountRepository) ListSignals(ctx context.Context, companyProfileID, accountID uuid.UUID) ([]models.Signal, bool, error) {
	const query = `
SELECT
    a.id,
    s.id, s.type, s.title, s.body, s.strength, s.relevance,
    s.signal_date, s.signal_date_raw, s.freshness_label,
    s.evidence_status, s.verified, s.is_active, s.score_eligible,
    s.source_document_id,
    sd.source_name, sd.source_type, sd.url
FROM target_account AS a
LEFT JOIN signal AS s ON s.account_id = a.id
LEFT JOIN source_document AS sd
    ON sd.id = s.source_document_id
   AND sd.account_id = a.id
WHERE a.id = $1 AND a.company_profile_id = $2
ORDER BY s.signal_date DESC NULLS LAST, s.created_at DESC NULLS LAST, s.id DESC`

	rows, err := r.pool.Query(ctx, query, accountID, companyProfileID)
	if err != nil {
		return nil, false, fmt.Errorf("query account signals: %w", err)
	}
	defer rows.Close()

	items := make([]models.Signal, 0)
	accountFound := false
	for rows.Next() {
		var accountIDMarker uuid.UUID
		var signalID *uuid.UUID
		var signal models.Signal
		var body, relevance, signalDateRaw, freshnessLabel *string
		var signalType, signalTitle, strength, evidenceStatus *string
		var verified, isActive, scoreEligible *bool
		var signalDate *time.Time
		var sourceDocumentID *uuid.UUID
		var sourceName, sourceType *string
		var sourceURL *string

		if err := rows.Scan(&accountIDMarker, &signalID, &signalType, &signalTitle, &body, &strength, &relevance,
			&signalDate, &signalDateRaw, &freshnessLabel, &evidenceStatus, &verified,
			&isActive, &scoreEligible, &sourceDocumentID, &sourceName, &sourceType, &sourceURL); err != nil {
			return nil, false, fmt.Errorf("scan account signal: %w", err)
		}
		_ = accountIDMarker
		accountFound = true
		if signalID == nil {
			continue
		}
		if signalType == nil || signalTitle == nil || strength == nil || evidenceStatus == nil ||
			verified == nil || isActive == nil || scoreEligible == nil {
			return nil, false, fmt.Errorf("signal %s has missing required persisted fields", *signalID)
		}
		signal.ID = *signalID
		signal.Type, signal.Title, signal.Strength = *signalType, *signalTitle, *strength
		signal.EvidenceStatus = *evidenceStatus
		signal.Verified, signal.IsActive, signal.ScoreEligible = *verified, *isActive, *scoreEligible
		signal.Body, signal.Relevance = body, relevance
		signal.SignalDateRaw, signal.FreshnessLabel = signalDateRaw, freshnessLabel
		if signalDate != nil {
			formatted := signalDate.Format("2006-01-02")
			signal.SignalDate = &formatted
		}
		if sourceDocumentID != nil {
			if sourceURL == nil {
				return nil, false, fmt.Errorf("signal %s references an invalid source document", signal.ID)
			}
			signal.Source = &models.SignalSource{Name: sourceName, Type: sourceType, URL: *sourceURL}
		}
		items = append(items, signal)
	}
	if err := rows.Err(); err != nil {
		return nil, false, fmt.Errorf("iterate account signals: %w", err)
	}
	return items, accountFound, nil
}

func (r *AccountRepository) GetCommunicationDNA(ctx context.Context, companyProfileID, accountID uuid.UUID) (*models.CommunicationDNA, bool, bool, error) {
	const query = `
SELECT a.id, d.id, d.tone, d.vocabulary, d.value_propositions, d.problem_framing,
       d.proof_style, d.cta_patterns, d.recurring_phrases, d.do_rules, d.dont_rules, d.created_at
FROM target_account a
LEFT JOIN LATERAL (
  SELECT id, tone, vocabulary, value_propositions, problem_framing, proof_style,
         cta_patterns, recurring_phrases, do_rules, dont_rules, created_at
  FROM communication_dna WHERE account_id = a.id
  ORDER BY created_at DESC, id DESC LIMIT 1
) d ON TRUE
WHERE a.id = $1 AND a.company_profile_id = $2`
	var marker, dnaID *uuid.UUID
	var tone, vocabulary, propositions, problem, proof, cta, phrases, doRules, dontRules []byte
	var createdAt *time.Time
	err := r.pool.QueryRow(ctx, query, accountID, companyProfileID).Scan(&marker, &dnaID, &tone, &vocabulary, &propositions, &problem, &proof, &cta, &phrases, &doRules, &dontRules, &createdAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, false, false, nil
		}
		return nil, false, false, fmt.Errorf("query communication DNA: %w", err)
	}
	if dnaID == nil {
		return nil, true, false, nil
	}
	for name, value := range map[string][]byte{"tone": tone, "vocabulary": vocabulary, "value_propositions": propositions, "problem_framing": problem, "proof_style": proof, "cta_patterns": cta, "recurring_phrases": phrases, "do_rules": doRules, "dont_rules": dontRules} {
		if value == nil {
			return nil, false, false, fmt.Errorf("communication DNA %s is null", name)
		}
	}
	if createdAt == nil {
		return nil, false, false, fmt.Errorf("communication DNA created_at is null")
	}
	dna, err := parseCommunicationDNA(*dnaID, accountID, tone, vocabulary, propositions, problem, proof, cta, phrases, doRules, dontRules, *createdAt)
	if err != nil {
		return nil, false, false, err
	}
	return dna, true, true, nil
}

func parseCommunicationDNA(id, accountID uuid.UUID, tone, vocabulary, propositions, problem, proof, cta, phrases, doRules, dontRules []byte, createdAt time.Time) (*models.CommunicationDNA, error) {
	var result models.CommunicationDNA
	result.ID, result.AccountID, result.CreatedAt = id, accountID, createdAt.UTC().Format(time.RFC3339)
	decode := func(name string, raw []byte, target any) error {
		if err := json.Unmarshal(raw, target); err != nil {
			return fmt.Errorf("invalid communication DNA %s: %w", name, err)
		}
		return nil
	}
	if err := decode("tone", tone, &result.Tone); err != nil {
		return nil, err
	}
	if err := validateStyle(result.Tone, "tone"); err != nil {
		return nil, err
	}
	if err := decode("vocabulary", vocabulary, &result.Vocabulary); err != nil {
		return nil, err
	}
	if err := validateStatus(result.Vocabulary.Status, "vocabulary.status"); err != nil {
		return nil, err
	}
	for i := range result.Vocabulary.Terms {
		if result.Vocabulary.Terms[i].Term == "" {
			return nil, fmt.Errorf("invalid vocabulary term")
		}
		if err := validateSources(result.Vocabulary.Terms[i].Sources); err != nil {
			return nil, err
		}
	}
	if err := decode("value_propositions", propositions, &result.ValuePropositions); err != nil {
		return nil, err
	}
	for i := range result.ValuePropositions {
		if result.ValuePropositions[i].Quote == "" {
			return nil, fmt.Errorf("invalid value proposition")
		}
		if err := validateStatus(result.ValuePropositions[i].Status, "value_proposition.status"); err != nil {
			return nil, err
		}
		if err := validateSources(result.ValuePropositions[i].Sources); err != nil {
			return nil, err
		}
	}
	if err := decode("problem_framing", problem, &result.ProblemFraming); err != nil {
		return nil, err
	}
	if err := validateStatus(result.ProblemFraming.Status, "problem_framing.status"); err != nil {
		return nil, err
	}
	if err := validateSources(result.ProblemFraming.Sources); err != nil {
		return nil, err
	}
	if err := decode("proof_style", proof, &result.ProofStyle); err != nil {
		return nil, err
	}
	if err := validateStyle(result.ProofStyle, "proof_style"); err != nil {
		return nil, err
	}
	if err := decode("cta_patterns", cta, &result.CTAPatterns); err != nil {
		return nil, err
	}
	if err := validateStatus(result.CTAPatterns.Status, "cta_patterns.status"); err != nil {
		return nil, err
	}
	if err := validateSources(result.CTAPatterns.Sources); err != nil {
		return nil, err
	}
	if err := decode("recurring_phrases", phrases, &result.RecurringPhrases); err != nil {
		return nil, err
	}
	for i := range result.RecurringPhrases {
		if result.RecurringPhrases[i].Quote == "" {
			return nil, fmt.Errorf("invalid recurring phrase")
		}
		if err := validateStatus(result.RecurringPhrases[i].Status, "recurring_phrase.status"); err != nil {
			return nil, err
		}
		if err := validateSources(result.RecurringPhrases[i].Sources); err != nil {
			return nil, err
		}
	}
	if err := decodeStringArray("do_rules", doRules, &result.DoRules); err != nil {
		return nil, err
	}
	if err := decodeStringArray("dont_rules", dontRules, &result.DontRules); err != nil {
		return nil, err
	}
	return &result, nil
}

func validateStatus(status, field string) error {
	if status != "SOURCE_BACKED" && status != "DERIVED" && status != "INSUFFICIENT_DATA" {
		return fmt.Errorf("invalid %s status", field)
	}
	return nil
}
func validateSources(sources []models.SignalSource) error {
	if sources == nil {
		return fmt.Errorf("DNA sources must be an array")
	}
	return nil
}
func validateStyle(style models.DNAStyle, field string) error {
	if err := validateStatus(style.Status, field+".status"); err != nil {
		return err
	}
	return validateSources(style.Sources)
}
func decodeStringArray(name string, raw []byte, out *[]string) error {
	if err := json.Unmarshal(raw, out); err != nil || *out == nil {
		return fmt.Errorf("invalid communication DNA %s", name)
	}
	return nil
}

func NewAccountRepository(pool *pgxpool.Pool) *AccountRepository {
	return &AccountRepository{pool: pool}
}

func (r *AccountRepository) List(ctx context.Context, companyProfileID uuid.UUID) ([]models.Account, error) {
	const query = `
SELECT
    a.id,
    a.name,
    a.industry,
    a.hq,
    a.lifecycle,
    latest.icp_score,
    latest.icp_fit,
    latest.signal_score,
    latest.resonance_score,
    latest.tier,
    latest.next_best_action,
    COALESCE(active_signal.active_count, 0) AS active_signal_count
FROM target_account AS a
LEFT JOIN LATERAL (
    SELECT icp_score, icp_fit, signal_score, resonance_score, tier, next_best_action
    FROM account_analysis
    WHERE account_id = a.id
    ORDER BY created_at DESC, id DESC
    LIMIT 1
) AS latest ON TRUE
LEFT JOIN (
    SELECT account_id, COUNT(*)::int AS active_count
    FROM signal
    WHERE is_active = TRUE
    GROUP BY account_id
) AS active_signal ON active_signal.account_id = a.id
WHERE a.company_profile_id = $1
ORDER BY
    CASE
        WHEN latest.tier = 'Focus Accounts' THEN 1
        WHEN latest.tier = 'Tier 1' THEN 2
        WHEN latest.tier = 'Tier 2' THEN 3
        WHEN latest.tier = 'Below ICP' THEN 4
        ELSE 5
    END,
    CASE WHEN latest.tier IS NULL THEN 0 ELSE latest.resonance_score END DESC NULLS LAST,
    a.name ASC,
    a.id ASC`

	rows, err := r.pool.Query(ctx, query, companyProfileID)
	if err != nil {
		return nil, fmt.Errorf("query accounts: %w", err)
	}
	defer rows.Close()

	accounts := make([]models.Account, 0)
	for rows.Next() {
		var account models.Account
		var industry, hq *string
		var icpScore, signalScore, resonanceScore *string
		var icpFit, tier *string
		var nextBestAction []byte
		var activeSignalCount int

		if err := rows.Scan(&account.ID, &account.Name, &industry, &hq, &account.Lifecycle,
			&icpScore, &icpFit, &signalScore, &resonanceScore, &tier, &nextBestAction, &activeSignalCount); err != nil {
			return nil, fmt.Errorf("scan account: %w", err)
		}
		account.Industry = industry
		account.HQ = hq
		account.ActiveSignalCount = activeSignalCount

		if tier != nil {
			if icpScore == nil || icpFit == nil || signalScore == nil || resonanceScore == nil {
				return nil, fmt.Errorf("account %s has incomplete analysis data", account.ID)
			}
			analysis, err := parseAnalysis(*icpScore, *icpFit, *signalScore, *resonanceScore, *tier, nextBestAction)
			if err != nil {
				return nil, fmt.Errorf("parse analysis for account %s: %w", account.ID, err)
			}
			account.Analysis = analysis
		}
		accounts = append(accounts, account)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate accounts: %w", err)
	}
	return accounts, nil
}

func parseRevenue(amount, currency *string) (*models.Revenue, error) {
	if amount == nil && currency == nil {
		return nil, nil
	}
	revenue := &models.Revenue{Currency: currency}
	if amount != nil {
		parsed, err := strconv.ParseFloat(*amount, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid revenue_m: %w", err)
		}
		revenue.AmountM = &parsed
	}
	return revenue, nil
}

func parseOverviewAnalysis(icpScore, icpFit, signalScore, resonanceScore, tier string, whyThisAccount, whyNow *string, rawAction []byte) (*models.AccountOverviewAnalysis, error) {
	parse := func(value, field string) (float64, error) {
		parsed, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return 0, fmt.Errorf("invalid %s: %w", field, err)
		}
		return parsed, nil
	}
	result := &models.AccountOverviewAnalysis{ICPFit: icpFit, Tier: tier, WhyThisAccount: whyThisAccount, WhyNow: whyNow}
	var err error
	if result.ICPScore, err = parse(icpScore, "icp_score"); err != nil {
		return nil, err
	}
	if result.SignalScore, err = parse(signalScore, "signal_score"); err != nil {
		return nil, err
	}
	if result.ResonanceScore, err = parse(resonanceScore, "resonance_score"); err != nil {
		return nil, err
	}
	if rawAction == nil {
		return result, nil
	}
	var action struct {
		Action     *string `json:"action"`
		Rationale  *string `json:"rationale"`
		TimeWindow *string `json:"timeWindow"`
		Priority   *string `json:"priority"`
	}
	if err := json.Unmarshal(rawAction, &action); err != nil || action.Action == nil {
		if err == nil {
			err = fmt.Errorf("missing action")
		}
		return nil, fmt.Errorf("invalid next_best_action: %w", err)
	}
	result.NextBestAction = &models.NextBestAction{Action: *action.Action, Rationale: action.Rationale, TimeWindow: action.TimeWindow, Priority: action.Priority}
	return result, nil
}

func parseAnalysis(icpScore, icpFit, signalScore, resonanceScore, tier string, rawAction []byte) (*models.AccountAnalysis, error) {
	parse := func(value, field string) (float64, error) {
		parsed, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return 0, fmt.Errorf("invalid %s: %w", field, err)
		}
		return parsed, nil
	}
	result := &models.AccountAnalysis{ICPFit: icpFit, Tier: tier}
	var err error
	if result.ICPScore, err = parse(icpScore, "icp_score"); err != nil {
		return nil, err
	}
	if result.SignalScore, err = parse(signalScore, "signal_score"); err != nil {
		return nil, err
	}
	if result.ResonanceScore, err = parse(resonanceScore, "resonance_score"); err != nil {
		return nil, err
	}
	if rawAction == nil {
		return result, nil
	}
	var action struct {
		Action *string `json:"action"`
	}
	if err := json.Unmarshal(rawAction, &action); err != nil || action.Action == nil {
		if err == nil {
			err = fmt.Errorf("missing action")
		}
		return nil, fmt.Errorf("invalid next_best_action: %w", err)
	}
	result.NextBestAction = action.Action
	return result, nil
}
