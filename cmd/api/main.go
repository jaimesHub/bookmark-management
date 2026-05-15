package main

import (
	"github.com/google/uuid"
	"github.com/jaimesHub/bookmark-management/internal/api"
	"github.com/jaimesHub/bookmark-management/internal/service"
	"github.com/jaimesHub/bookmark-management/pkg/logger"
	pkgredis "github.com/jaimesHub/bookmark-management/pkg/redis"
	"github.com/kelseyhightower/envconfig"
	"github.com/rs/zerolog/log"
)

// @title           Bookmark Management API
// @version         1.0
// @description     API for managing bookmarks
// @host            localhost:8080
// @BasePath
func main() {
	// Load API config first (chứa APP_ENV + LOG_LEVEL)
	cfg, err := api.NewConfig()
	if err != nil {
		// Logger chưa init → dùng panic là chấp nhận được ở bootstrap stage này
		panic(err)
	}

	// Init logger NGAY SAU khi load config — mọi log sau đây sẽ structured
	logger.Init(cfg.Env, cfg.LogLevel)

	var svcCfg service.Config
	if err := envconfig.Process("", &svcCfg); err != nil {
		log.Fatal().Err(err).Msg("failed to load service config")
	}

	// UUID fallback
	if svcCfg.InstanceID == "" {
		svcCfg.InstanceID = uuid.New().String()
	}

	log.Info().
		Str("app_port", cfg.AppPort).
		Str("service_name", svcCfg.ServiceName).
		Str("instance_id", svcCfg.InstanceID).
		Str("env", cfg.Env).
		Str("log_level", cfg.LogLevel).
		Msg("configuration loaded")

	redisClient, err := pkgredis.NewClient("")
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create redis client")
	}

	app := api.NewEngine(cfg, &svcCfg, redisClient)

	if err := app.Start(); err != nil {
		log.Fatal().Err(err).Msg("api server stopped with error")
	}
}
