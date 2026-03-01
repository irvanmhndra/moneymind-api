package repository

import (
	"database/sql"
	"fmt"
	"moneymind-backend/internal/models"
	"strings"
	"time"
)

type ExpenseRepository struct {
	db *sql.DB
}

func NewExpenseRepository(db *sql.DB) *ExpenseRepository {
	return &ExpenseRepository{db: db}
}

func (r *ExpenseRepository) CreateExpense(userID int, expense *models.ExpenseCreate) (*models.Expense, error) {
	if expense.Currency == "" {
		expense.Currency = "USD"
	}

	query := `
		INSERT INTO expenses (user_id, account_id, category_id, amount, currency, description, notes, location, date, is_recurring, recurring_frequency, tax_deductible, tax_category)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING id, user_id, account_id, category_id, amount, currency, description, notes, location, receipt_path, date, is_recurring, recurring_frequency, tax_deductible, tax_category, created_at, updated_at
	`

	var newExpense models.Expense
	err := r.db.QueryRow(
		query,
		userID,
		expense.AccountID,
		expense.CategoryID,
		expense.Amount,
		expense.Currency,
		expense.Description,
		expense.Notes,
		expense.Location,
		expense.Date,
		expense.IsRecurring,
		expense.RecurringFrequency,
		expense.TaxDeductible,
		expense.TaxCategory,
	).Scan(
		&newExpense.ID,
		&newExpense.UserID,
		&newExpense.AccountID,
		&newExpense.CategoryID,
		&newExpense.Amount,
		&newExpense.Currency,
		&newExpense.Description,
		&newExpense.Notes,
		&newExpense.Location,
		&newExpense.ReceiptPath,
		&newExpense.Date,
		&newExpense.IsRecurring,
		&newExpense.RecurringFrequency,
		&newExpense.TaxDeductible,
		&newExpense.TaxCategory,
		&newExpense.CreatedAt,
		&newExpense.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create expense: %w", err)
	}

	return &newExpense, nil
}

