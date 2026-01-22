package api

import (
	"net/http"

	"github.com/anyzearch/admin/core/internal/service"
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
