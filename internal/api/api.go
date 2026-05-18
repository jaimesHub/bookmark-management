// Package api contains the API engine and route initialization for the HTTP server.
package api

import (
	"fmt"
	"net/http"

	_ "github.com/jaimesHub/bookmark-management/docs"
	"github.com/redis/go-redis/v9"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/gin-gonic/gin"
	"github.com/jaimesHub/bookmark-management/internal/handler"
	"github.com/jaimesHub/bookmark-management/internal/repository"
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
	app         *gin.Engine
	cfg         *Config
	svcCfg      *service.Config
	redisClient *redis.Client
}

// NewEngine creates and returns a new API Engine instance with all routes initialized.
func NewEngine(cfg *Config, svcCfg *service.Config, redisClient *redis.Client) Engine {
	app := &engine{
		app:         gin.Default(),
		cfg:         cfg,
		svcCfg:      svcCfg,
		redisClient: redisClient,
	}

	app.initRoutes()
	return app
}

// Start starts the API server on the configured port.
func (e *engine) Start() error {
	return e.app.Run(fmt.Sprintf(":%s", e.cfg.ContainerPort))
}

// ServeHTTP implements the http.Handler interface for testing and request processing.
func (e *engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	e.app.ServeHTTP(w, req)
}

// initRoutes initializes all API routes and handlers.
func (e *engine) initRoutes() {
	urlRepo := repository.NewUrlStorage(e.redisClient)
	pingRepo := repository.NewPingRepo(e.redisClient)

	// check health handler
	checkHealthSvc := service.NewHealthCheck(e.svcCfg, pingRepo)
	checkHealthHandler := handler.NewHealthCheck(checkHealthSvc)

	if e.cfg.SwaggerEnabled {
		e.app.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	e.app.GET("/health-check", checkHealthHandler.CheckHealth)

	// shorten URL
	shortenSvc := service.NewShortenService(urlRepo)
	shortenHandler := handler.NewShorten(shortenSvc)
	redirectHandler := handler.NewRedirect(shortenSvc)

	v1 := e.app.Group("/v1")
	{
		links := v1.Group("/links")
		{
			links.POST("/shorten", shortenHandler.ShortenURL)
			links.GET("/redirect/:code", redirectHandler.Redirect)
		}

	}
}
