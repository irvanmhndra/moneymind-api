package repository

import (
	"database/sql"
	"fmt"
	"moneymind-backend/internal/models"
	"strings"
	"time"
)

type CategoryRepository struct {
	db *sql.DB
}

func NewCategoryRepository(db *sql.DB) *CategoryRepository {
	return &CategoryRepository{db: db}
}

func (r *CategoryRepository) GetCategoriesByUserID(userID int) ([]models.Category, error) {
	query := `
		SELECT id, user_id, name, color, icon, is_system, parent_id, is_active, created_at, updated_at
		FROM categories 
		WHERE (user_id = $1 OR is_system = true) AND is_active = true
		ORDER BY CASE WHEN parent_id IS NULL THEN 0 ELSE 1 END, name ASC
	`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get categories: %w", err)
	}
	defer rows.Close()

	var categories []models.Category
	categoryMap := make(map[int]*models.Category)

	for rows.Next() {
		var category models.Category
		err := rows.Scan(
			&category.ID,
			&category.UserID,
			&category.Name,
			&category.Color,
			&category.Icon,
			&category.IsSystem,
			&category.ParentID,
			&category.IsActive,
			&category.CreatedAt,
			&category.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan category: %w", err)
		}

		categories = append(categories, category)
		categoryMap[category.ID] = &categories[len(categories)-1]
	}

	// Build parent-child relationships
	for i := range categories {
		if categories[i].ParentID != nil {
			if parent, exists := categoryMap[*categories[i].ParentID]; exists {
				parent.Subcategories = append(parent.Subcategories, categories[i])
				categories[i].Parent = &models.Category{
					ID:   parent.ID,
					Name: parent.Name,
				}
			}
		}
	}

	return categories, nil
}

func (r *CategoryRepository) GetCategoryByID(id, userID int) (*models.Category, error) {
	query := `
		SELECT id, user_id, name, color, icon, is_system, parent_id, is_active, created_at, updated_at
		FROM categories 
		WHERE id = $1 AND (user_id = $2 OR is_system = true) AND is_active = true
	`

	var category models.Category
	err := r.db.QueryRow(query, id, userID).Scan(
		&category.ID,
		&category.UserID,
		&category.Name,
		&category.Color,
		&category.Icon,
		&category.IsSystem,
		&category.ParentID,
		&category.IsActive,
		&category.CreatedAt,
		&category.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("category not found")
		}
		return nil, fmt.Errorf("failed to get category: %w", err)
	}

	return &category, nil
}

func (r *CategoryRepository) CreateCategory(userID int, category *models.CategoryCreate) (*models.Category, error) {
	if category.ParentID != nil {
		// Validate parent exists and belongs to user or is system
		_, err := r.GetCategoryByID(*category.ParentID, userID)
		if err != nil {
			return nil, fmt.Errorf("parent category not found or not accessible")
		}
	}

	query := `
		INSERT INTO categories (user_id, name, color, icon, parent_id, is_system, is_active)
		VALUES ($1, $2, $3, $4, $5, false, true)
		RETURNING id, user_id, name, color, icon, is_system, parent_id, is_active, created_at, updated_at
	`

	var newCategory models.Category
	err := r.db.QueryRow(
		query,
		userID,
		category.Name,
		category.Color,
		category.Icon,
		category.ParentID,
	).Scan(
		&newCategory.ID,
		&newCategory.UserID,
		&newCategory.Name,
		&newCategory.Color,
		&newCategory.Icon,
		&newCategory.IsSystem,
		&newCategory.ParentID,
		&newCategory.IsActive,
		&newCategory.CreatedAt,
		&newCategory.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create category: %w", err)
	}

	return &newCategory, nil
}

func (r *CategoryRepository) UpdateCategory(id, userID int, updates *models.CategoryUpdate) (*models.Category, error) {
	// First check if category exists and belongs to user (not system category)
	existing, err := r.GetCategoryByID(id, userID)
	if err != nil {
		return nil, err
	}

	if existing.IsSystem {
		return nil, fmt.Errorf("cannot update system category")
	}

	if existing.UserID == nil || *existing.UserID != userID {
		return nil, fmt.Errorf("category not found")
	}

	setParts := []string{}
	args := []interface{}{}
	argIndex := 1

	if updates.Name != "" {
		setParts = append(setParts, fmt.Sprintf("name = $%d", argIndex))
		args = append(args, updates.Name)
		argIndex++
	}
	if updates.Color != "" {
		setParts = append(setParts, fmt.Sprintf("color = $%d", argIndex))
		args = append(args, updates.Color)
		argIndex++
	}
	if updates.Icon != "" {
		setParts = append(setParts, fmt.Sprintf("icon = $%d", argIndex))
		args = append(args, updates.Icon)
		argIndex++
	}
	if updates.IsActive != nil {
		setParts = append(setParts, fmt.Sprintf("is_active = $%d", argIndex))
		args = append(args, *updates.IsActive)
		argIndex++
	}

	if len(setParts) == 0 {
		return existing, nil
	}

	setParts = append(setParts, fmt.Sprintf("updated_at = $%d", argIndex))
	args = append(args, time.Now())
	argIndex++

	args = append(args, id, userID)

	query := fmt.Sprintf(`
		UPDATE categories 
		SET %s 
		WHERE id = $%d AND user_id = $%d
		RETURNING id, user_id, name, color, icon, is_system, parent_id, is_active, created_at, updated_at
	`, strings.Join(setParts, ", "), argIndex, argIndex+1)

	var category models.Category
	err = r.db.QueryRow(query, args...).Scan(
		&category.ID,
		&category.UserID,
		&category.Name,
		&category.Color,
		&category.Icon,
		&category.IsSystem,
		&category.ParentID,
		&category.IsActive,
		&category.CreatedAt,
		&category.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to update category: %w", err)
	}

	return &category, nil
}

func (r *CategoryRepository) DeleteCategory(id, userID int) error {
	// Check if category exists and belongs to user
	existing, err := r.GetCategoryByID(id, userID)
	if err != nil {
		return err
	}

	if existing.IsSystem {
		return fmt.Errorf("cannot delete system category")
	}

	if existing.UserID == nil || *existing.UserID != userID {
		return fmt.Errorf("category not found")
	}

	// Check if category is being used by expenses
	var expenseCount int
	err = r.db.QueryRow("SELECT COUNT(*) FROM expenses WHERE category_id = $1", id).Scan(&expenseCount)
	if err != nil {
		return fmt.Errorf("failed to check category usage: %w", err)
	}

	if expenseCount > 0 {
		return fmt.Errorf("cannot delete category that is being used by expenses")
	}

	// Check if category has subcategories
	var subcategoryCount int
	err = r.db.QueryRow("SELECT COUNT(*) FROM categories WHERE parent_id = $1 AND is_active = true", id).Scan(&subcategoryCount)
	if err != nil {
		return fmt.Errorf("failed to check subcategories: %w", err)
	}

	if subcategoryCount > 0 {
		return fmt.Errorf("cannot delete category that has subcategories")
	}

	// Soft delete by setting is_active to false
	query := `UPDATE categories SET is_active = false, updated_at = $1 WHERE id = $2 AND user_id = $3`
	result, err := r.db.Exec(query, time.Now(), id, userID)
	if err != nil {
		return fmt.Errorf("failed to delete category: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("category not found")
	}

	return nil
}