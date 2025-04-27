package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"github.com/the-digital-watchdog-initiative/middleware"
	"github.com/the-digital-watchdog-initiative/models"
	"github.com/the-digital-watchdog-initiative/services"
	"github.com/the-digital-watchdog-initiative/utils"
)

// TransactionController handles transaction-related operations
type TransactionController struct {
	transactionService services.TransactionService
	fundService        services.FundService
	entityService      services.EntityService
	auditService       services.AuditService
	logger             *utils.NamedLogger
	config             *utils.Config
}

// NewTransactionController creates a new transaction controller
func NewTransactionController(transactionService services.TransactionService, fundService services.FundService, entityService services.EntityService, auditService services.AuditService) *TransactionController {
	return &TransactionController{
		transactionService: transactionService,
		fundService:        fundService,
		entityService:      entityService,
		auditService:       auditService,
		logger:             utils.NewLogger("transaction-handler"),
		config:             utils.GetConfig(),
	}
}

// CreateTransactionRequest represents data needed to create a transaction
type CreateTransactionRequest struct {
	TransactionType  models.TransactionType `json:"transactionType" binding:"required"`
	Amount           float64                `json:"amount" binding:"required,gt=0"`
	Currency         string                 `json:"currency" binding:"required"`
	Description      string                 `json:"description" binding:"required"`
	SourceID         string                 `json:"sourceId" binding:"required"`
	DestinationID    string                 `json:"destinationId" binding:"required"`
	FundID           string                 `json:"fundId" binding:"required"`
	BudgetLineItemID string                 `json:"budgetLineItemId"`
	DocumentRef      string                 `json:"documentRef"`
}

// Create creates a new transaction
func (tc *TransactionController) Create(c *gin.Context) {
	// Get current user from context
	userID, _ := c.Get("userID")
	userIDStr := userID.(string)

	var req CreateTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		tc.logger.Warn("Invalid transaction creation request", map[string]interface{}{
			"error":  err.Error(),
			"userID": userIDStr,
		})
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tc.logger.Info("Processing transaction creation request", map[string]interface{}{
		"userID": userIDStr,
		"amount": req.Amount,
		"type":   req.TransactionType,
		"fundID": req.FundID,
	})

	// Convert float64 to decimal.Decimal
	amountDecimal := decimal.NewFromFloat(req.Amount)

	// Create transaction using the service
	transaction, blockchainTxHash, err := tc.transactionService.CreateTransaction(
		userIDStr,
		req.TransactionType,
		amountDecimal,
		req.Currency,
		req.Description,
		req.SourceID,
		req.DestinationID,
		req.FundID,
		req.BudgetLineItemID,
		req.DocumentRef,
	)

	if err != nil {
		tc.logger.Error("Failed to create transaction", map[string]interface{}{
			"error":  err.Error(),
			"userID": userIDStr,
		})

		// Check for specific error types
		if validationErr, ok := err.(*utils.ValidationError); ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create transaction"})
		return
	}

	// Return both database and blockchain transaction details
	c.JSON(http.StatusCreated, gin.H{
		"message":     "Transaction created successfully",
		"transaction": transaction,
		"blockchain": gin.H{
			"transactionHash":  blockchainTxHash,
			"networkUrl":       tc.config.BlockchainServiceURL,
			"blockExplorerUrl": fmt.Sprintf("%s/tx/%s", tc.config.BlockchainServiceURL, blockchainTxHash),
		},
	})
}

// GetAll retrieves all transactions with pagination and filtering
func (tc *TransactionController) GetAll(c *gin.Context) {
	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}

	// Get filter parameters
	filter := map[string]interface{}{
		"type":          c.Query("type"),
		"status":        c.Query("status"),
		"fundId":        c.Query("fundId"),
		"sourceId":      c.Query("sourceId"),
		"destinationId": c.Query("destinationId"),
		"startDate":     c.Query("startDate"),
		"endDate":       c.Query("endDate"),
		"aiFlagged":     c.Query("aiFlagged"),
	}

	tc.logger.Info("Retrieving transactions", map[string]interface{}{
		"page":    page,
		"limit":   limit,
		"filters": filter,
	})

	transactions, total, err := tc.transactionService.GetAllTransactions(page, limit, filter)
	if err != nil {
		tc.logger.Error("Failed to retrieve transactions", map[string]interface{}{
			"error": err.Error(),
		})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve transactions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"transactions": transactions,
		"total":        total,
		"page":         page,
		"limit":        limit,
	})
}

// GetByID retrieves a transaction by its ID
func (tc *TransactionController) GetByID(c *gin.Context) {
	id := c.Param("id")

	tc.logger.Info("Retrieving transaction by ID", map[string]interface{}{
		"transactionID": id,
	})

	transaction, err := tc.transactionService.GetTransactionByID(id)
	if err != nil {
		tc.logger.Error("Failed to retrieve transaction", map[string]interface{}{
			"error":         err.Error(),
			"transactionID": id,
		})

		if err == utils.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Transaction not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve transaction"})
		}
		return
	}

	c.JSON(http.StatusOK, transaction)
}

