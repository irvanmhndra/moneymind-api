package models

import "time"

type Goal struct {
	ID            int       `json:"id" db:"id"`
	UserID        int       `json:"user_id" db:"user_id"`
	Name          string    `json:"name" db:"name" validate:"required,min=1,max=100"`
	Description   string    `json:"description" db:"description"`
	TargetAmount  float64   `json:"target_amount" db:"target_amount" validate:"required,gt=0"`
	CurrentAmount float64   `json:"current_amount" db:"current_amount"`
	TargetDate    *time.Time `json:"target_date" db:"target_date"`
	IsAchieved    bool      `json:"is_achieved" db:"is_achieved"`
	GoalType      string    `json:"goal_type" db:"goal_type" validate:"required,oneof=savings debt_payoff expense_reduction other"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
}

type GoalCreate struct {
	Name          string     `json:"name" validate:"required,min=1,max=100"`
	Description   string     `json:"description"`
	TargetAmount  float64    `json:"target_amount" validate:"required,gt=0"`
	CurrentAmount float64    `json:"current_amount" validate:"omitempty,gte=0"`
	TargetDate    *time.Time `json:"target_date"`
	GoalType      string     `json:"goal_type" validate:"required,oneof=savings debt_payoff expense_reduction other"`
}

type GoalUpdate struct {
	Name          string     `json:"name" validate:"omitempty,min=1,max=100"`
	Description   string     `json:"description"`
	TargetAmount  *float64   `json:"target_amount" validate:"omitempty,gt=0"`
	CurrentAmount *float64   `json:"current_amount" validate:"omitempty,gte=0"`
	TargetDate    *time.Time `json:"target_date"`
	IsAchieved    *bool      `json:"is_achieved"`
	GoalType      string     `json:"goal_type" validate:"omitempty,oneof=savings debt_payoff expense_reduction other"`
}

type GoalProgress struct {
	Amount      float64 `json:"amount" validate:"required,gt=0"`
	Description string  `json:"description"`
}

type GoalWithProgress struct {
	Goal
	ProgressPercentage float64 `json:"progress_percentage"`
	RemainingAmount    float64 `json:"remaining_amount"`
	DaysRemaining      *int    `json:"days_remaining"`
	DailyTargetAmount  *float64 `json:"daily_target_amount"`
	OnTrack            bool    `json:"on_track"`
}

type GoalSummary struct {
	TotalGoals       int     `json:"total_goals"`
	AchievedGoals    int     `json:"achieved_goals"`
	ActiveGoals      int     `json:"active_goals"`
	TotalTargetAmount float64 `json:"total_target_amount"`
	TotalCurrentAmount float64 `json:"total_current_amount"`
	OverallProgress   float64 `json:"overall_progress"`
}