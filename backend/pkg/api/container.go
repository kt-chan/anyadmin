package api

import (
	"log"
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

	// Async execution
	go func() {
		if err := service.ControlContainer(req.Name, req.Action, req.NodeIP); err != nil {
			log.Printf("[Container] Async action %s failed for %s: %v", req.Action, req.Name, err)
		}
	}()

	username, _ := c.Get("username")
	service.RecordLog(username.(string), "容器控制", "对节点 "+req.NodeIP+" 的容器 "+req.Name+" 执行了 "+req.Action+" 操作", "Info")

	c.JSON(http.StatusOK, gin.H{"message": "Action triggered successfully"})
}
