package controllers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"github.com/the-digital-watchdog-initiative/middleware"
	"github.com/the-digital-watchdog-initiative/models"
	"github.com/the-digital-watchdog-initiative/repository"
	"github.com/the-digital-watchdog-initiative/utils"
)

// TransactionController handles transaction-related operations
type TransactionController struct {
	transactionRepo repository.TransactionRepository
	fundRepo        repository.FundRepository
	auditRepo       repository.AuditLogRepository
	entityRepo      repository.EntityRepository
}

// NewTransactionController creates a new transaction controller
func NewTransactionController() *TransactionController {
	return &TransactionController{
		transactionRepo: repository.NewTransactionRepository(),
		fundRepo:        repository.NewFundRepository(),
		auditRepo:       repository.NewAuditLogRepository(),
		entityRepo:      repository.NewEntityRepository(),
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
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate fund exists
	_, err := tc.fundRepo.FindByID(req.FundID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid fund ID: " + req.FundID})
		return
	}

	// Validate source entity exists
	_, err = tc.entityRepo.FindByID(req.SourceID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid source entity ID: " + req.SourceID})
		return
	}

	// Validate destination entity exists
	_, err = tc.entityRepo.FindByID(req.DestinationID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid destination entity ID: " + req.DestinationID})
		return
	}

	// Convert float64 to decimal.Decimal
	amountDecimal := decimal.NewFromFloat(req.Amount)

	// Create the transaction
	transaction := &models.Transaction{
		TransactionType:  req.TransactionType,
		Amount:           amountDecimal,
		Currency:         req.Currency,
		Status:           models.TransactionPending,
		Description:      req.Description,
		SourceID:         req.SourceID,
		DestinationID:    req.DestinationID,
		FundID:           req.FundID,
		BudgetLineItemID: req.BudgetLineItemID,
		CreatedByID:      &userIDStr,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	if err := tc.transactionRepo.Create(transaction); err != nil {
		utils.Logger.Error("Failed to create transaction:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create transaction: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":     "Transaction created successfully",
		"transaction": transaction,
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

	transactions, total, err := tc.transactionRepo.GetAll(page, limit, filter)
	if err != nil {
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

	transaction, err := tc.transactionRepo.GetByID(id)
	if err != nil {
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

	// Check if the user has the required role
	if userRole != models.RoleAdmin && userRole != models.RoleAuditor && userRole != models.RoleFinanceOfficer && userRole != models.RoleManager {
		c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions to approve transaction"})
		return
	}

	userID, _ := c.Get("userID")
	userIDStr := userID.(string)

	// First, verify that the transaction exists and is in a state that can be approved
	transaction, err := tc.transactionRepo.GetByID(id)
	if err != nil {
		if err == utils.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Transaction not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve transaction"})
		}
		return
	}

	// Check if transaction is in a state that can be approved
	if transaction.Status != models.TransactionPending && transaction.Status != models.TransactionFlagged {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Transaction cannot be approved in its current state"})
		return
	}

	// Use the repository's Approve method to approve the transaction
	if err := tc.transactionRepo.Approve(id, userIDStr); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update transaction"})
		return
	}

	// Get the updated transaction to return in the response
	updatedTransaction, err := tc.transactionRepo.GetByID(id)
	if err != nil {
		utils.Logger.Warn("Transaction approved but failed to retrieve updated transaction")
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "Transaction approved successfully",
		"transaction": updatedTransaction,
	})
}

// Reject rejects a transaction
func (tc *TransactionController) Reject(c *gin.Context) {
	id := c.Param("id")
	userRole, _ := c.Get("userRole")

	// Check if the user has the required role
	if userRole != models.RoleAdmin && userRole != models.RoleAuditor && userRole != models.RoleFinanceOfficer && userRole != models.RoleManager {
		c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions to reject transaction"})
		return
	}

	userID, _ := c.Get("userID")
	userIDStr := userID.(string)

	var req struct {
		Reason string `json:"reason" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	transaction, err := tc.transactionRepo.GetByID(id)
	if err != nil {
		if err == utils.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Transaction not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve transaction"})
		}
		return
	}

	// Check if transaction is in a state that can be rejected
	if transaction.Status != models.TransactionPending && transaction.Status != models.TransactionFlagged {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Transaction cannot be rejected in its current state"})
		return
	}

	// Use the repository's Reject method
	if err := tc.transactionRepo.Reject(id, userIDStr, req.Reason); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update transaction"})
		return
	}

	// Get the updated transaction to return in the response
	updatedTransaction, err := tc.transactionRepo.GetByID(id)
	if err != nil {
		utils.Logger.Warn("Transaction rejected but failed to retrieve updated transaction")
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "Transaction rejected successfully",
		"transaction": updatedTransaction,
	})
}

// Complete marks a transaction as completed
func (tc *TransactionController) Complete(c *gin.Context) {
	id := c.Param("id")
	userRole, _ := c.Get("userRole")

	// Check if the user has the required role
	if userRole != models.RoleAdmin && userRole != models.RoleAuditor && userRole != models.RoleFinanceOfficer && userRole != models.RoleManager {
		c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions to complete transaction"})
		return
	}

	userID, _ := c.Get("userID")
	userIDStr := userID.(string)

	transaction, err := tc.transactionRepo.GetByID(id)
	if err != nil {
		if err == utils.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Transaction not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve transaction"})
		}
		return
	}

	// Check if transaction is in a state that can be completed
	if transaction.Status != models.TransactionApproved {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Transaction must be approved before it can be completed"})
		return
	}

	// Use the repository's Complete method
	if err := tc.transactionRepo.Complete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to complete transaction"})
		return
	}

	// Get the updated transaction to return in the response
	updatedTransaction, err := tc.transactionRepo.GetByID(id)
	if err != nil {
		utils.Logger.Warn("Transaction completed but failed to retrieve updated transaction")
	}

	// Create audit log
	auditLog := &models.AuditLog{
		UserID:     userIDStr,
		Action:     "TRANSACTION_COMPLETED",
		EntityType: "Transaction",
		EntityID:   id,
		Detail:     "Transaction marked as completed",
		IP:         c.ClientIP(),
		Timestamp:  time.Now(),
	}

	if err := tc.auditRepo.Create(auditLog); err != nil {
		utils.Logger.Warn("Failed to create audit log for transaction completion")
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "Transaction completed successfully",
		"transaction": updatedTransaction,
	})
}

// RegisterRoutes registers routes for the transaction controller
func (tc *TransactionController) RegisterRoutes(rg *gin.RouterGroup) {
	transactions := rg.Group("/transactions")

	// Public routes
	transactions.GET("", tc.GetAll) // All users can view transactions
	transactions.GET("/:id", tc.GetByID)

	// Protected routes
	protected := transactions.Use(middleware.RequireAuth())
	{
		// Finance officers can create transactions
		financeRoutes := protected.Use(middleware.RequireRole(models.RoleFinanceOfficer, models.RoleAdmin))
		{
			financeRoutes.POST("", middleware.AuditLog("TRANSACTION_CREATED"), tc.Create)
		}

		// Managers can approve transactions
		managerRoutes := protected.Use(middleware.RequireRole(models.RoleManager, models.RoleAdmin))
		{
			managerRoutes.PUT("/:id/approve", middleware.AuditLog("TRANSACTION_APPROVED"), tc.Approve)
			managerRoutes.PUT("/:id/reject", middleware.AuditLog("TRANSACTION_REJECTED"), tc.Reject)
		}

		// Finance officers can mark transactions as completed
		financeRoutes.PUT("/:id/complete", middleware.AuditLog("TRANSACTION_COMPLETED"), tc.Complete)
	}
}
