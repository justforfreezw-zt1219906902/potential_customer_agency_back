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

func TestAccountServiceGetReturnsNotFoundForScopedMissingAccount(t *testing.T) {
	service := NewAccountService(accountReaderStub{}, uuid.New())
	_, err := service.Get(context.Background(), uuid.New())
	appErr, ok := errors.AsAppError(err)
	if !ok || appErr.Status != 404 || appErr.Message != "account not found" {
		t.Fatalf("unexpected error: %#v", err)
	}
}
