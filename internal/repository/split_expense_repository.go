package repository

import (
	"database/sql"
	"fmt"
	"moneymind-backend/internal/models"
	"strings"
	"time"
)

type SplitExpenseRepository struct {
	db *sql.DB
}

func NewSplitExpenseRepository(db *sql.DB) *SplitExpenseRepository {
	return &SplitExpenseRepository{db: db}
}

func (r *SplitExpenseRepository) CreateSplitExpense(payerUserID int, splitExpense *models.SplitExpenseCreate) (*models.SplitExpense, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Create split expense
	query := `
		INSERT INTO split_expenses (expense_id, payer_user_id, total_amount, split_type, description, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, expense_id, payer_user_id, total_amount, split_type, description, created_at, updated_at
	`

	now := time.Now()
	var newSplitExpense models.SplitExpense
	err = tx.QueryRow(
		query,
		splitExpense.ExpenseID,
		payerUserID,
		splitExpense.TotalAmount,
		splitExpense.SplitType,
		splitExpense.Description,
		now,
		now,
	).Scan(
		&newSplitExpense.ID,
		&newSplitExpense.ExpenseID,
		&newSplitExpense.PayerUserID,
		&newSplitExpense.TotalAmount,
		&newSplitExpense.SplitType,
		&newSplitExpense.Description,
		&newSplitExpense.CreatedAt,
		&newSplitExpense.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create split expense: %w", err)
	}

	// Create participants
	participantQuery := `
		INSERT INTO split_participants (split_expense_id, user_id, email, name, amount_owed, amount_paid, is_settled, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, split_expense_id, user_id, email, name, amount_owed, amount_paid, is_settled, settled_at, created_at, updated_at
	`

	var participants []models.SplitParticipant
	for _, participant := range splitExpense.Participants {
		var amountOwed float64
		switch splitExpense.SplitType {
		case "equal":
			amountOwed = splitExpense.TotalAmount / float64(len(splitExpense.Participants))
		case "percentage":
			if participant.Percentage != nil {
				amountOwed = splitExpense.TotalAmount * (*participant.Percentage / 100.0)
			}
		case "amount":
			if participant.AmountOwed != nil {
				amountOwed = *participant.AmountOwed
			}
		}

		var newParticipant models.SplitParticipant
		err = tx.QueryRow(
			participantQuery,
			newSplitExpense.ID,
			participant.UserID,
			participant.Email,
			participant.Name,
			amountOwed,
			0.0, // initial amount paid is 0
			false, // initially not settled
			now,
			now,
		).Scan(
			&newParticipant.ID,
			&newParticipant.SplitExpenseID,
			&newParticipant.UserID,
			&newParticipant.Email,
			&newParticipant.Name,
			&newParticipant.AmountOwed,
			&newParticipant.AmountPaid,
			&newParticipant.IsSettled,
			&newParticipant.SettledAt,
			&newParticipant.CreatedAt,
			&newParticipant.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create participant: %w", err)
		}
		participants = append(participants, newParticipant)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	newSplitExpense.Participants = participants
	return &newSplitExpense, nil
}

func (r *SplitExpenseRepository) GetSplitExpenseByID(id int, userID int) (*models.SplitExpenseWithBalance, error) {
	query := `
		SELECT se.id, se.expense_id, se.payer_user_id, se.total_amount, se.split_type, se.description, se.created_at, se.updated_at,
		       e.description as expense_description, e.amount as expense_amount, e.date as expense_date,
		       CONCAT(u.first_name, ' ', u.last_name) as payer_name, u.email as payer_email
		FROM split_expenses se
		LEFT JOIN expenses e ON se.expense_id = e.id
		LEFT JOIN users u ON se.payer_user_id = u.id
		WHERE se.id = $1 AND (se.payer_user_id = $2 OR EXISTS (
			SELECT 1 FROM split_participants sp WHERE sp.split_expense_id = se.id AND (sp.user_id = $2 OR sp.email = (SELECT email FROM users WHERE id = $2))
		))
	`

	var splitExpense models.SplitExpenseWithBalance
	var expense models.Expense
	var payerUser models.User

	err := r.db.QueryRow(query, id, userID).Scan(
		&splitExpense.ID,
		&splitExpense.ExpenseID,
		&splitExpense.PayerUserID,
		&splitExpense.TotalAmount,
		&splitExpense.SplitType,
		&splitExpense.Description,
		&splitExpense.CreatedAt,
		&splitExpense.UpdatedAt,
		&expense.Description,
		&expense.Amount,
		&expense.Date,
		&payerUser.FirstName, // This will store the full name from CONCAT
		&payerUser.Email,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get split expense: %w", err)
	}

	splitExpense.Expense = &expense
	splitExpense.PayerUser = &payerUser

	// Get participants
	participants, err := r.GetParticipantsBySplitExpenseID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get participants: %w", err)
	}
	splitExpense.Participants = participants

	// Calculate balances
	var totalOwed, totalPaid float64
	for _, p := range participants {
		totalOwed += p.AmountOwed
		totalPaid += p.AmountPaid
	}
	splitExpense.TotalOwed = totalOwed
	splitExpense.TotalPaid = totalPaid
	splitExpense.Balance = totalOwed - totalPaid
	splitExpense.IsFullyPaid = splitExpense.Balance <= 0.01 // account for floating point precision

	return &splitExpense, nil
}

