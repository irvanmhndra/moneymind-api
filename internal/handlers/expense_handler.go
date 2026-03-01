package handlers

import (
	"fmt"
	"io"
	"moneymind-backend/internal/models"
	"moneymind-backend/internal/services"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// parseFlexibleDate parses dates in multiple formats
func parseFlexibleDate(s string) (*time.Time, error) {
	if s == "" {
		return nil, nil
	}

	// Try various date formats
	formats := []string{
		"2006-01-02",           // Simple date format
		"2006-01-02T15:04:05Z", // RFC3339 UTC
		"2006-01-02T15:04:05Z07:00", // RFC3339 with timezone
		"2006-01-02 15:04:05",  // SQL timestamp format
	}

	for _, format := range formats {
		if t, err := time.Parse(format, s); err == nil {
			return &t, nil
		}
	}

	return nil, fmt.Errorf("invalid date format: %s", s)
}

type ExpenseHandler struct {
	expenseService *services.ExpenseService
	validator      *validator.Validate
}

func NewExpenseHandler(expenseService *services.ExpenseService) *ExpenseHandler {
	return &ExpenseHandler{
		expenseService: expenseService,
		validator:      validator.New(),
	}
}

func (h *ExpenseHandler) GetExpenses(c *gin.Context) {
	// Get user from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "User not authenticated",
		})
		return
	}

	// Parse query parameters for filtering
	var filter models.ExpenseFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid query parameters",
			"error":   err.Error(),
		})
		return
	}

	// Manually parse date parameters
	if startDateStr := c.Query("start_date"); startDateStr != "" {
		startDate, err := parseFlexibleDate(startDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Invalid start_date format",
				"error":   err.Error(),
			})
			return
		}
		filter.StartDate = startDate
	}

	if endDateStr := c.Query("end_date"); endDateStr != "" {
		endDate, err := parseFlexibleDate(endDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Invalid end_date format",
				"error":   err.Error(),
			})
			return
		}
		filter.EndDate = endDate
	}

	// Set default limit if not provided
	if filter.Limit == 0 {
		filter.Limit = 50
	}

	// Get expenses
	expenses, err := h.expenseService.GetExpenses(userID.(int), &filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to get expenses",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Expenses retrieved successfully",
		"data":    expenses,
	})
}

func (h *ExpenseHandler) CreateExpense(c *gin.Context) {
	// Get user from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "User not authenticated",
		})
		return
	}

	var expense models.ExpenseCreate
	if err := c.ShouldBindJSON(&expense); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request data",
			"error":   err.Error(),
		})
		return
	}

	// Validate input
	if err := h.validator.Struct(&expense); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Validation failed",
			"error":   err.Error(),
		})
		return
	}

	// Create expense
	newExpense, err := h.expenseService.CreateExpense(userID.(int), &expense)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to create expense",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Expense created successfully",
		"data":    newExpense,
	})
}

func (h *ExpenseHandler) GetExpense(c *gin.Context) {
	// Get user from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "User not authenticated",
		})
		return
	}

	// Get expense ID from URL
	expenseID := c.Param("id")
	id, err := strconv.Atoi(expenseID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid expense ID",
		})
		return
	}

	// Get expense
	expense, err := h.expenseService.GetExpenseByID(id, userID.(int))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Expense retrieved successfully",
		"data":    expense,
	})
}

func (h *ExpenseHandler) UpdateExpense(c *gin.Context) {
	// Get user from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "User not authenticated",
		})
		return
	}

	// Get expense ID from URL
	expenseID := c.Param("id")
	id, err := strconv.Atoi(expenseID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid expense ID",
		})
		return
	}

	var updates models.ExpenseUpdate
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request data",
			"error":   err.Error(),
		})
		return
	}

	// Validate input
	if err := h.validator.Struct(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Validation failed",
			"error":   err.Error(),
		})
		return
	}

	// Update expense
	expense, err := h.expenseService.UpdateExpense(id, userID.(int), &updates)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to update expense",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Expense updated successfully",
		"data":    expense,
	})
}

