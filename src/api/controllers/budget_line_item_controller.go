package controllers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"github.com/the-digital-watchdog-initiative/models"
	"github.com/the-digital-watchdog-initiative/repository"
	"github.com/the-digital-watchdog-initiative/utils"
)

// BudgetLineItemController handles budget line item operations
type BudgetLineItemController struct {
	budgetRepo repository.BudgetLineItemRepository
	fundRepo   repository.FundRepository
	auditRepo  repository.AuditLogRepository
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
func NewBudgetLineItemController() *BudgetLineItemController {
	return &BudgetLineItemController{
		budgetRepo: repository.NewBudgetLineItemRepository(),
		fundRepo:   repository.NewFundRepository(),
		auditRepo:  repository.NewAuditLogRepository(),
	}
}

// Create creates a new budget line item
func (bc *BudgetLineItemController) Create(c *gin.Context) {
	var req CreateBudgetLineItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verify fund exists
	fund, err := bc.fundRepo.FindByID(req.FundID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid fund ID"})
		return
	}

	// Convert float64 to decimal.Decimal
	amountDecimal := decimal.NewFromFloat(req.Amount)

	// Create budget line item
	item := &models.BudgetLineItem{
		Name:        req.Name,
		Description: req.Description,
		Code:        req.Code,
		Amount:      amountDecimal,
		FundID:      req.FundID,
		CreatedByID: c.GetString("userID"),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := bc.budgetRepo.Create(item); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create budget line item"})
		return
	}

	// Create audit log
	auditLog := &models.AuditLog{
		Action:     "BUDGET_LINE_ITEM_CREATED",
		UserID:     c.GetString("userID"),
		EntityType: "BudgetLineItem",
		EntityID:   item.ID,
		Detail:     "Created budget line item: " + item.Name + " for fund: " + fund.Name,
		IP:         c.ClientIP(),
		Timestamp:  time.Now(),
	}

	if err := bc.auditRepo.Create(auditLog); err != nil {
		utils.Logger.Warn("Failed to create audit log for budget line item creation")
	}

	c.JSON(http.StatusCreated, item)
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

	items, total, err := bc.budgetRepo.FindAll(page, limit, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve budget line items"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items": items,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

// GetByID retrieves a budget line item by ID
func (bc *BudgetLineItemController) GetByID(c *gin.Context) {
	id := c.Param("id")
	item, err := bc.budgetRepo.FindByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Budget line item not found"})
		return
	}

	c.JSON(http.StatusOK, item)
}

// Update updates a budget line item
func (bc *BudgetLineItemController) Update(c *gin.Context) {
	id := c.Param("id")
	var req CreateBudgetLineItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	item, err := bc.budgetRepo.FindByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Budget line item not found"})
		return
	}

	// Convert float64 to decimal.Decimal
	amountDecimal := decimal.NewFromFloat(req.Amount)

	// Update fields
	item.Name = req.Name
	item.Description = req.Description
	item.Code = req.Code
	item.Amount = amountDecimal
	item.UpdatedAt = time.Now()

	if err := bc.budgetRepo.Update(item); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update budget line item"})
		return
	}

	// Create audit log
	auditLog := &models.AuditLog{
		Action:     "BUDGET_LINE_ITEM_UPDATED",
		UserID:     c.GetString("userID"),
		EntityType: "BudgetLineItem",
		EntityID:   item.ID,
		Detail:     "Updated budget line item: " + item.Name,
		IP:         c.ClientIP(),
		Timestamp:  time.Now(),
	}

	if err := bc.auditRepo.Create(auditLog); err != nil {
		utils.Logger.Warn("Failed to create audit log for budget line item update")
	}

	c.JSON(http.StatusOK, item)
}

// Delete deletes a budget line item
func (bc *BudgetLineItemController) Delete(c *gin.Context) {
	id := c.Param("id")

	item, err := bc.budgetRepo.FindByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Budget line item not found"})
		return
	}

	if err := bc.budgetRepo.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete budget line item"})
		return
	}

	// Create audit log
	auditLog := &models.AuditLog{
		Action:     "BUDGET_LINE_ITEM_DELETED",
		UserID:     c.GetString("userID"),
		EntityType: "BudgetLineItem",
		EntityID:   id,
		Detail:     "Deleted budget line item: " + item.Name,
		IP:         c.ClientIP(),
		Timestamp:  time.Now(),
	}

	if err := bc.auditRepo.Create(auditLog); err != nil {
		utils.Logger.Warn("Failed to create audit log for budget line item deletion")
	}

	c.JSON(http.StatusOK, gin.H{"message": "Budget line item deleted successfully"})
}