func (r *ExpenseRepository) GetExpensesByUserID(userID int, filter *models.ExpenseFilter) ([]models.Expense, error) {
	whereConditions := []string{"e.user_id = $1"}
	args := []interface{}{userID}
	argIndex := 2

	if filter != nil {
		if filter.StartDate != nil {
			whereConditions = append(whereConditions, fmt.Sprintf("e.date >= $%d", argIndex))
			args = append(args, *filter.StartDate)
			argIndex++
		}
		if filter.EndDate != nil {
			whereConditions = append(whereConditions, fmt.Sprintf("e.date <= $%d", argIndex))
			args = append(args, *filter.EndDate)
			argIndex++
		}
		if filter.CategoryID != nil {
			whereConditions = append(whereConditions, fmt.Sprintf("e.category_id = $%d", argIndex))
			args = append(args, *filter.CategoryID)
			argIndex++
		}
		if filter.AccountID != nil {
			whereConditions = append(whereConditions, fmt.Sprintf("e.account_id = $%d", argIndex))
			args = append(args, *filter.AccountID)
			argIndex++
		}
		if filter.MinAmount != nil {
			whereConditions = append(whereConditions, fmt.Sprintf("e.amount >= $%d", argIndex))
			args = append(args, *filter.MinAmount)
			argIndex++
		}
		if filter.MaxAmount != nil {
			whereConditions = append(whereConditions, fmt.Sprintf("e.amount <= $%d", argIndex))
			args = append(args, *filter.MaxAmount)
			argIndex++
		}
		if filter.Search != "" {
			whereConditions = append(whereConditions, fmt.Sprintf("(e.description ILIKE $%d OR e.notes ILIKE $%d)", argIndex, argIndex))
			args = append(args, "%"+filter.Search+"%")
			argIndex++
		}
		if filter.TaxDeductible != nil {
			whereConditions = append(whereConditions, fmt.Sprintf("e.tax_deductible = $%d", argIndex))
			args = append(args, *filter.TaxDeductible)
			argIndex++
		}
	}

	orderBy := "ORDER BY e.date DESC"
	if filter != nil && filter.SortBy != "" {
		orderDirection := "DESC"
		if filter.SortOrder == "asc" {
			orderDirection = "ASC"
		}
		orderBy = fmt.Sprintf("ORDER BY e.%s %s", filter.SortBy, orderDirection)
	}

	limit := ""
	if filter != nil && filter.Limit > 0 {
		limit = fmt.Sprintf("LIMIT $%d", argIndex)
		args = append(args, filter.Limit)
		argIndex++

		if filter.Page > 0 {
			offset := (filter.Page - 1) * filter.Limit
			limit += fmt.Sprintf(" OFFSET $%d", argIndex)
			args = append(args, offset)
		}
	}

	query := fmt.Sprintf(`
		SELECT e.id, e.user_id, e.account_id, e.category_id, e.amount, e.currency, 
			   e.description, e.notes, e.location, e.receipt_path, e.date, 
			   e.is_recurring, e.recurring_frequency, e.tax_deductible, e.tax_category, 
			   e.created_at, e.updated_at,
			   a.name as account_name, c.name as category_name
		FROM expenses e
		LEFT JOIN accounts a ON e.account_id = a.id
		LEFT JOIN categories c ON e.category_id = c.id
		WHERE %s
		%s %s
	`, strings.Join(whereConditions, " AND "), orderBy, limit)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get expenses: %w", err)
	}
	defer rows.Close()

	var expenses []models.Expense
	for rows.Next() {
		var expense models.Expense
		var accountName, categoryName sql.NullString

		err := rows.Scan(
			&expense.ID,
			&expense.UserID,
			&expense.AccountID,
			&expense.CategoryID,
			&expense.Amount,
			&expense.Currency,
			&expense.Description,
			&expense.Notes,
			&expense.Location,
			&expense.ReceiptPath,
			&expense.Date,
			&expense.IsRecurring,
			&expense.RecurringFrequency,
			&expense.TaxDeductible,
			&expense.TaxCategory,
			&expense.CreatedAt,
			&expense.UpdatedAt,
			&accountName,
			&categoryName,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan expense: %w", err)
		}

		if accountName.Valid {
			expense.Account = &models.Account{Name: accountName.String}
		}
		if categoryName.Valid {
			expense.Category = &models.Category{Name: categoryName.String}
		}

		expenses = append(expenses, expense)
	}

	return expenses, nil
}

func (r *ExpenseRepository) GetExpenseByID(id, userID int) (*models.Expense, error) {
	query := `
		SELECT e.id, e.user_id, e.account_id, e.category_id, e.amount, e.currency, 
			   e.description, e.notes, e.location, e.receipt_path, e.date, 
			   e.is_recurring, e.recurring_frequency, e.tax_deductible, e.tax_category, 
			   e.created_at, e.updated_at,
			   a.name as account_name, c.name as category_name
		FROM expenses e
		LEFT JOIN accounts a ON e.account_id = a.id
		LEFT JOIN categories c ON e.category_id = c.id
		WHERE e.id = $1 AND e.user_id = $2
	`

	var expense models.Expense
	var accountName, categoryName sql.NullString

	err := r.db.QueryRow(query, id, userID).Scan(
		&expense.ID,
		&expense.UserID,
		&expense.AccountID,
		&expense.CategoryID,
		&expense.Amount,
		&expense.Currency,
		&expense.Description,
		&expense.Notes,
		&expense.Location,
		&expense.ReceiptPath,
		&expense.Date,
		&expense.IsRecurring,
		&expense.RecurringFrequency,
		&expense.TaxDeductible,
		&expense.TaxCategory,
		&expense.CreatedAt,
		&expense.UpdatedAt,
		&accountName,
		&categoryName,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("expense not found")
		}
		return nil, fmt.Errorf("failed to get expense: %w", err)
	}

	if accountName.Valid {
		expense.Account = &models.Account{Name: accountName.String}
	}
	if categoryName.Valid {
		expense.Category = &models.Category{Name: categoryName.String}
	}

	return &expense, nil
}

