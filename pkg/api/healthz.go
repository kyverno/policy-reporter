package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type HealthCheck = func() error

// HealthzHandler for the Halthz REST API
func HealthzHandler(checks []HealthCheck) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		for _, c := range checks {
			if err := c(); err != nil {
				zap.L().Warn("health check failed", zap.Error(err))
				ctx.AbortWithError(http.StatusServiceUnavailable, err)
				return
			}
		}

		ctx.JSON(http.StatusOK, gin.H{})
	}
}
