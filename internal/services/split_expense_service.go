package services

import (
	"fmt"
	"moneymind-backend/internal/models"
	"moneymind-backend/internal/repository"
)

type SplitExpenseService struct {
	splitExpenseRepo *repository.SplitExpenseRepository
	expenseRepo      *repository.ExpenseRepository
}

func NewSplitExpenseService(splitExpenseRepo *repository.SplitExpenseRepository, expenseRepo *repository.ExpenseRepository) *SplitExpenseService {
	return &SplitExpenseService{
		splitExpenseRepo: splitExpenseRepo,
		expenseRepo:      expenseRepo,
	}
}

func (s *SplitExpenseService) CreateSplitExpense(userID int, splitExpense *models.SplitExpenseCreate) (*models.SplitExpense, error) {
	// Validate that the expense exists and belongs to the user
	expense, err := s.expenseRepo.GetExpenseByID(splitExpense.ExpenseID, userID)
	if err != nil {
		return nil, fmt.Errorf("expense not found or access denied: %w", err)
	}

	// Validate total amount matches expense amount
	if splitExpense.TotalAmount != expense.Amount {
		return nil, fmt.Errorf("split total amount (%.2f) must match expense amount (%.2f)", splitExpense.TotalAmount, expense.Amount)
	}

	// Validate split calculations
	err = s.validateSplitCalculations(splitExpense)
	if err != nil {
		return nil, err
	}

	return s.splitExpenseRepo.CreateSplitExpense(userID, splitExpense)
}

func (s *SplitExpenseService) validateSplitCalculations(splitExpense *models.SplitExpenseCreate) error {
	if len(splitExpense.Participants) < 2 {
		return fmt.Errorf("split expense must have at least 2 participants")
	}

	switch splitExpense.SplitType {
	case "equal":
		// Equal split - no additional validation needed
		return nil

	case "percentage":
		totalPercentage := 0.0
		for _, participant := range splitExpense.Participants {
			if participant.Percentage == nil {
				return fmt.Errorf("percentage is required for each participant in percentage split")
			}
			if *participant.Percentage <= 0 || *participant.Percentage > 100 {
				return fmt.Errorf("percentage must be between 0 and 100")
			}
			totalPercentage += *participant.Percentage
		}
		if totalPercentage < 99.99 || totalPercentage > 100.01 { // Allow for floating point precision
			return fmt.Errorf("total percentage must equal 100%%, got %.2f%%", totalPercentage)
		}
		return nil

	case "amount":
		totalAmount := 0.0
		for _, participant := range splitExpense.Participants {
			if participant.AmountOwed == nil {
				return fmt.Errorf("amount_owed is required for each participant in amount split")
			}
			if *participant.AmountOwed <= 0 {
				return fmt.Errorf("amount_owed must be greater than 0")
			}
			totalAmount += *participant.AmountOwed
		}
		if totalAmount < splitExpense.TotalAmount-0.01 || totalAmount > splitExpense.TotalAmount+0.01 { // Allow for floating point precision
			return fmt.Errorf("sum of participant amounts (%.2f) must equal total amount (%.2f)", totalAmount, splitExpense.TotalAmount)
		}
		return nil

	default:
		return fmt.Errorf("invalid split type: %s. Must be one of: equal, percentage, amount", splitExpense.SplitType)
	}
}

func (s *SplitExpenseService) GetSplitExpenseByID(id, userID int) (*models.SplitExpenseWithBalance, error) {
	return s.splitExpenseRepo.GetSplitExpenseByID(id, userID)
}

func (s *SplitExpenseService) GetSplitExpensesByUser(userID int, limit, offset int) ([]models.SplitExpenseWithBalance, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}
	
	return s.splitExpenseRepo.GetSplitExpensesByUser(userID, limit, offset)
}

func (s *SplitExpenseService) UpdateParticipantPayment(participantID int, amountPaid float64, userID int) error {
	if amountPaid <= 0 {
		return fmt.Errorf("payment amount must be greater than 0")
	}

	return s.splitExpenseRepo.UpdateParticipantPayment(participantID, amountPaid, userID)
}

func (s *SplitExpenseService) DeleteSplitExpense(id, userID int) error {
	return s.splitExpenseRepo.DeleteSplitExpense(id, userID)
}

func (s *SplitExpenseService) GetSplitSummaryByUser(userID int) ([]models.SplitSummary, error) {
	return s.splitExpenseRepo.GetSplitSummaryByUser(userID)
}

func (s *SplitExpenseService) SettleExpense(splitExpenseID, userID int) error {
	// Get the split expense to verify access and get details
	splitExpense, err := s.splitExpenseRepo.GetSplitExpenseByID(splitExpenseID, userID)
	if err != nil {
		return fmt.Errorf("split expense not found or access denied: %w", err)
	}

	// Find the participant for this user
	var participantID int
	var amountOwed, amountPaid float64
	found := false

	for _, participant := range splitExpense.Participants {
		if (participant.UserID == userID) || 
		   (participant.User != nil && participant.User.ID == userID) {
			participantID = participant.ID
			amountOwed = participant.AmountOwed
			amountPaid = participant.AmountPaid
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("user is not a participant in this split expense")
	}

	// Calculate remaining amount to be paid
	remainingAmount := amountOwed - amountPaid
	if remainingAmount <= 0.01 { // Already settled (allow for floating point precision)
		return fmt.Errorf("participant has already settled their portion")
	}

	// Update the participant payment to settle the full amount
	return s.splitExpenseRepo.UpdateParticipantPayment(participantID, remainingAmount, userID)
}