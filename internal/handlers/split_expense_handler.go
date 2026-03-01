package handlers

import (
	"moneymind-backend/internal/models"
	"moneymind-backend/internal/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type SplitExpenseHandler struct {
	splitExpenseService *services.SplitExpenseService
	validator           *validator.Validate
}

func NewSplitExpenseHandler(splitExpenseService *services.SplitExpenseService) *SplitExpenseHandler {
	return &SplitExpenseHandler{
		splitExpenseService: splitExpenseService,
		validator:           validator.New(),
	}
}

// CreateSplitExpense creates a new split expense
func (h *SplitExpenseHandler) CreateSplitExpense(c *gin.Context) {
	// Get user from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "User not authenticated",
		})
		return
	}

	// Parse request body
	var splitExpense models.SplitExpenseCreate
	if err := c.ShouldBindJSON(&splitExpense); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request body",
			"error":   err.Error(),
		})
		return
	}

	// Validate request
	if err := h.validator.Struct(&splitExpense); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Validation failed",
			"error":   err.Error(),
		})
		return
	}

	// Create split expense
	newSplitExpense, err := h.splitExpenseService.CreateSplitExpense(userID.(int), &splitExpense)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Failed to create split expense",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Split expense created successfully",
		"data":    newSplitExpense,
	})
}

// GetSplitExpenses gets all split expenses for a user
func (h *SplitExpenseHandler) GetSplitExpenses(c *gin.Context) {
	// Get user from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "User not authenticated",
		})
		return
	}

	// Parse query parameters
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 10
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		offset = 0
	}

	// Get split expenses
	splitExpenses, err := h.splitExpenseService.GetSplitExpensesByUser(userID.(int), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to get split expenses",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    splitExpenses,
	})
}

// GetSplitExpenseByID gets a specific split expense by ID
func (h *SplitExpenseHandler) GetSplitExpenseByID(c *gin.Context) {
	// Get user from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "User not authenticated",
		})
		return
	}

	// Parse ID from URL
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid split expense ID",
		})
		return
	}

	// Get split expense
	splitExpense, err := h.splitExpenseService.GetSplitExpenseByID(id, userID.(int))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "Split expense not found",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    splitExpense,
	})
}

// UpdateParticipantPayment updates a participant's payment
func (h *SplitExpenseHandler) UpdateParticipantPayment(c *gin.Context) {
	// Get user from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "User not authenticated",
		})
		return
	}

	// Parse participant ID from URL
	participantIDStr := c.Param("participant_id")
	participantID, err := strconv.Atoi(participantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid participant ID",
		})
		return
	}

	// Parse request body
	var paymentData struct {
		AmountPaid float64 `json:"amount_paid" validate:"required,gt=0"`
	}
	if err := c.ShouldBindJSON(&paymentData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request body",
			"error":   err.Error(),
		})
		return
	}

	// Validate request
	if err := h.validator.Struct(&paymentData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Validation failed",
			"error":   err.Error(),
		})
		return
	}

	// Update payment
	err = h.splitExpenseService.UpdateParticipantPayment(participantID, paymentData.AmountPaid, userID.(int))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Failed to update payment",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Payment updated successfully",
	})
}

// SettleExpense settles a participant's portion of the expense
func (h *SplitExpenseHandler) SettleExpense(c *gin.Context) {
	// Get user from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "User not authenticated",
		})
		return
	}

	// Parse split expense ID from URL
	splitExpenseIDStr := c.Param("id")
	splitExpenseID, err := strconv.Atoi(splitExpenseIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid split expense ID",
		})
		return
	}

	// Settle the expense
	err = h.splitExpenseService.SettleExpense(splitExpenseID, userID.(int))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Failed to settle expense",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Expense settled successfully",
	})
}

// DeleteSplitExpense deletes a split expense
func (h *SplitExpenseHandler) DeleteSplitExpense(c *gin.Context) {
	// Get user from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "User not authenticated",
		})
		return
	}

	// Parse ID from URL
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid split expense ID",
		})
		return
	}

	// Delete split expense
	err = h.splitExpenseService.DeleteSplitExpense(id, userID.(int))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Failed to delete split expense",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Split expense deleted successfully",
	})
}

// GetSplitSummary gets a summary of all split expenses for a user
func (h *SplitExpenseHandler) GetSplitSummary(c *gin.Context) {
	// Get user from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "User not authenticated",
		})
		return
	}

	// Get split summary
	summary, err := h.splitExpenseService.GetSplitSummaryByUser(userID.(int))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to get split summary",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    summary,
	})
}