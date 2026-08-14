package handler

import (
	"context"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	apperrors "github.com/justforfreezw-zt1219906902/potential_customer_agency_back/internal/errors"
	"github.com/justforfreezw-zt1219906902/potential_customer_agency_back/internal/models"
	"net/http"
)

type OutreachGeneratorService interface {
	Generate(context.Context, uuid.UUID, models.OutreachGenerationRequest) (models.OutreachGenerationResponse, error)
}
type OutreachHandler struct{ service OutreachGeneratorService }

func NewOutreachHandler(service OutreachGeneratorService) *OutreachHandler {
	return &OutreachHandler{service: service}
}
func (h *OutreachHandler) Generate(c *gin.Context) {
	accountID, err := uuid.Parse(c.Param("accountId"))
	if err != nil {
		_ = c.Error(apperrors.BadRequest("accountId must be a valid UUID", err))
		return
	}
	var payload map[string]json.RawMessage
	if err := c.ShouldBindJSON(&payload); err != nil {
		_ = c.Error(apperrors.BadRequest("invalid JSON request body", err))
		return
	}
	var request models.OutreachGenerationRequest
	raw, err := json.Marshal(payload)
	if err != nil || json.Unmarshal(raw, &request) != nil {
		_ = c.Error(apperrors.BadRequest("invalid JSON request body", err))
		return
	}
	var draft map[string]json.RawMessage
	if value, ok := payload["currentDraft"]; !ok || json.Unmarshal(value, &draft) != nil || len(draft) != 4 {
		_ = c.Error(apperrors.BadRequest("currentDraft must contain subject, opening, value, and cta", nil))
		return
	}
	for _, field := range []string{"subject", "opening", "value", "cta"} {
		value, ok := draft[field]
		var text string
		if !ok || json.Unmarshal(value, &text) != nil {
			_ = c.Error(apperrors.BadRequest("currentDraft fields must be strings", nil))
			return
		}
	}
	if request.Persona != "marketing" && request.Persona != "sales" && request.Persona != "exec" {
		_ = c.Error(apperrors.BadRequest("persona must be marketing, sales, or exec", nil))
		return
	}
	if len(request.Parts) == 0 {
		_ = c.Error(apperrors.BadRequest("parts must be a non-empty array", nil))
		return
	}
	seen := map[string]struct{}{}
	for _, part := range request.Parts {
		if part != "subject" && part != "opening" && part != "value" && part != "cta" {
			_ = c.Error(apperrors.BadRequest("unsupported generation part", nil))
			return
		}
		if _, ok := seen[part]; ok {
			_ = c.Error(apperrors.BadRequest("parts must not contain duplicates", nil))
			return
		}
		seen[part] = struct{}{}
	}
	if request.AnchorSignalID == uuid.Nil {
		_ = c.Error(apperrors.BadRequest("anchorSignalId must be a valid UUID", nil))
		return
	}
	response, err := h.service.Generate(c.Request.Context(), accountID, request)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, response)
}