func (r *SplitExpenseRepository) GetParticipantsBySplitExpenseID(splitExpenseID int) ([]models.SplitParticipant, error) {
	query := `
		SELECT sp.id, sp.split_expense_id, sp.user_id, sp.email, sp.name, sp.amount_owed, sp.amount_paid, sp.is_settled, sp.settled_at, sp.created_at, sp.updated_at,
		       CONCAT(u.first_name, ' ', u.last_name) as user_name, u.email as user_email
		FROM split_participants sp
		LEFT JOIN users u ON sp.user_id = u.id
		WHERE sp.split_expense_id = $1
		ORDER BY sp.created_at
	`

	rows, err := r.db.Query(query, splitExpenseID)
	if err != nil {
		return nil, fmt.Errorf("failed to query participants: %w", err)
	}
	defer rows.Close()

	var participants []models.SplitParticipant
	for rows.Next() {
		var participant models.SplitParticipant
		var user models.User
		var userName, userEmail sql.NullString

		err := rows.Scan(
			&participant.ID,
			&participant.SplitExpenseID,
			&participant.UserID,
			&participant.Email,
			&participant.Name,
			&participant.AmountOwed,
			&participant.AmountPaid,
			&participant.IsSettled,
			&participant.SettledAt,
			&participant.CreatedAt,
			&participant.UpdatedAt,
			&userName,
			&userEmail,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan participant: %w", err)
		}

		if userName.Valid && userEmail.Valid {
			user.FirstName = userName.String // Store the full name from CONCAT
			user.Email = userEmail.String
			participant.User = &user
		}

		participants = append(participants, participant)
	}

	return participants, nil
}