func (h *ExpenseHandler) DeleteExpense(c *gin.Context) {
	// Get user from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "User not authenticated",
		})
		return
	}

	// Get expense ID from URL
	expenseID := c.Param("id")
	id, err := strconv.Atoi(expenseID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid expense ID",
		})
		return
	}

	// Delete expense
	err = h.expenseService.DeleteExpense(id, userID.(int))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to delete expense",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Expense deleted successfully",
	})
}

func (h *ExpenseHandler) UploadReceipt(c *gin.Context) {
	// Get user from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "User not authenticated",
		})
		return
	}

	// Get expense ID from URL
	expenseID := c.Param("id")
	id, err := strconv.Atoi(expenseID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid expense ID",
		})
		return
	}

	// Check if expense exists and belongs to user
	_, err = h.expenseService.GetExpenseByID(id, userID.(int))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	// Parse multipart form
	file, header, err := c.Request.FormFile("receipt")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Receipt file is required",
			"error":   err.Error(),
		})
		return
	}
	defer file.Close()

	// Validate file type
	allowedTypes := map[string]bool{
		"image/jpeg":      true,
		"image/jpg":       true,
		"image/png":       true,
		"application/pdf": true,
	}

	contentType := header.Header.Get("Content-Type")
	if !allowedTypes[contentType] {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid file type. Only JPEG, PNG, and PDF files are allowed",
		})
		return
	}

	// Validate file size (10MB max)
	const maxSize = 10 << 20 // 10MB
	if header.Size > maxSize {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "File size too large. Maximum size is 10MB",
		})
		return
	}

	// Generate unique filename
	ext := filepath.Ext(header.Filename)
	if ext == "" {
		switch contentType {
		case "image/jpeg", "image/jpg":
			ext = ".jpg"
		case "image/png":
			ext = ".png"
		case "application/pdf":
			ext = ".pdf"
		}
	}

	filename := fmt.Sprintf("receipt_%d_%d_%d%s", 
		userID.(int), id, time.Now().Unix(), ext)

	// Ensure uploads directory exists
	uploadDir := "./uploads/receipts"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to create upload directory",
			"error":   err.Error(),
		})
		return
	}

	// Create file path
	filePath := filepath.Join(uploadDir, filename)

	// Create the file
	dst, err := os.Create(filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to create file",
			"error":   err.Error(),
		})
		return
	}
	defer dst.Close()

	// Copy file content
	if _, err := io.Copy(dst, file); err != nil {
		// Clean up the file if copy failed
		os.Remove(filePath)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to save file",
			"error":   err.Error(),
		})
		return
	}

	// Update expense with receipt path
	receiptPath := strings.Replace(filePath, "./", "/", 1)
	updatedExpense, err := h.expenseService.UpdateExpenseReceiptPath(id, userID.(int), receiptPath)
	if err != nil {
		// Clean up the file if database update failed
		os.Remove(filePath)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to update expense with receipt path",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Receipt uploaded successfully",
		"data": gin.H{
			"expense":      updatedExpense,
			"receipt_path": receiptPath,
			"file_size":    header.Size,
			"file_type":    contentType,
		},
	})
}

func (h *ExpenseHandler) GetAnalytics(c *gin.Context) {
	// Get user from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "User not authenticated",
		})
		return
	}

	// Parse query parameters for date range
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	period := c.DefaultQuery("period", "month") // month, week, year
	categoryID := c.Query("category_id")

	// Get analytics data
	analytics, err := h.expenseService.GetAnalytics(userID.(int), startDate, endDate, period, categoryID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to get analytics",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Analytics retrieved successfully",
		"data":    analytics,
	})
}

func (h *ExpenseHandler) ExportExpenses(c *gin.Context) {
	c.JSON(http.StatusMovedPermanently, gin.H{
		"success": false,
		"message": "Export functionality has been moved to /api/v1/export/expenses",
		"redirect_url": "/api/v1/export/expenses",
	})
}