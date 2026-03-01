package repository

import (
	"database/sql"
	"fmt"
	"moneymind-backend/internal/models"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) CreateUser(user *models.UserRegistration) (*models.User, error) {
	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Set defaults
	if user.Currency == "" {
		user.Currency = "USD"
	}
	if user.Timezone == "" {
		user.Timezone = "UTC"
	}

	query := `
		INSERT INTO users (email, password_hash, first_name, last_name, currency, timezone)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, email, first_name, last_name, currency, timezone, is_active, email_verified, last_login_at, created_at, updated_at
	`

	var newUser models.User
	err = r.db.QueryRow(
		query,
		user.Email,
		string(hashedPassword),
		user.FirstName,
		user.LastName,
		user.Currency,
		user.Timezone,
	).Scan(
		&newUser.ID,
		&newUser.Email,
		&newUser.FirstName,
		&newUser.LastName,
		&newUser.Currency,
		&newUser.Timezone,
		&newUser.IsActive,
		&newUser.EmailVerified,
		&newUser.LastLoginAt,
		&newUser.CreatedAt,
		&newUser.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return &newUser, nil
}

func (r *UserRepository) GetUserByEmail(email string) (*models.User, error) {
	query := `
		SELECT id, email, password_hash, first_name, last_name, currency, timezone, 
			   is_active, email_verified, last_login_at, created_at, updated_at
		FROM users 
		WHERE email = $1 AND is_active = true
	`

	var user models.User
	err := r.db.QueryRow(query, email).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.FirstName,
		&user.LastName,
		&user.Currency,
		&user.Timezone,
		&user.IsActive,
		&user.EmailVerified,
		&user.LastLoginAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

func (r *UserRepository) GetUserByID(id int) (*models.User, error) {
	query := `
		SELECT id, email, first_name, last_name, currency, timezone, 
			   is_active, email_verified, last_login_at, created_at, updated_at
		FROM users 
		WHERE id = $1 AND is_active = true
	`

	var user models.User
	err := r.db.QueryRow(query, id).Scan(
		&user.ID,
		&user.Email,
		&user.FirstName,
		&user.LastName,
		&user.Currency,
		&user.Timezone,
		&user.IsActive,
		&user.EmailVerified,
		&user.LastLoginAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

func (r *UserRepository) UpdateUser(id int, updates *models.UserUpdate) (*models.User, error) {
	setParts := []string{}
	args := []interface{}{}
	argIndex := 1

	if updates.FirstName != "" {
		setParts = append(setParts, fmt.Sprintf("first_name = $%d", argIndex))
		args = append(args, updates.FirstName)
		argIndex++
	}
	if updates.LastName != "" {
		setParts = append(setParts, fmt.Sprintf("last_name = $%d", argIndex))
		args = append(args, updates.LastName)
		argIndex++
	}
	if updates.Currency != "" {
		setParts = append(setParts, fmt.Sprintf("currency = $%d", argIndex))
		args = append(args, updates.Currency)
		argIndex++
	}
	if updates.Timezone != "" {
		setParts = append(setParts, fmt.Sprintf("timezone = $%d", argIndex))
		args = append(args, updates.Timezone)
		argIndex++
	}

	if len(setParts) == 0 {
		return r.GetUserByID(id)
	}

	setParts = append(setParts, fmt.Sprintf("updated_at = $%d", argIndex))
	args = append(args, time.Now())
	argIndex++

	args = append(args, id)
	setClause := strings.Join(setParts, ", ")
	
	query := fmt.Sprintf(`
		UPDATE users 
		SET %s 
		WHERE id = $%d AND is_active = true
		RETURNING id, email, first_name, last_name, currency, timezone, is_active, email_verified, last_login_at, created_at, updated_at
	`, setClause, argIndex)

	var user models.User
	err := r.db.QueryRow(query, args...).Scan(
		&user.ID,
		&user.Email,
		&user.FirstName,
		&user.LastName,
		&user.Currency,
		&user.Timezone,
		&user.IsActive,
		&user.EmailVerified,
		&user.LastLoginAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return &user, nil
}

func (r *UserRepository) UpdateLastLogin(id int) error {
	query := `UPDATE users SET last_login_at = $1, updated_at = $1 WHERE id = $2`
	
	_, err := r.db.Exec(query, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update last login: %w", err)
	}

	return nil
}

func (r *UserRepository) ValidatePassword(hashedPassword, plainPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(plainPassword))
	return err == nil
}

func (r *UserRepository) ChangePassword(id int, oldPassword, newPassword string) error {
	// First get the current password hash
	var currentHash string
	query := `SELECT password_hash FROM users WHERE id = $1 AND is_active = true`
	err := r.db.QueryRow(query, id).Scan(&currentHash)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("user not found")
		}
		return fmt.Errorf("failed to get user password: %w", err)
	}

	// Validate current password
	if !r.ValidatePassword(currentHash, oldPassword) {
		return fmt.Errorf("current password is incorrect")
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash new password: %w", err)
	}

	// Update password
	updateQuery := `UPDATE users SET password_hash = $1, updated_at = $2 WHERE id = $3`
	_, err = r.db.Exec(updateQuery, string(hashedPassword), time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}

func (r *UserRepository) DeactivateUser(id int) error {
	query := `UPDATE users SET is_active = false, updated_at = $1 WHERE id = $2`
	result, err := r.db.Exec(query, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to deactivate user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

func (r *UserRepository) VerifyEmail(id int) error {
	query := `UPDATE users SET email_verified = true, updated_at = $1 WHERE id = $2`
	result, err := r.db.Exec(query, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to verify email: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}