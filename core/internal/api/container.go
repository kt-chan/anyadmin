package api

import (
	"net/http"

	"github.com/anyzearch/admin/core/internal/service"
	"github.com/gin-gonic/gin"
)

type ControlRequest struct {
	Name   string `json:"name" binding:"required"`
	Action string `json:"action" binding:"required"` // start, stop, restart
}

func ControlContainer(c *gin.Context) {
	var req ControlRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	if err := service.ControlContainer(req.Name, req.Action); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	username, _ := c.Get("username")
	service.RecordLog(username.(string), "容器控制", "对容器 "+req.Name+" 执行了 "+req.Action+" 操作", "Info")

	c.JSON(http.StatusOK, gin.H{"message": "Action executed successfully"})
}
