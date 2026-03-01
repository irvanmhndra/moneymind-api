package repository

import (
	"database/sql"
	"fmt"
	"moneymind-backend/internal/models"
	"time"
)

type BudgetRepository struct {
	db *sql.DB
}

func NewBudgetRepository(db *sql.DB) *BudgetRepository {
	return &BudgetRepository{db: db}
}

func (r *BudgetRepository) CreateBudget(userID int, budget *models.BudgetCreate) (*models.Budget, error) {
	query := `
		INSERT INTO budgets (user_id, name, category_id, amount, period, start_date, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, user_id, name, category_id, amount, period, start_date, end_date, is_active, created_at, updated_at
	`

	now := time.Now()
	var newBudget models.Budget
	err := r.db.QueryRow(
		query,
		userID,
		budget.Name,
		budget.CategoryID,
		budget.Amount,
		budget.Period,
		budget.StartDate,
		true, // new budgets are active by default
		now,
		now,
	).Scan(
		&newBudget.ID,
		&newBudget.UserID,
		&newBudget.Name,
		&newBudget.CategoryID,
		&newBudget.Amount,
		&newBudget.Period,
		&newBudget.StartDate,
		&newBudget.EndDate,
		&newBudget.IsActive,
		&newBudget.CreatedAt,
		&newBudget.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create budget: %w", err)
	}

	return &newBudget, nil
}

func (r *BudgetRepository) GetBudgetsByUserID(userID int) ([]models.Budget, error) {
	query := `
		SELECT b.id, b.user_id, b.name, b.category_id, b.amount, b.period, b.start_date, b.end_date, b.is_active, b.created_at, b.updated_at,
		       c.name as category_name, c.color as category_color, c.icon as category_icon
		FROM budgets b
		LEFT JOIN categories c ON b.category_id = c.id
		WHERE b.user_id = $1
		ORDER BY b.created_at DESC
	`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query budgets: %w", err)
	}
	defer rows.Close()

	var budgets []models.Budget
	for rows.Next() {
		var budget models.Budget
		var category models.Category
		var categoryName, categoryColor, categoryIcon sql.NullString

		err := rows.Scan(
			&budget.ID,
			&budget.UserID,
			&budget.Name,
			&budget.CategoryID,
			&budget.Amount,
			&budget.Period,
			&budget.StartDate,
			&budget.EndDate,
			&budget.IsActive,
			&budget.CreatedAt,
			&budget.UpdatedAt,
			&categoryName,
			&categoryColor,
			&categoryIcon,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan budget: %w", err)
		}

		if categoryName.Valid {
			category.Name = categoryName.String
			category.Color = categoryColor.String
			category.Icon = categoryIcon.String
			budget.Category = &category
		}

		budgets = append(budgets, budget)
	}

	return budgets, nil
}

func (r *BudgetRepository) GetBudgetByID(id, userID int) (*models.Budget, error) {
	query := `
		SELECT b.id, b.user_id, b.name, b.category_id, b.amount, b.period, b.start_date, b.end_date, b.is_active, b.created_at, b.updated_at,
		       c.name as category_name, c.color as category_color, c.icon as category_icon
		FROM budgets b
		LEFT JOIN categories c ON b.category_id = c.id
		WHERE b.id = $1 AND b.user_id = $2
	`

	var budget models.Budget
	var category models.Category
	var categoryName, categoryColor, categoryIcon sql.NullString

	err := r.db.QueryRow(query, id, userID).Scan(
		&budget.ID,
		&budget.UserID,
		&budget.Name,
		&budget.CategoryID,
		&budget.Amount,
		&budget.Period,
		&budget.StartDate,
		&budget.EndDate,
		&budget.IsActive,
		&budget.CreatedAt,
		&budget.UpdatedAt,
		&categoryName,
		&categoryColor,
		&categoryIcon,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get budget: %w", err)
	}

	if categoryName.Valid {
		category.Name = categoryName.String
		category.Color = categoryColor.String
		category.Icon = categoryIcon.String
		budget.Category = &category
	}

	return &budget, nil
}

