package repository

import (
	"database/sql"
	"fmt"
	"moneymind-backend/internal/models"
	"time"
)

type GoalRepository struct {
	db *sql.DB
}

func NewGoalRepository(db *sql.DB) *GoalRepository {
	return &GoalRepository{db: db}
}

func (r *GoalRepository) CreateGoal(userID int, goal *models.GoalCreate) (*models.Goal, error) {
	query := `
		INSERT INTO goals (user_id, name, description, target_amount, current_amount, target_date, goal_type, is_achieved, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, user_id, name, description, target_amount, current_amount, target_date, goal_type, is_achieved, created_at, updated_at
	`

	now := time.Now()
	var newGoal models.Goal
	err := r.db.QueryRow(
		query,
		userID,
		goal.Name,
		goal.Description,
		goal.TargetAmount,
		goal.CurrentAmount,
		goal.TargetDate,
		goal.GoalType,
		false, // new goals are not achieved
		now,
		now,
	).Scan(
		&newGoal.ID,
		&newGoal.UserID,
		&newGoal.Name,
		&newGoal.Description,
		&newGoal.TargetAmount,
		&newGoal.CurrentAmount,
		&newGoal.TargetDate,
		&newGoal.GoalType,
		&newGoal.IsAchieved,
		&newGoal.CreatedAt,
		&newGoal.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create goal: %w", err)
	}

	return &newGoal, nil
}

