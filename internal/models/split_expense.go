package models

import "time"

type SplitExpense struct {
	ID          int       `json:"id" db:"id"`
	ExpenseID   int       `json:"expense_id" db:"expense_id"`
	PayerUserID int       `json:"payer_user_id" db:"payer_user_id"`
	TotalAmount float64   `json:"total_amount" db:"total_amount"`
	SplitType   string    `json:"split_type" db:"split_type"`
	Description string    `json:"description" db:"description"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`

	// Related data
	Expense      *Expense             `json:"expense,omitempty"`
	PayerUser    *User                `json:"payer_user,omitempty"`
	Participants []SplitParticipant   `json:"participants,omitempty"`
}

type SplitParticipant struct {
	ID             int       `json:"id" db:"id"`
	SplitExpenseID int       `json:"split_expense_id" db:"split_expense_id"`
	UserID         int       `json:"user_id" db:"user_id"`
	Email          string    `json:"email" db:"email"`
	Name           string    `json:"name" db:"name"`
	AmountOwed     float64   `json:"amount_owed" db:"amount_owed"`
	AmountPaid     float64   `json:"amount_paid" db:"amount_paid"`
	IsSettled      bool      `json:"is_settled" db:"is_settled"`
	SettledAt      *time.Time `json:"settled_at" db:"settled_at"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`

	// Related data
	User *User `json:"user,omitempty"`
}

type SplitExpenseCreate struct {
	ExpenseID   int                        `json:"expense_id" validate:"required"`
	TotalAmount float64                    `json:"total_amount" validate:"required,gt=0"`
	SplitType   string                     `json:"split_type" validate:"required,oneof=equal percentage amount"`
	Description string                     `json:"description" validate:"max=500"`
	Participants []SplitParticipantCreate  `json:"participants" validate:"required,min=2,dive"`
}

type SplitParticipantCreate struct {
	UserID         *int    `json:"user_id"`
	Email          string  `json:"email" validate:"required,email"`
	Name           string  `json:"name" validate:"required,max=100"`
	AmountOwed     *float64 `json:"amount_owed"`
	Percentage     *float64 `json:"percentage"`
}

type SplitExpenseUpdate struct {
	Description string                     `json:"description" validate:"omitempty,max=500"`
	Participants []SplitParticipantUpdate  `json:"participants,omitempty"`
}

type SplitParticipantUpdate struct {
	ID         int      `json:"id" validate:"required"`
	AmountOwed *float64 `json:"amount_owed"`
	AmountPaid *float64 `json:"amount_paid"`
	IsSettled  *bool    `json:"is_settled"`
}

type SplitExpenseWithBalance struct {
	SplitExpense
	TotalOwed    float64 `json:"total_owed"`
	TotalPaid    float64 `json:"total_paid"`
	Balance      float64 `json:"balance"`
	IsFullyPaid  bool    `json:"is_fully_paid"`
}

type SplitSummary struct {
	UserID          int     `json:"user_id"`
	UserName        string  `json:"user_name"`
	UserEmail       string  `json:"user_email"`
	TotalOwedByUser float64 `json:"total_owed_by_user"`
	TotalPaidByUser float64 `json:"total_paid_by_user"`
	NetBalance      float64 `json:"net_balance"`
}