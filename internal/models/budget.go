package models

import "time"

type Budget struct {
	ID         int       `json:"id" db:"id"`
	UserID     int       `json:"user_id" db:"user_id"`
	Name       string    `json:"name" db:"name" validate:"required,min=1,max=100"`
	CategoryID *int      `json:"category_id" db:"category_id"`
	Amount     float64   `json:"amount" db:"amount" validate:"required,gt=0"`
	Period     string    `json:"period" db:"period" validate:"required,oneof=weekly monthly quarterly yearly"`
	StartDate  time.Time `json:"start_date" db:"start_date" validate:"required"`
	EndDate    *time.Time `json:"end_date" db:"end_date"`
	IsActive   bool      `json:"is_active" db:"is_active"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`

	// Related data
	Category *Category `json:"category,omitempty"`
}

type BudgetCreate struct {
	Name       string    `json:"name" validate:"required,min=1,max=100"`
	CategoryID *int      `json:"category_id"`
	Amount     float64   `json:"amount" validate:"required,gt=0"`
	Period     string    `json:"period" validate:"required,oneof=weekly monthly quarterly yearly"`
	StartDate  time.Time `json:"start_date" validate:"required"`
}

type BudgetUpdate struct {
	Name       string     `json:"name" validate:"omitempty,min=1,max=100"`
	CategoryID *int       `json:"category_id"`
	Amount     *float64   `json:"amount" validate:"omitempty,gt=0"`
	Period     string     `json:"period" validate:"omitempty,oneof=weekly monthly quarterly yearly"`
	StartDate  *time.Time `json:"start_date"`
	IsActive   *bool      `json:"is_active"`
}

type BudgetStatus struct {
	Budget
	SpentAmount      float64 `json:"spent_amount" db:"spent_amount"`
	RemainingAmount  float64 `json:"remaining_amount" db:"remaining_amount"`
	PercentageUsed   float64 `json:"percentage_used" db:"percentage_used"`
	DaysRemaining    int     `json:"days_remaining" db:"days_remaining"`
	IsOverBudget     bool    `json:"is_over_budget" db:"is_over_budget"`
	ExpenseCount     int     `json:"expense_count" db:"expense_count"`
	DailyAverage     float64 `json:"daily_average" db:"daily_average"`
	ProjectedAmount  float64 `json:"projected_amount" db:"projected_amount"`
}

type BudgetSummary struct {
	TotalBudgets      int     `json:"total_budgets"`
	ActiveBudgets     int     `json:"active_budgets"`
	TotalBudgetAmount float64 `json:"total_budget_amount"`
	TotalSpentAmount  float64 `json:"total_spent_amount"`
	OverBudgetCount   int     `json:"over_budget_count"`
	WarningCount      int     `json:"warning_count"` // >80% spent
}