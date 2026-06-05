package repository_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jaimesHub/bookmark-management/internal/model"
	"github.com/jaimesHub/bookmark-management/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// newTestDB opens an in-memory SQLite, AutoMigrate User schema, returns *gorm.DB.
// DSN ":memory:" (default cache=private) → mỗi gorm.Open get fresh isolated DB.
// KHÔNG dùng "?cache=shared" — share state process-wide → parallel tests collide
// trên AutoMigrate ("table users already exists").
func newTestDB(t *testing.T) *gorm.DB {
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

// fixtureUser tạo User với defaults phía test, override field qua functional opts.
func fixtureUser(t *testing.T, opts ...func(*model.User)) *model.User {
	t.Helper()
	u := &model.User{
		ID:           "id-" + t.Name(),
		Username:     "alice",
		Email:        "alice@example.com",
		DisplayName:  "Alice",
		PasswordHash: "hash",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	for _, opt := range opts {
		opt(u)
	}
	return u
}

func TestUserRepository_Create_GetByEmail_Success(t *testing.T) {
	t.Parallel()
	ctx := t.Context()

	db := newTestDB(t)
	repo := repository.NewUserRepository(db)

	u := fixtureUser(t)
	require.NoError(t, repo.Create(ctx, u))

	got, err := repo.GetByEmail(ctx, "alice@example.com")
	require.NoError(t, err)
	assert.Equal(t, "alice", got.Username)
	assert.Equal(t, "Alice", got.DisplayName)
	assert.Equal(t, "hash", got.PasswordHash)
}

func TestUserRepository_GetByEmail_NotFound(t *testing.T) {
	t.Parallel()
	ctx := t.Context()

	db := newTestDB(t)
	repo := repository.NewUserRepository(db)

	_, err := repo.GetByEmail(ctx, "missing@example.com")
	require.Error(t, err)
	assert.True(t, errors.Is(err, repository.ErrUserNotFound), "must surface ErrUserNotFound sentinel")
}

func TestUserRepository_GetByUsername_Success(t *testing.T) {
	t.Parallel()
	ctx := t.Context()

	db := newTestDB(t)
	repo := repository.NewUserRepository(db)

	u := fixtureUser(t)
	require.NoError(t, repo.Create(ctx, u))

	got, err := repo.GetByUsername(ctx, "alice")
	require.NoError(t, err)
	assert.Equal(t, "alice@example.com", got.Email)
}

func TestUserRepository_GetByUsername_NotFound(t *testing.T) {
	t.Parallel()
	ctx := t.Context()

	db := newTestDB(t)
	repo := repository.NewUserRepository(db)

	_, err := repo.GetByUsername(ctx, "missing")
	require.Error(t, err)
	assert.True(t, errors.Is(err, repository.ErrUserNotFound))
}

func TestUserRepository_Create_DuplicateEmail(t *testing.T) {
	t.Parallel()
	ctx := t.Context()

	db := newTestDB(t)
	repo := repository.NewUserRepository(db)

	u1 := fixtureUser(t, func(u *model.User) { u.ID = "id-1"; u.Username = "alice" })
	u2 := fixtureUser(t, func(u *model.User) { u.ID = "id-2"; u.Username = "bob" })
	// Cả 2 share Email -> uniqueIndex(Email) phải reject u2.

	require.NoError(t, repo.Create(ctx, u1))
	err := repo.Create(ctx, u2)
	assert.Error(t, err, "uniqueIndex Email must reject duplicate")
}

func TestUserRepository_Create_DuplicateUsername(t *testing.T) {
	t.Parallel()
	ctx := t.Context()

	db := newTestDB(t)
	repo := repository.NewUserRepository(db)

	u1 := fixtureUser(t, func(u *model.User) { u.ID = "id-1"; u.Email = "a@x.com" })
	u2 := fixtureUser(t, func(u *model.User) { u.ID = "id-2"; u.Email = "b@x.com" })
	// Cả 2 share Username -> uniqueIndex(Username) phải reject u2.

	require.NoError(t, repo.Create(ctx, u1))
	err := repo.Create(ctx, u2)
	assert.Error(t, err, "uniqueIndex Username must reject duplicate")
}

func TestUserRepository_Create_RespectsContextCancellation(t *testing.T) {
	t.Parallel()

	db := newTestDB(t)
	repo := repository.NewUserRepository(db)

	ctx, cancel := context.WithCancel(t.Context())
	cancel() // pre-cancel TRƯỚC khi gọi Create.

	u := fixtureUser(t)
	err := repo.Create(ctx, u)
	require.Error(t, err, "Create với ctx cancelled phải fail")
	// SQLite có thể trả "interrupted" hoặc "context canceled" tuỳ driver version —
	// quan trọng là KHÔNG silent success.
}

func TestUserRepository_CrossMethod_Consistency(t *testing.T) {
	t.Parallel()
	ctx := t.Context()

	db := newTestDB(t)
	repo := repository.NewUserRepository(db)

	u := fixtureUser(t)
	require.NoError(t, repo.Create(ctx, u))

	gotByEmail, err := repo.GetByEmail(ctx, "alice@example.com")
	require.NoError(t, err)

	gotByUsername, err := repo.GetByUsername(ctx, "alice")
	require.NoError(t, err)

	assert.Equal(t, gotByEmail.ID, gotByUsername.ID, "cùng user phải có ID giống nhau qua 2 query path")
}
