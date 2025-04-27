package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/the-digital-watchdog-initiative/models"
	"github.com/the-digital-watchdog-initiative/services"
	"github.com/the-digital-watchdog-initiative/utils"
)

// EntityController handles entity-related operations
type EntityController struct {
	entityService services.EntityService
	auditService  services.AuditService
	logger        *utils.NamedLogger
}

// NewEntityController creates a new entity controller
func NewEntityController(entityService services.EntityService, auditService services.AuditService) *EntityController {
	return &EntityController{
		entityService: entityService,
		auditService:  auditService,
		logger:        utils.NewLogger("entity-handler"),
	}
}

// Create creates a new entity
func (ec *EntityController) Create(c *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
		Type        string `json:"type" binding:"required"`
		Location    string `json:"location"`
		ContactInfo string `json:"contactInfo"`
		Code        string `json:"code"`	
		IsGovernment bool   `json:"isGovernment"`
		IsActive    bool   `json:"isActive"`	

	}

	if err := c.ShouldBindJSON(&req); err != nil {
		ec.logger.Warn("Invalid entity creation request", map[string]interface{}{
			"error": err.Error(),
		})
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, _ := c.Get("userID")
	userRole, _ := c.Get("userRole")
	userIDStr := userID.(string)

	ec.logger.Info("Processing entity creation request", map[string]interface{}{
		"userID":   userIDStr,
		"userRole": userRole,
		"name":     req.Name,
		"type":     req.Type,
	})

	// Only admins can create entities
	if userRole != models.RoleAdmin {
		ec.logger.Warn("Insufficient permissions to create entity", map[string]interface{}{
			"userID":   userIDStr,
			"userRole": userRole,
		})
		c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions to create entity"})
		return
	}

	// Create the entity using the service
	entity := &models.Entity{
		Name:        req.Name,
		Description: req.Description,
		Type:        req.Type,
		Location:    req.Location,
		ContactInfo: req.ContactInfo,
		CreatedByID: userIDStr,
		Code:        req.Code,
		IsGovernment: req.IsGovernment,
		IsActive:    req.IsActive,
	}

	err := ec.entityService.CreateEntity(entity)
	if err != nil {
		ec.logger.Error("Failed to create entity", map[string]interface{}{
			"error":  err.Error(),
			"userID": userIDStr,
			"name":   req.Name,
		})

		// Check for specific error types
		switch err := err.(type) {
		case *utils.ValidationError:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		case *utils.ConflictError:
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create entity"})
		return
	}

	c.JSON(http.StatusCreated, entity)
}

// GetAll retrieves entities with filtering and pagination
func (ec *EntityController) GetAll(c *gin.Context) {
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
	filter := map[string]interface{}{
		"name":     c.Query("name"),
		"type":     c.Query("type"),
		"location": c.Query("location"),
		"isActive": c.Query("isActive"),
	}

	// Remove empty filters
	for k, v := range filter {
		if v == "" {
			delete(filter, k)
		}
	}

	ec.logger.Info("Retrieving entities", map[string]interface{}{
		"page":    page,
		"limit":   limit,
		"filters": filter,
	})

	// Get entities using the service
	entities, total, err := ec.entityService.GetEntities(page, limit, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve entities"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"entities":   entities,
		"total":      total,
		"page":       page,
		"limit":      limit,
		"totalPages": (int(total) + limit - 1) / limit,
	})
}

// GetByID retrieves an entity by its ID
func (ec *EntityController) GetByID(c *gin.Context) {
	id := c.Param("id")

	ec.logger.Info("Retrieving entity by ID", map[string]interface{}{
		"entityID": id,
	})

	entity, err := ec.entityService.GetEntityByID(id)
	if err != nil {
		ec.logger.Error("Failed to retrieve entity", map[string]interface{}{
			"error":    err.Error(),
			"entityID": id,
		})

		if err == utils.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Entity not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve entity"})
		}
		return
	}

	c.JSON(http.StatusOK, entity)
}

// Update updates an entity
func (ec *EntityController) Update(c *gin.Context) {
	id := c.Param("id")
	userID, _ := c.Get("userID")
	userRole, _ := c.Get("userRole")
	userIDStr := userID.(string)

	ec.logger.Info("Processing entity update request", map[string]interface{}{
		"entityID":  id,
		"userID":    userIDStr,
		"userRole":  userRole,
	})

	// Only admins can update entities
	if userRole != models.RoleAdmin {
		ec.logger.Warn("Insufficient permissions to update entity", map[string]interface{}{
			"userID":   userIDStr,
			"userRole": userRole,
		})
		c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions to update entity"})
		return
	}

	// Parse request body
	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Type        string `json:"type"`
		Location    string `json:"location"`
		ContactInfo string `json:"contactInfo"`
		IsActive    *bool  `json:"isActive"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		ec.logger.Warn("Invalid entity update request", map[string]interface{}{
			"error":    err.Error(),
			"entityID": id,
		})
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Create update data map
	updateData := map[string]interface{}{}
	if req.Name != "" {
		updateData["name"] = req.Name
	}
	if req.Description != "" {
		updateData["description"] = req.Description
	}
	if req.Type != "" {
		updateData["type"] = req.Type
	}
	if req.Location != "" {
		updateData["location"] = req.Location
	}
	if req.ContactInfo != "" {
		updateData["contactInfo"] = req.ContactInfo
	}
	if req.IsActive != nil {
		updateData["isActive"] = *req.IsActive
	}

	// Update the entity using the service
	updatedEntity, err := ec.entityService.UpdateEntity(id, updateData)
	if err != nil {
		ec.logger.Error("Failed to update entity", map[string]interface{}{
			"error":    err.Error(),
			"entityID": id,
			"userID":   userIDStr,
		})

		// Check for specific error types
		if err == utils.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Entity not found"})
			return
		}
		
		switch err := err.(type) {
		case *utils.ValidationError:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		case *utils.ConflictError:
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update entity"})
		}
		return
	}

	c.JSON(http.StatusOK, updatedEntity)
}

// Delete deletes an entity (soft delete by marking as inactive)
func (ec *EntityController) Delete(c *gin.Context) {
	id := c.Param("id")
	userID, _ := c.Get("userID")
	userRole, _ := c.Get("userRole")
	userIDStr := userID.(string)

	ec.logger.Info("Processing entity deletion request", map[string]interface{}{
		"entityID":  id,
		"userID":    userIDStr,
		"userRole":  userRole,
	})

	// Only admins can delete entities
	if userRole != models.RoleAdmin {
		ec.logger.Warn("Insufficient permissions to delete entity", map[string]interface{}{
			"userID":   userIDStr,
			"userRole": userRole,
		})
		c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions to delete entity"})
		return
	}

	// Delete the entity using the service
	err := ec.entityService.DeleteEntity(id)
	if err != nil {
		ec.logger.Error("Failed to delete entity", map[string]interface{}{
			"error":    err.Error(),
			"entityID": id,
			"userID":   userIDStr,
		})

		// Check for specific error types
		if err == utils.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Entity not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete entity"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Entity deleted successfully"})
}
