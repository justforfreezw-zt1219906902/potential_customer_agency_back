package service

import (
	"context"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/justforfreezw-zt1219906902/potential_customer_agency_back/internal/models"
)

type SignalPulseReader interface {
	ListSignalPulseRows(context.Context, uuid.UUID) ([]models.SignalPulseRow, error)
}

func (s *AccountService) SignalPulse(ctx context.Context) (models.SignalPulseResponse, error) {
	rows, err := s.repository.ListSignalPulseRows(ctx, s.companyID)
	if err != nil {
		return models.SignalPulseResponse{}, err
	}
	return BuildSignalPulse(rows, time.Now().UTC()), nil
}

func BuildSignalPulse(rows []models.SignalPulseRow, now time.Time) models.SignalPulseResponse {
	today := dateOnly(now.UTC())
	accounts := make(map[uuid.UUID]*pulseAccumulator)
	for _, row := range rows {
		account := accounts[row.AccountID]
		if account == nil {
			account = &pulseAccumulator{row: row, previews: make([]models.SignalPulsePreview, 0)}
			accounts[row.AccountID] = account
		}
		if row.SignalID == nil {
			continue
		}
		if !*row.IsActive {
			continue
		}
		account.active++
		if *row.SignalStrength == "high" {
			account.high++
		}
		if row.SignalDate == nil {
			account.hasUndated = true
		} else {
			d := dateOnly(*row.SignalDate)
			if account.latest == nil || d.After(*account.latest) {
				account.latest = &d
			}
			if !d.Before(today.AddDate(0, 0, -6)) && !d.After(today) {
				account.newThisWeek++
			}
		}
		if len(account.previews) < 5 {
			previewDate := formatDate(row.SignalDate)
			account.previews = append(account.previews, models.SignalPulsePreview{ID: *row.SignalID, Type: *row.SignalType, Title: *row.SignalTitle, Strength: *row.SignalStrength, SignalDate: previewDate})
		}
	}
	result := models.SignalPulseResponse{Accounts: make([]models.SignalPulseAccount, 0), Metrics: models.SignalPulseMetrics{}}
	for _, account := range accounts {
		cold := account.active == 0 || (!account.hasUndated && account.latest != nil && account.latest.Before(today.AddDate(0, 0, -56)))
		urgency := "warm"
		if account.high >= 2 {
			urgency = "hot"
		} else if cold {
			urgency = "cold"
		}
		item := models.SignalPulseAccount{AccountID: account.row.AccountID, Name: account.row.Name, Industry: account.row.Industry, Tier: account.row.Tier, Urgency: urgency, ActiveSignalCount: account.active, HighActiveSignalCount: account.high, NextBestAction: account.row.NextBestAction, Signals: account.previews}
		item.LatestActiveSignalDate = formatDate(account.latest)
		result.Accounts = append(result.Accounts, item)
		result.Metrics.ActiveSignals += account.active
		result.Metrics.NewThisWeek += account.newThisWeek
		if account.high >= 2 {
			result.Metrics.HotAccounts++
		}
		if cold {
			result.Metrics.GoingCold++
		}
	}
	sort.Slice(result.Accounts, func(i, j int) bool {
		a, b := result.Accounts[i], result.Accounts[j]
		rank := func(v string) int {
			if v == "hot" {
				return 1
			}
			if v == "warm" {
				return 2
			}
			return 3
		}
		if rank(a.Urgency) != rank(b.Urgency) {
			return rank(a.Urgency) < rank(b.Urgency)
		}
		if a.HighActiveSignalCount != b.HighActiveSignalCount {
			return a.HighActiveSignalCount > b.HighActiveSignalCount
		}
		if a.ActiveSignalCount != b.ActiveSignalCount {
			return a.ActiveSignalCount > b.ActiveSignalCount
		}
		if a.LatestActiveSignalDate == nil && b.LatestActiveSignalDate != nil {
			return false
		}
		if a.LatestActiveSignalDate != nil && b.LatestActiveSignalDate == nil {
			return true
		}
		if a.LatestActiveSignalDate != nil && *a.LatestActiveSignalDate != *b.LatestActiveSignalDate {
			return *a.LatestActiveSignalDate > *b.LatestActiveSignalDate
		}
		if a.Name != b.Name {
			return a.Name < b.Name
		}
		return a.AccountID.String() < b.AccountID.String()
	})
	return result
}

type pulseAccumulator struct {
	row                       models.SignalPulseRow
	active, high, newThisWeek int
	latest                    *time.Time
	hasUndated                bool
	previews                  []models.SignalPulsePreview
}

func dateOnly(value time.Time) time.Time {
	y, m, d := value.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}
func formatDate(value *time.Time) *string {
	if value == nil {
		return nil
	}
	result := dateOnly(*value).Format("2006-01-02")
	return &result
}
