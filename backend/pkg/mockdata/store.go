package mockdata

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"

	"anyadmin-backend/pkg/global"
)

var (
	Users         []global.User
	ImportTasks   []global.ImportTask
	BackupRecords []global.BackupRecord
	InferenceCfgs []global.InferenceConfig

	// Deployment Nodes
	DeploymentNodes []string

	// Management Info
	MgmtHost string
	MgmtPort string

	// Mutex for thread-safe updates
	Mu sync.Mutex

	// DataFile path
	DataFile = "data.json"
)

func init() {
	// Find data.json in backend directory
	cwd, _ := os.Getwd()
	checkPaths := []string{
		filepath.Join(cwd, "backend", "data.json"),
		filepath.Join(cwd, "..", "backend", "data.json"),
		filepath.Join(cwd, "..", "..", "backend", "data.json"),
		filepath.Join(cwd, "..", "..", "..", "backend", "data.json"),
		filepath.Join(cwd, "data.json"),
	}

	for _, p := range checkPaths {
		if _, err := os.Stat(p); err == nil {
			DataFile = p
			break
		}
	}
}

type DataStore struct {
	Users           []global.User            `json:"users"`
	ImportTasks     []global.ImportTask      `json:"import_tasks"`
	BackupRecords   []global.BackupRecord    `json:"backup_records"`
	InferenceCfgs   []global.InferenceConfig `json:"inference_cfgs"`
	DeploymentNodes []string                 `json:"deployment_nodes"`
	MgmtHost        string                   `json:"mgmt_host"`
	MgmtPort        string                   `json:"mgmt_port"`
}

func InitData() {
	// Try to load from file first
	LoadFromFile()

	// Initialize Users if empty
	if len(Users) == 0 {
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
	}

	// Initialize DeploymentNodes if empty
	if len(DeploymentNodes) == 0 {
		DeploymentNodes = []string{"1.1.1.1:20", "1.1.1.2:20", "1.1.1.3:20"}
	}

	// Initialize ImportTasks if empty
	if len(ImportTasks) == 0 {
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
	}

	// Initialize BackupRecords if empty
	if len(BackupRecords) == 0 {
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
	}

	// Initialize MgmtHost and MgmtPort if empty
	if MgmtHost == "" {
		MgmtHost = "172.20.0.1"
	}
	if MgmtPort == "" {
		MgmtPort = "8080"
	}

	SaveToFile()
}

func SaveToFile() error {
	Mu.Lock()
	defer Mu.Unlock()

	data := DataStore{
		Users:           Users,
		ImportTasks:     ImportTasks,
		BackupRecords:   BackupRecords,
		InferenceCfgs:   InferenceCfgs,
		DeploymentNodes: DeploymentNodes,
		MgmtHost:        MgmtHost,
		MgmtPort:        MgmtPort,
	}

	file, err := os.Create(DataFile)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

func LoadFromFile() error {
	Mu.Lock()
	defer Mu.Unlock()

	file, err := os.Open(DataFile)
	if err != nil {
		return err
	}
	defer file.Close()

	var data DataStore
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&data); err != nil {
		return err
	}

	Users = data.Users
	ImportTasks = data.ImportTasks
	BackupRecords = data.BackupRecords
	InferenceCfgs = data.InferenceCfgs
	DeploymentNodes = data.DeploymentNodes
	MgmtHost = data.MgmtHost
	MgmtPort = data.MgmtPort

	return nil
}
