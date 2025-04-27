package handlers

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"github.com/the-digital-watchdog-initiative/models"
	"github.com/the-digital-watchdog-initiative/services"
	"github.com/the-digital-watchdog-initiative/utils"
)

// FundController handles fund-related operations
type FundController struct {
	fundService services.FundService
	auditService services.AuditService
	logger      *utils.NamedLogger
}

// NewFundController creates a new fund controller
func NewFundController(fundService services.FundService, auditService services.AuditService) *FundController {
	return &FundController{
		fundService: fundService,
		auditService: auditService,
		logger:      utils.NewLogger("fund-handler"),
	}
}

// Create creates a new fund
func (fc *FundController) Create(c *gin.Context) {
	var fund models.Fund

	if err := c.ShouldBindJSON(&fund); err != nil {
		fc.logger.Warn("Invalid fund creation request", map[string]interface{}{
			"error": err.Error(),
		})
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
	userRole, _ := c.Get("userRole")
	userIDStr := userID.(string)

	fc.logger.Info("Processing fund creation request", map[string]interface{}{
		"userID":   userIDStr,
		"userRole": userRole,
		"name":     fund.Name,
		"entityID": fund.EntityID,
	})

	// Only admins can create funds
	if userRole != models.RoleAdmin {
		fc.logger.Warn("Insufficient permissions to create fund", map[string]interface{}{
			"userID":   userIDStr,
			"userRole": userRole,
		})
		c.JSON(http.StatusForbidden, gin.H{"error": "Only administrators can create funds"})
		return
	}

	// Set created by ID
	fund.CreatedByID = userIDStr
	fund.CreatedAt = time.Now()
	fund.UpdatedAt = time.Now()

	// Create the fund using the service
	err := fc.fundService.CreateFund(&fund)
	if err != nil {
		fc.logger.Error("Failed to create fund", map[string]interface{}{
			"error":    err.Error(),
			"userID":   userIDStr,
			"name":     fund.Name,
			"entityID": fund.EntityID,
		})

		// Check for specific error types
		switch err := err.(type) {
		case *utils.ValidationError:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		case *utils.ConflictError:
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create fund"})
			return
		}
	}

	// Log the fund creation activity
	if err := fc.auditService.LogActivity(userIDStr, "FUND_CREATED", "Fund", fund.ID, map[string]interface{}{
		"name": fund.Name,
	}); err != nil {
		utils.Logger.Warn("Failed to create audit log for fund creation")
	}

	c.JSON(http.StatusCreated, fund)
}

// GetAll retrieves all funds with pagination
func (fc *FundController) GetAll(c *gin.Context) {
	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	// Get filter parameters
	entityID := c.Query("entityId")
	isActive := c.Query("isActive")
	search := c.Query("search")

	// Build filter map
	filter := make(map[string]interface{})
	if entityID != "" {
		filter["entity_id"] = entityID
	}
	if isActive != "" {
		active := isActive == "true"
		filter["is_active"] = active
	}
	if search != "" {
		// For simple search, we'll just look for the term in the name
		filter["name_contains"] = search
	}

	fc.logger.Info("Retrieving funds", map[string]interface{}{
		"page":    page,
		"limit":   limit,
		"filters": filter,
	})

	// Get funds using the service
	var funds []models.Fund
	var total int64
	var err error

	if entityID != "" {
		// If entity ID is provided, use the specific method
		funds, total, err = fc.fundService.GetFundsByEntityID(entityID, page, limit)
	} else {
		// Otherwise use the general method with filters
		funds, total, err = fc.fundService.GetFunds(page, limit, filter)
	}

	if err != nil {
		fc.logger.Error("Failed to retrieve funds", map[string]interface{}{
			"error":  err.Error(),
			"page":   page,
			"limit":  limit,
			"filter": filter,
		})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve funds"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"funds":  funds,
		"total":  total,
		"page":   page,
		"limit":  limit,
		"filter": filter,
	})
}

// GetByID retrieves a fund by its ID
func (fc *FundController) GetByID(c *gin.Context) {
	id := c.Param("id")

	fc.logger.Info("Retrieving fund by ID", map[string]interface{}{
		"fundID": id,
	})

	fund, err := fc.fundService.GetFundByID(id)
	if err != nil {
		fc.logger.Error("Failed to retrieve fund", map[string]interface{}{
			"error":  err.Error(),
			"fundID": id,
		})

		if err == utils.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Fund not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve fund"})
		}
		return
	}

	c.JSON(http.StatusOK, fund)
}

