-- Migration: 002_update_token_fields.sql
-- Description: Update token fields to handle JWT token lengths properly
-- Date: 2025-01-09

-- Update token fields to proper length for JWT tokens
-- Standard JWT tokens are typically 150-1000 characters
ALTER TABLE sessions ALTER COLUMN access_token TYPE VARCHAR(1000);
ALTER TABLE sessions ALTER COLUMN refresh_token TYPE VARCHAR(1000);

-- Update password reset and email verification tokens as well
ALTER TABLE password_reset_tokens ALTER COLUMN token TYPE VARCHAR(1000);
ALTER TABLE email_verification_tokens ALTER COLUMN token TYPE VARCHAR(1000);

-- Record migration
INSERT INTO schema_migrations (version, description) 
VALUES ('002_update_token_fields', 'Update token fields to handle JWT token lengths') 
ON CONFLICT DO NOTHING;