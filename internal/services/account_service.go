package services

import (
	"moneymind-backend/internal/models"
	"moneymind-backend/internal/repository"
)

type AccountService struct {
	accountRepo *repository.AccountRepository
}

func NewAccountService(accountRepo *repository.AccountRepository) *AccountService {
	return &AccountService{
		accountRepo: accountRepo,
	}
}

func (s *AccountService) GetAccounts(userID int) ([]models.Account, error) {
	return s.accountRepo.GetAccountsByUserID(userID)
}

func (s *AccountService) GetAccountByID(id, userID int) (*models.Account, error) {
	return s.accountRepo.GetAccountByID(id, userID)
}

func (s *AccountService) CreateAccount(userID int, account *models.AccountCreate) (*models.Account, error) {
	return s.accountRepo.CreateAccount(userID, account)
}

func (s *AccountService) UpdateAccount(id, userID int, updates *models.AccountUpdate) (*models.Account, error) {
	return s.accountRepo.UpdateAccount(id, userID, updates)
}

func (s *AccountService) DeleteAccount(id, userID int) error {
	return s.accountRepo.DeleteAccount(id, userID)
}