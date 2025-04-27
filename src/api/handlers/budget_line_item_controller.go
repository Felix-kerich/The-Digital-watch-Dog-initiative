package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"github.com/the-digital-watchdog-initiative/models"
	"github.com/the-digital-watchdog-initiative/services"
	"github.com/the-digital-watchdog-initiative/utils"
)

// BudgetLineItemController handles budget line item operations
type BudgetLineItemController struct {
	budgetService services.BudgetLineItemService
	auditService  services.AuditService
	logger        *logrus.Logger
}

// CreateBudgetLineItemRequest represents the request to create a new budget line item
type CreateBudgetLineItemRequest struct {
	Name        string  `json:"name" binding:"required"`
	Description string  `json:"description"`
	Code        string  `json:"code" binding:"required"`
	Amount      float64 `json:"amount" binding:"required"`
	FundID      string  `json:"fundId" binding:"required"`
}

// NewBudgetLineItemController creates a new budget line item controller
func NewBudgetLineItemController(budgetService services.BudgetLineItemService, auditService services.AuditService) *BudgetLineItemController {
	return &BudgetLineItemController{
		budgetService: budgetService,
		auditService:  auditService,
		logger:        utils.Logger,
	}
}

// Create creates a new budget line item
func (bc *BudgetLineItemController) Create(c *gin.Context) {
	bc.logger.Info("Handling create budget line item request")
	
	var req CreateBudgetLineItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		bc.logger.WithError(err).Error("Invalid request data")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Extract user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		bc.logger.Error("User ID not found in context")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userIDStr := userID.(string)

	// Create budget line item
	amount, _ := decimal.NewFromFloat(req.Amount).Round(2).Float64()
	budgetLineItem := models.BudgetLineItem{
		Name:        req.Name,
		Description: req.Description,
		Code:        req.Code,
		Amount:      decimal.NewFromFloat(amount),
		FundID:      req.FundID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := bc.budgetService.CreateBudgetLineItem(&budgetLineItem); err != nil {
		bc.logger.WithError(err).Error("Failed to create budget line item")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create budget line item"})
		return
	}

	// Log the activity
	if err := bc.auditService.LogActivity(userIDStr, "BUDGET_LINE_ITEM_CREATED", "BudgetLineItem", budgetLineItem.ID, map[string]interface{}{
		"name":   budgetLineItem.Name,
		"code":   budgetLineItem.Code,
		"amount": budgetLineItem.Amount,
		"fundID": budgetLineItem.FundID,
	}); err != nil {
		// Log error but continue
		bc.logger.WithError(err).Error("Failed to log budget line item creation activity")
	}

	bc.logger.WithField("id", budgetLineItem.ID).Info("Budget line item created successfully")
	c.JSON(http.StatusCreated, gin.H{"data": budgetLineItem, "message": "Budget line item created successfully"})
}

// GetAll retrieves all budget line items with pagination and filtering
func (bc *BudgetLineItemController) GetAll(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	fundID := c.Query("fundId")

	filter := make(map[string]interface{})
	if fundID != "" {
		filter["fund_id"] = fundID
	}

	budgetLineItems, total, err := bc.budgetService.GetBudgetLineItems(page, limit, filter)
	if err != nil {
		bc.logger.WithError(err).Error("Failed to retrieve budget line items")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve budget line items"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items": budgetLineItems,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

// GetByID retrieves a budget line item by ID
func (bc *BudgetLineItemController) GetByID(c *gin.Context) {
	id := c.Param("id")
	bc.logger.WithField("id", id).Info("Handling get budget line item by ID request")
	
	budgetLineItem, err := bc.budgetService.GetBudgetLineItemByID(id)
	if err != nil {
		bc.logger.WithError(err).WithField("id", id).Error("Budget line item not found")
		c.JSON(http.StatusNotFound, gin.H{"error": "Budget line item not found"})
		return
	}

	bc.logger.WithField("id", id).Info("Budget line item retrieved successfully")
	c.JSON(http.StatusOK, gin.H{"data": budgetLineItem})
}

// Update updates a budget line item
func (bc *BudgetLineItemController) Update(c *gin.Context) {
	id := c.Param("id")
	bc.logger.WithField("id", id).Info("Handling update budget line item request")

	// Check if budget line item exists
	existingItem, err := bc.budgetService.GetBudgetLineItemByID(id)
	if err != nil {
		bc.logger.WithError(err).WithField("id", id).Error("Budget line item not found")
		c.JSON(http.StatusNotFound, gin.H{"error": "Budget line item not found"})
		return
	}

	var req CreateBudgetLineItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		bc.logger.WithError(err).Error("Invalid request data")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Extract user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		bc.logger.Error("User ID not found in context")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userIDStr := userID.(string)

	// Update budget line item
	amount, _ := decimal.NewFromFloat(req.Amount).Round(2).Float64()
	existingItem.Name = req.Name
	existingItem.Description = req.Description
	existingItem.Code = req.Code
	existingItem.Amount = decimal.NewFromFloat(amount)
	existingItem.FundID = req.FundID
	existingItem.UpdatedAt = time.Now()

	if err := bc.budgetService.UpdateBudgetLineItem(existingItem); err != nil {
		bc.logger.WithError(err).Error("Failed to update budget line item")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update budget line item"})
		return
	}

	// Log the activity
	if err := bc.auditService.LogActivity(userIDStr, "BUDGET_LINE_ITEM_UPDATED", "BudgetLineItem", existingItem.ID, map[string]interface{}{
		"name":   existingItem.Name,
		"code":   existingItem.Code,
		"amount": existingItem.Amount,
		"fundID": existingItem.FundID,
	}); err != nil {
		// Log error but continue
		bc.logger.WithError(err).Error("Failed to log budget line item update activity")
	}

	bc.logger.WithField("id", id).Info("Budget line item updated successfully")
	c.JSON(http.StatusOK, gin.H{"data": existingItem, "message": "Budget line item updated successfully"})
}

// Delete deletes a budget line item
func (bc *BudgetLineItemController) Delete(c *gin.Context) {
	id := c.Param("id")
	bc.logger.WithField("id", id).Info("Handling delete budget line item request")

	// Check if budget line item exists and get its name for the audit log
	existingItem, err := bc.budgetService.GetBudgetLineItemByID(id)
	if err != nil {
		bc.logger.WithError(err).WithField("id", id).Error("Budget line item not found")
		c.JSON(http.StatusNotFound, gin.H{"error": "Budget line item not found"})
		return
	}

	// Extract user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		bc.logger.Error("User ID not found in context")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userIDStr := userID.(string)

	// Store item name for audit log
	itemName := existingItem.Name

	// Delete budget line item
	if err := bc.budgetService.DeleteBudgetLineItem(id); err != nil {
		bc.logger.WithError(err).Error("Failed to delete budget line item")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete budget line item"})
		return
	}

	// Log the activity
	if err := bc.auditService.LogActivity(userIDStr, "BUDGET_LINE_ITEM_DELETED", "BudgetLineItem", id, map[string]interface{}{
		"name": itemName,
	}); err != nil {
		// Log error but continue
		bc.logger.WithError(err).Error("Failed to log budget line item deletion activity")
	}

	bc.logger.WithField("id", id).Info("Budget line item deleted successfully")
	c.JSON(http.StatusOK, gin.H{"message": "Budget line item deleted successfully"})
}
