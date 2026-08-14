package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/justforfreezw-zt1219906902/potential_customer_agency_back/internal/errors"
	"github.com/justforfreezw-zt1219906902/potential_customer_agency_back/internal/models"
)

type accountReaderStub struct{}

func (accountReaderStub) List(context.Context, uuid.UUID) ([]models.Account, error) { return nil, nil }
func (accountReaderStub) GetByID(context.Context, uuid.UUID, uuid.UUID) (*models.AccountOverview, bool, error) {
	return nil, false, nil
}
func (accountReaderStub) ListSignals(context.Context, uuid.UUID, uuid.UUID) ([]models.Signal, bool, error) {
	return nil, false, nil
}
func (accountReaderStub) GetCommunicationDNA(context.Context, uuid.UUID, uuid.UUID) (*models.CommunicationDNA, bool, bool, error) {
	return nil, true, false, nil
}
func (accountReaderStub) ListSignalPulseRows(context.Context, uuid.UUID) ([]models.SignalPulseRow, error) {
	return nil, nil
}

func TestAccountServiceListSignalsBuildsSummary(t *testing.T) {
	reader := signalReaderStub{items: []models.Signal{
		{Type: "Job Posting", IsActive: true},
		{Type: "Job Posting", IsActive: false},
		{Type: "News & Events", IsActive: true},
	}}
	service := NewAccountService(reader, uuid.New())
	response, err := service.ListSignals(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if response.Summary.Total != 3 || response.Summary.Active != 2 {
		t.Fatalf("unexpected summary: %+v", response.Summary)
	}
	if response.Summary.ByType["Job Posting"] != 2 || response.Summary.ByType["News & Events"] != 1 {
		t.Fatalf("unexpected byType: %+v", response.Summary.ByType)
	}
}

type signalReaderStub struct{ items []models.Signal }

func (s signalReaderStub) List(context.Context, uuid.UUID) ([]models.Account, error) { return nil, nil }
func (s signalReaderStub) GetByID(context.Context, uuid.UUID, uuid.UUID) (*models.AccountOverview, bool, error) {
	return nil, true, nil
}
func (s signalReaderStub) ListSignals(context.Context, uuid.UUID, uuid.UUID) ([]models.Signal, bool, error) {
	return s.items, true, nil
}
func (s signalReaderStub) GetCommunicationDNA(context.Context, uuid.UUID, uuid.UUID) (*models.CommunicationDNA, bool, bool, error) {
	return nil, true, false, nil
}
func (s signalReaderStub) ListSignalPulseRows(context.Context, uuid.UUID) ([]models.SignalPulseRow, error) {
	return nil, nil
}

func TestAccountServiceMissingDNAReturnsNull(t *testing.T) {
	response, err := NewAccountService(accountReaderStub{}, uuid.New()).GetCommunicationDNA(context.Background(), uuid.New())
	if err != nil || response.Data != nil {
		t.Fatalf("response=%+v err=%v", response, err)
	}
}

func TestDedupeSignalSourcesPreservesFirstURL(t *testing.T) {
	a, b := "A", "B"
	signals := []models.Signal{{Source: &models.SignalSource{URL: "https://a", Name: &a}}, {Source: &models.SignalSource{URL: "https://a"}}, {Source: &models.SignalSource{URL: "https://b", Name: &b}}, {Source: nil}}
	result := dedupeSignalSources(signals)
	if len(result) != 2 || result[0].URL != "https://a" || result[1].URL != "https://b" {
		t.Fatalf("unexpected sources: %+v", result)
	}
}

func TestBuildSignalPulseRules(t *testing.T) {
	today := time.Date(2026, 8, 14, 12, 0, 0, 0, time.UTC)
	date := func(days int) *time.Time { value := today.AddDate(0, 0, days); return &value }
	active := true
	inactive := false
	created := today
	high, medium := "high", "medium"
	typ, title := "Job Posting", "Hiring"
	rows := []models.SignalPulseRow{
		{AccountID: uuid.MustParse("00000000-0000-0000-0000-000000000101"), Name: "Hot", SignalID: ptrUUID("00000000-0000-0000-0000-000000002001"), SignalType: &typ, SignalTitle: &title, SignalStrength: &high, SignalDate: date(0), SignalCreatedAt: &created, IsActive: &active},
		{AccountID: uuid.MustParse("00000000-0000-0000-0000-000000000101"), Name: "Hot", SignalID: ptrUUID("00000000-0000-0000-0000-000000002002"), SignalType: &typ, SignalTitle: &title, SignalStrength: &high, SignalDate: date(-6), SignalCreatedAt: &created, IsActive: &active},
		{AccountID: uuid.MustParse("00000000-0000-0000-0000-000000000102"), Name: "Unknown Date", SignalID: ptrUUID("00000000-0000-0000-0000-000000002003"), SignalType: &typ, SignalTitle: &title, SignalStrength: &medium, SignalDate: nil, SignalCreatedAt: &created, IsActive: &active},
		{AccountID: uuid.MustParse("00000000-0000-0000-0000-000000000103"), Name: "Cold", SignalID: ptrUUID("00000000-0000-0000-0000-000000002004"), SignalType: &typ, SignalTitle: &title, SignalStrength: &medium, SignalDate: date(-57), SignalCreatedAt: &created, IsActive: &active},
		{AccountID: uuid.MustParse("00000000-0000-0000-0000-000000000104"), Name: "Empty", SignalID: nil, IsActive: nil},
		{AccountID: uuid.MustParse("00000000-0000-0000-0000-000000000105"), Name: "Inactive", SignalID: ptrUUID("00000000-0000-0000-0000-000000002005"), SignalType: &typ, SignalTitle: &title, SignalStrength: &high, SignalDate: date(0), SignalCreatedAt: &created, IsActive: &inactive},
	}
	response := BuildSignalPulse(rows, today)
	if response.Metrics.ActiveSignals != 4 || response.Metrics.NewThisWeek != 2 || response.Metrics.HotAccounts != 1 || response.Metrics.GoingCold != 3 {
		t.Fatalf("unexpected metrics: %+v", response.Metrics)
	}
	if response.Accounts[0].Urgency != "hot" || response.Accounts[1].Urgency != "warm" || response.Accounts[2].Urgency != "cold" {
		t.Fatalf("unexpected ordering: %+v", response.Accounts)
	}
}

func TestBuildSignalPulseDateBoundariesAndStrengthCounts(t *testing.T) {
	today := time.Date(2026, 8, 14, 12, 0, 0, 0, time.UTC)
	rows := []models.SignalPulseRow{
		pulseRow("00000000-0000-0000-0000-000000000201", "Boundary", "medium", today.AddDate(0, 0, -7), true),
		pulseRow("00000000-0000-0000-0000-000000000201", "Boundary", "low", today.AddDate(0, 0, 1), true),
		pulseRow("00000000-0000-0000-0000-000000000202", "Unknown date", "medium", today.AddDate(0, 0, -57), true),
		pulseRow("00000000-0000-0000-0000-000000000202", "Unknown date", "low", time.Time{}, true),
		pulseRow("00000000-0000-0000-0000-000000000203", "Inactive", "high", today, false),
	}
	response := BuildSignalPulse(rows, today)
	if response.Metrics.NewThisWeek != 0 {
		t.Fatalf("boundary/future signals counted as new: %+v", response.Metrics)
	}
	if response.Metrics.ActiveSignals != 4 {
		t.Fatalf("inactive signal counted: %+v", response.Metrics)
	}
	for _, account := range response.Accounts {
		if account.Name == "Unknown date" && account.Urgency != "warm" {
			t.Fatalf("undated active signal incorrectly cold: %+v", account)
		}
	}
	for _, account := range response.Accounts {
		if account.Name == "Boundary" && account.HighActiveSignalCount != 0 {
			t.Fatalf("medium/low signal increased high count: %+v", account)
		}
	}
}

func TestBuildSignalPulseLimitsPreviewsAndKeepsAllActiveCount(t *testing.T) {
	today := time.Date(2026, 8, 14, 12, 0, 0, 0, time.UTC)
	rows := make([]models.SignalPulseRow, 0, 6)
	for i := 0; i < 6; i++ {
		rows = append(rows, pulseRow("00000000-0000-0000-0000-000000000204", "Many Signals", "medium", today.AddDate(0, 0, -i), true))
	}
	response := BuildSignalPulse(rows, today)
	if len(response.Accounts) != 1 || len(response.Accounts[0].Signals) != 5 || response.Accounts[0].ActiveSignalCount != 6 {
		t.Fatalf("unexpected preview/count: %+v", response.Accounts)
	}
}

func TestBuildSignalPulseZeroRowsReturnsEmptyArrayAndZeroMetrics(t *testing.T) {
	response := BuildSignalPulse(nil, time.Date(2026, 8, 14, 0, 0, 0, 0, time.UTC))
	if len(response.Accounts) != 0 || response.Accounts == nil || response.Metrics != (models.SignalPulseMetrics{}) {
		t.Fatalf("unexpected empty response: %+v", response)
	}
}

func TestBuildSignalPulseIncludesAccountWithoutAnalysis(t *testing.T) {
	accountID := uuid.MustParse("00000000-0000-0000-0000-000000000299")
	response := BuildSignalPulse([]models.SignalPulseRow{{AccountID: accountID, Name: "Unanalyzed"}}, time.Date(2026, 8, 14, 0, 0, 0, 0, time.UTC))
	if len(response.Accounts) != 1 || response.Accounts[0].Tier != nil || response.Accounts[0].NextBestAction != nil {
		t.Fatalf("unexpected unanalyzed account: %+v", response.Accounts)
	}
}

func TestBuildSignalPulseDeterministicTieBreakers(t *testing.T) {
	today := time.Date(2026, 8, 14, 0, 0, 0, 0, time.UTC)
	rows := []models.SignalPulseRow{
		pulseRow("00000000-0000-0000-0000-000000000301", "Warm More Active", "medium", today, true),
		pulseRow("00000000-0000-0000-0000-000000000301", "Warm More Active", "low", today.AddDate(0, 0, -1), true),
		pulseRow("00000000-0000-0000-0000-000000000302", "Warm One", "medium", today, true),
		pulseRow("00000000-0000-0000-0000-000000000303", "Same Date Zulu", "medium", today, true),
		pulseRow("00000000-0000-0000-0000-000000000304", "Same Date Alpha", "medium", today, true),
		pulseRow("00000000-0000-0000-0000-000000000305", "Cold", "medium", today.AddDate(0, 0, -57), true),
	}
	high := "high"
	rows = append(rows, pulseRow("00000000-0000-0000-0000-000000000306", "Hot", high, today, true), pulseRow("00000000-0000-0000-0000-000000000306", "Hot", high, today.AddDate(0, 0, -1), true))
	response := BuildSignalPulse(rows, today)
	order := make([]string, len(response.Accounts))
	for i, account := range response.Accounts {
		order[i] = account.Name
	}
	expected := []string{"Hot", "Warm More Active", "Same Date Alpha", "Same Date Zulu", "Warm One", "Cold"}
	for i := range expected {
		if order[i] != expected[i] {
			t.Fatalf("unexpected order: %v", order)
		}
	}
}

func TestBuildSignalPulseUsesLatestActiveDateAsTieBreaker(t *testing.T) {
	today := time.Date(2026, 8, 14, 0, 0, 0, 0, time.UTC)
	rows := []models.SignalPulseRow{
		pulseRow("00000000-0000-0000-0000-000000000307", "Warm Older", "medium", today.AddDate(0, 0, -1), true),
		pulseRow("00000000-0000-0000-0000-000000000308", "Warm Newer", "medium", today, true),
	}
	response := BuildSignalPulse(rows, today)
	if len(response.Accounts) != 2 || response.Accounts[0].Name != "Warm Newer" || response.Accounts[1].Name != "Warm Older" {
		t.Fatalf("latest active date tie-breaker failed: %+v", response.Accounts)
	}
}

func pulseRow(accountID, name, strength string, date time.Time, active bool) models.SignalPulseRow {
	id := uuid.New()
	typ, title := "Job Posting", "Signal"
	created := time.Date(2026, 8, 14, 0, 0, 0, 0, time.UTC)
	var signalDate *time.Time
	if !date.IsZero() {
		signalDate = &date
	}
	return models.SignalPulseRow{AccountID: uuid.MustParse(accountID), Name: name, SignalID: &id, SignalType: &typ, SignalTitle: &title, SignalStrength: &strength, SignalDate: signalDate, SignalCreatedAt: &created, IsActive: &active}
}

func ptrUUID(value string) *uuid.UUID { parsed := uuid.MustParse(value); return &parsed }

func TestAccountServiceGetReturnsNotFoundForScopedMissingAccount(t *testing.T) {
	service := NewAccountService(accountReaderStub{}, uuid.New())
	_, err := service.Get(context.Background(), uuid.New())
	appErr, ok := errors.AsAppError(err)
	if !ok || appErr.Status != 404 || appErr.Message != "account not found" {
		t.Fatalf("unexpected error: %#v", err)
	}
}
