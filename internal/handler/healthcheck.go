package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jaimesHub/bookmark-management/internal/service"
)

// HealthCheck defines the interface for health check HTTP request handling.
type HealthCheck interface {
	// CheckHealth handles HTTP requests to check the service health status.
	CheckHealth(c *gin.Context)
}

type healthCheckHandler struct {
	healthCheckService service.HealthCheck
}

// NewHealthCheck creates and returns a new HealthCheck handler instance.
func NewHealthCheck(healthCheckSvc service.HealthCheck) HealthCheck {
	return &healthCheckHandler{
		healthCheckService: healthCheckSvc,
	}
}

func (s *healthCheckHandler) CheckHealth(c *gin.Context) {
	res, err := s.healthCheckService.Check()
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error!"})
		return

	}

	c.JSON(http.StatusOK, res)
}
