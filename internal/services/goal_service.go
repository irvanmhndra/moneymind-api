package services

import (
	"moneymind-backend/internal/models"
	"moneymind-backend/internal/repository"
)

type GoalService struct {
	goalRepo *repository.GoalRepository
}

func NewGoalService(goalRepo *repository.GoalRepository) *GoalService {
	return &GoalService{
		goalRepo: goalRepo,
	}
}

func (s *GoalService) GetGoals(userID int) ([]models.Goal, error) {
	return s.goalRepo.GetGoalsByUserID(userID)
}

func (s *GoalService) CreateGoal(userID int, goal *models.GoalCreate) (*models.Goal, error) {
	return s.goalRepo.CreateGoal(userID, goal)
}

func (s *GoalService) UpdateGoal(id, userID int, updates *models.GoalUpdate) (*models.Goal, error) {
	return s.goalRepo.UpdateGoal(id, userID, updates)
}

func (s *GoalService) DeleteGoal(id, userID int) error {
	return s.goalRepo.DeleteGoal(id, userID)
}

func (s *GoalService) GetGoalByID(id, userID int) (*models.Goal, error) {
	return s.goalRepo.GetGoalByID(id, userID)
}

func (s *GoalService) UpdateGoalProgress(id, userID int, progress *models.GoalProgress) (*models.Goal, error) {
	return s.goalRepo.UpdateGoalProgress(id, userID, progress)
}

func (s *GoalService) GetGoalsWithProgress(userID int) ([]models.GoalWithProgress, error) {
	return s.goalRepo.GetGoalsWithProgress(userID)
}

func (s *GoalService) GetGoalSummary(userID int) (*models.GoalSummary, error) {
	return s.goalRepo.GetGoalSummary(userID)
}