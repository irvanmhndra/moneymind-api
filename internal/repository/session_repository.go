package repository

import (
	"database/sql"
	"fmt"
	"moneymind-backend/internal/models"
	"time"

	"github.com/google/uuid"
)

type SessionRepository struct {
	db *sql.DB
}

func NewSessionRepository(db *sql.DB) *SessionRepository {
	return &SessionRepository{db: db}
}

func (r *SessionRepository) CreateSession(userID int, deviceInfo, ipAddress, userAgent string, expiresAt, refreshExpiresAt time.Time) (*models.Session, error) {
	accessToken := uuid.New().String()
	refreshToken := uuid.New().String()

	// Convert device info to JSON format
	deviceInfoJSON := fmt.Sprintf(`{"user_agent": "%s", "ip_address": "%s"}`, userAgent, ipAddress)

	query := `
		INSERT INTO sessions (user_id, access_token, refresh_token, device_info, ip_address, user_agent, expires_at, refresh_expires_at)
		VALUES ($1, $2, $3, $4::jsonb, $5, $6, $7, $8)
		RETURNING id, user_id, access_token, refresh_token, device_info, ip_address, user_agent, is_active, expires_at, refresh_expires_at, last_used_at, created_at
	`

	var session models.Session
	err := r.db.QueryRow(
		query,
		userID,
		accessToken,
		refreshToken,
		deviceInfoJSON,
		ipAddress,
		userAgent,
		expiresAt,
		refreshExpiresAt,
	).Scan(
		&session.ID,
		&session.UserID,
		&session.AccessToken,
		&session.RefreshToken,
		&session.DeviceInfo,
		&session.IPAddress,
		&session.UserAgent,
		&session.IsActive,
		&session.ExpiresAt,
		&session.RefreshExpiresAt,
		&session.LastUsedAt,
		&session.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return &session, nil
}

func (r *SessionRepository) GetSessionByAccessToken(accessToken string) (*models.Session, error) {
	query := `
		SELECT id, user_id, access_token, refresh_token, device_info, ip_address, user_agent, 
			   is_active, expires_at, refresh_expires_at, last_used_at, created_at
		FROM sessions 
		WHERE access_token = $1 AND is_active = true AND expires_at > NOW()
	`

	var session models.Session
	err := r.db.QueryRow(query, accessToken).Scan(
		&session.ID,
		&session.UserID,
		&session.AccessToken,
		&session.RefreshToken,
		&session.DeviceInfo,
		&session.IPAddress,
		&session.UserAgent,
		&session.IsActive,
		&session.ExpiresAt,
		&session.RefreshExpiresAt,
		&session.LastUsedAt,
		&session.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("session not found or expired")
		}
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	return &session, nil
}

func (r *SessionRepository) GetSessionByRefreshToken(refreshToken string) (*models.Session, error) {
	query := `
		SELECT id, user_id, access_token, refresh_token, device_info, ip_address, user_agent, 
			   is_active, expires_at, refresh_expires_at, last_used_at, created_at
		FROM sessions 
		WHERE refresh_token = $1 AND is_active = true AND refresh_expires_at > NOW()
	`

	var session models.Session
	err := r.db.QueryRow(query, refreshToken).Scan(
		&session.ID,
		&session.UserID,
		&session.AccessToken,
		&session.RefreshToken,
		&session.DeviceInfo,
		&session.IPAddress,
		&session.UserAgent,
		&session.IsActive,
		&session.ExpiresAt,
		&session.RefreshExpiresAt,
		&session.LastUsedAt,
		&session.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("session not found or expired")
		}
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	return &session, nil
}

func (r *SessionRepository) UpdateSessionTokens(sessionID int, accessToken, refreshToken string, expiresAt, refreshExpiresAt time.Time) error {
	query := `
		UPDATE sessions 
		SET access_token = $1, refresh_token = $2, expires_at = $3, refresh_expires_at = $4, last_used_at = NOW()
		WHERE id = $5 AND is_active = true
	`

	result, err := r.db.Exec(query, accessToken, refreshToken, expiresAt, refreshExpiresAt, sessionID)
	if err != nil {
		return fmt.Errorf("failed to update session: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("session not found or inactive")
	}

	return nil
}

func (r *SessionRepository) UpdateLastUsed(sessionID int) error {
	query := `UPDATE sessions SET last_used_at = NOW() WHERE id = $1 AND is_active = true`
	
	_, err := r.db.Exec(query, sessionID)
	if err != nil {
		return fmt.Errorf("failed to update last used: %w", err)
	}

	return nil
}

func (r *SessionRepository) InvalidateSession(sessionID int) error {
	query := `UPDATE sessions SET is_active = false WHERE id = $1`
	
	_, err := r.db.Exec(query, sessionID)
	if err != nil {
		return fmt.Errorf("failed to invalidate session: %w", err)
	}

	return nil
}

func (r *SessionRepository) InvalidateAllUserSessions(userID int) error {
	query := `UPDATE sessions SET is_active = false WHERE user_id = $1`
	
	_, err := r.db.Exec(query, userID)
	if err != nil {
		return fmt.Errorf("failed to invalidate user sessions: %w", err)
	}

	return nil
}

func (r *SessionRepository) CleanupExpiredSessions() error {
	query := `
		UPDATE sessions 
		SET is_active = false 
		WHERE (expires_at < NOW() OR refresh_expires_at < NOW()) AND is_active = true
	`
	
	_, err := r.db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to cleanup expired sessions: %w", err)
	}

	return nil
}

func (r *SessionRepository) GetUserActiveSessions(userID int) ([]models.Session, error) {
	query := `
		SELECT id, user_id, access_token, refresh_token, device_info, ip_address, user_agent, 
			   is_active, expires_at, refresh_expires_at, last_used_at, created_at
		FROM sessions 
		WHERE user_id = $1 AND is_active = true AND refresh_expires_at > NOW()
		ORDER BY last_used_at DESC
	`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user sessions: %w", err)
	}
	defer rows.Close()

	var sessions []models.Session
	for rows.Next() {
		var session models.Session
		err := rows.Scan(
			&session.ID,
			&session.UserID,
			&session.AccessToken,
			&session.RefreshToken,
			&session.DeviceInfo,
			&session.IPAddress,
			&session.UserAgent,
			&session.IsActive,
			&session.ExpiresAt,
			&session.RefreshExpiresAt,
			&session.LastUsedAt,
			&session.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan session: %w", err)
		}
		sessions = append(sessions, session)
	}

	return sessions, nil
}