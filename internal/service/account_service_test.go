package service

import (
	"context"
	"testing"

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

func TestAccountServiceGetReturnsNotFoundForScopedMissingAccount(t *testing.T) {
	service := NewAccountService(accountReaderStub{}, uuid.New())
	_, err := service.Get(context.Background(), uuid.New())
	appErr, ok := errors.AsAppError(err)
	if !ok || appErr.Status != 404 || appErr.Message != "account not found" {
		t.Fatalf("unexpected error: %#v", err)
	}
}
