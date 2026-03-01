package services

import (
	"fmt"
	"moneymind-backend/internal/models"
	"moneymind-backend/internal/repository"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type AuthService struct {
	userRepo    *repository.UserRepository
	sessionRepo *repository.SessionRepository
	jwtSecret   string
}

type JWTClaims struct {
	UserID    int    `json:"user_id"`
	Email     string `json:"email"`
	SessionID int    `json:"session_id"`
	jwt.RegisteredClaims
}

func NewAuthService(userRepo *repository.UserRepository, sessionRepo *repository.SessionRepository, jwtSecret string) *AuthService {
	return &AuthService{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		jwtSecret:   jwtSecret,
	}
}

func (s *AuthService) Register(registration *models.UserRegistration) (*models.AuthResponse, error) {
	// Check if user already exists
	existingUser, err := s.userRepo.GetUserByEmail(registration.Email)
	if err == nil && existingUser != nil {
		return nil, fmt.Errorf("user with this email already exists")
	}

	// Create user
	user, err := s.userRepo.CreateUser(registration)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Create session and tokens
	return s.createAuthResponse(user, "", "", "")
}

func (s *AuthService) Login(login *models.UserLogin, deviceInfo, ipAddress, userAgent string) (*models.AuthResponse, error) {
	// Get user by email
	user, err := s.userRepo.GetUserByEmail(login.Email)
	if err != nil {
		return nil, fmt.Errorf("invalid email or password")
	}

	// Validate password
	if !s.userRepo.ValidatePassword(user.PasswordHash, login.Password) {
		return nil, fmt.Errorf("invalid email or password")
	}

	// Update last login
	err = s.userRepo.UpdateLastLogin(user.ID)
	if err != nil {
		// Log error but don't fail login
		fmt.Printf("Failed to update last login: %v\n", err)
	}

	// Create session and tokens
	return s.createAuthResponse(user, deviceInfo, ipAddress, userAgent)
}

func (s *AuthService) RefreshToken(refreshToken string) (*models.AuthResponse, error) {
	// Get session by refresh token
	session, err := s.sessionRepo.GetSessionByRefreshToken(refreshToken)
	if err != nil {
		return nil, fmt.Errorf("invalid or expired refresh token")
	}

	// Get user
	user, err := s.userRepo.GetUserByID(session.UserID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	// Generate new tokens
	accessToken, expiresIn, err := s.generateJWT(user.ID, user.Email, session.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	newRefreshToken := session.RefreshToken // Keep same refresh token
	expiresAt := time.Now().Add(time.Duration(expiresIn) * time.Second)
	refreshExpiresAt := time.Now().Add(30 * 24 * time.Hour) // 30 days

	// Update session with new tokens
	err = s.sessionRepo.UpdateSessionTokens(session.ID, accessToken, newRefreshToken, expiresAt, refreshExpiresAt)
	if err != nil {
		return nil, fmt.Errorf("failed to update session: %w", err)
	}

	return &models.AuthResponse{
		User:         *user,
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		ExpiresIn:    expiresIn,
	}, nil
}

func (s *AuthService) Logout(accessToken string) error {
	// Get session by access token
	session, err := s.sessionRepo.GetSessionByAccessToken(accessToken)
	if err != nil {
		return fmt.Errorf("invalid session")
	}

	// Invalidate session
	return s.sessionRepo.InvalidateSession(session.ID)
}

func (s *AuthService) ValidateToken(tokenString string) (*models.User, *models.Session, error) {
	// Parse token
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		return nil, nil, fmt.Errorf("invalid token: %w", err)
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, nil, fmt.Errorf("invalid token claims")
	}

	// Get session
	session, err := s.sessionRepo.GetSessionByAccessToken(tokenString)
	if err != nil {
		return nil, nil, fmt.Errorf("session not found or expired")
	}

	// Get user
	user, err := s.userRepo.GetUserByID(claims.UserID)
	if err != nil {
		return nil, nil, fmt.Errorf("user not found")
	}

	// Update last used
	err = s.sessionRepo.UpdateLastUsed(session.ID)
	if err != nil {
		// Log error but don't fail validation
		fmt.Printf("Failed to update last used: %v\n", err)
	}

	return user, session, nil
}

func (s *AuthService) LogoutAll(userID int) error {
	return s.sessionRepo.InvalidateAllUserSessions(userID)
}

func (s *AuthService) GetUserActiveSessions(userID int) ([]models.Session, error) {
	return s.sessionRepo.GetUserActiveSessions(userID)
}

func (s *AuthService) createAuthResponse(user *models.User, deviceInfo, ipAddress, userAgent string) (*models.AuthResponse, error) {
	// Create session first (temporary tokens)
	expiresAt := time.Now().Add(24 * time.Hour)        // 24 hours
	refreshExpiresAt := time.Now().Add(30 * 24 * time.Hour) // 30 days

	session, err := s.sessionRepo.CreateSession(user.ID, deviceInfo, ipAddress, userAgent, expiresAt, refreshExpiresAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	// Generate JWT
	accessToken, expiresIn, err := s.generateJWT(user.ID, user.Email, session.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	// Update session with real access token
	realExpiresAt := time.Now().Add(time.Duration(expiresIn) * time.Second)
	err = s.sessionRepo.UpdateSessionTokens(session.ID, accessToken, session.RefreshToken, realExpiresAt, refreshExpiresAt)
	if err != nil {
		return nil, fmt.Errorf("failed to update session tokens: %w", err)
	}

	return &models.AuthResponse{
		User:         *user,
		AccessToken:  accessToken,
		RefreshToken: session.RefreshToken,
		ExpiresIn:    expiresIn,
	}, nil
}

func (s *AuthService) generateJWT(userID int, email string, sessionID int) (string, int64, error) {
	expiresIn := int64(24 * 60 * 60) // 24 hours in seconds
	expirationTime := time.Now().Add(time.Duration(expiresIn) * time.Second)

	claims := &JWTClaims{
		UserID:    userID,
		Email:     email,
		SessionID: sessionID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "moneymind-backend",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", 0, err
	}

	return tokenString, expiresIn, nil
}