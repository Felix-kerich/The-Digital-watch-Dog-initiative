package controllers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/the-digital-watchdog-initiative/models"
	"github.com/the-digital-watchdog-initiative/repository"
	"github.com/the-digital-watchdog-initiative/utils"
)

// EntityController handles entity-related operations
type EntityController struct {
	entityRepo repository.EntityRepository
	auditRepo  repository.AuditLogRepository
}

// NewEntityController creates a new entity controller
func NewEntityController() *EntityController {
	return &EntityController{
		entityRepo: repository.NewEntityRepository(),
		auditRepo:  repository.NewAuditLogRepository(),
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
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, _ := c.Get("userID")
	userRole, _ := c.Get("userRole")

	// Only admins can create entities
	if userRole != models.RoleAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions to create entity"})
		return
	}

	// Create the entity
	entity := models.Entity{
		Name:        req.Name,
		Description: req.Description,
		Type:        req.Type,
		Location:    req.Location,
		ContactInfo: req.ContactInfo,
		IsActive:    true,
		CreatedByID: userID.(string),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Save to database
	if err := ec.entityRepo.Create(&entity); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create entity"})
		return
	}

	// Create audit log
	auditLog := models.AuditLog{
		Action:     "ENTITY_CREATED",
		UserID:     userID.(string),
		EntityID:   entity.ID,
		EntityType: "Entity",
		Detail:     "Created entity: " + entity.Name,
		IP:         c.ClientIP(),
		Timestamp:  time.Now(),
	}

	if err := ec.auditRepo.Create(&auditLog); err != nil {
		utils.Logger.Warn("Failed to create audit log for entity creation")
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
	entityType := c.Query("type")
	location := c.Query("location")
	isActive := c.Query("isActive")
	search := c.Query("search")

	// Build filter map
	filter := make(map[string]interface{})
	if entityType != "" {
		filter["type"] = entityType
	}
	if location != "" {
		filter["location"] = location
	}
	if isActive != "" {
		active := isActive == "true"
		filter["is_active"] = active
	}
	if search != "" {
		// Complex filter like LIKE queries need special handling
		// For now, we'll skip this and handle in the repository implementation if needed
	}

	// Get entities with pagination and filters
	entities, total, err := ec.entityRepo.List(page, limit, filter)
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

	entity, err := ec.entityRepo.FindByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Entity not found"})
		return
	}

	c.JSON(http.StatusOK, entity)
}

// Update updates an entity
func (ec *EntityController) Update(c *gin.Context) {
	id := c.Param("id")

	entity, err := ec.entityRepo.FindByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Entity not found"})
		return
	}

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Type        string `json:"type"`
		Location    string `json:"location"`
		ContactInfo string `json:"contactInfo"`
		IsActive    *bool  `json:"isActive"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, _ := c.Get("userID")
	userRole, _ := c.Get("userRole")

	// Verify user has permission to update this entity
	if userRole != models.RoleAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions to update entity"})
		return
	}

	// Update fields if provided
	if req.Name != "" {
		entity.Name = req.Name
	}
	if req.Description != "" {
		entity.Description = req.Description
	}
	if req.Type != "" {
		entity.Type = req.Type
	}
	if req.Location != "" {
		entity.Location = req.Location
	}
	if req.ContactInfo != "" {
		entity.ContactInfo = req.ContactInfo
	}
	if req.IsActive != nil {
		entity.IsActive = *req.IsActive
	}

	entity.UpdatedAt = time.Now()

	// Save to database
	if err := ec.entityRepo.Update(entity); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update entity"})
		return
	}

	// Create audit log
	auditLog := models.AuditLog{
		Action:     "ENTITY_UPDATED",
		UserID:     userID.(string),
		EntityID:   entity.ID,
		EntityType: "Entity",
		Detail:     "Updated entity: " + entity.Name,
		IP:         c.ClientIP(),
		Timestamp:  time.Now(),
	}

	if err := ec.auditRepo.Create(&auditLog); err != nil {
		utils.Logger.Warn("Failed to create audit log for entity update")
	}

	c.JSON(http.StatusOK, entity)
}

// Delete deletes an entity (soft delete by marking as inactive)
func (ec *EntityController) Delete(c *gin.Context) {
	id := c.Param("id")

	entity, err := ec.entityRepo.FindByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Entity not found"})
		return
	}

	userID, _ := c.Get("userID")
	userRole, _ := c.Get("userRole")

	// Verify user has permission to delete this entity
	if userRole != models.RoleAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions to delete entity"})
		return
	}

	// Soft delete by marking as inactive
	if err := ec.entityRepo.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete entity"})
		return
	}

	// Create audit log
	auditLog := models.AuditLog{
		Action:     "ENTITY_DELETED",
		UserID:     userID.(string),
		EntityID:   entity.ID,
		EntityType: "Entity",
		Detail:     "Deleted entity: " + entity.Name,
		IP:         c.ClientIP(),
		Timestamp:  time.Now(),
	}

	if err := ec.auditRepo.Create(&auditLog); err != nil {
		utils.Logger.Warn("Failed to create audit log for entity deletion")
	}

	c.JSON(http.StatusOK, gin.H{"message": "Entity deleted successfully"})
}
