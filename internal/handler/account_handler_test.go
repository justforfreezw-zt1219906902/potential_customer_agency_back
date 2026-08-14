package handler

import (
	"context"
	"log"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/justforfreezw-zt1219906902/potential_customer_agency_back/internal/middleware"
	"github.com/justforfreezw-zt1219906902/potential_customer_agency_back/internal/models"
)

type accountHandlerTestService struct {
	getCalls         int
	listSignalsCalls int
	dnaCalls         int
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
	return models.SignalPulseResponse{}, nil
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
