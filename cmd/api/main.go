package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/justforfreezw-zt1219906902/potential_customer_agency_back/config"
	"github.com/justforfreezw-zt1219906902/potential_customer_agency_back/internal/email"
	"github.com/justforfreezw-zt1219906902/potential_customer_agency_back/internal/handler"
	"github.com/justforfreezw-zt1219906902/potential_customer_agency_back/internal/hubspot"
	"github.com/justforfreezw-zt1219906902/potential_customer_agency_back/internal/middleware"
	"github.com/justforfreezw-zt1219906902/potential_customer_agency_back/internal/service"
	"github.com/justforfreezw-zt1219906902/potential_customer_agency_back/routes"
)

func main() {
	logger := log.New(os.Stdout, "[potential-customer-agency] ", log.LstdFlags|log.Lshortfile)

	cfg := config.Load()
	if cfg.HubSpotAccessToken == "" {
		logger.Fatal("HUBSPOT_ACCESS_TOKEN is required")
	}

	hubSpotClient := hubspot.NewClient(cfg.HubSpotAccessToken, logger)
	emailService := email.NewResendEmailService(cfg.ResendAPIKey, cfg.NotificationEmails, logger)
	leadService := service.NewLeadService(hubSpotClient, emailService, logger)
	leadHandler := handler.NewLeadHandler(leadService, logger)

	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(middleware.CORS(cfg.CORSAllowedOrigins))
	router.Use(middleware.ErrorHandler(logger))

	routes.Register(router, leadHandler)

	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	logger.Printf("server starting on port %s", cfg.Port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Fatalf("server failed: %v", err)
	}
}