func (r *ExpenseRepository) UpdateExpense(id, userID int, updates *models.ExpenseUpdate) (*models.Expense, error) {
	setParts := []string{}
	args := []interface{}{}
	argIndex := 1

	if updates.Amount != nil {
		setParts = append(setParts, fmt.Sprintf("amount = $%d", argIndex))
		args = append(args, *updates.Amount)
		argIndex++
	}
	if updates.Currency != "" {
		setParts = append(setParts, fmt.Sprintf("currency = $%d", argIndex))
		args = append(args, updates.Currency)
		argIndex++
	}
	if updates.Description != "" {
		setParts = append(setParts, fmt.Sprintf("description = $%d", argIndex))
		args = append(args, updates.Description)
		argIndex++
	}
	if updates.Notes != "" {
		setParts = append(setParts, fmt.Sprintf("notes = $%d", argIndex))
		args = append(args, updates.Notes)
		argIndex++
	}
	if updates.Location != "" {
		setParts = append(setParts, fmt.Sprintf("location = $%d", argIndex))
		args = append(args, updates.Location)
		argIndex++
	}
	if updates.Date != nil {
		setParts = append(setParts, fmt.Sprintf("date = $%d", argIndex))
		args = append(args, *updates.Date)
		argIndex++
	}
	if updates.AccountID != nil {
		setParts = append(setParts, fmt.Sprintf("account_id = $%d", argIndex))
		args = append(args, *updates.AccountID)
		argIndex++
	}
	if updates.CategoryID != nil {
		setParts = append(setParts, fmt.Sprintf("category_id = $%d", argIndex))
		args = append(args, *updates.CategoryID)
		argIndex++
	}
	if updates.IsRecurring != nil {
		setParts = append(setParts, fmt.Sprintf("is_recurring = $%d", argIndex))
		args = append(args, *updates.IsRecurring)
		argIndex++
	}
	if updates.RecurringFrequency != "" {
		setParts = append(setParts, fmt.Sprintf("recurring_frequency = $%d", argIndex))
		args = append(args, updates.RecurringFrequency)
		argIndex++
	}
	if updates.TaxDeductible != nil {
		setParts = append(setParts, fmt.Sprintf("tax_deductible = $%d", argIndex))
		args = append(args, *updates.TaxDeductible)
		argIndex++
	}
	if updates.TaxCategory != "" {
		setParts = append(setParts, fmt.Sprintf("tax_category = $%d", argIndex))
		args = append(args, updates.TaxCategory)
		argIndex++
	}

	if len(setParts) == 0 {
		return r.GetExpenseByID(id, userID)
	}

	setParts = append(setParts, fmt.Sprintf("updated_at = $%d", argIndex))
	args = append(args, time.Now())
	argIndex++

	args = append(args, id, userID)
	setClause := strings.Join(setParts, ", ")

	query := fmt.Sprintf(`
		UPDATE expenses 
		SET %s 
		WHERE id = $%d AND user_id = $%d
	`, setClause, argIndex, argIndex+1)

	result, err := r.db.Exec(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to update expense: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return nil, fmt.Errorf("expense not found")
	}

	return r.GetExpenseByID(id, userID)
}

