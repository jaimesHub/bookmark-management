package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jaimesHub/bookmark-management/internal/model"
	"github.com/jaimesHub/bookmark-management/internal/repository"
	"github.com/jaimesHub/bookmark-management/internal/repository/mocks"
	"github.com/jaimesHub/bookmark-management/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

const (
	testBcryptCost = 4 // ADR-4: cost 4 cho test → ~10ms/case vs ~100ms cost 12.
	testUserID     = "test-uuid-fixed-1234"
)

var testFixedTime = time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)

// newTestSvc — DRY helper: bcryptCost=4 + fixed time/ID cho deterministic assertions.
// T6 inject qua NewUserServiceForTest (T5 § 4.4 line 124-134).
func newTestSvc(repo *mocks.UserRepository) service.UserService {
	return service.NewUserServiceForTest(
		repo,
		testBcryptCost,
		func() time.Time { return testFixedTime },
		func() string { return testUserID },
	)
}

// TC-L6-S001 — Happy path: all fields normalized và passed tới repo.
func TestUserService_Register_HappyPath(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	repo := mocks.NewUserRepository(t)

	req := model.RegisterRequest{
		Username: "alice", Email: "Alice@Example.com",
		Password: "Passw0rd!", DisplayName: "Alice",
	}

	repo.On("GetByEmail", ctx, "alice@example.com").
		Return(nil, repository.ErrUserNotFound).Once()
	repo.On("GetByUsername", ctx, "alice").
		Return(nil, repository.ErrUserNotFound).Once()
	repo.On("Create", ctx, mock.MatchedBy(func(u *model.User) bool {
		return u.Email == "alice@example.com" && // lowercase normalized
			u.Username == "alice" &&
			u.DisplayName == "Alice" &&
			u.ID == testUserID && // deterministic UUID
			u.CreatedAt.Equal(testFixedTime) &&
			u.UpdatedAt.Equal(testFixedTime)
	})).Return(nil).Once()

	svc := newTestSvc(repo)
	u, err := svc.Register(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, u)
	assert.Equal(t, testUserID, u.ID)
	assert.Equal(t, "alice@example.com", u.Email)
	assert.Equal(t, testFixedTime, u.CreatedAt)
}

// TC-L6-S002 — Email uppercase → GetByEmail gọi với lowercase (ADR-7 normalization).
func TestUserService_Register_EmailUppercase_NormalizedLowercase(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	repo := mocks.NewUserRepository(t)

	req := model.RegisterRequest{
		Username: "user", Email: "USER@DOMAIN.COM",
		Password: "Passw0rd!", DisplayName: "User",
	}

	repo.On("GetByEmail", ctx, "user@domain.com").
		Return(nil, repository.ErrUserNotFound).Once()
	repo.On("GetByUsername", ctx, "user").
		Return(nil, repository.ErrUserNotFound).Once()
	repo.On("Create", ctx, mock.MatchedBy(func(u *model.User) bool {
		return u.Email == "user@domain.com"
	})).Return(nil).Once()

	svc := newTestSvc(repo)
	_, err := svc.Register(ctx, req)
	require.NoError(t, err)
}

// TC-L6-S003 — Email với leading/trailing space → trimmed + lowercased.
func TestUserService_Register_EmailWithSpaces_TrimmedAndLowercased(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	repo := mocks.NewUserRepository(t)

	req := model.RegisterRequest{
		Username: "user", Email: "  Alice@Example.com  ",
		Password: "Passw0rd!", DisplayName: "Alice",
	}

	repo.On("GetByEmail", ctx, "alice@example.com").
		Return(nil, repository.ErrUserNotFound).Once()
	repo.On("GetByUsername", ctx, "user").
		Return(nil, repository.ErrUserNotFound).Once()
	repo.On("Create", ctx, mock.MatchedBy(func(u *model.User) bool {
		return u.Email == "alice@example.com"
	})).Return(nil).Once()

	svc := newTestSvc(repo)
	_, err := svc.Register(ctx, req)
	require.NoError(t, err)
}

