package service

import (
	"context"

	"github.com/google/uuid"
	apperrors "github.com/justforfreezw-zt1219906902/potential_customer_agency_back/internal/errors"
	"github.com/justforfreezw-zt1219906902/potential_customer_agency_back/internal/models"
)

type AccountReader interface {
	List(ctx context.Context, companyProfileID uuid.UUID) ([]models.Account, error)
	GetByID(ctx context.Context, companyProfileID, accountID uuid.UUID) (*models.AccountOverview, bool, error)
	ListSignals(ctx context.Context, companyProfileID, accountID uuid.UUID) ([]models.Signal, bool, error)
}

func (s *AccountService) ListSignals(ctx context.Context, accountID uuid.UUID) (models.AccountSignalsResponse, error) {
	items, found, err := s.repository.ListSignals(ctx, s.companyID, accountID)
	if err != nil {
		return models.AccountSignalsResponse{}, err
	}
	if !found {
		return models.AccountSignalsResponse{}, apperrors.NotFound("account not found", nil)
	}

	response := models.AccountSignalsResponse{
		Summary: models.SignalSummary{ByType: make(map[string]int)},
		Items:   items,
	}
	for _, item := range items {
		response.Summary.Total++
		if item.IsActive {
			response.Summary.Active++
		}
		response.Summary.ByType[item.Type]++
	}
	return response, nil
}

func (s *AccountService) Get(ctx context.Context, accountID uuid.UUID) (models.AccountOverview, error) {
	account, found, err := s.repository.GetByID(ctx, s.companyID, accountID)
	if err != nil {
		return models.AccountOverview{}, err
	}
	if !found {
		return models.AccountOverview{}, apperrors.NotFound("account not found", nil)
	}
	return *account, nil
}

type AccountService struct {
	repository AccountReader
	companyID  uuid.UUID
}

func NewAccountService(repository AccountReader, companyID uuid.UUID) *AccountService {
	return &AccountService{repository: repository, companyID: companyID}
}

func (s *AccountService) List(ctx context.Context) (models.AccountListResponse, error) {
	items, err := s.repository.List(ctx, s.companyID)
	if err != nil {
		return models.AccountListResponse{}, err
	}
	return models.AccountListResponse{Items: items}, nil
}
