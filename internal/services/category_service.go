package services

import (
	"moneymind-backend/internal/models"
	"moneymind-backend/internal/repository"
)

type CategoryService struct {
	categoryRepo *repository.CategoryRepository
}

func NewCategoryService(categoryRepo *repository.CategoryRepository) *CategoryService {
	return &CategoryService{
		categoryRepo: categoryRepo,
	}
}

func (s *CategoryService) GetCategories(userID int) ([]models.Category, error) {
	return s.categoryRepo.GetCategoriesByUserID(userID)
}

func (s *CategoryService) CreateCategory(userID int, category *models.CategoryCreate) (*models.Category, error) {
	return s.categoryRepo.CreateCategory(userID, category)
}

func (s *CategoryService) UpdateCategory(id, userID int, updates *models.CategoryUpdate) (*models.Category, error) {
	return s.categoryRepo.UpdateCategory(id, userID, updates)
}

func (s *CategoryService) DeleteCategory(id, userID int) error {
	return s.categoryRepo.DeleteCategory(id, userID)
}