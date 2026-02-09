package api

import (
	"net/http"

	"anyadmin-backend/pkg/global"
	"anyadmin-backend/pkg/service"
	"anyadmin-backend/pkg/utils"

	"github.com/gin-gonic/gin"
)

func CreateImportTask(c *gin.Context) {
	var task global.ImportTask
	if err := c.ShouldBindJSON(&task); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	task.Status = "Running"

	utils.ExecuteWrite(func() {
		// Assign a mock ID if needed, or just append
		utils.ImportTasks = append(utils.ImportTasks, task)
	}, true)

	username, _ := c.Get("username")
	service.RecordLog(username.(string), "数据导入", "启动了新的数据导入任务: "+task.Name, "Info")

	// 异步开始任务
	go service.StartImportTask(task.ID)

	c.JSON(http.StatusOK, task)
}

func GetImportTasks(c *gin.Context) {
	utils.ExecuteRead(func() {
		c.JSON(http.StatusOK, utils.ImportTasks)
	})
}
