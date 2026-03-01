package models

import "time"

type Category struct {
	ID        int       `json:"id" db:"id"`
	UserID    *int      `json:"user_id" db:"user_id"`
	Name      string    `json:"name" db:"name" validate:"required,min=1,max=100"`
	Color     string    `json:"color" db:"color" validate:"omitempty,hexcolor"`
	Icon      string    `json:"icon" db:"icon"`
	IsSystem  bool      `json:"is_system" db:"is_system"`
	ParentID  *int      `json:"parent_id" db:"parent_id"`
	IsActive  bool      `json:"is_active" db:"is_active"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`

	// Related data
	Parent       *Category   `json:"parent,omitempty"`
	Subcategories []Category `json:"subcategories,omitempty"`
}

type CategoryCreate struct {
	Name     string `json:"name" validate:"required,min=1,max=100"`
	Color    string `json:"color" validate:"omitempty,hexcolor"`
	Icon     string `json:"icon"`
	ParentID *int   `json:"parent_id"`
}

type CategoryUpdate struct {
	Name     string `json:"name" validate:"omitempty,min=1,max=100"`
	Color    string `json:"color" validate:"omitempty,hexcolor"`
	Icon     string `json:"icon"`
	IsActive *bool  `json:"is_active"`
}

type CategoryWithStats struct {
	Category
	ExpenseCount  int     `json:"expense_count" db:"expense_count"`
	TotalAmount   float64 `json:"total_amount" db:"total_amount"`
	LastExpenseAt *time.Time `json:"last_expense_at" db:"last_expense_at"`
}