package mockdata

import (
	"sync"
	"time"

	"anyadmin-backend/internal/global"
)

var (
	Users         []global.User
	ImportTasks   []global.ImportTask
	BackupRecords []global.BackupRecord
	InferenceCfgs []global.InferenceConfig
	OperationLogs []global.OperationLog
	
	// Mutex for thread-safe updates
	Mu sync.Mutex
)

func InitData() {
	// Initialize Users
	Users = []global.User{
		{
			Username: "admin",
			Password: "password",
			Role:     "admin",
		},
		{
			Username: "operator_01",
			Password: "password",
			Role:     "operator",
		},
	}

	// Initialize Inference Configs (Models)
	InferenceCfgs = []global.InferenceConfig{
		{
			Name:           "llama-3-8b-instruct",
			Engine:         "MindIE",
			ModelPath:      "/models/llama3",
			IP:             "10.0.1.5",
			Port:           "8000",
			MaxConcurrency: 64,
			TokenLimit:     8192,
		},
		{
			Name:      "bge-large-zh-v1.5",
			Engine:    "Embedding",
			IP:        "10.0.1.5",
			Port:      "8001",
		},
        {
            Name: "milvus-standalone",
            Engine: "Vector DB",
            IP: "10.0.1.8",
            Port: "19530",
        },
	}

	// Initialize Import Tasks
	ImportTasks = []global.ImportTask{
		{
			Name:       "文档全量同步",
			SourceType: "NFS",
			SourcePath: "/mnt/nfs/docs/v1",
			Status:     "Processing",
			Progress:   56,
			TotalFiles: 15000,
			Processed:  8432,
		},
		{
			Name:       "图片资源归档",
			SourceType: "S3",
			SourcePath: "s3://company-assets/images",
			Status:     "Paused",
			Progress:   42,
			TotalFiles: 5000,
			Processed:  2100,
		},
	}

	// Initialize Backups
	BackupRecords = []global.BackupRecord{
		{
			Name:   "backup_20240520_full.tar.gz",
			Path:   "/backups/backup_20240520_full.tar.gz",
			Size:   107374182400, // 100GB
			Type:   "Full",
			Status: "Success",
		},
		{
			Name:   "backup_20240519_inc.tar.gz",
			Path:   "/backups/backup_20240519_inc.tar.gz",
			Size:   10737418240, // 10GB
			Type:   "Incremental",
			Status: "Success",
		},
	}

	// Initialize Logs
	OperationLogs = []global.OperationLog{
		{
			Username: "admin",
			Action:   "用户登录",
			Detail:   "IP: 192.168.1.102 | Method: JWT Auth",
			Level:    "Info",
		},
		{
			Username: "system",
			Action:   "系统备份",
			Detail:   "系统自动执行全量备份",
			Level:    "Info",
		},
	}
    // Set timestamps
    now := time.Now()
    for i := range OperationLogs {
        OperationLogs[i].CreatedAt = now.Add(-time.Duration(i) * time.Hour)
    }
}
