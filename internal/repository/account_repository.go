package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/justforfreezw-zt1219906902/potential_customer_agency_back/internal/models"
)

type AccountRepository struct {
	pool *pgxpool.Pool
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
