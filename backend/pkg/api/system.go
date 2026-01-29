package api

import (
	"net/http"

	"anyadmin-backend/pkg/service"
	"github.com/gin-gonic/gin"
)

func GetSystemStats(c *gin.Context) {
	stats, err := service.GetSystemStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, stats)
}
