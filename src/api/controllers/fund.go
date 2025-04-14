package controllers

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/the-digital-watchdog-initiative/models"
	"github.com/the-digital-watchdog-initiative/repository"
	"github.com/the-digital-watchdog-initiative/utils"
)

// FundController handles fund-related operations
type FundController struct {
	fundRepo  repository.FundRepository
	auditRepo repository.AuditLogRepository
}

// NewFundController creates a new fund controller
func NewFundController() *FundController {
	return &FundController{
		fundRepo:  repository.NewFundRepository(),
		auditRepo: repository.NewAuditLogRepository(),
	}
}

// Create creates a new fund
func (fc *FundController) Create(c *gin.Context) {
	var fund models.Fund

	if err := c.ShouldBindJSON(&fund); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate fund
	if fund.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Fund name is required"})
		return
	}

	if fund.EntityID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Entity ID is required"})
		return
	}

	// Get user ID from context
	userID, _ := c.Get("userID")

	fund.CreatedByID = userID.(string)
	fund.CreatedAt = time.Now()
	fund.UpdatedAt = time.Now()

	if err := fc.fundRepo.Create(&fund); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create fund"})
		return
	}

	// Create audit log
	auditLog := models.AuditLog{
		Action:     "FUND_CREATED",
		UserID:     userID.(string),
		EntityID:   fund.ID,
		EntityType: "Fund",
		Detail:     "Created fund: " + fund.Name,
		IP:         c.ClientIP(),
		Timestamp:  time.Now(),
	}

	if err := fc.auditRepo.Create(&auditLog); err != nil {
		utils.Logger.Warn("Failed to create audit log for fund creation")
	}

	c.JSON(http.StatusCreated, fund)
}

// GetAll retrieves all funds with pagination
func (fc *FundController) GetAll(c *gin.Context) {
	// Parse pagination parameters
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "10")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		limit = 10
	}

	// Parse filters
	filters := make(map[string]interface{})

	// Filter by entity ID if provided
	if entityID := c.Query("entityId"); entityID != "" {
		filters["entity_id"] = entityID
	}

	// Filter by category if provided
	if category := c.Query("category"); category != "" {
		filters["category"] = category
	}

	// Filter by status if provided
	if status := c.Query("status"); status != "" {
		filters["status"] = status
	}

	// Filter by name (partial match) if provided
	if name := c.Query("name"); name != "" {
		filters["name_like"] = name
	}

	// Get funds with pagination
	funds, total, err := fc.fundRepo.List(page, limit, filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve funds"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"funds": funds,
		"pagination": gin.H{
			"total": total,
			"page":  page,
			"limit": limit,
		},
	})
}

// GetByID retrieves a fund by its ID
func (fc *FundController) GetByID(c *gin.Context) {
	id := c.Param("id")

	fund, err := fc.fundRepo.FindByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Fund not found"})
		return
	}

	c.JSON(http.StatusOK, fund)
}

// Update updates a fund
func (fc *FundController) Update(c *gin.Context) {
	id := c.Param("id")

	fund, err := fc.fundRepo.FindByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Fund not found"})
		return
	}

	// Check authorization
	userID, _ := c.Get("userID")
	userRole, _ := c.Get("userRole")
	if userRole != models.RoleAdmin && fund.CreatedByID != userID.(string) {
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to update this fund"})
		return
	}

	// Store the old name for audit log
	oldName := fund.Name

	// Bind updated fund data
	var updatedFund models.Fund
	if err := c.ShouldBindJSON(&updatedFund); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update fields
	fund.Name = updatedFund.Name
	fund.Description = updatedFund.Description
	fund.Category = updatedFund.Category
	fund.Status = updatedFund.Status
	fund.Amount = updatedFund.Amount
	fund.Currency = updatedFund.Currency
	fund.UpdatedAt = time.Now()

	if err := fc.fundRepo.Update(fund); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update fund"})
		return
	}

	// Create audit log
	changes := []string{}
	if oldName != fund.Name {
		changes = append(changes, "name changed from '"+oldName+"' to '"+fund.Name+"'")
	}

	auditLog := models.AuditLog{
		Action:     "FUND_UPDATED",
		UserID:     userID.(string),
		EntityID:   fund.ID,
		EntityType: "Fund",
		Detail:     "Updated fund: " + strings.Join(changes, ", "),
		IP:         c.ClientIP(),
		Timestamp:  time.Now(),
	}

	if err := fc.auditRepo.Create(&auditLog); err != nil {
		utils.Logger.Warn("Failed to create audit log for fund update")
	}

	c.JSON(http.StatusOK, fund)
}

// Delete deletes a fund
func (fc *FundController) Delete(c *gin.Context) {
	id := c.Param("id")

	fund, err := fc.fundRepo.FindByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Fund not found"})
		return
	}

	// Check authorization
	userID, _ := c.Get("userID")
	userRole, _ := c.Get("userRole")
	if userRole != models.RoleAdmin && fund.CreatedByID != userID.(string) {
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to delete this fund"})
		return
	}

	// Delete the fund
	if err := fc.fundRepo.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete fund"})
		return
	}

	// Create audit log
	auditLog := models.AuditLog{
		Action:     "FUND_DELETED",
		UserID:     userID.(string),
		EntityID:   fund.ID,
		EntityType: "Fund",
		Detail:     "Deleted fund: " + fund.Name,
		IP:         c.ClientIP(),
		Timestamp:  time.Now(),
	}

	if err := fc.auditRepo.Create(&auditLog); err != nil {
		utils.Logger.Warn("Failed to create audit log for fund deletion")
	}

	c.JSON(http.StatusOK, gin.H{"message": "Fund deleted successfully"})
}
