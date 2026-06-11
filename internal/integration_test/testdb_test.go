package integration_test

import (
	"net/http"
	"testing"

	"github.com/jaimesHub/bookmark-management/internal/api"
	"github.com/jaimesHub/bookmark-management/internal/handler"
	"github.com/jaimesHub/bookmark-management/internal/model"
	"github.com/jaimesHub/bookmark-management/internal/repository"
	"github.com/jaimesHub/bookmark-management/internal/service"
	"github.com/jaimesHub/bookmark-management/internal/testutil"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

const (
	// testBcryptCost = 4 — bcrypt min, ~10ms/hash thay vì ~100ms cost 12.
	// Integration test ~11 cases × 4 cases có bcrypt path = ~40ms tổng.
	testBcryptCost = 4
)

// SetupTestDB opens an in-memory SQLite + AutoMigrate User schema.
// DSN ":memory:" (default cache=private) → mỗi gorm.Open isolated → safe cho
// shared DB pattern trong test (multiple subtests reuse same DB instance).
// Match T4 pattern (lesson: "?cache=shared" bị parallel collision).
func SetupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err, "open sqlite in-memory")
	require.NoError(t, db.AutoMigrate(&model.User{}), "auto-migrate User")
	t.Cleanup(func() {
		sqlDB, _ := db.DB()
		_ = sqlDB.Close()
	})
	return db
}

// BuildEngineWithDB wires the full Gin engine over an injected *gorm.DB +
// mock Redis (existing integration test pattern Lec-3).
// Tách hàm này khỏi production main.go để test inject SQLite + miniredis.
//
// Wire path (match main.go T11):
//
//	userRepo := repository.NewUserRepository(db)
//	userSvc := service.NewUserService(userRepo, testBcryptCost)
//	userHandler := handler.NewUserHandler(userSvc)
//	engine := api.NewEngine(cfg, svcCfg, redisMock, userHandler)
//
// KHÔNG wire RSA load (T11) — integration test scope: register endpoint flow only.
// KHÔNG wire AutoMigrate cho User (đã làm ở SetupTestDB).
func BuildEngineWithDB(t *testing.T, db *gorm.DB) http.Handler {
	t.Helper()

	cfg := &api.Config{ContainerPort: "8080", BcryptCost: testBcryptCost}
	svcCfg := &service.Config{
		ServiceName: "bookmark-it-test",
		InstanceID:  "550e8400-e29b-41d4-a716-446655440000",
		Hostname:    "integration-test-host",
	}

	userRepo := repository.NewUserRepository(db)
	userSvc := service.NewUserService(userRepo, cfg.BcryptCost)
	userHandler := handler.NewUserHandler(userSvc)

	return api.NewEngine(cfg, svcCfg, testutil.InitMockRedis(t), userHandler)
}
