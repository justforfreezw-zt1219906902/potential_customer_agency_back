package handler

import (
	"context"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	apperrors "github.com/justforfreezw-zt1219906902/potential_customer_agency_back/internal/errors"
	"github.com/justforfreezw-zt1219906902/potential_customer_agency_back/internal/models"
)

type LeadCreator interface {
	CreateLead(ctx context.Context, lead models.LeadRequest) (models.LeadResponse, error)
}

type LeadHandler struct {
	service LeadCreator
	logger  *log.Logger
}

func NewLeadHandler(service LeadCreator, logger *log.Logger) *LeadHandler {
	return &LeadHandler{
		service: service,
		logger:  logger,
	}
}

func (h *LeadHandler) CreateLead(c *gin.Context) {
	var request models.LeadRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		_ = c.Error(apperrors.BadRequest("invalid JSON request body", err))
		return
	}

	if err := request.Validate(); err != nil {
		_ = c.Error(apperrors.BadRequest(err.Error(), err))
		return
	}

	response, err := h.service.CreateLead(c.Request.Context(), request)
	if err != nil {
		_ = c.Error(err)
		return
	}

	h.logger.Printf("lead created: email=%s company=%s", request.Email, request.Company)
	c.JSON(http.StatusOK, response)
}
