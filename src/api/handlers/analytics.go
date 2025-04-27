package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/the-digital-watchdog-initiative/services"
	"github.com/the-digital-watchdog-initiative/utils"
)

// AnalyticsController handles analytics-related operations
type AnalyticsController struct {
	analyticsService services.AnalyticsService
	auditService     services.AuditService
	logger           *utils.NamedLogger
}

// NewAnalyticsController creates a new analytics controller
func NewAnalyticsController(analyticsService services.AnalyticsService, auditService services.AuditService) *AnalyticsController {
	return &AnalyticsController{
		analyticsService: analyticsService,
		auditService:     auditService,
		logger:           utils.NewLogger("analytics-controller"),
	}
}

// GetTransactionSummary returns a summary of transactions
func (ac *AnalyticsController) GetTransactionSummary(c *gin.Context) {
	ac.logger.Info("Getting transaction summary", nil)

	// Get filter parameters
	filter := map[string]interface{}{
		"entityId": c.Query("entityId"),
		"year":     c.Query("year"),
		"type":     c.Query("type"),
	}

	summary, err := ac.analyticsService.GetTransactionSummary(filter)
	if err != nil {
		ac.logger.Error("Failed to retrieve transaction summary", map[string]interface{}{"error": err.Error()})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve transaction summary"})
		return
	}

	// Log the activity
	userID, exists := c.Get("userID")
	if exists {
		metadata := map[string]interface{}{
			"filter": filter,
		}

		if err := ac.auditService.LogActivity(
			userID.(string),
			"ANALYTICS_TRANSACTION_SUMMARY_VIEWED",
			"Analytics",
			"",
			metadata,
		); err != nil {
			ac.logger.Warn("Failed to log analytics activity", map[string]interface{}{"error": err.Error()})
		}
	}

	c.JSON(http.StatusOK, summary)
}

// GetUserActivitySummary returns a summary of user activities
func (ac *AnalyticsController) GetUserActivitySummary(c *gin.Context) {
	ac.logger.Info("Getting user activity summary", nil)

	// Get filter parameters
	filter := map[string]interface{}{
		"entityId": c.Query("entityId"),
		"fromDate": c.Query("fromDate"),
		"toDate":   c.Query("toDate"),
		"userId":   c.Query("userId"),
	}

	summary, err := ac.analyticsService.GetUserActivitySummary(filter)
	if err != nil {
		ac.logger.Error("Failed to retrieve user activity summary", map[string]interface{}{"error": err.Error()})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user activity summary"})
		return
	}

	// Log the activity
	userID, exists := c.Get("userID")
	if exists {
		metadata := map[string]interface{}{
			"filter": filter,
		}

		if err := ac.auditService.LogActivity(
			userID.(string),
			"ANALYTICS_USER_ACTIVITY_VIEWED",
			"Analytics",
			"",
			metadata,
		); err != nil {
			ac.logger.Warn("Failed to log analytics activity", map[string]interface{}{"error": err.Error()})
		}
	}

	c.JSON(http.StatusOK, summary)
}

// GetFundUtilizationReport returns a report of fund utilization
func (ac *AnalyticsController) GetFundUtilizationReport(c *gin.Context) {
	ac.logger.Info("Getting fund utilization report", nil)

	fundID := c.Query("fundId")
	entityID := c.Query("entityId")
	fiscalYear := c.Query("fiscalYear")

	if fundID == "" && entityID == "" {
		ac.logger.Warn("Missing required parameters", map[string]interface{}{
			"fundID":   fundID,
			"entityID": entityID,
		})
		c.JSON(http.StatusBadRequest, gin.H{"error": "Either fundId or entityId must be provided"})
		return
	}

	report, err := ac.analyticsService.GetFundUtilizationReport(fundID, entityID, fiscalYear)
	if err != nil {
		ac.logger.Error("Failed to retrieve fund utilization report", map[string]interface{}{"error": err.Error()})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve fund utilization report"})
		return
	}

	// Log the activity
	userID, exists := c.Get("userID")
	if exists {
		metadata := map[string]interface{}{
			"fundID":     fundID,
			"entityID":   entityID,
			"fiscalYear": fiscalYear,
		}

		if err := ac.auditService.LogActivity(
			userID.(string),
			"ANALYTICS_FUND_UTILIZATION_VIEWED",
			"Analytics",
			"",
			metadata,
		); err != nil {
			ac.logger.Warn("Failed to log analytics activity", map[string]interface{}{"error": err.Error()})
		}
	}

	c.JSON(http.StatusOK, report)
}

// GetSystemStats returns system statistics
func (ac *AnalyticsController) GetSystemStats(c *gin.Context) {
	ac.logger.Info("Getting system statistics", nil)

	stats, err := ac.analyticsService.GetSystemStats()
	if err != nil {
		ac.logger.Error("Failed to retrieve system statistics", map[string]any{"error": err.Error()})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve system statistics"})
		return
	}

	// Log the activity
	userID, exists := c.Get("userID")
	if exists {
		metadata := map[string]interface{}{
			"stats": stats,
		}

		if err := ac.auditService.LogActivity(
			userID.(string),
			"ANALYTICS_SYSTEM_STATS_VIEWED",
			"Analytics",
			"",
			metadata,
		); err != nil {
			ac.logger.Warn("Failed to log analytics activity", map[string]any{"error": err.Error()})
		}
	}

	c.JSON(http.StatusOK, stats)
}

// GetRecentActivity returns recent system activity
func (ac *AnalyticsController) GetRecentActivity(c *gin.Context) {
	ac.logger.Info("Getting recent activity", nil)

	activity, err := ac.analyticsService.GetRecentActivity(10)
	if err != nil {
		ac.logger.Error("Failed to retrieve recent activity", map[string]any{"error": err.Error()})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve recent activity"})
		return
	}

	// Log the activity
	userID, exists := c.Get("userID")
	if exists {
		metadata := map[string]interface{}{
			"activity": activity,
		}

		if err := ac.auditService.LogActivity(
			userID.(string),
			"ANALYTICS_RECENT_ACTIVITY_VIEWED",
			"Analytics",
			"",
			metadata,
		); err != nil {
			ac.logger.Warn("Failed to log analytics activity", map[string]any{"error": err.Error()})
		}
	}

	c.JSON(http.StatusOK, activity)
}

// GetUserRegistrationTrends returns trends in user registrations
func (ac *AnalyticsController) GetUserRegistrationTrends(c *gin.Context) {
	ac.logger.Info("Getting user registration trends", nil)

	trends, err := ac.analyticsService.GetUserRegistrationTrends(12)
	if err != nil {
		ac.logger.Error("Failed to retrieve user registration trends", map[string]any{"error": err.Error()})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user registration trends"})
		return
	}

	// Log the activity
	userID, exists := c.Get("userID")
	if exists {
		metadata := map[string]interface{}{
			"trends": trends,
		}

		if err := ac.auditService.LogActivity(
			userID.(string),
			"ANALYTICS_USER_REGISTRATION_TRENDS_VIEWED",
			"Analytics",
			"",
			metadata,
		); err != nil {
			ac.logger.Warn("Failed to log analytics activity", map[string]any{"error": err.Error()})
		}
	}

	c.JSON(http.StatusOK, trends)
}







