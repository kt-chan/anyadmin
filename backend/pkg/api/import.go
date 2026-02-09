package api

import (
	"net/http"

	"anyadmin-backend/pkg/global"
	"anyadmin-backend/pkg/utils"
	"anyadmin-backend/pkg/service"
	"github.com/gin-gonic/gin"
)

func CreateImportTask(c *gin.Context) {
	var task global.ImportTask
	if err := c.ShouldBindJSON(&task); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	task.Status = "Running"
	
	utils.Mu.Lock()
	// Assign a mock ID if needed, or just append
	utils.ImportTasks = append(utils.ImportTasks, task)
	utils.Mu.Unlock()

	username, _ := c.Get("username")
	service.RecordLog(username.(string), "数据导入", "启动了新的批量导入任务: "+task.Name, "Info")

	// 异步开始任务
	go service.StartImportTask(task.ID)

	c.JSON(http.StatusOK, task)
}

func GetImportTasks(c *gin.Context) {
	utils.Mu.Lock()
	defer utils.Mu.Unlock()
	c.JSON(http.StatusOK, utils.ImportTasks)
}
