package main

import (
	"os"

	"github.com/google/uuid"
	"github.com/jaimesHub/bookmark-management/internal/api"
	"github.com/jaimesHub/bookmark-management/internal/auth"
	"github.com/jaimesHub/bookmark-management/internal/handler"
	"github.com/jaimesHub/bookmark-management/internal/model"
	"github.com/jaimesHub/bookmark-management/internal/repository"
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

	// Hostname fallback: nếu APP_HOSTNAME không set, dùng os.Hostname() để
	// vẫn định danh được instance khi chạy local (chưa cần config thêm).
	if svcCfg.Hostname == "" {
		if h, err := os.Hostname(); err == nil {
			svcCfg.Hostname = h
		} else {
			svcCfg.Hostname = "unknown"
			log.Warn().Err(err).Msg("os.Hostname() failed, using 'unknown'")
		}
	}

	log.Info().
		Str("container_port", cfg.ContainerPort).
		Str("service_name", svcCfg.ServiceName).
		Str("instance_id", svcCfg.InstanceID).
		Str("hostname", svcCfg.Hostname).
		Str("env", cfg.Env).
		Str("log_level", cfg.LogLevel).
		Msg("configuration loaded")

	redisClient, err := pkgredis.NewClient("")
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create redis client")
	}

	// ─── Lec-6: RSA keypair load (T10 loader + T11 wire) ────────
	// Eager load + fail-fast per ADR-10. Lec-6 scope: chỉ verify load OK;
	// Lec-7 sẽ wire rsaKeys vào JWT sign/verify (login service).
	// Placement TRƯỚC DB wire: nếu RSA missing → fail trước khi mở Postgres conn,
	// tránh waste resources restart loop.
	var rsaCfg api.RSAConfig
	if err := envconfig.Process("", &rsaCfg); err != nil {
		log.Fatal().Err(err).Msg("failed to load rsa config")
	}

	rsaKeys, err := auth.LoadKeys(rsaCfg.PrivatePath, rsaCfg.PublicPath)
	if err != nil {
		log.Fatal().Err(err).
			Str("private_path", rsaCfg.PrivatePath).
			Str("public_path", rsaCfg.PublicPath).
			Msg("load rsa keys")
	}
	_ = rsaKeys // TODO Lec-7: wire vào JWT signer/verifier
	log.Info().Msg("rsa keys loaded")
	// ────────────────────────────────────────────────────────────

	// ─── Lec-6: DB + User feature wire ──────────────────────────
	// Load api.DBConfig với prefix "" → env vars DB_HOST/PORT/...
	// trực tiếp (KHÔNG qua "api" prefix của cfg). Tránh cascade rename
	// docker-compose.yml + .env.example sang DB_HOST → API_DB_HOST.
	// Architecture: api.DBConfig có envconfig tags (api boundary);
	// repository.DBConfig là pure construction params (KHÔNG depend envconfig)
	// — main.go copy fields giữa 2 structs.
	var apiDBCfg api.DBConfig
	if err := envconfig.Process("", &apiDBCfg); err != nil {
		log.Fatal().Err(err).Msg("failed to load db config")
	}

	db, err := repository.NewPostgresDB(repository.DBConfig{
		Host:     apiDBCfg.Host,
		Port:     apiDBCfg.Port,
		User:     apiDBCfg.User,
		Password: apiDBCfg.Password,
		Name:     apiDBCfg.Name,
		SSLMode:  apiDBCfg.SSLMode,
	})
	if err != nil {
		log.Fatal().Err(err).Msg("connect postgres")
	}

	// AutoMigrate User schema (Lec-6 spec mandate via PDF).
	// Idempotent — chạy lại không hỏng schema; production an toàn.
	if err := db.AutoMigrate(&model.User{}); err != nil {
		log.Fatal().Err(err).Msg("auto migrate")
	}
	log.Info().Msg("postgres connected + user schema migrated")

	// Wire user feature: repo → svc → handler
	userRepo := repository.NewUserRepository(db)
	userSvc := service.NewUserService(userRepo, cfg.BcryptCost)
	userHandler := handler.NewUserHandler(userSvc)
	// ────────────────────────────────────────────────────────────

	app := api.NewEngine(cfg, &svcCfg, redisClient, userHandler)

	if err := app.Start(); err != nil {
		log.Fatal().Err(err).Msg("api server stopped with error")
	}
}
