package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/justforfreezw-zt1219906902/potential_customer_agency_back/internal/handler"
)

func Register(router *gin.Engine, leadHandler *handler.LeadHandler) {
	api := router.Group("/api")
	{
		api.POST("/lead", leadHandler.CreateLead)
	}
}
