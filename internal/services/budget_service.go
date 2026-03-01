package services

import (
	"moneymind-backend/internal/models"
	"moneymind-backend/internal/repository"
)

type BudgetService struct {
	budgetRepo *repository.BudgetRepository
}

func NewBudgetService(budgetRepo *repository.BudgetRepository) *BudgetService {
	return &BudgetService{
		budgetRepo: budgetRepo,
	}
}

func (s *BudgetService) GetBudgets(userID int) ([]models.Budget, error) {
	return s.budgetRepo.GetBudgetsByUserID(userID)
}

func (s *BudgetService) CreateBudget(userID int, budget *models.BudgetCreate) (*models.Budget, error) {
	return s.budgetRepo.CreateBudget(userID, budget)
}

func (s *BudgetService) UpdateBudget(id, userID int, updates *models.BudgetUpdate) (*models.Budget, error) {
	return s.budgetRepo.UpdateBudget(id, userID, updates)
}

func (s *BudgetService) DeleteBudget(id, userID int) error {
	return s.budgetRepo.DeleteBudget(id, userID)
}

func (s *BudgetService) GetBudgetByID(id, userID int) (*models.Budget, error) {
	return s.budgetRepo.GetBudgetByID(id, userID)
}

func (s *BudgetService) GetBudgetStatus(userID int) ([]models.BudgetStatus, error) {
	return s.budgetRepo.GetBudgetStatus(userID)
}

func (s *BudgetService) GetBudgetSummary(userID int) (*models.BudgetSummary, error) {
	return s.budgetRepo.GetBudgetSummary(userID)
}