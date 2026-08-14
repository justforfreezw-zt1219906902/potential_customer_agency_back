package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/justforfreezw-zt1219906902/potential_customer_agency_back/internal/models"
)

func (r *AccountRepository) ListSignalPulseRows(ctx context.Context, companyProfileID uuid.UUID) ([]models.SignalPulseRow, error) {
	const query = `
SELECT a.id, a.name, a.industry, latest.tier, latest.next_best_action,
       s.id, s.type, s.title, s.strength, s.signal_date, s.created_at, s.is_active
FROM target_account a
LEFT JOIN LATERAL (
  SELECT tier, next_best_action FROM account_analysis
  WHERE account_id = a.id ORDER BY created_at DESC, id DESC LIMIT 1
) latest ON TRUE
LEFT JOIN signal s ON s.account_id = a.id
WHERE a.company_profile_id = $1
ORDER BY a.id, s.signal_date DESC NULLS LAST, s.created_at DESC NULLS LAST, s.id DESC`
	rows, err := r.pool.Query(ctx, query, companyProfileID)
	if err != nil {
		return nil, fmt.Errorf("query signal pulse: %w", err)
	}
	defer rows.Close()
	result := make([]models.SignalPulseRow, 0)
	for rows.Next() {
		var row models.SignalPulseRow
		var rawAction []byte
		if err := rows.Scan(&row.AccountID, &row.Name, &row.Industry, &row.Tier, &rawAction, &row.SignalID, &row.SignalType, &row.SignalTitle, &row.SignalStrength, &row.SignalDate, &row.SignalCreatedAt, &row.IsActive); err != nil {
			return nil, fmt.Errorf("scan signal pulse: %w", err)
		}
		if row.Tier != nil {
			if rawAction == nil {
				row.NextBestAction = nil
			} else {
				var action struct {
					Action *string `json:"action"`
				}
				if err := json.Unmarshal(rawAction, &action); err != nil || action.Action == nil {
					return nil, fmt.Errorf("invalid signal pulse next_best_action")
				}
				row.NextBestAction = action.Action
			}
		}
		if row.SignalID != nil && (row.SignalType == nil || row.SignalTitle == nil || row.SignalStrength == nil || row.SignalCreatedAt == nil || row.IsActive == nil) {
			return nil, fmt.Errorf("signal %s has missing required fields", row.SignalID)
		}
		result = append(result, row)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate signal pulse: %w", err)
	}
	return result, nil
}
