package service

import (
	"fmt"
	"log"
	"time"

	"anyadmin-backend/pkg/global"
	"anyadmin-backend/pkg/utils"
)

func CreateBackup() (*global.BackupRecord, error) {
	log.Println("[Backup] 开始执行系统备份 (MOCKED)...")
	timestamp := time.Now().Format("20060102_150405")
	
	fileName := fmt.Sprintf("backup_%s.tar.gz", timestamp)
	filePath := "/backups/" + fileName

	// Simulate delay
	time.Sleep(1 * time.Second)

	record := global.BackupRecord{
		Name:   fileName,
		Path:   filePath,
		Size:   1024 * 1024 * 500, // 500MB
		Type:   "Full",
		Status: "Success",
	}

	utils.ExecuteWrite(func() {
		utils.BackupRecords = append([]global.BackupRecord{record}, utils.BackupRecords...)
	}, true)
	
	return &record, nil
}

func RestoreBackup(id uint) error {
	// Mock implementation
	fmt.Printf("Restoring backup ID: %d (MOCKED)\n", id)
	time.Sleep(1 * time.Second)
	return nil
}

