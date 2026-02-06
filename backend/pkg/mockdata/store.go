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
	
	// Deployment Nodes (Nested Structure)
	DeploymentNodes []global.DeploymentNode

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
	DeploymentNodes []global.DeploymentNode  `json:"deployment_nodes"`
	MgmtHost        string                   `json:"mgmt_host"`
	MgmtPort        string                   `json:"mgmt_port"`
}

func InitData() {
	// Try to load from file first
	if err := LoadFromFile(); err != nil {
		// If load fails (e.g. file doesn't match), we initialize defaults
		// but we should probably try to migrate if possible. 
		// For now, let's assume we start fresh or file is correct.
	}

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

	SaveToFile()
}

func SaveToFile() error {
	Mu.Lock()
	defer Mu.Unlock()

	data := DataStore{
		Users:           Users,
		ImportTasks:     ImportTasks,
		BackupRecords:   BackupRecords,
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
	DeploymentNodes = data.DeploymentNodes
	MgmtHost = data.MgmtHost
	MgmtPort = data.MgmtPort

	return nil
}