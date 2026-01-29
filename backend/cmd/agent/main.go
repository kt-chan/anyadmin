package main

import (
	"anyadmin-backend/pkg/agent"
	"flag"
	"log"
	"os"
	"time"
)

var (
	mgmtURL string
	nodeIP  string
)

func init() {
	flag.StringVar(&mgmtURL, "server", "http://127.0.0.1:8080", "Management Server URL")
	flag.StringVar(&nodeIP, "ip", "127.0.0.1", "Node IP Address")
}

func main() {
	flag.Parse()

	hostname, _ := os.Hostname()

	log.Printf("Starting Agent on %s (%s)...", hostname, nodeIP)
	log.Printf("Management Server: %s", mgmtURL)

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	// Initial heartbeat
	if err := agent.SendHeartbeat(mgmtURL, nodeIP, hostname); err != nil {
		log.Println(err)
	}

	for range ticker.C {
		if err := agent.SendHeartbeat(mgmtURL, nodeIP, hostname); err != nil {
			log.Println(err)
		}
	}
}
