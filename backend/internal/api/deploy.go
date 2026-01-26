package api

import (
	"log"
	"net/http"

	"anyadmin-backend/internal/global"
	"anyadmin-backend/internal/mockdata"
	"anyadmin-backend/internal/service"
	"github.com/gin-gonic/gin"
)

type DeployRequest struct {
	global.InferenceConfig
	DeployMode string `json:"deploy_mode"`
}

func DeployService(c *gin.Context) {
	var req DeployRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var containerID string
	var path string
	var err error

	// 仅当选择“全新部署”时才拉起 Docker 容器
	if req.DeployMode == "new" {
		path, containerID, err = service.GenerateAndStart(req.InferenceConfig)
		if err != nil {
			log.Printf("[部署失败] 模型: %s, 错误: %v", req.Name, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "path": path})
			return
		}
	} else {
		log.Printf("[接入服务] 正在登记现有服务: %s (%s:%s)", req.Name, req.IP, req.Port)
	}

	// 将配置持久化到数据库 (Mock Store)
	config := req.InferenceConfig
	
	mockdata.Mu.Lock()
	found := false
	for i, cfg := range mockdata.InferenceCfgs {
		if cfg.Name == config.Name {
			mockdata.InferenceCfgs[i] = config
			found = true
			break
		}
	}
	if !found {
		mockdata.InferenceCfgs = append(mockdata.InferenceCfgs, config)
	}
	mockdata.Mu.Unlock()

	// 记录审计日志
	action := "服务部署"
	detail := "成功部署并启动了新容器: " + config.Name
	if req.DeployMode == "existing" {
		action = "服务接入"
		detail = "成功接入现有外部服务: " + config.Name + " (" + config.IP + ":" + config.Port + ")"
	} else if containerID != "" {
		detail = "容器已拉起, 名称: " + config.Name + ", ID: " + containerID[:12]
	}
	service.RecordLog(c.GetString("username"), action, detail, "Info")

	c.JSON(http.StatusOK, gin.H{
		"message":      "Success",
		"container_id": containerID,
	})
}