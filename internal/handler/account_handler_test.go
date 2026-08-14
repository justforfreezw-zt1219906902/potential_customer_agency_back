package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/justforfreezw-zt1219906902/potential_customer_agency_back/internal/middleware"
	"github.com/justforfreezw-zt1219906902/potential_customer_agency_back/internal/models"
)

type accountHandlerTestService struct {
	getCalls          int
	listSignalsCalls  int
	dnaCalls          int
	pulseCalls        int
	pulseResponse     models.SignalPulseResponse
	portfolioResponse models.DNAPortfolioResponse
	compareResponse   models.DNACompareResponse
	portfolioCalls    int
	compareCalls      int
}

func (s *accountHandlerTestService) List(context.Context) (models.AccountListResponse, error) {
	return models.AccountListResponse{}, nil
}
func (s *accountHandlerTestService) Get(context.Context, uuid.UUID) (models.AccountOverview, error) {
	s.getCalls++
	return models.AccountOverview{}, nil
}
func (s *accountHandlerTestService) ListSignals(context.Context, uuid.UUID) (models.AccountSignalsResponse, error) {
	s.listSignalsCalls++
	return models.AccountSignalsResponse{}, nil
}
func (s *accountHandlerTestService) GetCommunicationDNA(context.Context, uuid.UUID) (models.CommunicationDNAResponse, error) {
	s.dnaCalls++
	return models.CommunicationDNAResponse{}, nil
}

func (s *accountHandlerTestService) SignalPulse(context.Context) (models.SignalPulseResponse, error) {
	s.pulseCalls++
	return s.pulseResponse, nil
}
func (s *accountHandlerTestService) ListDNAPortfolio(context.Context) (models.DNAPortfolioResponse, error) {
	s.portfolioCalls++
	return s.portfolioResponse, nil
}
func (s *accountHandlerTestService) CompareDNAPortfolio(context.Context, []uuid.UUID) (models.DNACompareResponse, error) {
	s.compareCalls++
	return s.compareResponse, nil
}

func TestDNAPortfolioHandlersReturnRepresentativeResponses(t *testing.T) {
	gin.SetMode(gin.TestMode)
	accountID := uuid.MustParse("00000000-0000-0000-0000-000000000501")
	service := &accountHandlerTestService{portfolioResponse: models.DNAPortfolioResponse{Summary: models.DNAPortfolioSummary{TotalProfiles: 1, ByTier: map[string]int{}, ByIndustry: map[string]int{}}, Items: []models.DNAPortfolioItem{{AccountID: accountID, Name: "Example", Vocabulary: []string{}, DoRules: []string{}, DontRules: []string{}, SignalTypes: []string{}}}}, compareResponse: models.DNACompareResponse{SelectedCount: 2, DominantTone: []models.DNAValueCount{}, SharedVocabulary: []models.DNAValueCount{}, UniqueVocabulary: []models.DNAValueCount{}, ProofStyles: []models.DNAValueCount{}, CTAStyles: []models.DNAValueCount{}, DoRules: []models.DNAValueCount{}, DontRules: []models.DNAValueCount{}, SignalTypes: []models.DNAValueCount{}, ProblemFraming: []models.DNAProblemFramingItem{}}}
	router := gin.New()
	router.Use(middleware.ErrorHandler(log.Default()))
	h := NewAccountHandler(service, log.Default())
	router.GET("/api/dna-portfolio", h.ListDNAPortfolio)
	router.POST("/api/dna-portfolio/compare", h.CompareDNAPortfolio)
	get := httptest.NewRecorder()
	router.ServeHTTP(get, httptest.NewRequest(http.MethodGet, "/api/dna-portfolio", nil))
	if get.Code != http.StatusOK || service.portfolioCalls != 1 || !bytes.Contains(get.Body.Bytes(), []byte(`"totalProfiles":1`)) {
		t.Fatalf("unexpected portfolio response: %d %s", get.Code, get.Body.String())
	}
	post := httptest.NewRecorder()
	body := strings.NewReader(`{"accountIds":["00000000-0000-0000-0000-000000000501","00000000-0000-0000-0000-000000000502"]}`)
	req := httptest.NewRequest(http.MethodPost, "/api/dna-portfolio/compare", body)
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(post, req)
	if post.Code != http.StatusOK || service.compareCalls != 1 || !bytes.Contains(post.Body.Bytes(), []byte(`"selectedCount":2`)) {
		t.Fatalf("unexpected compare response: %d %s", post.Code, post.Body.String())
	}
}

