package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jaimesHub/bookmark-management/internal/model"
	"gorm.io/gorm"
)

// ErrUserNotFound surfaces khi GORM trả ErrRecordNotFound. Service layer
// dùng để phân biệt "không tồn tại" vs lỗi DB thực sự (transient / config).
var ErrUserNotFound = errors.New("user not found")

//go:generate mockery --name UserRepository --filename user_repository.go
type UserRepository interface {
	Create(ctx context.Context, u *model.User) error
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	GetByUsername(ctx context.Context, username string) (*model.User, error)
}

type userRepository struct {
	db *gorm.DB
}

// NewUserRepository constructs the GORM-backed implementation.
// Trả interface để service layer mock dễ; mock generated qua mockery (T11).
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

// Create inserts a new user row. Trả lỗi thô từ GORM —
// service layer đã pre-check uniqueness trước khi gọi Create.
func (r *userRepository) Create(ctx context.Context, u *model.User) error {
	if err := r.db.WithContext(ctx).Create(u).Error; err != nil {
		return fmt.Errorf("user repo create: %w", err)
	}
	return nil
}

// GetByEmail returns ErrUserNotFound khi không tồn tại;
// lỗi khác wrap với context để debug trace.
func (r *userRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	var u model.User
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&u).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("user repo get by email: %w", err)
	}
	return &u, nil
}

// GetByUsername — symmetric với GetByEmail.
func (r *userRepository) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	var u model.User
	err := r.db.WithContext(ctx).Where("username = ?", username).First(&u).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("user repo get by username: %w", err)
	}
	return &u, nil
}
