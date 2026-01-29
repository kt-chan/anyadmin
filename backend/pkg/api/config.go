package api

import (
	"net/http"

	"anyadmin-backend/pkg/global"
	"anyadmin-backend/pkg/mockdata"
	"anyadmin-backend/pkg/service"
	"github.com/gin-gonic/gin"
)

func SaveInferenceConfig(c *gin.Context) {
	var config global.InferenceConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

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

	username, _ := c.Get("username")
	service.RecordLog(username.(string), "修改配置", "保存了模型 "+config.Name+" 的推理参数", "Info")

	c.JSON(http.StatusOK, config)
}

func GetInferenceConfigs(c *gin.Context) {
	mockdata.Mu.Lock()
	defer mockdata.Mu.Unlock()
	c.JSON(http.StatusOK, mockdata.InferenceCfgs)
}

func DeleteInferenceConfig(c *gin.Context) {
	// Mock delete
	// In real logic we would delete from DB and stop container
	username, _ := c.Get("username")
	service.RecordLog(username.(string), "删除服务", "彻底移除了模型配置及其关联容器", "Warning")

	c.JSON(http.StatusOK, gin.H{"message": "服务已彻底删除"})
}