func TestCompareDNAPortfolioRejectsInvalidRequests(t *testing.T) {
	cases := []struct{ name, body string }{
		{"malformed JSON", "{"}, {"missing accountIds", `{}`}, {"null accountIds", `{"accountIds":null}`}, {"empty accountIds", `{"accountIds":[]}`},
		{"one UUID", `{"accountIds":["00000000-0000-0000-0000-000000000601"]}`},
		{"invalid UUID", `{"accountIds":["bad","00000000-0000-0000-0000-000000000602"]}`},
		{"duplicate UUID", `{"accountIds":["00000000-0000-0000-0000-000000000601","00000000-0000-0000-0000-000000000601"]}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			service := &accountHandlerTestService{}
			router := gin.New()
			router.Use(middleware.ErrorHandler(log.Default()))
			router.POST("/api/dna-portfolio/compare", NewAccountHandler(service, log.Default()).CompareDNAPortfolio)
			req := httptest.NewRequest(http.MethodPost, "/api/dna-portfolio/compare", strings.NewReader(tc.body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)
			if rec.Code != http.StatusBadRequest {
				t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
			}
			if service.compareCalls != 0 {
				t.Fatalf("service called %d times", service.compareCalls)
			}
		})
	}
}

func TestSignalPulseHandlerReturnsResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	service := &accountHandlerTestService{pulseResponse: models.SignalPulseResponse{
		Metrics:  models.SignalPulseMetrics{ActiveSignals: 2},
		Accounts: []models.SignalPulseAccount{{AccountID: uuid.MustParse("00000000-0000-0000-0000-000000000101"), Name: "Focus Account", Urgency: "hot", ActiveSignalCount: 2}},
	}}
	router := gin.New()
	router.Use(middleware.ErrorHandler(log.Default()))
	router.GET("/api/signal-pulse", NewAccountHandler(service, log.Default()).SignalPulse)

	recording := httptest.NewRecorder()
	router.ServeHTTP(recording, httptest.NewRequest(http.MethodGet, "/api/signal-pulse", nil))
	if recording.Code != http.StatusOK {
		t.Fatalf("status=%d, want 200", recording.Code)
	}
	if service.pulseCalls != 1 {
		t.Fatalf("pulseCalls=%d, want 1", service.pulseCalls)
	}
	var response struct {
		Metrics  map[string]int   `json:"metrics"`
		Accounts []map[string]any `json:"accounts"`
	}
	if err := json.Unmarshal(recording.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if response.Metrics["activeSignals"] != 2 || len(response.Accounts) != 1 || response.Accounts[0]["accountId"] != "00000000-0000-0000-0000-000000000101" || response.Accounts[0]["urgency"] != "hot" || response.Accounts[0]["activeSignalCount"] != float64(2) {
		t.Fatalf("unexpected response: %s", recording.Body.String())
	}
}

func TestGetAccountRejectsMalformedUUIDBeforeServiceCall(t *testing.T) {
	gin.SetMode(gin.TestMode)
	service := &accountHandlerTestService{}
	router := gin.New()
	router.Use(middleware.ErrorHandler(log.Default()))
	router.GET("/api/accounts/:accountId", NewAccountHandler(service, log.Default()).GetAccount)

	recording := httptest.NewRecorder()
	request := httptest.NewRequest("GET", "/api/accounts/not-a-uuid", nil)
	router.ServeHTTP(recording, request)

	if recording.Code != 400 {
		t.Fatalf("handler returned status %d, want 400", recording.Code)
	}
	if service.getCalls != 0 {
		t.Fatal("service should not be called for malformed UUID")
	}
}

func TestListSignalsRejectsMalformedUUIDBeforeServiceCall(t *testing.T) {
	gin.SetMode(gin.TestMode)
	service := &accountHandlerTestService{}
	router := gin.New()
	router.Use(middleware.ErrorHandler(log.Default()))
	router.GET("/api/accounts/:accountId/signals", NewAccountHandler(service, log.Default()).ListSignals)

	recording := httptest.NewRecorder()
	router.ServeHTTP(recording, httptest.NewRequest("GET", "/api/accounts/not-a-uuid/signals", nil))
	if recording.Code != 400 {
		t.Fatalf("handler returned status %d, want 400", recording.Code)
	}
	if service.listSignalsCalls != 0 {
		t.Fatal("service should not be called for malformed UUID")
	}
}

func TestCommunicationDNARejectsMalformedUUIDBeforeServiceCall(t *testing.T) {
	gin.SetMode(gin.TestMode)
	service := &accountHandlerTestService{}
	router := gin.New()
	router.Use(middleware.ErrorHandler(log.Default()))
	router.GET("/api/accounts/:accountId/communication-dna", NewAccountHandler(service, log.Default()).GetCommunicationDNA)
	recording := httptest.NewRecorder()
	router.ServeHTTP(recording, httptest.NewRequest("GET", "/api/accounts/not-a-uuid/communication-dna", nil))
	if recording.Code != 400 || service.dnaCalls != 0 {
		t.Fatalf("status=%d dnaCalls=%d", recording.Code, service.dnaCalls)
	}
}
