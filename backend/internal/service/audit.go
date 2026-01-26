package service

import (
	"anyadmin-backend/internal/global"
	"anyadmin-backend/internal/mockdata"
)

func RecordLog(username, action, detail, level string) {
	log := global.OperationLog{
		Username: username,
		Action:   action,
		Detail:   detail,
		Level:    level,
	}
	mockdata.Mu.Lock()
	mockdata.OperationLogs = append([]global.OperationLog{log}, mockdata.OperationLogs...)
	if len(mockdata.OperationLogs) > 100 {
		mockdata.OperationLogs = mockdata.OperationLogs[:100]
	}
	mockdata.Mu.Unlock()
}

func GetRecentLogs(limit int) []global.OperationLog {
	mockdata.Mu.Lock()
	defer mockdata.Mu.Unlock()
	if len(mockdata.OperationLogs) < limit {
		return mockdata.OperationLogs
	}
	return mockdata.OperationLogs[:limit]
}
