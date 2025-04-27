package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/the-digital-watchdog-initiative/utils"
)

// HealthCheck handles the API health check endpoint
func HealthCheck(c *gin.Context) {
	// Check database connection
	dbStatus := "ok"
	if err := utils.DB.Exec("SELECT 1").Error; err != nil {
		dbStatus = "error: " + err.Error()
	}

	c.JSON(http.StatusOK, gin.H{
		"status":    "ok",
		"timestamp": time.Now().Format(time.RFC3339),
		"version":   "1.0.0",
		"services": gin.H{
			"database": dbStatus,
			"api":      "ok",
		},
	})
}