func (r *SplitExpenseRepository) GetSplitExpensesByUser(userID int, limit, offset int) ([]models.SplitExpenseWithBalance, error) {
	query := `
		SELECT DISTINCT se.id, se.expense_id, se.payer_user_id, se.total_amount, se.split_type, se.description, se.created_at, se.updated_at,
		       e.description as expense_description, e.amount as expense_amount, e.date as expense_date,
		       CONCAT(u.first_name, ' ', u.last_name) as payer_name, u.email as payer_email
		FROM split_expenses se
		LEFT JOIN expenses e ON se.expense_id = e.id
		LEFT JOIN users u ON se.payer_user_id = u.id
		LEFT JOIN split_participants sp ON se.id = sp.split_expense_id
		WHERE se.payer_user_id = $1 OR (sp.user_id = $1 OR sp.email = (SELECT email FROM users WHERE id = $1))
		ORDER BY se.created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query split expenses: %w", err)
	}
	defer rows.Close()

	var splitExpenses []models.SplitExpenseWithBalance
	for rows.Next() {
		var splitExpense models.SplitExpenseWithBalance
		var expense models.Expense
		var payerUser models.User

		err := rows.Scan(
			&splitExpense.ID,
			&splitExpense.ExpenseID,
			&splitExpense.PayerUserID,
			&splitExpense.TotalAmount,
			&splitExpense.SplitType,
			&splitExpense.Description,
			&splitExpense.CreatedAt,
			&splitExpense.UpdatedAt,
			&expense.Description,
			&expense.Amount,
			&expense.Date,
			&payerUser.FirstName, // This will store the full name from CONCAT
			&payerUser.Email,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan split expense: %w", err)
		}

		splitExpense.Expense = &expense
		splitExpense.PayerUser = &payerUser

		// Get participants for each split expense
		participants, err := r.GetParticipantsBySplitExpenseID(splitExpense.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get participants for split expense %d: %w", splitExpense.ID, err)
		}
		splitExpense.Participants = participants

		// Calculate balances
		var totalOwed, totalPaid float64
		for _, p := range participants {
			totalOwed += p.AmountOwed
			totalPaid += p.AmountPaid
		}
		splitExpense.TotalOwed = totalOwed
		splitExpense.TotalPaid = totalPaid
		splitExpense.Balance = totalOwed - totalPaid
		splitExpense.IsFullyPaid = splitExpense.Balance <= 0.01

		splitExpenses = append(splitExpenses, splitExpense)
	}

	return splitExpenses, nil
}

func (r *SplitExpenseRepository) UpdateParticipantPayment(participantID int, amountPaid float64, userID int) error {
	// First verify the user has permission to update this participant
	verifyQuery := `
		SELECT sp.amount_owed, sp.amount_paid, se.payer_user_id
		FROM split_participants sp
		JOIN split_expenses se ON sp.split_expense_id = se.id
		WHERE sp.id = $1 AND (se.payer_user_id = $2 OR sp.user_id = $2 OR sp.email = (SELECT email FROM users WHERE id = $2))
	`

	var currentAmountOwed, currentAmountPaid float64
	var payerUserID int
	err := r.db.QueryRow(verifyQuery, participantID, userID).Scan(&currentAmountOwed, &currentAmountPaid, &payerUserID)
	if err != nil {
		return fmt.Errorf("participant not found or access denied: %w", err)
	}

	newAmountPaid := currentAmountPaid + amountPaid
	isSettled := newAmountPaid >= currentAmountOwed

	var settledAt *time.Time
	if isSettled {
		now := time.Now()
		settledAt = &now
	}

	updateQuery := `
		UPDATE split_participants 
		SET amount_paid = $1, is_settled = $2, settled_at = $3, updated_at = $4
		WHERE id = $5
	`

	_, err = r.db.Exec(updateQuery, newAmountPaid, isSettled, settledAt, time.Now(), participantID)
	if err != nil {
		return fmt.Errorf("failed to update participant payment: %w", err)
	}

	return nil
}

func (r *SplitExpenseRepository) DeleteSplitExpense(id int, userID int) error {
	// Only the payer can delete a split expense
	query := `DELETE FROM split_expenses WHERE id = $1 AND payer_user_id = $2`
	result, err := r.db.Exec(query, id, userID)
	if err != nil {
		return fmt.Errorf("failed to delete split expense: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("split expense not found or access denied")
	}

	return nil
}

func (r *SplitExpenseRepository) GetSplitSummaryByUser(userID int) ([]models.SplitSummary, error) {
	query := `
		SELECT 
			CASE 
				WHEN sp.user_id IS NOT NULL THEN sp.user_id
				ELSE 0
			END as user_id,
			CASE 
				WHEN u.first_name IS NOT NULL THEN CONCAT(u.first_name, ' ', u.last_name)
				ELSE sp.name
			END as user_name,
			CASE 
				WHEN u.email IS NOT NULL THEN u.email
				ELSE sp.email
			END as user_email,
			COALESCE(SUM(sp.amount_owed), 0) as total_owed_by_user,
			COALESCE(SUM(sp.amount_paid), 0) as total_paid_by_user,
			COALESCE(SUM(sp.amount_owed - sp.amount_paid), 0) as net_balance
		FROM split_participants sp
		JOIN split_expenses se ON sp.split_expense_id = se.id
		LEFT JOIN users u ON sp.user_id = u.id
		WHERE se.payer_user_id = $1 OR sp.user_id = $1 OR sp.email = (SELECT email FROM users WHERE id = $1)
		GROUP BY sp.user_id, u.first_name, u.last_name, u.email, sp.name, sp.email
		HAVING COALESCE(SUM(sp.amount_owed - sp.amount_paid), 0) != 0
		ORDER BY net_balance DESC
	`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query split summary: %w", err)
	}
	defer rows.Close()

	var summaries []models.SplitSummary
	for rows.Next() {
		var summary models.SplitSummary
		err := rows.Scan(
			&summary.UserID,
			&summary.UserName,
			&summary.UserEmail,
			&summary.TotalOwedByUser,
			&summary.TotalPaidByUser,
			&summary.NetBalance,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan split summary: %w", err)
		}
		summaries = append(summaries, summary)
	}

	return summaries, nil
}

func (r *SplitExpenseRepository) UpdateSplitExpense(id int, userID int, updates *models.SplitExpenseUpdate) (*models.SplitExpense, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// First verify the user owns this split expense
	var currentPayerID int
	checkQuery := `SELECT payer_user_id FROM split_expenses WHERE id = $1`
	err = tx.QueryRow(checkQuery, id).Scan(&currentPayerID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("split expense not found")
		}
		return nil, fmt.Errorf("failed to check split expense ownership: %w", err)
	}

	if currentPayerID != userID {
		return nil, fmt.Errorf("only the payer can update the split expense")
	}

	// Build dynamic update query
	var setParts []string
	var args []interface{}
	argIndex := 1

	if updates.Description != "" {
		setParts = append(setParts, fmt.Sprintf("description = $%d", argIndex))
		args = append(args, updates.Description)
		argIndex++
	}

	if len(setParts) == 0 && len(updates.Participants) == 0 {
		return nil, fmt.Errorf("no fields to update")
	}

	var updatedSplitExpense models.SplitExpense

	// Update the split expense description if provided
	if len(setParts) > 0 {
		// Add updated_at timestamp
		setParts = append(setParts, fmt.Sprintf("updated_at = $%d", argIndex))
		args = append(args, time.Now())
		argIndex++

		// Add WHERE conditions
		args = append(args, id)

		updateQuery := fmt.Sprintf(`
			UPDATE split_expenses
			SET %s
			WHERE id = $%d
			RETURNING id, expense_id, payer_user_id, total_amount, split_type, description, created_at, updated_at`,
			strings.Join(setParts, ", "), argIndex)

		err = tx.QueryRow(updateQuery, args...).Scan(
			&updatedSplitExpense.ID,
			&updatedSplitExpense.ExpenseID,
			&updatedSplitExpense.PayerUserID,
			&updatedSplitExpense.TotalAmount,
			&updatedSplitExpense.SplitType,
			&updatedSplitExpense.Description,
			&updatedSplitExpense.CreatedAt,
			&updatedSplitExpense.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to update split expense: %w", err)
		}
	} else {
		// Just get the current split expense data
		selectQuery := `
			SELECT id, expense_id, payer_user_id, total_amount, split_type, description, created_at, updated_at
			FROM split_expenses WHERE id = $1`
		err = tx.QueryRow(selectQuery, id).Scan(
			&updatedSplitExpense.ID,
			&updatedSplitExpense.ExpenseID,
			&updatedSplitExpense.PayerUserID,
			&updatedSplitExpense.TotalAmount,
			&updatedSplitExpense.SplitType,
			&updatedSplitExpense.Description,
			&updatedSplitExpense.CreatedAt,
			&updatedSplitExpense.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to get split expense: %w", err)
		}
	}

	// Update participants if provided
	if len(updates.Participants) > 0 {
		for _, participant := range updates.Participants {
			var participantSetParts []string
			var participantArgs []interface{}
			participantArgIndex := 1

			if participant.AmountOwed != nil {
				participantSetParts = append(participantSetParts, fmt.Sprintf("amount_owed = $%d", participantArgIndex))
				participantArgs = append(participantArgs, *participant.AmountOwed)
				participantArgIndex++
			}

			if participant.AmountPaid != nil {
				participantSetParts = append(participantSetParts, fmt.Sprintf("amount_paid = $%d", participantArgIndex))
				participantArgs = append(participantArgs, *participant.AmountPaid)
				participantArgIndex++
			}

			if participant.IsSettled != nil {
				participantSetParts = append(participantSetParts, fmt.Sprintf("is_settled = $%d", participantArgIndex))
				participantArgs = append(participantArgs, *participant.IsSettled)
				participantArgIndex++
			}

			if len(participantSetParts) > 0 {
				participantSetParts = append(participantSetParts, fmt.Sprintf("updated_at = $%d", participantArgIndex))
				participantArgs = append(participantArgs, time.Now())
				participantArgIndex++

				participantArgs = append(participantArgs, participant.ID, id)

				participantUpdateQuery := fmt.Sprintf(`
					UPDATE split_participants
					SET %s
					WHERE id = $%d AND split_expense_id = $%d`,
					strings.Join(participantSetParts, ", "), participantArgIndex, participantArgIndex+1)

				_, err = tx.Exec(participantUpdateQuery, participantArgs...)
				if err != nil {
					return nil, fmt.Errorf("failed to update participant %d: %w", participant.ID, err)
				}
			}
		}
	}

	err = tx.Commit()
	if err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &updatedSplitExpense, nil
}