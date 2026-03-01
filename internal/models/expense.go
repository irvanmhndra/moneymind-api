package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

type Expense struct {
	ID                  int       `json:"id" db:"id"`
	UserID              int       `json:"user_id" db:"user_id"`
	AccountID           *int      `json:"account_id" db:"account_id"`
	CategoryID          *int      `json:"category_id" db:"category_id"`
	Amount              float64   `json:"amount" db:"amount" validate:"required,gt=0"`
	Currency            string    `json:"currency" db:"currency" validate:"len=3"`
	Description         string    `json:"description" db:"description" validate:"required,max=500"`
	Notes               string    `json:"notes" db:"notes"`
	Location            string    `json:"location" db:"location"`
	ReceiptPath         string    `json:"receipt_path" db:"receipt_path"`
	Date                time.Time `json:"date" db:"date" validate:"required"`
	IsRecurring         bool      `json:"is_recurring" db:"is_recurring"`
	RecurringFrequency  string    `json:"recurring_frequency" db:"recurring_frequency"`
	TaxDeductible       bool      `json:"tax_deductible" db:"tax_deductible"`
	TaxCategory         string    `json:"tax_category" db:"tax_category"`
	CreatedAt           time.Time `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time `json:"updated_at" db:"updated_at"`

	// Related data (loaded via joins)
	Account  *Account  `json:"account,omitempty"`
	Category *Category `json:"category,omitempty"`
}

type ExpenseCreate struct {
	AccountID          *int      `json:"account_id"`
	CategoryID         *int      `json:"category_id"`
	Amount             float64   `json:"amount" validate:"required,gt=0"`
	Currency           string    `json:"currency" validate:"omitempty,len=3"`
	Description        string    `json:"description" validate:"required,max=500"`
	Notes              string    `json:"notes"`
	Location           string    `json:"location"`
	Date               time.Time `json:"date" validate:"required"`
	IsRecurring        bool      `json:"is_recurring"`
	RecurringFrequency string    `json:"recurring_frequency" validate:"omitempty,oneof=daily weekly monthly yearly"`
	TaxDeductible      bool      `json:"tax_deductible"`
	TaxCategory        string    `json:"tax_category"`
}

type ExpenseUpdate struct {
	AccountID          *int      `json:"account_id"`
	CategoryID         *int      `json:"category_id"`
	Amount             *float64  `json:"amount" validate:"omitempty,gt=0"`
	Currency           string    `json:"currency" validate:"omitempty,len=3"`
	Description        string    `json:"description" validate:"omitempty,max=500"`
	Notes              string    `json:"notes"`
	Location           string    `json:"location"`
	Date               *time.Time `json:"date"`
	IsRecurring        *bool     `json:"is_recurring"`
	RecurringFrequency string    `json:"recurring_frequency" validate:"omitempty,oneof=daily weekly monthly yearly"`
	TaxDeductible      *bool     `json:"tax_deductible"`
	TaxCategory        string    `json:"tax_category"`
}

// FlexibleDate handles multiple date formats for form binding
type FlexibleDate struct {
	*time.Time
}

// UnmarshalJSON handles JSON unmarshaling with flexible date formats
func (fd *FlexibleDate) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	return fd.parseDate(s)
}

// UnmarshalText handles form/query parameter binding with flexible date formats
func (fd *FlexibleDate) UnmarshalText(data []byte) error {
	return fd.parseDate(string(data))
}

func (fd *FlexibleDate) parseDate(s string) error {
	if s == "" {
		fd.Time = nil
		return nil
	}

	// Try various date formats
	formats := []string{
		"2006-01-02",           // Simple date format
		"2006-01-02T15:04:05Z", // RFC3339 UTC
		"2006-01-02T15:04:05Z07:00", // RFC3339 with timezone
		"2006-01-02 15:04:05",  // SQL timestamp format
	}

	for _, format := range formats {
		if t, err := time.Parse(format, s); err == nil {
			fd.Time = &t
			return nil
		}
	}

	return fmt.Errorf("invalid date format: %s", s)
}

type ExpenseFilter struct {
	StartDate    *time.Time `form:"-"` // Handle manually
	EndDate      *time.Time `form:"-"` // Handle manually
	CategoryID   *int       `form:"category_id"`
	AccountID    *int       `form:"account_id"`
	MinAmount    *float64   `form:"min_amount"`
	MaxAmount    *float64   `form:"max_amount"`
	Search       string     `form:"search"`
	TaxDeductible *bool     `form:"tax_deductible"`
	Page         int        `form:"page"`
	Limit        int        `form:"limit"`
	SortBy       string     `form:"sort_by"`
	SortOrder    string     `form:"sort_order"`
}

type ExpenseAnalytics struct {
	TotalExpenses     float64                    `json:"total_expenses"`
	TotalIncome       float64                    `json:"total_income"`
	NetAmount         float64                    `json:"net_amount"`
	ExpenseCount      int                        `json:"expense_count"`
	CategoryBreakdown []CategoryExpenseBreakdown `json:"category_breakdown"`
	MonthlyTrends     []MonthlyTrend             `json:"monthly_trends"`
	AccountBreakdown  []AccountExpenseBreakdown  `json:"account_breakdown"`
	TaxDeductibleSum  float64                    `json:"tax_deductible_sum"`
}

type CategoryExpenseBreakdown struct {
	CategoryID   int     `json:"category_id"`
	CategoryName string  `json:"category_name"`
	TotalAmount  float64 `json:"total_amount"`
	Count        int     `json:"count"`
	Percentage   float64 `json:"percentage"`
}

type MonthlyTrend struct {
	Month  time.Time `json:"month"`
	Amount float64   `json:"amount"`
	Count  int       `json:"count"`
}

type AccountExpenseBreakdown struct {
	AccountID   int     `json:"account_id"`
	AccountName string  `json:"account_name"`
	TotalAmount float64 `json:"total_amount"`
	Count       int     `json:"count"`
}

// Custom type for handling NULL values in database
type NullTime struct {
	Time  time.Time
	Valid bool
}

func (nt *NullTime) Scan(value interface{}) error {
	if value == nil {
		nt.Time, nt.Valid = time.Time{}, false
		return nil
	}
	nt.Valid = true
	switch v := value.(type) {
	case time.Time:
		nt.Time = v
	case []byte:
		t, err := time.Parse("2006-01-02 15:04:05", string(v))
		if err != nil {
			return err
		}
		nt.Time = t
	}
	return nil
}

func (nt NullTime) Value() (driver.Value, error) {
	if !nt.Valid {
		return nil, nil
	}
	return nt.Time, nil
}