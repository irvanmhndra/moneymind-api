-- Migration: 002_fix_split_expenses.sql
-- Description: Fix split expenses table to match model definitions
-- Date: 2025-09-29

-- Add missing columns to split_expenses table
ALTER TABLE split_expenses
ADD COLUMN IF NOT EXISTS split_type VARCHAR(20) DEFAULT 'equal' CHECK (split_type IN ('equal', 'percentage', 'amount')),
ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;

-- Update split_participants table to match model
ALTER TABLE split_participants
RENAME COLUMN participant_name TO name;

ALTER TABLE split_participants
RENAME COLUMN participant_email TO email;

ALTER TABLE split_participants
ADD COLUMN IF NOT EXISTS user_id INTEGER REFERENCES users(id),
ADD COLUMN IF NOT EXISTS amount_paid DECIMAL(15,2) DEFAULT 0.00,
ADD COLUMN IF NOT EXISTS created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;

-- Add trigger for split_expenses updated_at
CREATE TRIGGER update_split_expenses_updated_at
    BEFORE UPDATE ON split_expenses FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Add trigger for split_participants updated_at
CREATE TRIGGER update_split_participants_updated_at
    BEFORE UPDATE ON split_participants FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Create indexes for new columns
CREATE INDEX IF NOT EXISTS idx_split_expenses_split_type ON split_expenses(split_type);
CREATE INDEX IF NOT EXISTS idx_split_participants_user_id ON split_participants(user_id);

-- Record this migration
INSERT INTO schema_migrations (version, description)
VALUES ('002_fix_split_expenses', 'Fix split expenses table to match model definitions')
ON CONFLICT DO NOTHING;