// TC-L6-S004 — Password được hash (bcrypt verify); không store plain text.
func TestUserService_Register_PasswordHashed_NotPlainText(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	repo := mocks.NewUserRepository(t)

	const plainPwd = "Passw0rd!"
	req := model.RegisterRequest{
		Username: "alice", Email: "alice@x.com",
		Password: plainPwd, DisplayName: "Alice",
	}

	repo.On("GetByEmail", ctx, "alice@x.com").
		Return(nil, repository.ErrUserNotFound).Once()
	repo.On("GetByUsername", ctx, "alice").
		Return(nil, repository.ErrUserNotFound).Once()
	repo.On("Create", ctx, mock.MatchedBy(func(u *model.User) bool {
		return u.PasswordHash != plainPwd &&
			len(u.PasswordHash) > 0 &&
			bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(plainPwd)) == nil
	})).Return(nil).Once()

	svc := newTestSvc(repo)
	_, err := svc.Register(ctx, req)
	require.NoError(t, err)
}

// TC-L6-S005 — Email đã tồn tại → ErrEmailAlreadyExists + short-circuit (KHÔNG gọi GetByUsername/Create).
func TestUserService_Register_EmailAlreadyExists(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	repo := mocks.NewUserRepository(t)

	req := model.RegisterRequest{
		Username: "bob", Email: "taken@x.com",
		Password: "Passw0rd!", DisplayName: "Bob",
	}

	repo.On("GetByEmail", ctx, "taken@x.com").
		Return(&model.User{ID: "existing"}, nil).Once()

	svc := newTestSvc(repo)
	u, err := svc.Register(ctx, req)
	assert.ErrorIs(t, err, service.ErrEmailAlreadyExists)
	assert.Nil(t, u)
	repo.AssertNotCalled(t, "GetByUsername", mock.Anything, mock.Anything)
	repo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}

// TC-L6-S006 — Username đã tồn tại → ErrUsernameAlreadyExists + short-circuit Create.
func TestUserService_Register_UsernameAlreadyExists(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	repo := mocks.NewUserRepository(t)

	req := model.RegisterRequest{
		Username: "taken", Email: "new@x.com",
		Password: "Passw0rd!", DisplayName: "New",
	}

	repo.On("GetByEmail", ctx, "new@x.com").
		Return(nil, repository.ErrUserNotFound).Once()
	repo.On("GetByUsername", ctx, "taken").
		Return(&model.User{ID: "existing"}, nil).Once()

	svc := newTestSvc(repo)
	u, err := svc.Register(ctx, req)
	assert.ErrorIs(t, err, service.ErrUsernameAlreadyExists)
	assert.Nil(t, u)
	repo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}

// TC-L6-S007 — DB error trên GetByEmail → wrapped error (KHÔNG phải ErrEmailAlreadyExists).
func TestUserService_Register_GetByEmail_DBError(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	repo := mocks.NewUserRepository(t)

	req := model.RegisterRequest{
		Username: "u", Email: "u@x.com",
		Password: "Passw0rd!", DisplayName: "U",
	}

	dbErr := errors.New("db connection lost")
	repo.On("GetByEmail", ctx, "u@x.com").Return(nil, dbErr).Once()

	svc := newTestSvc(repo)
	u, err := svc.Register(ctx, req)
	require.Error(t, err)
	assert.False(t, errors.Is(err, service.ErrEmailAlreadyExists))
	assert.False(t, errors.Is(err, service.ErrUsernameAlreadyExists))
	assert.Nil(t, u)
}

// TC-L6-S008 — DB error trên GetByUsername → wrapped error (symmetric với S007).
func TestUserService_Register_GetByUsername_DBError(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	repo := mocks.NewUserRepository(t)

	req := model.RegisterRequest{
		Username: "u", Email: "u@x.com",
		Password: "Passw0rd!", DisplayName: "U",
	}

	dbErr := errors.New("timeout")
	repo.On("GetByEmail", ctx, "u@x.com").
		Return(nil, repository.ErrUserNotFound).Once()
	repo.On("GetByUsername", ctx, "u").Return(nil, dbErr).Once()

	svc := newTestSvc(repo)
	u, err := svc.Register(ctx, req)
	require.Error(t, err)
	assert.False(t, errors.Is(err, service.ErrUsernameAlreadyExists))
	assert.Nil(t, u)
}

