package service_test

import (
	"context"
	"errors"
	"testing"

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

// newTestSvc — DRY helper: bcryptCost=4 + fixed UUID cho deterministic assertions.
// CreatedAt/UpdatedAt KHÔNG inject — GORM auto-handle ở repository layer
// (per instructor feedback PR #13).
func newTestSvc(repo *mocks.UserRepository) service.UserService {
	return service.NewUserServiceForTest(
		repo,
		testBcryptCost,
		func() string { return testUserID },
	)
}

// TestUserService_Register — table-driven test phủ 12 scenarios
// (instructor feedback PR #13: consolidate 12 separate test funcs).
// Mapping TC-L6-S001 → S012 giữ nguyên qua subtest name "SNNN_<desc>".
func TestUserService_Register(t *testing.T) {
	t.Parallel()

	const plainPwd = "Passw0rd!"

	testCases := []struct {
		name      string
		req       model.RegisterRequest
		setupMock func(ctx context.Context, repo *mocks.UserRepository)
		assert    func(t *testing.T, repo *mocks.UserRepository, u *model.User, err error)
	}{
		{
			// TC-L6-S001 — Happy path: all fields normalized và passed tới repo.
			name: "S001_HappyPath",
			req: model.RegisterRequest{
				Username: "alice", Email: "Alice@Example.com",
				Password: plainPwd, DisplayName: "Alice",
			},
			setupMock: func(ctx context.Context, repo *mocks.UserRepository) {
				repo.On("GetByEmail", ctx, "alice@example.com").
					Return(nil, repository.ErrUserNotFound).Once()
				repo.On("GetByUsername", ctx, "alice").
					Return(nil, repository.ErrUserNotFound).Once()
				repo.On("Create", ctx, mock.MatchedBy(func(u *model.User) bool {
					return u.Email == "alice@example.com" && // lowercase normalized
						u.Username == "alice" &&
						u.DisplayName == "Alice" &&
						u.ID == testUserID // deterministic UUID
				})).Return(nil).Once()
			},
			assert: func(t *testing.T, _ *mocks.UserRepository, u *model.User, err error) {
				require.NoError(t, err)
				require.NotNil(t, u)
				assert.Equal(t, testUserID, u.ID)
				assert.Equal(t, "alice@example.com", u.Email)
			},
		},
		{
			// TC-L6-S002 — Email uppercase → GetByEmail gọi với lowercase (ADR-7 normalization).
			name: "S002_EmailUppercase_NormalizedLowercase",
			req: model.RegisterRequest{
				Username: "user", Email: "USER@DOMAIN.COM",
				Password: plainPwd, DisplayName: "User",
			},
			setupMock: func(ctx context.Context, repo *mocks.UserRepository) {
				repo.On("GetByEmail", ctx, "user@domain.com").
					Return(nil, repository.ErrUserNotFound).Once()
				repo.On("GetByUsername", ctx, "user").
					Return(nil, repository.ErrUserNotFound).Once()
				repo.On("Create", ctx, mock.MatchedBy(func(u *model.User) bool {
					return u.Email == "user@domain.com"
				})).Return(nil).Once()
			},
			assert: func(t *testing.T, _ *mocks.UserRepository, _ *model.User, err error) {
				require.NoError(t, err)
			},
		},
		{
			// TC-L6-S003 — Email với leading/trailing space → trimmed + lowercased.
			name: "S003_EmailWithSpaces_TrimmedAndLowercased",
			req: model.RegisterRequest{
				Username: "user", Email: "  Alice@Example.com  ",
				Password: plainPwd, DisplayName: "Alice",
			},
			setupMock: func(ctx context.Context, repo *mocks.UserRepository) {
				repo.On("GetByEmail", ctx, "alice@example.com").
					Return(nil, repository.ErrUserNotFound).Once()
				repo.On("GetByUsername", ctx, "user").
					Return(nil, repository.ErrUserNotFound).Once()
				repo.On("Create", ctx, mock.MatchedBy(func(u *model.User) bool {
					return u.Email == "alice@example.com"
				})).Return(nil).Once()
			},
			assert: func(t *testing.T, _ *mocks.UserRepository, _ *model.User, err error) {
				require.NoError(t, err)
			},
		},
		{
			// TC-L6-S004 — Password được hash (bcrypt verify); không store plain text.
			name: "S004_PasswordHashed_NotPlainText",
			req: model.RegisterRequest{
				Username: "alice", Email: "alice@x.com",
				Password: plainPwd, DisplayName: "Alice",
			},
			setupMock: func(ctx context.Context, repo *mocks.UserRepository) {
				repo.On("GetByEmail", ctx, "alice@x.com").
					Return(nil, repository.ErrUserNotFound).Once()
				repo.On("GetByUsername", ctx, "alice").
					Return(nil, repository.ErrUserNotFound).Once()
				repo.On("Create", ctx, mock.MatchedBy(func(u *model.User) bool {
					return u.PasswordHash != plainPwd &&
						len(u.PasswordHash) > 0 &&
						bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(plainPwd)) == nil
				})).Return(nil).Once()
			},
			assert: func(t *testing.T, _ *mocks.UserRepository, _ *model.User, err error) {
				require.NoError(t, err)
			},
		},
		{
			// TC-L6-S005 — Email đã tồn tại → ErrEmailAlreadyExists + short-circuit
			// (KHÔNG gọi GetByUsername/Create).
			name: "S005_EmailAlreadyExists",
			req: model.RegisterRequest{
				Username: "bob", Email: "taken@x.com",
				Password: plainPwd, DisplayName: "Bob",
			},
			setupMock: func(ctx context.Context, repo *mocks.UserRepository) {
				repo.On("GetByEmail", ctx, "taken@x.com").
					Return(&model.User{ID: "existing"}, nil).Once()
			},
			assert: func(t *testing.T, repo *mocks.UserRepository, u *model.User, err error) {
				assert.ErrorIs(t, err, service.ErrEmailAlreadyExists)
				assert.Nil(t, u)
				repo.AssertNotCalled(t, "GetByUsername", mock.Anything, mock.Anything)
				repo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
			},
		},
		{
			// TC-L6-S006 — Username đã tồn tại → ErrUsernameAlreadyExists + short-circuit Create.
			name: "S006_UsernameAlreadyExists",
			req: model.RegisterRequest{
				Username: "taken", Email: "new@x.com",
				Password: plainPwd, DisplayName: "New",
			},
			setupMock: func(ctx context.Context, repo *mocks.UserRepository) {
				repo.On("GetByEmail", ctx, "new@x.com").
					Return(nil, repository.ErrUserNotFound).Once()
				repo.On("GetByUsername", ctx, "taken").
					Return(&model.User{ID: "existing"}, nil).Once()
			},
			assert: func(t *testing.T, repo *mocks.UserRepository, u *model.User, err error) {
				assert.ErrorIs(t, err, service.ErrUsernameAlreadyExists)
				assert.Nil(t, u)
				repo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
			},
		},
		{
			// TC-L6-S007 — DB error trên GetByEmail → wrapped error (KHÔNG phải ErrEmailAlreadyExists).
			name: "S007_GetByEmail_DBError",
			req: model.RegisterRequest{
				Username: "u", Email: "u@x.com",
				Password: plainPwd, DisplayName: "U",
			},
			setupMock: func(ctx context.Context, repo *mocks.UserRepository) {
				repo.On("GetByEmail", ctx, "u@x.com").
					Return(nil, errors.New("db connection lost")).Once()
			},
			assert: func(t *testing.T, _ *mocks.UserRepository, u *model.User, err error) {
				require.Error(t, err)
				assert.False(t, errors.Is(err, service.ErrEmailAlreadyExists))
				assert.False(t, errors.Is(err, service.ErrUsernameAlreadyExists))
				assert.Nil(t, u)
			},
		},
		{
			// TC-L6-S008 — DB error trên GetByUsername → wrapped error (symmetric với S007).
			name: "S008_GetByUsername_DBError",
			req: model.RegisterRequest{
				Username: "u", Email: "u@x.com",
				Password: plainPwd, DisplayName: "U",
			},
			setupMock: func(ctx context.Context, repo *mocks.UserRepository) {
				repo.On("GetByEmail", ctx, "u@x.com").
					Return(nil, repository.ErrUserNotFound).Once()
				repo.On("GetByUsername", ctx, "u").
					Return(nil, errors.New("timeout")).Once()
			},
			assert: func(t *testing.T, _ *mocks.UserRepository, u *model.User, err error) {
				require.Error(t, err)
				assert.False(t, errors.Is(err, service.ErrUsernameAlreadyExists))
				assert.Nil(t, u)
			},
		},
		{
			// TC-L6-S009 — DB error trên Create → wrapped error.
			name: "S009_Create_DBError",
			req: model.RegisterRequest{
				Username: "u", Email: "u@x.com",
				Password: plainPwd, DisplayName: "U",
			},
			setupMock: func(ctx context.Context, repo *mocks.UserRepository) {
				repo.On("GetByEmail", ctx, "u@x.com").
					Return(nil, repository.ErrUserNotFound).Once()
				repo.On("GetByUsername", ctx, "u").
					Return(nil, repository.ErrUserNotFound).Once()
				repo.On("Create", ctx, mock.Anything).
					Return(errors.New("disk full")).Once()
			},
			assert: func(t *testing.T, _ *mocks.UserRepository, u *model.User, err error) {
				require.Error(t, err)
				assert.Nil(t, u)
			},
		},
		{
			// TC-L6-S010 — Returned user không có plain password (defense-in-depth trên top của json:"-").
			name: "S010_ReturnedUser_NoPlainPassword",
			req: model.RegisterRequest{
				Username: "u", Email: "u@x.com",
				Password: plainPwd, DisplayName: "U",
			},
			setupMock: func(ctx context.Context, repo *mocks.UserRepository) {
				repo.On("GetByEmail", ctx, "u@x.com").
					Return(nil, repository.ErrUserNotFound).Once()
				repo.On("GetByUsername", ctx, "u").
					Return(nil, repository.ErrUserNotFound).Once()
				repo.On("Create", ctx, mock.Anything).Return(nil).Once()
			},
			assert: func(t *testing.T, _ *mocks.UserRepository, u *model.User, err error) {
				require.NoError(t, err)
				require.NotNil(t, u)
				assert.NotEqual(t, plainPwd, u.PasswordHash)
				assert.NotEmpty(t, u.PasswordHash)
			},
		},
		{
			// TC-L6-S011 — ID được inject (deterministic) → assert exact match.
			name: "S011_DeterministicID",
			req: model.RegisterRequest{
				Username: "u", Email: "u@x.com",
				Password: plainPwd, DisplayName: "U",
			},
			setupMock: func(ctx context.Context, repo *mocks.UserRepository) {
				repo.On("GetByEmail", ctx, "u@x.com").
					Return(nil, repository.ErrUserNotFound).Once()
				repo.On("GetByUsername", ctx, "u").
					Return(nil, repository.ErrUserNotFound).Once()
				repo.On("Create", ctx, mock.Anything).Return(nil).Once()
			},
			assert: func(t *testing.T, _ *mocks.UserRepository, u *model.User, err error) {
				require.NoError(t, err)
				assert.Equal(t, testUserID, u.ID)
			},
		},
		{
			// TC-L6-S012 — Architectural boundary: service KHÔNG set timestamps,
			// GORM auto-handle via BeforeCreate hook ở repository layer (instructor feedback PR #13).
			name: "S012_TimestampsLeftToGORM",
			req: model.RegisterRequest{
				Username: "u", Email: "u@x.com",
				Password: plainPwd, DisplayName: "U",
			},
			setupMock: func(ctx context.Context, repo *mocks.UserRepository) {
				repo.On("GetByEmail", ctx, "u@x.com").
					Return(nil, repository.ErrUserNotFound).Once()
				repo.On("GetByUsername", ctx, "u").
					Return(nil, repository.ErrUserNotFound).Once()
				repo.On("Create", ctx, mock.Anything).Return(nil).Once()
			},
			assert: func(t *testing.T, _ *mocks.UserRepository, u *model.User, err error) {
				require.NoError(t, err)
				require.NotNil(t, u)
				assert.True(t, u.CreatedAt.IsZero(), "service must not set CreatedAt — GORM responsibility")
				assert.True(t, u.UpdatedAt.IsZero(), "service must not set UpdatedAt — GORM responsibility")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctx := t.Context()
			repo := mocks.NewUserRepository(t)
			tc.setupMock(ctx, repo)
			svc := newTestSvc(repo)
			u, err := svc.Register(ctx, tc.req)
			tc.assert(t, repo, u, err)
		})
	}
}

// _ = context.Background — keep "context" import alive for closure type sigs.
var _ = context.Background
