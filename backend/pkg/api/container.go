package api

import (
	"net/http"

	"anyadmin-backend/pkg/service"
	"github.com/gin-gonic/gin"
)

type ControlRequest struct {
	Name   string `json:"name" binding:"required"`
	Action string `json:"action" binding:"required"` // start, stop, restart
	NodeIP string `json:"node_ip"`
}

func ControlContainer(c *gin.Context) {
	var req ControlRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	if err := service.ControlContainer(req.Name, req.Action, req.NodeIP); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	username, _ := c.Get("username")
	service.RecordLog(username.(string), "容器控制", "对节点 "+req.NodeIP+" 的容器 "+req.Name+" 执行了 "+req.Action+" 操作", "Info")

	c.JSON(http.StatusOK, gin.H{"message": "Action executed successfully"})
}
