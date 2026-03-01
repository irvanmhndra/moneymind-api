package services

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"moneymind-backend/internal/models"
	"moneymind-backend/internal/repository"
	"strconv"
	"strings"
	"time"
)

type ExportService struct {
	expenseRepo *repository.ExpenseRepository
	budgetRepo  *repository.BudgetRepository
	goalRepo    *repository.GoalRepository
	categoryRepo *repository.CategoryRepository
}

func NewExportService(expenseRepo *repository.ExpenseRepository, budgetRepo *repository.BudgetRepository, goalRepo *repository.GoalRepository, categoryRepo *repository.CategoryRepository) *ExportService {
	return &ExportService{
		expenseRepo:  expenseRepo,
		budgetRepo:   budgetRepo,
		goalRepo:     goalRepo,
		categoryRepo: categoryRepo,
	}
}

type ExportRequest struct {
	Format    string    `json:"format" validate:"required,oneof=csv json"`
	DataTypes []string  `json:"data_types" validate:"required,min=1"`
	StartDate *time.Time `json:"start_date"`
	EndDate   *time.Time `json:"end_date"`
}

type ExportResult struct {
	Content     string `json:"content"`
	Filename    string `json:"filename"`
	ContentType string `json:"content_type"`
	Size        int    `json:"size"`
}

func (s *ExportService) ExportData(userID int, request *ExportRequest) (*ExportResult, error) {
	switch request.Format {
	case "csv":
		content, filename, err := s.exportToCSV(userID, request)
		if err != nil {
			return nil, err
		}
		return &ExportResult{
			Content:     content,
			Filename:    filename,
			ContentType: "text/csv",
			Size:        len(content),
		}, nil

	case "json":
		content, filename, err := s.exportToJSON(userID, request)
		if err != nil {
			return nil, err
		}
		return &ExportResult{
			Content:     content,
			Filename:    filename,
			ContentType: "application/json",
			Size:        len(content),
		}, nil

	default:
		return nil, fmt.Errorf("unsupported export format: %s", request.Format)
	}
}

func (s *ExportService) exportToCSV(userID int, request *ExportRequest) (string, string, error) {
	var csvData strings.Builder
	writer := csv.NewWriter(&csvData)
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("moneymind_export_%s.csv", timestamp)

	for _, dataType := range request.DataTypes {
		switch dataType {
		case "expenses":
			if err := s.exportExpensesToCSV(writer, userID, request.StartDate, request.EndDate); err != nil {
				return "", "", fmt.Errorf("failed to export expenses: %w", err)
			}

		case "budgets":
			if err := s.exportBudgetsToCSV(writer, userID); err != nil {
				return "", "", fmt.Errorf("failed to export budgets: %w", err)
			}

		case "goals":
			if err := s.exportGoalsToCSV(writer, userID); err != nil {
				return "", "", fmt.Errorf("failed to export goals: %w", err)
			}

		case "categories":
			if err := s.exportCategoriesToCSV(writer, userID); err != nil {
				return "", "", fmt.Errorf("failed to export categories: %w", err)
			}

		default:
			return "", "", fmt.Errorf("unsupported data type: %s", dataType)
		}
		
		// Add a blank line between data types
		writer.Write([]string{})
	}

	writer.Flush()
	return csvData.String(), filename, writer.Error()
}

