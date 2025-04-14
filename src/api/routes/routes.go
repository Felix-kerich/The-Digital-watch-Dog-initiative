package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/the-digital-watchdog-initiative/controllers"
	"github.com/the-digital-watchdog-initiative/middleware"
	"github.com/the-digital-watchdog-initiative/models"
)

// RegisterRoutes registers all routes with their respective controllers
func RegisterRoutes(router *gin.RouterGroup) {
	// Public routes - No authentication required

	// Health check
	router.GET("/health", controllers.HealthCheck)

	// Auth routes
	authController := controllers.NewAuthController()
	auth := router.Group("/auth")
	{
		auth.POST("/register", authController.Register)
		auth.POST("/login", authController.Login)
		auth.POST("/refresh", authController.RefreshToken)
		auth.POST("/logout", authController.Logout)
		// These endpoints will be implemented in the future
		// auth.POST("/forgot-password", authController.ForgotPassword)
		// auth.POST("/reset-password", authController.ResetPassword)
	}

	// Protected routes - Authentication required
	protected := router.Group("")
	protected.Use(middleware.RequireAuth())
	{
		// User routesa
		userController := controllers.NewUserController()
		users := protected.Group("/users")
		{
			// Based on UserController implementation
			users.GET("/profile", userController.GetProfile)
			users.PUT("/profile", userController.UpdateProfile)
			users.POST("/change-password", userController.ChangePassword)
		}

		// Transaction routes
		transactionController := controllers.NewTransactionController()
		transactions := protected.Group("/transactions")
		{
			transactions.POST("", transactionController.Create)
			transactions.GET("", transactionController.GetAll)
			transactions.GET("/:id", transactionController.GetByID)
			// Use methods that match the actual TransactionController implementation
			// transactions.PUT("/:id", transactionController.Update)
			// transactions.DELETE("/:id", transactionController.Delete)
			transactions.POST("/:id/approve", transactionController.Approve)
			transactions.POST("/:id/reject", transactionController.Reject)
			transactions.POST("/:id/complete", transactionController.Complete)
		}

		// Fund routes
		fundController := controllers.NewFundController()
		funds := protected.Group("/funds")
		{
			funds.POST("", fundController.Create)
			funds.GET("", fundController.GetAll)
			funds.GET("/:id", fundController.GetByID)
			funds.PUT("/:id", fundController.Update)
			funds.DELETE("/:id", fundController.Delete)
			// This method will be implemented in the future
			// funds.GET("/:id/transactions", fundController.GetTransactions)
		}

		// Budget Line Item routes
		budgetController := controllers.NewBudgetLineItemController()
		budgets := protected.Group("/budgets")
		{
			budgets.POST("", budgetController.Create)
			budgets.GET("", budgetController.GetAll)
			budgets.GET("/:id", budgetController.GetByID)
			budgets.PUT("/:id", budgetController.Update)
			budgets.DELETE("/:id", budgetController.Delete)
		}

		// Entity routes
		entityController := controllers.NewEntityController()
		entities := protected.Group("/entities")
		{
			entities.POST("", entityController.Create)
			entities.GET("", entityController.GetAll)
			entities.GET("/:id", entityController.GetByID)
			entities.PUT("/:id", entityController.Update)
			entities.DELETE("/:id", entityController.Delete)
			// This method will be implemented in the future
			// entities.GET("/:id/transactions", entityController.GetTransactions)
		}

		// File upload routes - Commented out until FileController is implemented

		fileController := controllers.NewFileController()
		files := protected.Group("/files")
		{
			files.POST("", fileController.Upload)
			files.GET("/:id", fileController.Download)
			files.DELETE("/:id", fileController.Delete)
		}

		// Analytics routes (admin and auditor only)
		analyticsRoutes := protected.Group("/analytics")
		analyticsRoutes.Use(middleware.RequireRole(models.RoleAdmin, models.RoleAuditor))
		{
			analyticsController := controllers.NewAnalyticsController()
			// Based on AnalyticsController implementation
			analyticsRoutes.GET("/transactions", analyticsController.GetTransactionSummary)
			analyticsRoutes.GET("/users", analyticsController.GetUserActivitySummary)
			analyticsRoutes.GET("/funds", analyticsController.GetFundUtilizationReport)
			// These endpoints will be implemented in the future
			// analyticsRoutes.GET("/dashboard", analyticsController.GetDashboard)
			// analyticsRoutes.GET("/suspicious-transactions", analyticsController.GetSuspiciousTransactions)
			// analyticsRoutes.GET("/audit-trail", analyticsController.GetAuditTrail)
			// analyticsRoutes.GET("/reports", analyticsController.GetReports)
		}

		// Admin routes (admin only)
		adminRoutes := protected.Group("/admin")
		adminRoutes.Use(middleware.RequireRole(models.RoleAdmin))
		{
			adminController := controllers.NewAdminController()
			// Based on AdminController implementation
			adminRoutes.GET("/users", adminController.GetUsers)
			adminRoutes.GET("/users/:id", adminController.GetUserByID)
			adminRoutes.POST("/users", adminController.CreateUser)
			adminRoutes.PUT("/users/:id", adminController.UpdateUser)
			adminRoutes.POST("/users/:id/reset-password", adminController.ResetUserPassword)
			adminRoutes.GET("/system-info", adminController.GetSystemInfo)
			// These endpoints will be implemented in the future
			// adminRoutes.GET("/settings", adminController.GetSettings)
			// adminRoutes.PUT("/settings", adminController.UpdateSettings)
			// adminRoutes.GET("/logs", adminController.GetLogs)
		}

		// User management routes (admin only)
		userManagementController := controllers.NewUserManagementController()
		userRoutes := protected.Group("/users")
		{
			userRoutes.POST("", userManagementController.CreateUser)
			userRoutes.GET("", userManagementController.GetUsers)
		}
	}
}
