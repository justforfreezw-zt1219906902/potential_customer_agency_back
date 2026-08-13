package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/justforfreezw-zt1219906902/potential_customer_agency_back/internal/handler"
)

func Register(router *gin.Engine, leadHandler *handler.LeadHandler, accountHandler *handler.AccountHandler) {
	api := router.Group("/api")
	{
		api.POST("/lead", leadHandler.CreateLead)
		api.GET("/accounts", accountHandler.ListAccounts)
		api.GET("/accounts/:accountId", accountHandler.GetAccount)
		api.GET("/accounts/:accountId/signals", accountHandler.ListSignals)
	}
}
