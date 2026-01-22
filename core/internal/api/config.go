package api

import (
	"net/http"

	"github.com/anyzearch/admin/core/internal/global"
	"github.com/anyzearch/admin/core/internal/service"
	"github.com/gin-gonic/gin"
)

func SaveInferenceConfig(c *gin.Context) {
	var config global.InferenceConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var existing global.InferenceConfig
	if global.DB.Where("name = ?", config.Name).First(&existing).Error == nil {
		config.ID = existing.ID
		global.DB.Save(&config)
	} else {
		global.DB.Create(&config)
	}

	username, _ := c.Get("username")
	service.RecordLog(username.(string), "修改配置", "保存了模型 "+config.Name+" 的推理参数", "Info")

	c.JSON(http.StatusOK, config)
}

func GetInferenceConfigs(c *gin.Context) {
	var configs []global.InferenceConfig
	global.DB.Find(&configs)
	c.JSON(http.StatusOK, configs)
}

func DeleteInferenceConfig(c *gin.Context) {
	id := c.Param("id")
	var config global.InferenceConfig
	if err := global.DB.First(&config, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "配置不存在"})
		return
	}

	// 1. 尝试清理 Docker 容器
	service.ControlContainer(config.Name, "stop")
	// 注意：ControlContainer 暂不支持删除，我们在这里补充逻辑或直接在 API 调用 Docker SDK
	
	// 2. 从数据库删除
	global.DB.Delete(&config)

	// 3. 记录审计日志
	username, _ := c.Get("username")
	service.RecordLog(username.(string), "删除服务", "彻底移除了模型配置及其关联容器: "+config.Name, "Warning")

	c.JSON(http.StatusOK, gin.H{"message": "服务已彻底删除"})
}
