package main

import (
	"log"
	"moneymind-backend/internal/config"
	"moneymind-backend/internal/database"
	"moneymind-backend/internal/handlers"
	"moneymind-backend/internal/middleware"
	"moneymind-backend/internal/repository"
	"moneymind-backend/internal/services"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load configuration:", err)
	}

	// Initialize database
	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	sessionRepo := repository.NewSessionRepository(db)
	expenseRepo := repository.NewExpenseRepository(db)
	categoryRepo := repository.NewCategoryRepository(db)
	accountRepo := repository.NewAccountRepository(db)
	budgetRepo := repository.NewBudgetRepository(db)
	goalRepo := repository.NewGoalRepository(db)
	splitExpenseRepo := repository.NewSplitExpenseRepository(db)

	// Initialize services
	authService := services.NewAuthService(userRepo, sessionRepo, cfg.JWTSecret)
	userService := services.NewUserService(userRepo)
	expenseService := services.NewExpenseService(expenseRepo)
	categoryService := services.NewCategoryService(categoryRepo)
	accountService := services.NewAccountService(accountRepo)
	budgetService := services.NewBudgetService(budgetRepo)
	goalService := services.NewGoalService(goalRepo)
	splitExpenseService := services.NewSplitExpenseService(splitExpenseRepo, expenseRepo)
	exportService := services.NewExportService(expenseRepo, budgetRepo, goalRepo, categoryRepo)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService)
	userHandler := handlers.NewUserHandler(userService)
	expenseHandler := handlers.NewExpenseHandler(expenseService)
	categoryHandler := handlers.NewCategoryHandler(categoryService)
	accountHandler := handlers.NewAccountHandler(accountService)
	budgetHandler := handlers.NewBudgetHandler(budgetService)
	goalHandler := handlers.NewGoalHandler(goalService)
	splitExpenseHandler := handlers.NewSplitExpenseHandler(splitExpenseService)
	exportHandler := handlers.NewExportHandler(exportService)

	// Setup Gin
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	
	router := gin.New()
	
	// Middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	
	// CORS configuration - best practice approach
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = cfg.AllowedOrigins
	corsConfig.AllowMethods = cfg.AllowedMethods
	corsConfig.AllowHeaders = cfg.AllowedHeaders
	corsConfig.AllowCredentials = true
	router.Use(cors.New(corsConfig))

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "moneymind-backend"})
	})

	// Static file serving for receipts
	router.Static("/uploads", "./uploads")

	// API routes
	api := router.Group("/api/v1")
	{
		// Public routes
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.RefreshToken)
			auth.POST("/logout", authHandler.Logout)
			auth.POST("/forgot-password", authHandler.ForgotPassword)
			auth.POST("/reset-password", authHandler.ResetPassword)
		}

		// Protected routes
		protected := api.Group("/")
		protected.Use(middleware.AuthRequired(authService))
		{
			// User routes
			users := protected.Group("/users")
			{
				users.GET("/profile", userHandler.GetProfile)
				users.PUT("/profile", userHandler.UpdateProfile)
				users.DELETE("/account", userHandler.DeleteAccount)
			}

			// Category routes
			categories := protected.Group("/categories")
			{
				categories.GET("/", categoryHandler.GetCategories)
			categories.GET("", categoryHandler.GetCategories)
				categories.POST("/", categoryHandler.CreateCategory)
			categories.POST("", categoryHandler.CreateCategory)
				categories.PUT("/:id", categoryHandler.UpdateCategory)
				categories.DELETE("/:id", categoryHandler.DeleteCategory)
			}

			// Account routes
			accounts := protected.Group("/accounts")
			{
				accounts.GET("/", accountHandler.GetAccounts)
			accounts.GET("", accountHandler.GetAccounts)
				accounts.POST("/", accountHandler.CreateAccount)
			accounts.POST("", accountHandler.CreateAccount)
				accounts.PUT("/:id", accountHandler.UpdateAccount)
				accounts.DELETE("/:id", accountHandler.DeleteAccount)
			}

			// Expense routes
			expenses := protected.Group("/expenses")
			{
				expenses.GET("/", expenseHandler.GetExpenses)
			expenses.GET("", expenseHandler.GetExpenses)
				expenses.POST("/", expenseHandler.CreateExpense)
			expenses.POST("", expenseHandler.CreateExpense)
				expenses.GET("/:id", expenseHandler.GetExpense)
				expenses.PUT("/:id", expenseHandler.UpdateExpense)
				expenses.DELETE("/:id", expenseHandler.DeleteExpense)
				expenses.POST("/:id/receipt", expenseHandler.UploadReceipt)
				expenses.GET("/analytics", expenseHandler.GetAnalytics)
				expenses.GET("/export", expenseHandler.ExportExpenses)
			}

			// Budget routes
			budgets := protected.Group("/budgets")
			{
				budgets.GET("/", budgetHandler.GetBudgets)
			budgets.GET("", budgetHandler.GetBudgets)
				budgets.POST("/", budgetHandler.CreateBudget)
			budgets.POST("", budgetHandler.CreateBudget)
				budgets.GET("/:id", budgetHandler.GetBudgetByID)
				budgets.PUT("/:id", budgetHandler.UpdateBudget)
				budgets.DELETE("/:id", budgetHandler.DeleteBudget)
				budgets.GET("/status", budgetHandler.GetBudgetStatus)
				budgets.GET("/summary", budgetHandler.GetBudgetSummary)
			}

			// Goal routes
			goals := protected.Group("/goals")
			{
				goals.GET("/", goalHandler.GetGoals)
			goals.GET("", goalHandler.GetGoals)
				goals.POST("/", goalHandler.CreateGoal)
			goals.POST("", goalHandler.CreateGoal)
				goals.GET("/:id", goalHandler.GetGoalByID)
				goals.PUT("/:id", goalHandler.UpdateGoal)
				goals.DELETE("/:id", goalHandler.DeleteGoal)
				goals.POST("/:id/progress", goalHandler.UpdateProgress)
				goals.GET("/summary", goalHandler.GetGoalSummary)
			}

			// Split expense routes
			splitExpenses := protected.Group("/split-expenses")
			{
				splitExpenses.GET("/", splitExpenseHandler.GetSplitExpenses)
			splitExpenses.GET("", splitExpenseHandler.GetSplitExpenses)
				splitExpenses.POST("/", splitExpenseHandler.CreateSplitExpense)
			splitExpenses.POST("", splitExpenseHandler.CreateSplitExpense)
				splitExpenses.GET("/:id", splitExpenseHandler.GetSplitExpenseByID)
				splitExpenses.DELETE("/:id", splitExpenseHandler.DeleteSplitExpense)
				splitExpenses.PUT("/participants/:participant_id/payment", splitExpenseHandler.UpdateParticipantPayment)
				splitExpenses.POST("/:id/settle", splitExpenseHandler.SettleExpense)
				splitExpenses.GET("/summary", splitExpenseHandler.GetSplitSummary)
			}

			// Export routes
			export := protected.Group("/export")
			{
				export.GET("/info", exportHandler.GetExportInfo)
				export.POST("/data", exportHandler.ExportData)
				export.GET("/expenses", exportHandler.ExportExpenses)
			}
		}
	}

	// Start server
	log.Printf("Starting MoneyMind backend server on port %s", cfg.Port)
	log.Printf("Environment: %s", cfg.Environment)
	
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}