package utils

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

	// Mutex for thread-safe updates (Private)
	dataMu sync.RWMutex

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

// ExecuteRead performs a thread-safe read operation
func ExecuteRead(fn func()) {
	dataMu.RLock()
	defer dataMu.RUnlock()
	fn()
}

// ExecuteWrite performs a thread-safe write operation and optionally persists to file
func ExecuteWrite(fn func(), persist bool) error {
	dataMu.Lock()
	defer dataMu.Unlock()
	
	fn()
	
	if persist {
		return saveToFile()
	}
	return nil
}

func InitData() {
	// Try to load from file first
	if err := LoadFromFile(); err != nil {
		// If load fails, initialize defaults
	}

	ExecuteWrite(func() {
		// Initialize Users if empty
		if len(Users) == 0 {
			adminPass, _ := EncryptPassword("password")
			operatorPass, _ := EncryptPassword("password")
			
			Users = []global.User{
				{
					Username: "admin",
					Password: adminPass,
					Role:     "admin",
				},
				{
					Username: "operator_01",
					Password: operatorPass,
					Role:     "operator",
				},
			}
		} else {
			// Migration: Ensure all users have encrypted passwords
			for i := range Users {
				// Try to decrypt. If it fails, it's likely plain text.
				// Also check if it's the specific plain text "password" to be sure, 
				// or just rely on decryption failure. 
				// A simple heuristic: real encrypted string is long (Base64 of 256 bytes = 344 chars).
				// "password" is short.
				if len(Users[i].Password) < 100 { 
					enc, err := EncryptPassword(Users[i].Password)
					if err == nil {
						Users[i].Password = enc
					}
				} else {
					// Double check by trying to decrypt
					_, err := DecryptPassword(Users[i].Password)
					if err != nil {
						// If decryption fails, maybe it's corrupted or just a long plain text?
						// Re-encrypting might be risky if it was already encrypted but with a different key.
						// Assume < 100 check covers most plain passwords. 
						// For now, trust the length check as RSA-2048 output is 256 bytes -> base64 is longer.
					}
				}
			}
			// Note: We don't explicit set 'updated' flag logic here because 'saveToFile' is called if persist=true
			// But ExecuteWrite calls saveToFile if persist is true.
		}

		// Initialize MgmtHost and MgmtPort if empty
		if MgmtHost == "" {
			MgmtHost = "172.20.0.1"
		}
		if MgmtPort == "" {
			MgmtPort = "8080"
		}
	}, true)
}

// saveToFile writes data to disk (internal, assumes lock held)
func saveToFile() error {
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

// SaveToFile Public alias for backward compatibility or explicit save if needed, 
// though ExecuteWrite is preferred.
func SaveToFile() error {
	return ExecuteWrite(func() {}, true)
}

func LoadFromFile() error {
	dataMu.Lock()
	defer dataMu.Unlock()

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