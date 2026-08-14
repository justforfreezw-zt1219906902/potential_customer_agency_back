package handler

import (
	"context"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/justforfreezw-zt1219906902/potential_customer_agency_back/internal/errors"
	"github.com/justforfreezw-zt1219906902/potential_customer_agency_back/internal/models"
)

type AccountLister interface {
	List(ctx context.Context) (models.AccountListResponse, error)
	Get(ctx context.Context, accountID uuid.UUID) (models.AccountOverview, error)
	ListSignals(ctx context.Context, accountID uuid.UUID) (models.AccountSignalsResponse, error)
	GetCommunicationDNA(ctx context.Context, accountID uuid.UUID) (models.CommunicationDNAResponse, error)
	SignalPulse(ctx context.Context) (models.SignalPulseResponse, error)
	ListDNAPortfolio(ctx context.Context) (models.DNAPortfolioResponse, error)
	CompareDNAPortfolio(ctx context.Context, ids []uuid.UUID) (models.DNACompareResponse, error)
}

func (h *AccountHandler) ListDNAPortfolio(c *gin.Context) {
	response, err := h.service.ListDNAPortfolio(c.Request.Context())
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, response)
}

func (h *AccountHandler) CompareDNAPortfolio(c *gin.Context) {
	var request struct {
		AccountIDs []string `json:"accountIds"`
	}
	if err := c.ShouldBindJSON(&request); err != nil || request.AccountIDs == nil || len(request.AccountIDs) < 2 {
		_ = c.Error(errors.BadRequest("accountIds must contain at least 2 valid UUIDs", err))
		return
	}
	ids := make([]uuid.UUID, len(request.AccountIDs))
	seen := map[uuid.UUID]struct{}{}
	for i, value := range request.AccountIDs {
		id, err := uuid.Parse(value)
		if err != nil {
			_ = c.Error(errors.BadRequest("accountIds must contain valid UUIDs", err))
			return
		}
		if _, ok := seen[id]; ok {
			_ = c.Error(errors.BadRequest("accountIds must not contain duplicates", nil))
			return
		}
		seen[id] = struct{}{}
		ids[i] = id
	}
	response, err := h.service.CompareDNAPortfolio(c.Request.Context(), ids)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, response)
}

func (h *AccountHandler) GetCommunicationDNA(c *gin.Context) {
	accountID, err := uuid.Parse(c.Param("accountId"))
	if err != nil {
		_ = c.Error(errors.BadRequest("accountId must be a valid UUID", err))
		return
	}
	response, err := h.service.GetCommunicationDNA(c.Request.Context(), accountID)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, response)
}

func (h *AccountHandler) SignalPulse(c *gin.Context) {
	response, err := h.service.SignalPulse(c.Request.Context())
	if err != nil {
		_ = c.Error(errors.Internal("failed to load signal pulse", err))
		return
	}
	c.JSON(http.StatusOK, response)
}

func (h *AccountHandler) ListSignals(c *gin.Context) {
	accountID, err := uuid.Parse(c.Param("accountId"))
	if err != nil {
		_ = c.Error(errors.BadRequest("accountId must be a valid UUID", err))
		return
	}
	response, err := h.service.ListSignals(c.Request.Context(), accountID)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, response)
}

func (h *AccountHandler) GetAccount(c *gin.Context) {
	accountID, err := uuid.Parse(c.Param("accountId"))
	if err != nil {
		_ = c.Error(errors.BadRequest("accountId must be a valid UUID", err))
		return
	}
	response, err := h.service.Get(c.Request.Context(), accountID)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, response)
}

type AccountHandler struct {
	service AccountLister
	logger  *log.Logger
}

func NewAccountHandler(service AccountLister, logger *log.Logger) *AccountHandler {
	return &AccountHandler{service: service, logger: logger}
}

func (h *AccountHandler) ListAccounts(c *gin.Context) {
	response, err := h.service.List(c.Request.Context())
	if err != nil {
		h.logger.Printf("failed to list accounts: %v", err)
		_ = c.Error(errors.Internal("failed to load accounts", err))
		return
	}
	c.JSON(http.StatusOK, response)
}
