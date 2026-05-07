package api

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/jaimesHub/bookmark-management/internal/handler"
	"github.com/jaimesHub/bookmark-management/internal/service"
)

// Engine defines the interface for the API engine.
type Engine interface {
	// Start starts the API server and begins listening for HTTP requests.
	Start() error
}

type engine struct {
	app    *gin.Engine
	cfg    *Config
	svcCfg *service.Config
}

// NewEngine creates and returns a new API Engine instance with all routes initialized.
func NewEngine(cfg *Config, svcCfg *service.Config) Engine {
	app := &engine{
		app:    gin.Default(),
		cfg:    cfg,
		svcCfg: svcCfg,
	}

	app.initRoutes()

	return app
}

// Start starts the API server on port 8080.
func (e *engine) Start() error {
	return e.app.Run(fmt.Sprintf(":%s", e.cfg.AppPort))
}

func (e *engine) initRoutes() {
	// check health handler
	checkHealthSvc := service.NewHealthCheck(e.svcCfg)
	checkHealthHandler := handler.NewHealthCheck(checkHealthSvc)

	e.app.GET("/health-check", checkHealthHandler.CheckHealth)
}