func (r *ExpenseRepository) DeleteExpense(id, userID int) error {
	query := `DELETE FROM expenses WHERE id = $1 AND user_id = $2`

	result, err := r.db.Exec(query, id, userID)
	if err != nil {
		return fmt.Errorf("failed to delete expense: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("expense not found")
	}

	return nil
}

func (r *ExpenseRepository) UpdateExpenseReceiptPath(id, userID int, receiptPath string) (*models.Expense, error) {
	query := `
		UPDATE expenses 
		SET receipt_path = $1, updated_at = $2 
		WHERE id = $3 AND user_id = $4
	`

	result, err := r.db.Exec(query, receiptPath, time.Now(), id, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to update expense receipt path: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return nil, fmt.Errorf("expense not found")
	}

	return r.GetExpenseByID(id, userID)
}

func (r *ExpenseRepository) GetAnalytics(userID int, startDate, endDate, period, categoryIDStr string) (*models.ExpenseAnalytics, error) {
	analytics := &models.ExpenseAnalytics{}

	// Build date filter
	whereClause := "WHERE e.user_id = $1"
	args := []interface{}{userID}
	argIndex := 2

	if startDate != "" {
		whereClause += fmt.Sprintf(" AND e.date >= $%d", argIndex)
		args = append(args, startDate)
		argIndex++
	}

	if endDate != "" {
		whereClause += fmt.Sprintf(" AND e.date <= $%d", argIndex)
		args = append(args, endDate)
		argIndex++
	}

	if categoryIDStr != "" {
		whereClause += fmt.Sprintf(" AND e.category_id = $%d", argIndex)
		args = append(args, categoryIDStr)
		argIndex++
	}

	// Get total expenses and count
	totalQuery := fmt.Sprintf(`
		SELECT 
			COALESCE(SUM(amount), 0) as total_expenses,
			COUNT(*) as expense_count,
			COALESCE(SUM(CASE WHEN tax_deductible THEN amount ELSE 0 END), 0) as tax_deductible_sum
		FROM expenses e
		%s
	`, whereClause)

	err := r.db.QueryRow(totalQuery, args...).Scan(
		&analytics.TotalExpenses,
		&analytics.ExpenseCount,
		&analytics.TaxDeductibleSum,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get total analytics: %w", err)
	}

	// Get category breakdown
	categoryQuery := fmt.Sprintf(`
		SELECT 
			COALESCE(e.category_id, 0) as category_id,
			COALESCE(c.name, 'Uncategorized') as category_name,
			SUM(e.amount) as total_amount,
			COUNT(e.*) as count
		FROM expenses e
		LEFT JOIN categories c ON e.category_id = c.id
		%s
		GROUP BY e.category_id, c.name
		ORDER BY total_amount DESC
	`, whereClause)

	rows, err := r.db.Query(categoryQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get category breakdown: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var breakdown models.CategoryExpenseBreakdown
		err := rows.Scan(
			&breakdown.CategoryID,
			&breakdown.CategoryName,
			&breakdown.TotalAmount,
			&breakdown.Count,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan category breakdown: %w", err)
		}

		// Calculate percentage
		if analytics.TotalExpenses > 0 {
			breakdown.Percentage = (breakdown.TotalAmount / analytics.TotalExpenses) * 100
		}

		analytics.CategoryBreakdown = append(analytics.CategoryBreakdown, breakdown)
	}

	// Get monthly trends
	monthlyQuery := fmt.Sprintf(`
		SELECT 
			DATE_TRUNC('month', e.date) as month,
			SUM(e.amount) as amount,
			COUNT(e.*) as count
		FROM expenses e
		%s
		GROUP BY DATE_TRUNC('month', e.date)
		ORDER BY month DESC
		LIMIT 12
	`, whereClause)

	rows, err = r.db.Query(monthlyQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get monthly trends: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var trend models.MonthlyTrend
		err := rows.Scan(
			&trend.Month,
			&trend.Amount,
			&trend.Count,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan monthly trend: %w", err)
		}

		analytics.MonthlyTrends = append(analytics.MonthlyTrends, trend)
	}

	// Get account breakdown
	accountQuery := fmt.Sprintf(`
		SELECT 
			COALESCE(e.account_id, 0) as account_id,
			COALESCE(a.name, 'No Account') as account_name,
			SUM(e.amount) as total_amount,
			COUNT(e.*) as count
		FROM expenses e
		LEFT JOIN accounts a ON e.account_id = a.id
		%s
		GROUP BY e.account_id, a.name
		ORDER BY total_amount DESC
	`, whereClause)

	rows, err = r.db.Query(accountQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get account breakdown: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var breakdown models.AccountExpenseBreakdown
		err := rows.Scan(
			&breakdown.AccountID,
			&breakdown.AccountName,
			&breakdown.TotalAmount,
			&breakdown.Count,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan account breakdown: %w", err)
		}

		analytics.AccountBreakdown = append(analytics.AccountBreakdown, breakdown)
	}

	// Set net amount (for now, same as total expenses since we don't have income tracking)
	analytics.NetAmount = -analytics.TotalExpenses // Negative because expenses reduce net worth
	analytics.TotalIncome = 0 // Placeholder for future income tracking

	return analytics, nil
}