package middleware

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	apperrors "github.com/justforfreezw-zt1219906902/potential_customer_agency_back/internal/errors"
	"github.com/justforfreezw-zt1219906902/potential_customer_agency_back/internal/models"
)

func ErrorHandler(logger *log.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) == 0 {
			return
		}

		err := c.Errors.Last().Err
		if appErr, ok := apperrors.AsAppError(err); ok {
			logger.Printf("request failed: code=%s status=%d error=%v", appErr.Code, appErr.Status, appErr)
			c.AbortWithStatusJSON(appErr.Status, models.ErrorResponse{Error: appErr.Message})
			return
		}

		logger.Printf("unexpected request error: %v", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, models.ErrorResponse{Error: "internal server error"})
	}
}
