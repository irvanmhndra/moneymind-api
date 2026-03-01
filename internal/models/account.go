package models

import "time"

type Account struct {
	ID        int       `json:"id" db:"id"`
	UserID    int       `json:"user_id" db:"user_id"`
	Name      string    `json:"name" db:"name" validate:"required,min=1,max=100"`
	Type      string    `json:"type" db:"type" validate:"required,oneof=checking savings credit_card cash investment other"`
	Balance   float64   `json:"balance" db:"balance"`
	Currency  string    `json:"currency" db:"currency" validate:"len=3"`
	IsActive  bool      `json:"is_active" db:"is_active"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type AccountCreate struct {
	Name     string  `json:"name" validate:"required,min=1,max=100"`
	Type     string  `json:"type" validate:"required,oneof=checking savings credit_card cash investment other"`
	Balance  float64 `json:"balance" validate:"omitempty"`
	Currency string  `json:"currency" validate:"omitempty,len=3"`
}

type AccountUpdate struct {
	Name     string   `json:"name" validate:"omitempty,min=1,max=100"`
	Type     string   `json:"type" validate:"omitempty,oneof=checking savings credit_card cash investment other"`
	Balance  *float64 `json:"balance"`
	Currency string   `json:"currency" validate:"omitempty,len=3"`
	IsActive *bool    `json:"is_active"`
}

type AccountWithStats struct {
	Account
	ExpenseCount    int     `json:"expense_count" db:"expense_count"`
	TotalExpenses   float64 `json:"total_expenses" db:"total_expenses"`
	LastExpenseAt   *time.Time `json:"last_expense_at" db:"last_expense_at"`
	MonthlyAverage  float64 `json:"monthly_average" db:"monthly_average"`
}