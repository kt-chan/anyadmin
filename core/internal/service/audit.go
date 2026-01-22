package service

import (
	"github.com/anyzearch/admin/core/internal/global"
)

func RecordLog(username, action, detail, level string) {
	log := global.OperationLog{
		Username: username,
		Action:   action,
		Detail:   detail,
		Level:    level,
	}
	global.DB.Create(&log)
}

func GetRecentLogs(limit int) []global.OperationLog {
	var logs []global.OperationLog
	global.DB.Order("created_at desc").Limit(limit).Find(&logs)
	return logs
}