func (r *GoalRepository) GetGoalsByUserID(userID int) ([]models.Goal, error) {
	query := `
		SELECT id, user_id, name, description, target_amount, current_amount, target_date, goal_type, is_achieved, created_at, updated_at
		FROM goals
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query goals: %w", err)
	}
	defer rows.Close()

	var goals []models.Goal
	for rows.Next() {
		var goal models.Goal
		err := rows.Scan(
			&goal.ID,
			&goal.UserID,
			&goal.Name,
			&goal.Description,
			&goal.TargetAmount,
			&goal.CurrentAmount,
			&goal.TargetDate,
			&goal.GoalType,
			&goal.IsAchieved,
			&goal.CreatedAt,
			&goal.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan goal: %w", err)
		}
		goals = append(goals, goal)
	}

	return goals, nil
}

func (r *GoalRepository) GetGoalByID(id, userID int) (*models.Goal, error) {
	query := `
		SELECT id, user_id, name, description, target_amount, current_amount, target_date, goal_type, is_achieved, created_at, updated_at
		FROM goals
		WHERE id = $1 AND user_id = $2
	`

	var goal models.Goal
	err := r.db.QueryRow(query, id, userID).Scan(
		&goal.ID,
		&goal.UserID,
		&goal.Name,
		&goal.Description,
		&goal.TargetAmount,
		&goal.CurrentAmount,
		&goal.TargetDate,
		&goal.GoalType,
		&goal.IsAchieved,
		&goal.CreatedAt,
		&goal.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get goal: %w", err)
	}

	return &goal, nil
}

func (r *GoalRepository) UpdateGoal(id, userID int, updates *models.GoalUpdate) (*models.Goal, error) {
	// Build dynamic query based on which fields are being updated
	setParts := []string{}
	args := []interface{}{}
	argIndex := 1

	if updates.Name != "" {
		setParts = append(setParts, fmt.Sprintf("name = $%d", argIndex))
		args = append(args, updates.Name)
		argIndex++
	}
	if updates.Description != "" {
		setParts = append(setParts, fmt.Sprintf("description = $%d", argIndex))
		args = append(args, updates.Description)
		argIndex++
	}
	if updates.TargetAmount != nil {
		setParts = append(setParts, fmt.Sprintf("target_amount = $%d", argIndex))
		args = append(args, *updates.TargetAmount)
		argIndex++
	}
	if updates.CurrentAmount != nil {
		setParts = append(setParts, fmt.Sprintf("current_amount = $%d", argIndex))
		args = append(args, *updates.CurrentAmount)
		argIndex++
	}
	if updates.TargetDate != nil {
		setParts = append(setParts, fmt.Sprintf("target_date = $%d", argIndex))
		args = append(args, updates.TargetDate)
		argIndex++
	}
	if updates.IsAchieved != nil {
		setParts = append(setParts, fmt.Sprintf("is_achieved = $%d", argIndex))
		args = append(args, *updates.IsAchieved)
		argIndex++
	}
	if updates.GoalType != "" {
		setParts = append(setParts, fmt.Sprintf("goal_type = $%d", argIndex))
		args = append(args, updates.GoalType)
		argIndex++
	}

	if len(setParts) == 0 {
		return r.GetGoalByID(id, userID) // No updates to apply
	}

	// Add updated_at
	setParts = append(setParts, fmt.Sprintf("updated_at = $%d", argIndex))
	args = append(args, time.Now())
	argIndex++

	// Add WHERE clause parameters
	args = append(args, id, userID)

	// Build the query properly
	setClause := ""
	for i, part := range setParts {
		if i > 0 {
			setClause += ", "
		}
		setClause += part
	}

	query := fmt.Sprintf(`
		UPDATE goals 
		SET %s
		WHERE id = $%d AND user_id = $%d
	`, setClause, argIndex, argIndex+1)

	_, err := r.db.Exec(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to update goal: %w", err)
	}

	return r.GetGoalByID(id, userID)
}

func (r *GoalRepository) DeleteGoal(id, userID int) error {
	query := `DELETE FROM goals WHERE id = $1 AND user_id = $2`
	result, err := r.db.Exec(query, id, userID)
	if err != nil {
		return fmt.Errorf("failed to delete goal: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("goal not found")
	}

	return nil
}

func (r *GoalRepository) UpdateGoalProgress(id, userID int, progress *models.GoalProgress) (*models.Goal, error) {
	// First get the current goal to add to the current amount
	goal, err := r.GetGoalByID(id, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get goal: %w", err)
	}

	newCurrentAmount := goal.CurrentAmount + progress.Amount
	isAchieved := newCurrentAmount >= goal.TargetAmount

	query := `
		UPDATE goals 
		SET current_amount = $1, is_achieved = $2, updated_at = $3
		WHERE id = $4 AND user_id = $5
	`

	_, err = r.db.Exec(query, newCurrentAmount, isAchieved, time.Now(), id, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to update goal progress: %w", err)
	}

	return r.GetGoalByID(id, userID)
}

func (r *GoalRepository) GetGoalsWithProgress(userID int) ([]models.GoalWithProgress, error) {
	query := `
		SELECT 
			id, user_id, name, description, target_amount, current_amount, target_date, goal_type, is_achieved, created_at, updated_at,
			CASE WHEN target_amount > 0 THEN (current_amount / target_amount * 100) ELSE 0 END as progress_percentage,
			(target_amount - current_amount) as remaining_amount,
			CASE 
				WHEN target_date IS NOT NULL THEN EXTRACT(DAY FROM target_date - CURRENT_DATE)::int
				ELSE NULL 
			END as days_remaining,
			CASE 
				WHEN target_date IS NOT NULL AND target_date > CURRENT_DATE THEN 
					(target_amount - current_amount) / GREATEST(EXTRACT(DAY FROM target_date - CURRENT_DATE), 1)
				ELSE NULL 
			END as daily_target_amount,
			CASE 
				WHEN target_date IS NULL THEN true
				WHEN target_date <= CURRENT_DATE THEN is_achieved
				ELSE (current_amount / target_amount) >= (EXTRACT(DAY FROM CURRENT_DATE - created_at) / EXTRACT(DAY FROM target_date - created_at))
			END as on_track
		FROM goals
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query goals with progress: %w", err)
	}
	defer rows.Close()

	var goals []models.GoalWithProgress
	for rows.Next() {
		var goal models.GoalWithProgress
		var daysRemaining sql.NullInt64
		var dailyTargetAmount sql.NullFloat64

		err := rows.Scan(
			&goal.ID,
			&goal.UserID,
			&goal.Name,
			&goal.Description,
			&goal.TargetAmount,
			&goal.CurrentAmount,
			&goal.TargetDate,
			&goal.GoalType,
			&goal.IsAchieved,
			&goal.CreatedAt,
			&goal.UpdatedAt,
			&goal.ProgressPercentage,
			&goal.RemainingAmount,
			&daysRemaining,
			&dailyTargetAmount,
			&goal.OnTrack,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan goal with progress: %w", err)
		}

		if daysRemaining.Valid {
			days := int(daysRemaining.Int64)
			goal.DaysRemaining = &days
		}
		if dailyTargetAmount.Valid {
			amount := dailyTargetAmount.Float64
			goal.DailyTargetAmount = &amount
		}

		goals = append(goals, goal)
	}

	return goals, nil
}

func (r *GoalRepository) GetGoalSummary(userID int) (*models.GoalSummary, error) {
	query := `
		SELECT 
			COUNT(*) as total_goals,
			COUNT(CASE WHEN is_achieved THEN 1 END) as achieved_goals,
			COUNT(CASE WHEN NOT is_achieved THEN 1 END) as active_goals,
			COALESCE(SUM(target_amount), 0) as total_target_amount,
			COALESCE(SUM(current_amount), 0) as total_current_amount,
			CASE 
				WHEN SUM(target_amount) > 0 THEN (SUM(current_amount) / SUM(target_amount) * 100)
				ELSE 0 
			END as overall_progress
		FROM goals
		WHERE user_id = $1
	`

	var summary models.GoalSummary
	err := r.db.QueryRow(query, userID).Scan(
		&summary.TotalGoals,
		&summary.AchievedGoals,
		&summary.ActiveGoals,
		&summary.TotalTargetAmount,
		&summary.TotalCurrentAmount,
		&summary.OverallProgress,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get goal summary: %w", err)
	}

	return &summary, nil
}