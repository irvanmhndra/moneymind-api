package services

import (
	"moneymind-backend/internal/models"
	"moneymind-backend/internal/repository"
)

type UserService struct {
	userRepo *repository.UserRepository
}

func NewUserService(userRepo *repository.UserRepository) *UserService {
	return &UserService{
		userRepo: userRepo,
	}
}

func (s *UserService) GetUserByID(id int) (*models.User, error) {
	return s.userRepo.GetUserByID(id)
}

func (s *UserService) UpdateUser(id int, updates *models.UserUpdate) (*models.User, error) {
	return s.userRepo.UpdateUser(id, updates)
}

func (s *UserService) DeactivateUser(id int) error {
	return s.userRepo.DeactivateUser(id)
}

func (s *UserService) ChangePassword(id int, oldPassword, newPassword string) error {
	return s.userRepo.ChangePassword(id, oldPassword, newPassword)
}

func (s *UserService) VerifyEmail(id int) error {
	return s.userRepo.VerifyEmail(id)
}