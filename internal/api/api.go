// Package api contains the API engine and route initialization for the HTTP server.
package api

import (
	"fmt"
	"net/http"

	_ "github.com/jaimesHub/bookmark-management/docs"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/gin-gonic/gin"
	"github.com/jaimesHub/bookmark-management/internal/handler"
	"github.com/jaimesHub/bookmark-management/internal/service"
)

// Engine defines the interface for the API engine.
type Engine interface {
	// Start starts the API server and begins listening for HTTP requests.
	Start() error
	ServeHTTP(w http.ResponseWriter, req *http.Request)
}

// engine implements the Engine interface and manages the Gin HTTP server.
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

// ServeHTTP implements the http.Handler interface for testing and request processing.
func (e *engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	e.app.ServeHTTP(w, req)
}

// initRoutes initializes all API routes and handlers.
func (e *engine) initRoutes() {
	// check health handler
	checkHealthSvc := service.NewHealthCheck(e.svcCfg)
	checkHealthHandler := handler.NewHealthCheck(checkHealthSvc)

	if e.cfg.SwaggerEnabled {
		e.app.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	e.app.GET("/health-check", checkHealthHandler.CheckHealth)
}
