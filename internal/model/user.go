package model

import "time"

// User là persisted entity. GORM tags drive AutoMigrate schema —
// model file KHÔNG import gorm package (tags là string, parse lúc AutoMigrate).
//
// PasswordHash KHÔNG bao giờ serialize ra JSON (json:"-") để tránh leak
// qua response shape (handler chỉ return User trong RegisterResponse).
type User struct {
	ID           string    `gorm:"type:varchar(36);primaryKey" json:"id"`
	Username     string    `gorm:"type:varchar(32);uniqueIndex;not null" json:"username"`
	Email        string    `gorm:"type:varchar(254);uniqueIndex;not null" json:"email"`
	DisplayName  string    `gorm:"type:varchar(64);not null" json:"display_name"`
	PasswordHash string    `gorm:"type:varchar(72);not null" json:"-"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// TableName explicit để tránh GORM pluralization auto-generate ("users")
// thành format khác với mong đợi nếu sau này đổi struct name.
func (User) TableName() string { return "users" }

// RegisterRequest — inbound DTO cho POST /v1/users/register.
// Tag binding áp dụng validation ở handler layer (HTTP 400 khi fail).
// Business rule (uniqueness, password hashing) enforce ở service layer.
type RegisterRequest struct {
	Username    string `json:"username"     binding:"required,min=3,max=32,alphanum"`
	Email       string `json:"email"        binding:"required,email,max=254"`
	Password    string `json:"password"     binding:"required,min=8,max=72"`
	DisplayName string `json:"display_name" binding:"required,min=1,max=64"`
}

// RegisterResponse wraps user dưới "data" + message — exact shape
// theo PDF assignment ("Register an user successfully!" — string lock).
type RegisterResponse struct {
	Data    User   `json:"data"`
	Message string `json:"message"`
}
