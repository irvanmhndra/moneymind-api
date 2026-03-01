package services

import (
	"moneymind-backend/internal/models"
	"moneymind-backend/internal/repository"
)

type ExpenseService struct {
	expenseRepo *repository.ExpenseRepository
}

func NewExpenseService(expenseRepo *repository.ExpenseRepository) *ExpenseService {
	return &ExpenseService{
		expenseRepo: expenseRepo,
	}
}

// TODO: Implement expense service methods
func (s *ExpenseService) CreateExpense(userID int, expense *models.ExpenseCreate) (*models.Expense, error) {
	return s.expenseRepo.CreateExpense(userID, expense)
}

func (s *ExpenseService) GetExpenses(userID int, filter *models.ExpenseFilter) ([]models.Expense, error) {
	return s.expenseRepo.GetExpensesByUserID(userID, filter)
}

func (s *ExpenseService) GetExpenseByID(id, userID int) (*models.Expense, error) {
	return s.expenseRepo.GetExpenseByID(id, userID)
}

func (s *ExpenseService) UpdateExpense(id, userID int, updates *models.ExpenseUpdate) (*models.Expense, error) {
	return s.expenseRepo.UpdateExpense(id, userID, updates)
}

func (s *ExpenseService) DeleteExpense(id, userID int) error {
	return s.expenseRepo.DeleteExpense(id, userID)
}

func (s *ExpenseService) UpdateExpenseReceiptPath(id, userID int, receiptPath string) (*models.Expense, error) {
	return s.expenseRepo.UpdateExpenseReceiptPath(id, userID, receiptPath)
}

func (s *ExpenseService) GetAnalytics(userID int, startDate, endDate, period, categoryID string) (*models.ExpenseAnalytics, error) {
	return s.expenseRepo.GetAnalytics(userID, startDate, endDate, period, categoryID)
}