package api

import (
	"net/http"

	"github.com/anyzearch/admin/core/internal/global"
	"github.com/anyzearch/admin/core/internal/service"
	"github.com/gin-gonic/gin"
)

func CreateImportTask(c *gin.Context) {
	var task global.ImportTask
	if err := c.ShouldBindJSON(&task); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	task.Status = "Running"
	global.DB.Create(&task)

	username, _ := c.Get("username")
	service.RecordLog(username.(string), "数据导入", "启动了新的批量导入任务: "+task.Name, "Info")

	// 异步开始任务
	go service.StartImportTask(task.ID)

	c.JSON(http.StatusOK, task)
}

func GetImportTasks(c *gin.Context) {
	var tasks []global.ImportTask
	global.DB.Order("created_at desc").Find(&tasks)
	c.JSON(http.StatusOK, tasks)
}
