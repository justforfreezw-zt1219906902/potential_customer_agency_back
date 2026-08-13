package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

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
