package handler

import (
	"context"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/justforfreezw-zt1219906902/potential_customer_agency_back/internal/errors"
	"github.com/justforfreezw-zt1219906902/potential_customer_agency_back/internal/models"
)

type AccountLister interface {
	List(ctx context.Context) (models.AccountListResponse, error)
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