// Approve approves a pending transaction
func (tc *TransactionController) Approve(c *gin.Context) {
	id := c.Param("id")
	userRole, _ := c.Get("userRole")
	userID, _ := c.Get("userID")
	userIDStr := userID.(string)

	tc.logger.Info("Processing transaction approval request", map[string]interface{}{
		"transactionID": id,
		"userID":        userIDStr,
		"userRole":      userRole,
	})

	// Check if the user has the required role
	if userRole != models.RoleAdmin && userRole != models.RoleAuditor && userRole != models.RoleFinanceOfficer && userRole != models.RoleManager {
		tc.logger.Warn("Insufficient permissions to approve transaction", map[string]interface{}{
			"userID":   userIDStr,
			"userRole": userRole,
		})
		c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions to approve transaction"})
		return
	}

	// Approve the transaction using the service
	err := tc.transactionService.ApproveTransaction(id, userIDStr)
	if err != nil {
		tc.logger.Error("Failed to approve transaction", map[string]interface{}{
			"error":         err.Error(),
			"transactionID": id,
			"userID":        userIDStr,
		})

		// Check for specific error types
		if err == utils.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Transaction not found"})
		} else if err == utils.ErrInvalidState {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Transaction cannot be approved in its current state"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to approve transaction"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Transaction approved successfully"})
}
func (tc *TransactionController) Reject(c *gin.Context) {
	id := c.Param("id")
	userRole, _ := c.Get("userRole")
	userID, _ := c.Get("userID")
	userIDStr := userID.(string)

	tc.logger.Info("Processing transaction rejection request", map[string]interface{}{
		"transactionID": id,
		"userID":        userIDStr,
		"userRole":      userRole,
	})

	// Check if the user has the required role
	if userRole != models.RoleAdmin && userRole != models.RoleAuditor && userRole != models.RoleFinanceOfficer && userRole != models.RoleManager {
		tc.logger.Warn("Insufficient permissions to reject transaction", map[string]interface{}{
			"userID":   userIDStr,
			"userRole": userRole,
		})
		c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions to reject transaction"})
		return
	}

	// Get rejection reason from request body
	var req struct {
		Reason string `json:"reason" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		tc.logger.Warn("Invalid rejection request", map[string]interface{}{
			"error":  err.Error(),
			"userID": userIDStr,
		})
		c.JSON(http.StatusBadRequest, gin.H{"error": "Rejection reason is required"})
		return
	}

	// Reject the transaction using the service
	err := tc.transactionService.RejectTransaction(id, userIDStr, req.Reason)
	if err != nil {
		tc.logger.Error("Failed to reject transaction", map[string]interface{}{
			"error":         err.Error(),
			"transactionID": id,
			"userID":        userIDStr,
		})

		// Check for specific error types
		if err == utils.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Transaction not found"})
		} else if err == utils.ErrInvalidState {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Transaction cannot be rejected in its current state"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reject transaction"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Transaction rejected successfully"})
}

// Complete marks a transaction as completed
func (tc *TransactionController) Complete(c *gin.Context) {
	id := c.Param("id")
	userRole, _ := c.Get("userRole")
	userID, _ := c.Get("userID")
	userIDStr := userID.(string)

	tc.logger.Info("Processing transaction completion request", map[string]interface{}{
		"transactionID": id,
		"userID":        userIDStr,
		"userRole":      userRole,
	})

	// Check if the user has the required role
	if userRole != models.RoleAdmin && userRole != models.RoleFinanceOfficer {
		tc.logger.Warn("Insufficient permissions to complete transaction", map[string]interface{}{
			"userID":   userIDStr,
			"userRole": userRole,
		})
		c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions to complete transaction"})
		return
	}

	// Complete the transaction using the service
	err := tc.transactionService.CompleteTransaction(id)
	if err != nil {
		tc.logger.Error("Failed to complete transaction", map[string]interface{}{
			"error":         err.Error(),
			"transactionID": id,
			"userID":        userIDStr,
		})

		// Check for specific error types
		if err == utils.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Transaction not found"})
		} else if err == utils.ErrInvalidState {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Transaction must be approved before it can be completed"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to complete transaction"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Transaction completed successfully"})
}

// RegisterRoutes registers routes for the transaction controller
func (tc *TransactionController) RegisterRoutes(rg *gin.RouterGroup) {
	routes := rg.Group("/transactions")

	// Public routes
	routes.GET("", tc.GetAll)
	routes.GET("/:id", tc.GetByID)

	// Protected routes
	protected := routes.Group("/")
	protected.Use(middleware.RequireAuth())

	// Routes for all authenticated users
	protected.POST("", middleware.AuditLog("create_transaction"), tc.Create)

	// Routes for admins, auditors, and finance officers
	protected.POST("/:id/approve", middleware.RequireRole(models.RoleAdmin, models.RoleAuditor, models.RoleFinanceOfficer, models.RoleManager), middleware.AuditLog("approve_transaction"), tc.Approve)
	protected.POST("/:id/reject", middleware.RequireRole(models.RoleAdmin, models.RoleAuditor, models.RoleFinanceOfficer, models.RoleManager), middleware.AuditLog("reject_transaction"), tc.Reject)

	// Routes for admins and finance officers only
	protected.POST("/:id/complete", middleware.RequireRole(models.RoleAdmin, models.RoleFinanceOfficer), middleware.AuditLog("complete_transaction"), tc.Complete)

	tc.logger.Info("Transaction routes registered", map[string]interface{}{})
}