// Update updates a fund
func (fc *FundController) Update(c *gin.Context) {
	id := c.Param("id")
	userID, _ := c.Get("userID")
	userRole, _ := c.Get("userRole")
	userIDStr := userID.(string)

	fc.logger.Info("Processing fund update request", map[string]interface{}{
		"fundID":   id,
		"userID":   userIDStr,
		"userRole": userRole,
	})

	// Only admins can update funds
	if userRole != models.RoleAdmin {
		fc.logger.Warn("Insufficient permissions to update fund", map[string]interface{}{
			"userID":   userIDStr,
			"userRole": userRole,
		})
		c.JSON(http.StatusForbidden, gin.H{"error": "Only administrators can update funds"})
		return
	}

	// Parse request body
	var updateData struct {
		Name        *string  `json:"name"`
		Description *string  `json:"description"`
		Amount      *float64 `json:"amount"`
		Currency    *string  `json:"currency"`
		StartDate   *string  `json:"startDate"`
		EndDate     *string  `json:"endDate"`
		IsActive    *bool    `json:"isActive"`
	}

	if err := c.ShouldBindJSON(&updateData); err != nil {
		fc.logger.Warn("Invalid fund update request", map[string]interface{}{
			"error":  err.Error(),
			"fundID": id,
		})
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get the existing fund
	existingFund, err := fc.fundService.GetFundByID(id)
	if err != nil {
		fc.logger.Error("Failed to retrieve fund for update", map[string]interface{}{
			"error":  err.Error(),
			"fundID": id,
		})

		if err == utils.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Fund not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve fund"})
		}
		return
	}

	// Update fields if provided
	if updateData.Name != nil {
		existingFund.Name = *updateData.Name
	}

	if updateData.Description != nil {
		existingFund.Description = *updateData.Description
	}

	if updateData.Amount != nil {
		// Convert float64 to decimal.Decimal
		decimalAmount := decimal.NewFromFloat(*updateData.Amount)
		existingFund.Amount = decimalAmount
	}

	if updateData.Currency != nil {
		existingFund.Currency = *updateData.Currency
	}

	// Only include these fields if they exist in your Fund model
	// If they don't exist, remove these blocks

	// Update the fund using the service
	err = fc.fundService.UpdateFund(existingFund)
	if err != nil {
		fc.logger.Error("Failed to update fund", map[string]interface{}{
			"error":  err.Error(),
			"fundID": id,
			"userID": userIDStr,
		})

		switch err := err.(type) {
		case *utils.ValidationError:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		case *utils.ConflictError:
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update fund"})
			return
		}
	}

	// Create audit log
	var changes []string
	if updateData.Name != nil {
		changes = append(changes, "name updated to '"+*updateData.Name+"'")
	}

	detail := "Updated fund"
	if len(changes) > 0 {
		detail = "Updated fund: " + strings.Join(changes, ", ")
	}

	if err := fc.auditService.LogActivity(userIDStr, "FUND_UPDATED", "Fund", existingFund.ID, map[string]interface{}{
		"changes": changes,
		"detail":  detail,
	}); err != nil {
		utils.Logger.Warn("Failed to create audit log for fund update")
	}

	c.JSON(http.StatusOK, existingFund)
}

// Delete deletes a fund
func (fc *FundController) Delete(c *gin.Context) {
	id := c.Param("id")
	userID, _ := c.Get("userID")
	userRole, _ := c.Get("userRole")
	userIDStr := userID.(string)

	fc.logger.Info("Processing fund deletion request", map[string]interface{}{
		"fundID":   id,
		"userID":   userIDStr,
		"userRole": userRole,
	})

	// Only admins can delete funds
	if userRole != models.RoleAdmin {
		fc.logger.Warn("Insufficient permissions to delete fund", map[string]interface{}{
			"userID":   userIDStr,
			"userRole": userRole,
		})
		c.JSON(http.StatusForbidden, gin.H{"error": "Only administrators can delete funds"})
		return
	}

	// Delete the fund using the service
	err := fc.fundService.DeleteFund(id)
	if err != nil {
		fc.logger.Error("Failed to delete fund", map[string]interface{}{
			"error":  err.Error(),
			"fundID": id,
			"userID": userIDStr,
		})

		if err == utils.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Fund not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete fund"})
		}
		return
	}

	// Log the deletion activity
	if err := fc.auditService.LogActivity(userIDStr, "FUND_DELETED", "Fund", id, map[string]interface{}{
		"detail": "Deleted fund",
	}); err != nil {
		utils.Logger.Warn("Failed to create audit log for fund deletion")
	}

	c.JSON(http.StatusOK, gin.H{"message": "Fund deleted successfully"})
}
