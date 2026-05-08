// Package handler contains HTTP request handlers for API endpoints.
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

// healthCheckHandler implements the HealthCheck interface.
type healthCheckHandler struct {
	healthCheckService service.HealthCheck
}

// NewHealthCheck creates and returns a new HealthCheck handler instance.
func NewHealthCheck(healthCheckSvc service.HealthCheck) HealthCheck {
	return &healthCheckHandler{
		healthCheckService: healthCheckSvc,
	}
}

// CheckHealth handles HTTP GET requests to the /health-check endpoint and returns the service health status.
//
// @Summary      Check health
// @Description  Returns service health status
// @Tags         health
// @Produce      json
// @Success      200  {object}  service.Response
// @Failure      500  {object}  map[string]string
// @Router       /health-check [get]
func (s *healthCheckHandler) CheckHealth(c *gin.Context) {
	res, err := s.healthCheckService.Check()
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error!"})
		return

	}

	c.JSON(http.StatusOK, res)
}