func (s *ExportService) exportExpensesToCSV(writer *csv.Writer, userID int, startDate, endDate *time.Time) error {
	// Build filter for expenses
	filter := &models.ExpenseFilter{
		StartDate: startDate,
		EndDate:   endDate,
		Limit:     1000, // Get all expenses
	}

	expenses, err := s.expenseRepo.GetExpensesByUserID(userID, filter)
	if err != nil {
		return err
	}

	// Write header for expenses
	writer.Write([]string{"Data Type", "ID", "Amount", "Currency", "Description", "Notes", "Category", "Account", "Location", "Date", "Is Recurring", "Recurring Frequency", "Tax Deductible", "Tax Category", "Created At"})

	// Write expense data
	for _, expense := range expenses {
		categoryName := ""
		if expense.Category != nil {
			categoryName = expense.Category.Name
		}
		
		accountName := ""
		if expense.Account != nil {
			accountName = expense.Account.Name
		}

		writer.Write([]string{
			"expense",
			strconv.Itoa(expense.ID),
			strconv.FormatFloat(expense.Amount, 'f', 2, 64),
			expense.Currency,
			expense.Description,
			expense.Notes,
			categoryName,
			accountName,
			expense.Location,
			expense.Date.Format("2006-01-02"),
			strconv.FormatBool(expense.IsRecurring),
			expense.RecurringFrequency,
			strconv.FormatBool(expense.TaxDeductible),
			expense.TaxCategory,
			expense.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	return nil
}

func (s *ExportService) exportBudgetsToCSV(writer *csv.Writer, userID int) error {
	budgets, err := s.budgetRepo.GetBudgetsByUserID(userID)
	if err != nil {
		return err
	}

	// Write header for budgets
	writer.Write([]string{"Data Type", "ID", "Name", "Category", "Amount", "Period", "Start Date", "End Date", "Is Active", "Created At"})

	// Write budget data
	for _, budget := range budgets {
		categoryName := ""
		if budget.Category != nil {
			categoryName = budget.Category.Name
		}
		
		endDate := ""
		if budget.EndDate != nil {
			endDate = budget.EndDate.Format("2006-01-02")
		}

		writer.Write([]string{
			"budget",
			strconv.Itoa(budget.ID),
			budget.Name,
			categoryName,
			strconv.FormatFloat(budget.Amount, 'f', 2, 64),
			budget.Period,
			budget.StartDate.Format("2006-01-02"),
			endDate,
			strconv.FormatBool(budget.IsActive),
			budget.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	return nil
}

func (s *ExportService) exportGoalsToCSV(writer *csv.Writer, userID int) error {
	goals, err := s.goalRepo.GetGoalsByUserID(userID)
	if err != nil {
		return err
	}

	// Write header for goals
	writer.Write([]string{"Data Type", "ID", "Name", "Description", "Goal Type", "Target Amount", "Current Amount", "Target Date", "Is Achieved", "Created At"})

	// Write goal data
	for _, goal := range goals {
		targetDate := ""
		if goal.TargetDate != nil {
			targetDate = goal.TargetDate.Format("2006-01-02")
		}

		writer.Write([]string{
			"goal",
			strconv.Itoa(goal.ID),
			goal.Name,
			goal.Description,
			goal.GoalType,
			strconv.FormatFloat(goal.TargetAmount, 'f', 2, 64),
			strconv.FormatFloat(goal.CurrentAmount, 'f', 2, 64),
			targetDate,
			strconv.FormatBool(goal.IsAchieved),
			goal.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	return nil
}

func (s *ExportService) exportCategoriesToCSV(writer *csv.Writer, userID int) error {
	categories, err := s.categoryRepo.GetCategoriesByUserID(userID)
	if err != nil {
		return err
	}

	// Write header for categories
	writer.Write([]string{"Data Type", "ID", "Name", "Color", "Icon", "Parent Category", "Created At"})

	// Write category data
	for _, category := range categories {
		parentName := ""
		if category.Parent != nil {
			parentName = category.Parent.Name
		}

		writer.Write([]string{
			"category",
			strconv.Itoa(category.ID),
			category.Name,
			category.Color,
			category.Icon,
			parentName,
			category.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	return nil
}

func (s *ExportService) exportToJSON(userID int, request *ExportRequest) (string, string, error) {
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("moneymind_export_%s.json", timestamp)

	exportData := make(map[string]interface{})
	exportData["export_timestamp"] = time.Now().Format("2006-01-02 15:04:05")
	exportData["export_format"] = "json"
	exportData["user_id"] = userID

	for _, dataType := range request.DataTypes {
		switch dataType {
		case "expenses":
			filter := &models.ExpenseFilter{
				StartDate: request.StartDate,
				EndDate:   request.EndDate,
				Limit:     1000,
			}
			expenses, err := s.expenseRepo.GetExpensesByUserID(userID, filter)
			if err != nil {
				return "", "", fmt.Errorf("failed to export expenses: %w", err)
			}
			exportData["expenses"] = expenses

		case "budgets":
			budgets, err := s.budgetRepo.GetBudgetsByUserID(userID)
			if err != nil {
				return "", "", fmt.Errorf("failed to export budgets: %w", err)
			}
			exportData["budgets"] = budgets

		case "goals":
			goals, err := s.goalRepo.GetGoalsByUserID(userID)
			if err != nil {
				return "", "", fmt.Errorf("failed to export goals: %w", err)
			}
			exportData["goals"] = goals

		case "categories":
			categories, err := s.categoryRepo.GetCategoriesByUserID(userID)
			if err != nil {
				return "", "", fmt.Errorf("failed to export categories: %w", err)
			}
			exportData["categories"] = categories

		default:
			return "", "", fmt.Errorf("unsupported data type: %s", dataType)
		}
	}

	jsonBytes, err := json.MarshalIndent(exportData, "", "  ")
	if err != nil {
		return "", "", fmt.Errorf("failed to marshal JSON: %w", err)
	}

	return string(jsonBytes), filename, nil
}