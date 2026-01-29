package api

import (
	"net/http"

	"anyadmin-backend/pkg/service"
	"github.com/gin-gonic/gin"
)

func GetDashboardStats(c *gin.Context) {
	systemStats, err := service.GetSystemStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	servicesHealth := service.GetServicesHealth()
	recentLogs := service.GetRecentLogs(10)

	c.JSON(http.StatusOK, gin.H{
		"system":   systemStats,
		"services": servicesHealth,
		"logs":     recentLogs,
	})
}
