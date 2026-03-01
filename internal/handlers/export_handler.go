package handlers

import (
	"moneymind-backend/internal/services"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type ExportHandler struct {
	exportService *services.ExportService
	validator     *validator.Validate
}

func NewExportHandler(exportService *services.ExportService) *ExportHandler {
	return &ExportHandler{
		exportService: exportService,
		validator:     validator.New(),
	}
}

// ExportData handles data export requests
func (h *ExportHandler) ExportData(c *gin.Context) {
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
	var exportRequest services.ExportRequest
	if err := c.ShouldBindJSON(&exportRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request body",
			"error":   err.Error(),
		})
		return
	}

	// Validate request
	if err := h.validator.Struct(&exportRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Validation failed",
			"error":   err.Error(),
		})
		return
	}

	// Validate data types
	validDataTypes := map[string]bool{
		"expenses":   true,
		"budgets":    true,
		"goals":      true,
		"categories": true,
	}

	for _, dataType := range exportRequest.DataTypes {
		if !validDataTypes[dataType] {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Invalid data type: " + dataType,
			})
			return
		}
	}

	// Export data
	result, err := h.exportService.ExportData(userID.(int), &exportRequest)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to export data",
			"error":   err.Error(),
		})
		return
	}

	// Set response headers for file download
	c.Header("Content-Disposition", "attachment; filename="+result.Filename)
	c.Header("Content-Type", result.ContentType)
	c.Header("Content-Length", string(rune(result.Size)))

	// Return the file content
	c.String(http.StatusOK, result.Content)
}

// GetExportInfo provides information about available export options
func (h *ExportHandler) GetExportInfo(c *gin.Context) {
	// Get user from context
	_, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "User not authenticated",
		})
		return
	}

	info := gin.H{
		"supported_formats": []string{"csv", "json"},
		"supported_data_types": []string{"expenses", "budgets", "goals", "categories"},
		"date_format": "2006-01-02",
		"max_records_per_type": 1000,
		"notes": []string{
			"CSV format provides tabular data suitable for spreadsheet applications",
			"JSON format provides structured data with full field details",
			"Date filters apply only to expenses (based on expense date)",
			"All other data types export all records regardless of date filters",
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    info,
	})
}

// ExportExpenses handles expense-specific export (backwards compatibility)
func (h *ExportHandler) ExportExpenses(c *gin.Context) {
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
	format := c.DefaultQuery("format", "csv")
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	var startDate, endDate *time.Time
	if startDateStr != "" {
		if parsed, err := time.Parse("2006-01-02", startDateStr); err == nil {
			startDate = &parsed
		}
	}
	if endDateStr != "" {
		if parsed, err := time.Parse("2006-01-02", endDateStr); err == nil {
			endDate = &parsed
		}
	}

	// Create export request for expenses only
	exportRequest := services.ExportRequest{
		Format:    format,
		DataTypes: []string{"expenses"},
		StartDate: startDate,
		EndDate:   endDate,
	}

	// Export data
	result, err := h.exportService.ExportData(userID.(int), &exportRequest)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to export expenses",
			"error":   err.Error(),
		})
		return
	}

	// Set response headers for file download
	c.Header("Content-Disposition", "attachment; filename="+result.Filename)
	c.Header("Content-Type", result.ContentType)
	c.Header("Content-Length", string(rune(result.Size)))

	// Return the file content
	c.String(http.StatusOK, result.Content)
}