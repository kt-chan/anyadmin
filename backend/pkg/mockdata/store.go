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
	RagAppCfgs    []global.RagAppConfig

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
	RagAppCfgs      []global.RagAppConfig    `json:"rag_app_cfgs"`
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

	// Initialize MgmtHost and MgmtPort if empty
	if MgmtHost == "" {
		MgmtHost = "172.20.0.1"
	}
	if MgmtPort == "" {
		MgmtPort = "8080"
	}

	if len(InferenceCfgs) == 0 {
		InferenceCfgs = []global.InferenceConfig{
			{
				Name:                 "default",
				Engine:               "vLLM",
				ModelName:            "Qwen3-1.7B",
				Mode:                 "balanced",
				MaxModelLen:          4096,
				MaxNumSeqs:           256,
				MaxNumBatchedTokens:  2048,
				GpuMemoryUtilization: 0.85,
			},
		}
	}

	if len(RagAppCfgs) == 0 {
		RagAppCfgs = []global.RagAppConfig{
			{
				Name:     "anythingllm",
				Host:     "172.20.0.10",
				Port:     "3001",
				VectorDB: "lancedb",
			},
		}
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
		RagAppCfgs:      RagAppCfgs,
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
	RagAppCfgs = data.RagAppCfgs
	DeploymentNodes = data.DeploymentNodes
	MgmtHost = data.MgmtHost
	MgmtPort = data.MgmtPort

	return nil
}
