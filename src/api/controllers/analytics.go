package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/the-digital-watchdog-initiative/repository"
)

// AnalyticsController handles analytics-related operations
type AnalyticsController struct {
	analyticsRepo repository.AnalyticsRepository
}

// NewAnalyticsController creates a new analytics controller
func NewAnalyticsController() *AnalyticsController {
	return &AnalyticsController{
		analyticsRepo: repository.NewAnalyticsRepository(),
	}
}

// GetTransactionSummary returns a summary of transactions
func (ac *AnalyticsController) GetTransactionSummary(c *gin.Context) {
	// Get filter parameters
	filter := map[string]interface{}{
		"entityId": c.Query("entityId"),
		"year":     c.Query("year"),
		"type":     c.Query("type"),
	}

	summary, err := ac.analyticsRepo.GetTransactionSummary(filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve transaction summary"})
		return
	}

	c.JSON(http.StatusOK, summary)
}

// GetUserActivitySummary returns a summary of user activities
func (ac *AnalyticsController) GetUserActivitySummary(c *gin.Context) {
	// Get filter parameters
	filter := map[string]interface{}{
		"entityId": c.Query("entityId"),
		"fromDate": c.Query("fromDate"),
		"toDate":   c.Query("toDate"),
		"userId":   c.Query("userId"),
	}

	summary, err := ac.analyticsRepo.GetUserActivitySummary(filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user activity summary"})
		return
	}

	c.JSON(http.StatusOK, summary)
}

// GetFundUtilizationReport returns a report of fund utilization
func (ac *AnalyticsController) GetFundUtilizationReport(c *gin.Context) {
	fundID := c.Query("fundId")
	entityID := c.Query("entityId")
	fiscalYear := c.Query("fiscalYear")

	if fundID == "" && entityID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Either fundId or entityId must be provided"})
		return
	}

	report, err := ac.analyticsRepo.GetFundUtilizationReport(fundID, entityID, fiscalYear)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve fund utilization report"})
		return
	}

	c.JSON(http.StatusOK, report)
}
