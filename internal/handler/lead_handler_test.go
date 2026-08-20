package handler

import (
	"context"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/justforfreezw-zt1219906902/potential_customer_agency_back/internal/middleware"
	"github.com/justforfreezw-zt1219906902/potential_customer_agency_back/internal/models"
)

type leadCreatorStub struct {
	called int
	lead   models.LeadRequest
}

func (s *leadCreatorStub) CreateLead(_ context.Context, lead models.LeadRequest) (models.LeadResponse, error) {
	s.called++
	s.lead = lead
	return models.LeadResponse{Message: "lead submitted successfully", HubSpotContactID: "contact-1"}, nil
}

func leadTestRouter(service LeadCreator) *gin.Engine {
	gin.SetMode(gin.TestMode)
	logger := log.New(io.Discard, "", 0)
	router := gin.New()
	router.Use(middleware.ErrorHandler(logger))
	router.POST("/api/lead", NewLeadHandler(service, logger).CreateLead)
	return router
}

func TestLeadHandlerAcceptsOptionalContext(t *testing.T) {
	cases := []struct {
		name        string
		body        string
		wantContext string
	}{
		{"with context", `{"firstName":"Tom","familyName":"Zhao","company":"Example Company","workEmail":"tom@example.com","context":"Interested in discussing the product."}`, "Interested in discussing the product."},
		{"without context", `{"firstName":"Tom","familyName":"Zhao","company":"Example Company","workEmail":"tom@example.com"}`, ""},
		{"empty context", `{"firstName":"Tom","familyName":"Zhao","company":"Example Company","workEmail":"tom@example.com","context":""}`, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			service := &leadCreatorStub{}
			request := httptest.NewRequest(http.MethodPost, "/api/lead", strings.NewReader(tc.body))
			request.Header.Set("Content-Type", "application/json")
			response := httptest.NewRecorder()
			leadTestRouter(service).ServeHTTP(response, request)
			if response.Code != http.StatusOK || service.called != 1 || service.lead.Context != tc.wantContext {
				t.Fatalf("status=%d called=%d lead=%+v body=%s", response.Code, service.called, service.lead, response.Body.String())
			}
		})
	}
}

func TestLeadHandlerStillRequiresFamilyName(t *testing.T) {
	service := &leadCreatorStub{}
	request := httptest.NewRequest(http.MethodPost, "/api/lead", strings.NewReader(`{"firstName":"Tom","company":"Example Company","workEmail":"tom@example.com"}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	leadTestRouter(service).ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest || service.called != 0 || !strings.Contains(response.Body.String(), "familyName is required") {
		t.Fatalf("status=%d called=%d body=%s", response.Code, service.called, response.Body.String())
	}
}
