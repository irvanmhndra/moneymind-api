package models

import (
	"time"
)

type User struct {
	ID            int       `json:"id" db:"id"`
	Email         string    `json:"email" db:"email" validate:"required,email"`
	PasswordHash  string    `json:"-" db:"password_hash"`
	FirstName     string    `json:"first_name" db:"first_name" validate:"required,min=2,max=50"`
	LastName      string    `json:"last_name" db:"last_name" validate:"required,min=2,max=50"`
	Currency      string    `json:"currency" db:"currency" validate:"len=3"`
	Timezone      string    `json:"timezone" db:"timezone"`
	IsActive      bool      `json:"is_active" db:"is_active"`
	EmailVerified bool      `json:"email_verified" db:"email_verified"`
	LastLoginAt   *time.Time `json:"last_login_at" db:"last_login_at"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
}

type UserRegistration struct {
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required,min=8"`
	FirstName string `json:"first_name" validate:"required,min=2,max=50"`
	LastName  string `json:"last_name" validate:"required,min=2,max=50"`
	Currency  string `json:"currency" validate:"omitempty,len=3"`
	Timezone  string `json:"timezone" validate:"omitempty"`
}

type UserLogin struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type UserUpdate struct {
	FirstName string `json:"first_name" validate:"omitempty,min=2,max=50"`
	LastName  string `json:"last_name" validate:"omitempty,min=2,max=50"`
	Currency  string `json:"currency" validate:"omitempty,len=3"`
	Timezone  string `json:"timezone" validate:"omitempty"`
}

type AuthResponse struct {
	User         User   `json:"user"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

type Session struct {
	ID                 int       `json:"id" db:"id"`
	UserID             int       `json:"user_id" db:"user_id"`
	AccessToken        string    `json:"access_token" db:"access_token"`
	RefreshToken       string    `json:"refresh_token" db:"refresh_token"`
	DeviceInfo         string    `json:"device_info" db:"device_info"`
	IPAddress          string    `json:"ip_address" db:"ip_address"`
	UserAgent          string    `json:"user_agent" db:"user_agent"`
	IsActive           bool      `json:"is_active" db:"is_active"`
	ExpiresAt          time.Time `json:"expires_at" db:"expires_at"`
	RefreshExpiresAt   time.Time `json:"refresh_expires_at" db:"refresh_expires_at"`
	LastUsedAt         time.Time `json:"last_used_at" db:"last_used_at"`
	CreatedAt          time.Time `json:"created_at" db:"created_at"`
}