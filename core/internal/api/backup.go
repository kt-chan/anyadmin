package api

import (
	"net/http"
	"strconv"

	"github.com/anyzearch/admin/core/internal/global"
	"github.com/anyzearch/admin/core/internal/service"
	"github.com/gin-gonic/gin"
)

func CreateBackup(c *gin.Context) {
	record, err := service.CreateBackup()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	username, _ := c.Get("username")
	service.RecordLog(username.(string), "系统备份", "手动触发了全量备份操作: "+record.Name, "Info")

	c.JSON(http.StatusOK, record)
}

func GetBackups(c *gin.Context) {
	var records []global.BackupRecord
	global.DB.Order("created_at desc").Find(&records)
	c.JSON(http.StatusOK, records)
}

func RestoreBackup(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := service.RestoreBackup(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	username, _ := c.Get("username")
	service.RecordLog(username.(string), "数据恢复", "手动触发了数据回滚操作", "Warning")

	c.JSON(http.StatusOK, gin.H{"message": "Restore initiated"})
}
