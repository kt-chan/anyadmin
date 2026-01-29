package main

import (
	"anyadmin-backend/pkg/agent"
	"flag"
	"log"
	"os"
	"time"
	"encoding/json"
	"io/ioutil"
)

var (
	mgmtURL string
	nodeIP  string
	configFile string
)

func init() {
	flag.StringVar(&mgmtURL, "server", "http://127.0.0.1:8080", "Management Server URL")
	flag.StringVar(&nodeIP, "ip", "127.0.0.1", "Node IP Address")
	flag.StringVar(&configFile, "config", "config.json", "Path to config file")
}

type Config struct {
	MgmtURL        string `json:"mgmt_url"`
	NodeIP         string `json:"node_ip"`
	DeploymentTime string `json:"deployment_time"`
}

var deploymentTime string

func loadConfig(path string) (*Config, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func main() {
	flag.Parse()

	// Try to load from config file first
	cfg, err := loadConfig(configFile)
	if err == nil {
		log.Printf("Loaded config from %s", configFile)
		mgmtURL = cfg.MgmtURL
		nodeIP = cfg.NodeIP
		deploymentTime = cfg.DeploymentTime
	} else {
		log.Printf("Config file %s not found or invalid, using flags or defaults", configFile)
	}

	hostname, _ := os.Hostname()

	log.Printf("Starting Agent on %s (%s)...", hostname, nodeIP)
	log.Printf("Management Server: %s", mgmtURL)

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	// Initial heartbeat
	if err := agent.SendHeartbeat(mgmtURL, nodeIP, hostname, deploymentTime); err != nil {
		log.Println("Heartbeat error:", err)
	}

	for range ticker.C {
		if err := agent.SendHeartbeat(mgmtURL, nodeIP, hostname, deploymentTime); err != nil {
			log.Println("Heartbeat error:", err)
		}
	}
}
