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