// TC-L6-S009 — DB error trên Create → wrapped error.
func TestUserService_Register_Create_DBError(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	repo := mocks.NewUserRepository(t)

	req := model.RegisterRequest{
		Username: "u", Email: "u@x.com",
		Password: "Passw0rd!", DisplayName: "U",
	}

	dbErr := errors.New("disk full")
	repo.On("GetByEmail", ctx, "u@x.com").
		Return(nil, repository.ErrUserNotFound).Once()
	repo.On("GetByUsername", ctx, "u").
		Return(nil, repository.ErrUserNotFound).Once()
	repo.On("Create", ctx, mock.Anything).Return(dbErr).Once()

	svc := newTestSvc(repo)
	u, err := svc.Register(ctx, req)
	require.Error(t, err)
	assert.Nil(t, u)
}

// TC-L6-S010 — Returned user không có plain password (defense-in-depth trên top của json:"-").
func TestUserService_Register_ReturnedUser_NoPlainPassword(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	repo := mocks.NewUserRepository(t)

	const plainPwd = "Passw0rd!"
	req := model.RegisterRequest{
		Username: "u", Email: "u@x.com",
		Password: plainPwd, DisplayName: "U",
	}

	repo.On("GetByEmail", ctx, "u@x.com").
		Return(nil, repository.ErrUserNotFound).Once()
	repo.On("GetByUsername", ctx, "u").
		Return(nil, repository.ErrUserNotFound).Once()
	repo.On("Create", ctx, mock.Anything).Return(nil).Once()

	svc := newTestSvc(repo)
	u, err := svc.Register(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, u)
	assert.NotEqual(t, plainPwd, u.PasswordHash)
	assert.NotEmpty(t, u.PasswordHash)
}

// TC-L6-S011 — ID được inject (deterministic) → assert exact match.
func TestUserService_Register_DeterministicID(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	repo := mocks.NewUserRepository(t)

	req := model.RegisterRequest{
		Username: "u", Email: "u@x.com",
		Password: "Passw0rd!", DisplayName: "U",
	}

	repo.On("GetByEmail", ctx, "u@x.com").
		Return(nil, repository.ErrUserNotFound).Once()
	repo.On("GetByUsername", ctx, "u").
		Return(nil, repository.ErrUserNotFound).Once()
	repo.On("Create", ctx, mock.Anything).Return(nil).Once()

	svc := newTestSvc(repo)
	u, err := svc.Register(ctx, req)
	require.NoError(t, err)
	assert.Equal(t, testUserID, u.ID)
}

// TC-L6-S012 — Timestamps được inject (deterministic + UTC).
func TestUserService_Register_DeterministicTimestamps(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	repo := mocks.NewUserRepository(t)

	req := model.RegisterRequest{
		Username: "u", Email: "u@x.com",
		Password: "Passw0rd!", DisplayName: "U",
	}

	repo.On("GetByEmail", ctx, "u@x.com").
		Return(nil, repository.ErrUserNotFound).Once()
	repo.On("GetByUsername", ctx, "u").
		Return(nil, repository.ErrUserNotFound).Once()
	repo.On("Create", ctx, mock.Anything).Return(nil).Once()

	svc := newTestSvc(repo)
	u, err := svc.Register(ctx, req)
	require.NoError(t, err)
	assert.Equal(t, testFixedTime, u.CreatedAt)
	assert.Equal(t, testFixedTime, u.UpdatedAt)
	assert.Equal(t, time.UTC, u.CreatedAt.Location())
}

// _ = context.Background — keep "context" import alive if future refactor drops t.Context().
var _ = context.Background
