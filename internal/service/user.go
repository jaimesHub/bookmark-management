package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jaimesHub/bookmark-management/internal/model"
	"github.com/jaimesHub/bookmark-management/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

// Errors surface to handler → HTTP 409.
// errors.Is(err, ErrEmailAlreadyExists) → handler map sang 409 với message.
var (
	ErrEmailAlreadyExists    = errors.New("email already registered")
	ErrUsernameAlreadyExists = errors.New("username already taken")
)

//go:generate mockery --name UserService --filename user_service.go
type UserService interface {
	Register(ctx context.Context, req model.RegisterRequest) (*model.User, error)
}

// userService giữ deps + inject points cho deterministic test.
// `now` + `newID` mặc định production; T6 override qua NewUserServiceForTest.
type userService struct {
	repo       repository.UserRepository
	bcryptCost int
	now        func() time.Time
	newID      func() string
}

// NewUserService — production constructor. bcryptCost từ config (default 12).
// Inject time.Now + uuid.NewString → deterministic-friendly architecture
// (override qua NewUserServiceForTest cho unit test).
func NewUserService(repo repository.UserRepository, bcryptCost int) UserService {
	return &userService{
		repo:       repo,
		bcryptCost: bcryptCost,
		now:        time.Now,
		newID:      uuid.NewString,
	}
}

// NewUserServiceForTest — test-only constructor exposing now + newID inject.
// Plan § 5.1 line 827 require: tránh build tag complexity. T6 service tests
// dùng cost=4 (bcrypt fast), fixed clock, fixed UUID cho assertion stable.
func NewUserServiceForTest(repo repository.UserRepository, bcryptCost int, nowFn func() time.Time, idFn func() string) UserService {
	return &userService{
		repo:       repo,
		bcryptCost: bcryptCost,
		now:        nowFn,
		newID:      idFn,
	}
}

// Register normalizes email, pre-checks uniqueness, hashes password, inserts.
//
// Flow:
//  1. Lowercase + trim email (simulate CITEXT for portability Postgres ↔ SQLite).
//  2. Pre-check email exists → ErrEmailAlreadyExists (handler → 409).
//  3. Pre-check username exists → ErrUsernameAlreadyExists.
//  4. bcrypt hash password với cost từ config.
//  5. Build User entity với gen ID + timestamp UTC.
//  6. Repo Create — race-rare khi 2 register chính xác simultaneous.
//     uniqueIndex ở DB layer là defense-in-depth.
func (s *userService) Register(ctx context.Context, req model.RegisterRequest) (*model.User, error) {
	email := strings.ToLower(strings.TrimSpace(req.Email))

	// Pre-check email (race-rare cho register flow; uniqueIndex là defense-in-depth).
	if _, err := s.repo.GetByEmail(ctx, email); err == nil {
		return nil, ErrEmailAlreadyExists
	} else if !errors.Is(err, repository.ErrUserNotFound) {
		return nil, fmt.Errorf("check email: %w", err)
	}

	// Pre-check username.
	if _, err := s.repo.GetByUsername(ctx, req.Username); err == nil {
		return nil, ErrUsernameAlreadyExists
	} else if !errors.Is(err, repository.ErrUserNotFound) {
		return nil, fmt.Errorf("check username: %w", err)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), s.bcryptCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	now := s.now().UTC()
	u := &model.User{
		ID:           s.newID(),
		Username:     req.Username,
		Email:        email,
		DisplayName:  req.DisplayName,
		PasswordHash: string(hash),
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := s.repo.Create(ctx, u); err != nil {
		return nil, fmt.Errorf("user service register: %w", err)
	}
	return u, nil
}
