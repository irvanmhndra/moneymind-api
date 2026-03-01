package repository

import (
	"database/sql"
	"fmt"
	"moneymind-backend/internal/models"
	"strings"
	"time"
)

type AccountRepository struct {
	db *sql.DB
}

func NewAccountRepository(db *sql.DB) *AccountRepository {
	return &AccountRepository{db: db}
}

// TODO: Implement account repository methods
func (r *AccountRepository) GetAccountsByUserID(userID int) ([]models.Account, error) {
	query := `
		SELECT id, user_id, name, type, balance, currency, is_active, created_at, updated_at
		FROM accounts
		WHERE user_id = $1
		ORDER BY created_at DESC`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accounts []models.Account
	for rows.Next() {
		var account models.Account
		err := rows.Scan(
			&account.ID,
			&account.UserID,
			&account.Name,
			&account.Type,
			&account.Balance,
			&account.Currency,
			&account.IsActive,
			&account.CreatedAt,
			&account.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, account)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return accounts, nil
}

func (r *AccountRepository) GetAccountByID(id, userID int) (*models.Account, error) {
	query := `
		SELECT id, user_id, name, type, balance, currency, is_active, created_at, updated_at
		FROM accounts
		WHERE id = $1 AND user_id = $2`

	var account models.Account
	err := r.db.QueryRow(query, id, userID).Scan(
		&account.ID,
		&account.UserID,
		&account.Name,
		&account.Type,
		&account.Balance,
		&account.Currency,
		&account.IsActive,
		&account.CreatedAt,
		&account.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("account not found")
		}
		return nil, fmt.Errorf("failed to get account: %w", err)
	}

	return &account, nil
}

func (r *AccountRepository) CreateAccount(userID int, account *models.AccountCreate) (*models.Account, error) {
	query := `
		INSERT INTO accounts (user_id, name, type, balance, currency, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, true, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		RETURNING id, user_id, name, type, balance, currency, is_active, created_at, updated_at`

	var createdAccount models.Account

	// Set default values
	balance := account.Balance
	currency := account.Currency
	if currency == "" {
		currency = "USD"
	}

	err := r.db.QueryRow(query, userID, account.Name, account.Type, balance, currency).Scan(
		&createdAccount.ID,
		&createdAccount.UserID,
		&createdAccount.Name,
		&createdAccount.Type,
		&createdAccount.Balance,
		&createdAccount.Currency,
		&createdAccount.IsActive,
		&createdAccount.CreatedAt,
		&createdAccount.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &createdAccount, nil
}

func (r *AccountRepository) UpdateAccount(id, userID int, updates *models.AccountUpdate) (*models.Account, error) {
	var setParts []string
	var args []interface{}
	argIndex := 1

	// Build dynamic UPDATE query based on provided fields
	if updates.Name != "" {
		setParts = append(setParts, fmt.Sprintf("name = $%d", argIndex))
		args = append(args, updates.Name)
		argIndex++
	}

	if updates.Type != "" {
		setParts = append(setParts, fmt.Sprintf("type = $%d", argIndex))
		args = append(args, updates.Type)
		argIndex++
	}

	if updates.Balance != nil {
		setParts = append(setParts, fmt.Sprintf("balance = $%d", argIndex))
		args = append(args, *updates.Balance)
		argIndex++
	}

	if updates.Currency != "" {
		setParts = append(setParts, fmt.Sprintf("currency = $%d", argIndex))
		args = append(args, updates.Currency)
		argIndex++
	}

	if updates.IsActive != nil {
		setParts = append(setParts, fmt.Sprintf("is_active = $%d", argIndex))
		args = append(args, *updates.IsActive)
		argIndex++
	}

	if len(setParts) == 0 {
		return nil, fmt.Errorf("no fields to update")
	}

	// Add updated_at timestamp
	setParts = append(setParts, fmt.Sprintf("updated_at = $%d", argIndex))
	args = append(args, time.Now())
	argIndex++

	// Add WHERE conditions
	args = append(args, id, userID)

	query := fmt.Sprintf(`
		UPDATE accounts
		SET %s
		WHERE id = $%d AND user_id = $%d
		RETURNING id, user_id, name, type, balance, currency, is_active, created_at, updated_at`,
		strings.Join(setParts, ", "), argIndex, argIndex+1)

	var updatedAccount models.Account
	err := r.db.QueryRow(query, args...).Scan(
		&updatedAccount.ID,
		&updatedAccount.UserID,
		&updatedAccount.Name,
		&updatedAccount.Type,
		&updatedAccount.Balance,
		&updatedAccount.Currency,
		&updatedAccount.IsActive,
		&updatedAccount.CreatedAt,
		&updatedAccount.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("account not found or access denied")
		}
		return nil, fmt.Errorf("failed to update account: %w", err)
	}

	return &updatedAccount, nil
}

func (r *AccountRepository) DeleteAccount(id, userID int) error {
	// Check if account has associated expenses
	var expenseCount int
	checkQuery := `SELECT COUNT(*) FROM expenses WHERE account_id = $1`
	err := r.db.QueryRow(checkQuery, id).Scan(&expenseCount)
	if err != nil {
		return fmt.Errorf("failed to check account dependencies: %w", err)
	}

	if expenseCount > 0 {
		return fmt.Errorf("cannot delete account: %d expenses are associated with this account", expenseCount)
	}

	// Delete the account
	query := `DELETE FROM accounts WHERE id = $1 AND user_id = $2`
	result, err := r.db.Exec(query, id, userID)
	if err != nil {
		return fmt.Errorf("failed to delete account: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("account not found or access denied")
	}

	return nil
}