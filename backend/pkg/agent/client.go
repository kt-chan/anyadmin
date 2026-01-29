package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// HeartbeatRequest defines the structure of the heartbeat payload
type HeartbeatRequest struct {
	NodeIP       string  `json:"node_ip"`
	Hostname     string  `json:"hostname"`
	OS           string  `json:"os"`
	Arch         string  `json:"arch"`
	Status       string  `json:"status"`
	CPUUsage     float64 `json:"cpu_usage"`
	MemoryUsage  float64 `json:"memory_usage"`
	DockerStatus string  `json:"docker_status"`
}

// SendHeartbeat sends a heartbeat to the management server
func SendHeartbeat(mgmtURL, nodeIP, hostname string) error {
	cpu, _ := getCPUUsage()
	mem, _ := getMemoryUsage()

	payload := HeartbeatRequest{
		NodeIP:       nodeIP,
		Hostname:     hostname,
		OS:           runtime.GOOS,
		Arch:         runtime.GOARCH,
		Status:       "online",
		CPUUsage:     cpu,
		MemoryUsage:  mem,
		DockerStatus: checkDocker(),
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("error marshalling heartbeat: %w", err)
	}

	url := fmt.Sprintf("%s/api/v1/agent/heartbeat", mgmtURL)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("failed to send heartbeat to %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned non-200 status: %d", resp.StatusCode)
	}

	log.Printf("Heartbeat sent successfully to %s", mgmtURL)
	return nil
}

// getCPUUsage calculates CPU usage by sampling /proc/stat
func getCPUUsage() (float64, error) {
	if runtime.GOOS != "linux" {
		return 0, nil // Return 0 on non-linux to avoid errors in dev/test
	}

	idle0, total0, err := readProcStat()
	if err != nil {
		return 0, err
	}

	time.Sleep(200 * time.Millisecond)

	idle1, total1, err := readProcStat()
	if err != nil {
		return 0, err
	}

	return CalculateCPUPercent(idle0, total0, idle1, total1), nil
}

func readProcStat() (uint64, uint64, error) {
	data, err := os.ReadFile("/proc/stat")
	if err != nil {
		return 0, 0, err
	}
	return ParseProcStat(string(data))
}

// ParseProcStat exports the parsing logic for /proc/stat
func ParseProcStat(content string) (uint64, uint64, error) {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) > 0 && fields[0] == "cpu" {
			var total uint64
			var idle uint64
			// fields[0] is "cpu"
			// Columns: user, nice, system, idle, iowait, irq, softirq, steal, guest, guest_nice
			for i, field := range fields[1:] {
				val, err := strconv.ParseUint(field, 10, 64)
				if err != nil {
					return 0, 0, err
				}
			total += val
			if i == 3 { // idle is the 4th value (index 3)
				idle = val
			}
			}
			return idle, total, nil
		}
	}
	return 0, 0, fmt.Errorf("cpu line not found")
}

// CalculateCPUPercent calculates the percentage usage between two samples
func CalculateCPUPercent(idle0, total0, idle1, total1 uint64) float64 {
	deltaTotal := float64(total1 - total0)
	deltaIdle := float64(idle1 - idle0)
	if deltaTotal == 0 {
		return 0.0
	}
	return (1.0 - deltaIdle/deltaTotal) * 100.0
}

func getMemoryUsage() (float64, error) {
	if runtime.GOOS != "linux" {
		return 0, nil
	}
	data, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return 0, err
	}
	return ParseMemInfo(string(data))
}

// ParseMemInfo parses /proc/meminfo content
func ParseMemInfo(content string) (float64, error) {
	var total, available float64
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		key := strings.TrimSuffix(fields[0], ":")
		val, err := strconv.ParseFloat(fields[1], 64)
		if err != nil {
			continue
		}

		if key == "MemTotal" {
			total = val
		} else if key == "MemAvailable" {
			available = val
		}
	}

	if total == 0 {
		return 0, fmt.Errorf("MemTotal not found")
	}

	// Usage %
	return ((total - available) / total) * 100.0, nil
}

func checkDocker() string {
	// Simple check using 'docker info'
	// Note: cmd.Run() returns error if exit code is non-zero
	cmd := exec.Command("docker", "info")
	if err := cmd.Run(); err != nil {
		return "inactive"
	}
	return "active"
}