func (r *BudgetRepository) UpdateBudget(id, userID int, updates *models.BudgetUpdate) (*models.Budget, error) {
	// Build dynamic query based on which fields are being updated
	setParts := []string{}
	args := []interface{}{}
	argIndex := 1

	if updates.Name != "" {
		setParts = append(setParts, fmt.Sprintf("name = $%d", argIndex))
		args = append(args, updates.Name)
		argIndex++
	}
	if updates.CategoryID != nil {
		setParts = append(setParts, fmt.Sprintf("category_id = $%d", argIndex))
		args = append(args, updates.CategoryID)
		argIndex++
	}
	if updates.Amount != nil {
		setParts = append(setParts, fmt.Sprintf("amount = $%d", argIndex))
		args = append(args, *updates.Amount)
		argIndex++
	}
	if updates.Period != "" {
		setParts = append(setParts, fmt.Sprintf("period = $%d", argIndex))
		args = append(args, updates.Period)
		argIndex++
	}
	if updates.StartDate != nil {
		setParts = append(setParts, fmt.Sprintf("start_date = $%d", argIndex))
		args = append(args, *updates.StartDate)
		argIndex++
	}
	if updates.IsActive != nil {
		setParts = append(setParts, fmt.Sprintf("is_active = $%d", argIndex))
		args = append(args, *updates.IsActive)
		argIndex++
	}

	if len(setParts) == 0 {
		return r.GetBudgetByID(id, userID) // No updates to apply
	}

	// Add updated_at
	setParts = append(setParts, fmt.Sprintf("updated_at = $%d", argIndex))
	args = append(args, time.Now())
	argIndex++

	// Add WHERE clause parameters
	args = append(args, id, userID)

	query := fmt.Sprintf(`
		UPDATE budgets 
		SET %s
		WHERE id = $%d AND user_id = $%d
	`, fmt.Sprintf("%s", setParts[0]), argIndex, argIndex+1)

	for i := 1; i < len(setParts); i++ {
		query = fmt.Sprintf(`
			UPDATE budgets 
			SET %s
			WHERE id = $%d AND user_id = $%d
		`, fmt.Sprintf("%s, %s", setParts[0], setParts[i]), argIndex, argIndex+1)
	}

	// Rebuild the query properly
	setClause := ""
	for i, part := range setParts {
		if i > 0 {
			setClause += ", "
		}
		setClause += part
	}

	query = fmt.Sprintf(`
		UPDATE budgets 
		SET %s
		WHERE id = $%d AND user_id = $%d
	`, setClause, argIndex, argIndex+1)

	_, err := r.db.Exec(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to update budget: %w", err)
	}

	return r.GetBudgetByID(id, userID)
}

func (r *BudgetRepository) DeleteBudget(id, userID int) error {
	query := `DELETE FROM budgets WHERE id = $1 AND user_id = $2`
	result, err := r.db.Exec(query, id, userID)
	if err != nil {
		return fmt.Errorf("failed to delete budget: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("budget not found")
	}

	return nil
}

func (r *BudgetRepository) GetBudgetStatus(userID int) ([]models.BudgetStatus, error) {
	query := `
		WITH budget_expenses AS (
			SELECT 
				b.id as budget_id,
				COALESCE(SUM(e.amount), 0) as spent_amount,
				COUNT(e.id) as expense_count
			FROM budgets b
			LEFT JOIN expenses e ON (
				(b.category_id IS NULL OR e.category_id = b.category_id) AND
				e.user_id = b.user_id AND
				e.date >= b.start_date AND
				(b.end_date IS NULL OR e.date <= b.end_date) AND
				e.date >= CASE 
					WHEN b.period = 'weekly' THEN DATE_TRUNC('week', CURRENT_DATE)
					WHEN b.period = 'monthly' THEN DATE_TRUNC('month', CURRENT_DATE)
					WHEN b.period = 'quarterly' THEN DATE_TRUNC('quarter', CURRENT_DATE)
					WHEN b.period = 'yearly' THEN DATE_TRUNC('year', CURRENT_DATE)
				END
			)
			WHERE b.user_id = $1 AND b.is_active = true
			GROUP BY b.id
		)
		SELECT 
			b.id, b.user_id, b.name, b.category_id, b.amount, b.period, b.start_date, b.end_date, b.is_active, b.created_at, b.updated_at,
			c.name as category_name, c.color as category_color, c.icon as category_icon,
			be.spent_amount,
			(b.amount - be.spent_amount) as remaining_amount,
			CASE WHEN b.amount > 0 THEN (be.spent_amount / b.amount * 100) ELSE 0 END as percentage_used,
			EXTRACT(DAY FROM (
				CASE 
					WHEN b.period = 'weekly' THEN DATE_TRUNC('week', CURRENT_DATE) + INTERVAL '1 week'
					WHEN b.period = 'monthly' THEN DATE_TRUNC('month', CURRENT_DATE) + INTERVAL '1 month'
					WHEN b.period = 'quarterly' THEN DATE_TRUNC('quarter', CURRENT_DATE) + INTERVAL '3 months'
					WHEN b.period = 'yearly' THEN DATE_TRUNC('year', CURRENT_DATE) + INTERVAL '1 year'
				END - CURRENT_DATE
			)) as days_remaining,
			(be.spent_amount > b.amount) as is_over_budget,
			be.expense_count,
			CASE WHEN be.expense_count > 0 THEN be.spent_amount / GREATEST(EXTRACT(DAY FROM CURRENT_DATE - b.start_date), 1) ELSE 0 END as daily_average,
			CASE 
				WHEN EXTRACT(DAY FROM CURRENT_DATE - b.start_date) > 0 THEN 
					(be.spent_amount / GREATEST(EXTRACT(DAY FROM CURRENT_DATE - b.start_date), 1)) * 
					EXTRACT(DAY FROM (
						CASE 
							WHEN b.period = 'weekly' THEN DATE_TRUNC('week', CURRENT_DATE) + INTERVAL '1 week'
							WHEN b.period = 'monthly' THEN DATE_TRUNC('month', CURRENT_DATE) + INTERVAL '1 month'
							WHEN b.period = 'quarterly' THEN DATE_TRUNC('quarter', CURRENT_DATE) + INTERVAL '3 months'
							WHEN b.period = 'yearly' THEN DATE_TRUNC('year', CURRENT_DATE) + INTERVAL '1 year'
						END - b.start_date
					))
				ELSE 0 
			END as projected_amount
		FROM budgets b
		LEFT JOIN categories c ON b.category_id = c.id
		LEFT JOIN budget_expenses be ON b.id = be.budget_id
		WHERE b.user_id = $1 AND b.is_active = true
		ORDER BY b.created_at DESC
	`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query budget status: %w", err)
	}
	defer rows.Close()

	var budgetStatuses []models.BudgetStatus
	for rows.Next() {
		var status models.BudgetStatus
		var category models.Category
		var categoryName, categoryColor, categoryIcon sql.NullString

		err := rows.Scan(
			&status.ID,
			&status.UserID,
			&status.Name,
			&status.CategoryID,
			&status.Amount,
			&status.Period,
			&status.StartDate,
			&status.EndDate,
			&status.IsActive,
			&status.CreatedAt,
			&status.UpdatedAt,
			&categoryName,
			&categoryColor,
			&categoryIcon,
			&status.SpentAmount,
			&status.RemainingAmount,
			&status.PercentageUsed,
			&status.DaysRemaining,
			&status.IsOverBudget,
			&status.ExpenseCount,
			&status.DailyAverage,
			&status.ProjectedAmount,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan budget status: %w", err)
		}

		if categoryName.Valid {
			category.Name = categoryName.String
			category.Color = categoryColor.String
			category.Icon = categoryIcon.String
			status.Category = &category
		}

		budgetStatuses = append(budgetStatuses, status)
	}

	return budgetStatuses, nil
}

