package config

import (
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	// Server configuration
	Port        string
	Environment string
	GinMode     string

	// Database configuration
	DatabaseURL string

	// JWT configuration
	JWTSecret              string
	JWTExpiresIn           time.Duration
	RefreshTokenExpiresIn  time.Duration

	// File upload configuration
	UploadDir       string
	MaxUploadSize   int64

	// CORS configuration
	AllowedOrigins []string
	AllowedMethods []string
	AllowedHeaders []string

	// Email configuration
	SMTPHost string
	SMTPPort int
	SMTPUser string
	SMTPPass string
	FromEmail string
	FromName  string

	// Security configuration
	BcryptCost        int
	RateLimitRequests int
	RateLimitDuration time.Duration

	// Logging configuration
	LogLevel  string
	LogFormat string
}

func Load() (*Config, error) {
	// Load .env file if it exists
	_ = godotenv.Load()

	cfg := &Config{
		// Server defaults
		Port:        getEnv("PORT", "8080"),
		Environment: getEnv("ENVIRONMENT", "development"),
		GinMode:     getEnv("GIN_MODE", "debug"),

		// Database
		DatabaseURL: getEnv("DATABASE_URL", "postgresql://postgres:password@localhost:5432/moneymind"),

		// JWT
		JWTSecret:              getEnv("JWT_SECRET", "your-super-secret-jwt-key"),
		JWTExpiresIn:           parseDuration("JWT_EXPIRES_IN", "24h"),
		RefreshTokenExpiresIn:  parseDuration("REFRESH_TOKEN_EXPIRES_IN", "720h"),

		// File upload
		UploadDir:     getEnv("UPLOAD_DIR", "./uploads"),
		MaxUploadSize: parseSize("MAX_UPLOAD_SIZE", "10MB"),

		// CORS
		AllowedOrigins: strings.Split(getEnv("ALLOWED_ORIGINS", "http://localhost:3000,http://localhost:5173"), ","),
		AllowedMethods: strings.Split(getEnv("ALLOWED_METHODS", "GET,POST,PUT,DELETE,OPTIONS"), ","),
		AllowedHeaders: strings.Split(getEnv("ALLOWED_HEADERS", "Content-Type,Authorization"), ","),

		// Email
		SMTPHost:  getEnv("SMTP_HOST", "smtp.gmail.com"),
		SMTPPort:  parseInt("SMTP_PORT", "587"),
		SMTPUser:  getEnv("SMTP_USER", ""),
		SMTPPass:  getEnv("SMTP_PASS", ""),
		FromEmail: getEnv("FROM_EMAIL", "noreply@moneymind.com"),
		FromName:  getEnv("FROM_NAME", "MoneyMind"),

		// Security
		BcryptCost:        parseInt("BCRYPT_COST", "12"),
		RateLimitRequests: parseInt("RATE_LIMIT_REQUESTS", "100"),
		RateLimitDuration: parseDuration("RATE_LIMIT_DURATION", "1h"),

		// Logging
		LogLevel:  getEnv("LOG_LEVEL", "info"),
		LogFormat: getEnv("LOG_FORMAT", "json"),
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func parseInt(key, defaultValue string) int {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			return parsed
		}
	}
	parsed, _ := strconv.Atoi(defaultValue)
	return parsed
}

func parseDuration(key, defaultValue string) time.Duration {
	if value := os.Getenv(key); value != "" {
		if parsed, err := time.ParseDuration(value); err == nil {
			return parsed
		}
	}
	parsed, _ := time.ParseDuration(defaultValue)
	return parsed
}

func parseSize(key, defaultValue string) int64 {
	value := getEnv(key, defaultValue)
	
	// Remove spaces and convert to lowercase
	value = strings.ToLower(strings.ReplaceAll(value, " ", ""))
	
	// Parse the numeric part
	var size int64
	var unit string
	
	if strings.HasSuffix(value, "gb") {
		unit = "gb"
		value = strings.TrimSuffix(value, "gb")
	} else if strings.HasSuffix(value, "mb") {
		unit = "mb"
		value = strings.TrimSuffix(value, "mb")
	} else if strings.HasSuffix(value, "kb") {
		unit = "kb"
		value = strings.TrimSuffix(value, "kb")
	} else {
		unit = "b"
	}
	
	if parsed, err := strconv.ParseInt(value, 10, 64); err == nil {
		size = parsed
	} else {
		size = 10 // default to 10MB
		unit = "mb"
	}
	
	// Convert to bytes
	switch unit {
	case "gb":
		return size * 1024 * 1024 * 1024
	case "mb":
		return size * 1024 * 1024
	case "kb":
		return size * 1024
	default:
		return size
	}
}