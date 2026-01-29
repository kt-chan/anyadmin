package service

import (
	"log"
	"time"

	"anyadmin-backend/pkg/global"
)

func RecordLog(username, action, detail, level string) {
	// Standard Log using the unified logger
	log.Printf("[AUDIT] - %s [%s] \"%s %s\" %s\n", 
		username, 
		time.Now().Format("02/Jan/2006:15:04:05 -0700"), 
		action, 
		detail, 
		level)

	// In the future, this could still populate a database or other persistent store
}

func GetRecentLogs(limit int) []global.OperationLog {
	// Returns empty for now as OperationLogs was removed from mockdata
	return []global.OperationLog{}
}