func (r *BudgetRepository) GetBudgetSummary(userID int) (*models.BudgetSummary, error) {
	query := `
		WITH budget_stats AS (
			SELECT 
				COUNT(*) as total_budgets,
				COUNT(CASE WHEN is_active THEN 1 END) as active_budgets,
				COALESCE(SUM(CASE WHEN is_active THEN amount ELSE 0 END), 0) as total_budget_amount
			FROM budgets 
			WHERE user_id = $1
		),
		spending_stats AS (
			SELECT 
				COALESCE(SUM(be.spent_amount), 0) as total_spent_amount,
				COUNT(CASE WHEN be.spent_amount > b.amount THEN 1 END) as over_budget_count,
				COUNT(CASE WHEN be.spent_amount > (b.amount * 0.8) AND be.spent_amount <= b.amount THEN 1 END) as warning_count
			FROM budgets b
			LEFT JOIN (
				SELECT 
					b.id as budget_id,
					COALESCE(SUM(e.amount), 0) as spent_amount
				FROM budgets b
				LEFT JOIN expenses e ON (
					(b.category_id IS NULL OR e.category_id = b.category_id) AND
					e.user_id = b.user_id AND
					e.date >= b.start_date AND
					(b.end_date IS NULL OR e.date <= b.end_date) AND
					e.date >= CASE 
						WHEN b.period = 'weekly' THEN DATE_TRUNC('week', CURRENT_DATE)
						WHEN b.period = 'monthly' THEN DATE_TRUNC('month', CURRENT_DATE)
						WHEN b.period = 'quarterly' THEN DATE_TRUNC('quarter', CURRENT_DATE)
						WHEN b.period = 'yearly' THEN DATE_TRUNC('year', CURRENT_DATE)
					END
				)
				WHERE b.user_id = $1 AND b.is_active = true
				GROUP BY b.id
			) be ON b.id = be.budget_id
			WHERE b.user_id = $1 AND b.is_active = true
		)
		SELECT 
			bs.total_budgets,
			bs.active_budgets,
			bs.total_budget_amount,
			ss.total_spent_amount,
			ss.over_budget_count,
			ss.warning_count
		FROM budget_stats bs, spending_stats ss
	`

	var summary models.BudgetSummary
	err := r.db.QueryRow(query, userID).Scan(
		&summary.TotalBudgets,
		&summary.ActiveBudgets,
		&summary.TotalBudgetAmount,
		&summary.TotalSpentAmount,
		&summary.OverBudgetCount,
		&summary.WarningCount,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get budget summary: %w", err)
	}

	return &summary, nil
}