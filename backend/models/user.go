package models

import (
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
)

const (
	RoleAdmin  = "admin"
	RoleEditor = "editor"

	AuthProviderLocal = "local"

	MinPasswordLength = 8
)

var ErrPasswordTooShort = errors.New("password must be at least 8 characters")

// User is an account that can sign in to the admin panel.
// Admins can manage everything; editors can only manage events (changelog entries).
type User struct {
	ID           uint       `json:"id" gorm:"primaryKey"`
	Username     string     `json:"username" gorm:"not null;uniqueIndex"`
	Email        string     `json:"email" gorm:"index"`
	DisplayName  string     `json:"display_name"`
	PasswordHash string     `json:"-"`
	Role         string     `json:"role" gorm:"not null;default:editor"`
	Active       bool       `json:"active" gorm:"not null;default:true"`
	AuthProvider string     `json:"auth_provider" gorm:"not null;default:local"` // "local" for now; reserved for SSO providers such as Entra ID
	ExternalID   *string    `json:"external_id,omitempty" gorm:"uniqueIndex"`    // ID of the user at the SSO provider (e.g. Entra object ID)
	LastLoginAt  *time.Time `json:"last_login_at"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

func IsValidRole(role string) bool {
	return role == RoleAdmin || role == RoleEditor
}

// SetPassword hashes and stores the given password.
func (u *User) SetPassword(password string) error {
	if len(password) < MinPasswordLength {
		return ErrPasswordTooShort
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.PasswordHash = string(hash)
	return nil
}

// CheckPassword reports whether the given password matches the stored hash.
func (u *User) CheckPassword(password string) bool {
	if u.PasswordHash == "" {
		return false
	}
	return bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)) == nil
}
