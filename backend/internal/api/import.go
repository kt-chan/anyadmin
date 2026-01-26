package api

import (
	"net/http"

	"anyadmin-backend/internal/global"
	"anyadmin-backend/internal/mockdata"
	"anyadmin-backend/internal/service"
	"github.com/gin-gonic/gin"
)

func CreateImportTask(c *gin.Context) {
	var task global.ImportTask
	if err := c.ShouldBindJSON(&task); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	task.Status = "Running"
	
	mockdata.Mu.Lock()
	// Assign a mock ID if needed, or just append
	mockdata.ImportTasks = append(mockdata.ImportTasks, task)
	mockdata.Mu.Unlock()

	username, _ := c.Get("username")
	service.RecordLog(username.(string), "数据导入", "启动了新的批量导入任务: "+task.Name, "Info")

	// 异步开始任务
	go service.StartImportTask(task.ID)

	c.JSON(http.StatusOK, task)
}

func GetImportTasks(c *gin.Context) {
	mockdata.Mu.Lock()
	defer mockdata.Mu.Unlock()
	c.JSON(http.StatusOK, mockdata.ImportTasks)
}
