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

type accountHandlerTestService struct{ getCalls int }

func (s *accountHandlerTestService) List(context.Context) (models.AccountListResponse, error) {
	return models.AccountListResponse{}, nil
}
func (s *accountHandlerTestService) Get(context.Context, uuid.UUID) (models.AccountOverview, error) {
	s.getCalls++
	return models.AccountOverview{}, nil
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
