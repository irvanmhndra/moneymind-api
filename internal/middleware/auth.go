package middleware

import (
	"moneymind-backend/internal/models"
	"moneymind-backend/internal/services"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func AuthRequired(authService *services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Authorization header required",
			})
			c.Abort()
			return
		}

		// Check if it's a Bearer token
		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Invalid authorization header format",
			})
			c.Abort()
			return
		}

		// Extract token
		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Token required",
			})
			c.Abort()
			return
		}

		// Validate token
		user, session, err := authService.ValidateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Invalid or expired token",
			})
			c.Abort()
			return
		}

		// Set user and session in context
		c.Set("user", user)
		c.Set("session", session)
		c.Set("user_id", user.ID)

		c.Next()
	}
}

func OptionalAuth(authService *services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		// Check if it's a Bearer token
		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.Next()
			return
		}

		// Extract token
		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == "" {
			c.Next()
			return
		}

		// Try to validate token
		user, session, err := authService.ValidateToken(token)
		if err == nil && user != nil {
			// Set user and session in context if valid
			c.Set("user", user)
			c.Set("session", session)
			c.Set("user_id", user.ID)
		}

		c.Next()
	}
}

// Helper function to get user from context
func GetUserFromContext(c *gin.Context) (*models.User, bool) {
	if user, exists := c.Get("user"); exists {
		if u, ok := user.(*models.User); ok {
			return u, true
		}
	}
	return nil, false
}

// Helper function to get user ID from context
func GetUserIDFromContext(c *gin.Context) (int, bool) {
	if userID, exists := c.Get("user_id"); exists {
		if id, ok := userID.(int); ok {
			return id, true
		}
	}
	return 0, false
}