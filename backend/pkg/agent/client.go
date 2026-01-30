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

// DockerServiceStatus represents the status of a specific docker container
type DockerServiceStatus struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Image  string `json:"image"`
	Status string `json:"status"`
	State  string `json:"state"`
	Uptime string `json:"uptime"`
}

// HeartbeatRequest defines the structure of the heartbeat payload
type HeartbeatRequest struct {
	NodeIP         string                `json:"node_ip"`
	Hostname       string                `json:"hostname"`
	Status         string                `json:"status"`
	CPUUsage       float64               `json:"cpu_usage"`
	CPUCapacity    string                `json:"cpu_capacity"` // e.g. "16 vCPUs"
	MemoryUsage    float64               `json:"memory_usage"`
	MemoryCapacity string                `json:"memory_capacity"` // e.g. "32 GB"
	DockerStatus   string                `json:"docker_status"`
	DeploymentTime string                `json:"deployment_time"`
	OSSpec         string                `json:"os_spec"`
	GPUStatus      string                `json:"gpu_status"`
	Services       []DockerServiceStatus `json:"services"`
}

// SendHeartbeat sends a heartbeat to the management server
func SendHeartbeat(mgmtURL, nodeIP, hostname, deploymentTime string) error {
	cpu, _ := getCPUUsage()
	mem, _ := getMemoryUsage()
	gpu := getGPUStatus()
	osSpec := getOSDescription()
	cpuCap := getCPUCapacity()
	memCap := getMemoryCapacity()

	payload := HeartbeatRequest{
		NodeIP:         nodeIP,
		Hostname:       hostname,
		Status:         "online",
		CPUUsage:       cpu,
		CPUCapacity:    cpuCap,
		MemoryUsage:    mem,
		MemoryCapacity: memCap,
		DockerStatus:   checkDocker(),
		DeploymentTime: deploymentTime,
		OSSpec:         osSpec,
		GPUStatus:      gpu,
		Services:       getDockerServices(),
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

// getCPUCapacity returns the number of CPU cores using nproc
func getCPUCapacity() string {
	if runtime.GOOS != "linux" {
		return fmt.Sprintf("%d vCPUs", runtime.NumCPU())
	}
	cmd := exec.Command("nproc")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Sprintf("%d vCPUs", runtime.NumCPU())
	}
	return strings.TrimSpace(string(output))
}

// getMemoryCapacity returns the total system memory using free -h
func getMemoryCapacity() string {
	if runtime.GOOS != "linux" {
		return "Unknown"
	}
	// Command: free -h | grep Mem: | awk '{print $2}'
	cmd := exec.Command("bash", "-c", "free -h | grep Mem: | awk '{print $2}'")
	output, err := cmd.Output()
	if err != nil {
		// Fallback to reading /proc/meminfo if command fails
		data, err := os.ReadFile("/proc/meminfo")
		if err != nil {
			return "Unknown"
		}
		
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "MemTotal:") {
				parts := strings.Fields(line)
				if len(parts) >= 2 {
					kb, _ := strconv.ParseFloat(parts[1], 64)
					return fmt.Sprintf("%.1f GB", kb/(1024*1024))
				}
			}
		}
		return "Unknown"
	}
	return strings.TrimSpace(string(output))
}

func getOSDescription() string {
	if runtime.GOOS == "windows" {
		return "Windows " + runtime.GOARCH
	}

	// Try /etc/os-release first (standard)
	data, err := os.ReadFile("/etc/os-release")
	if err == nil {
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "PRETTY_NAME=") {
				return strings.Trim(strings.TrimPrefix(line, "PRETTY_NAME="), "\"")
			}
		}
	}

	// Fallback to lsb_release if available
	cmd := exec.Command("lsb_release", "-ds")
	output, err := cmd.Output()
	if err == nil {
		return strings.TrimSpace(string(output))
	}

	return fmt.Sprintf("%s %s", runtime.GOOS, runtime.GOARCH)
}

func getGPUStatus() string {
	// Try nvidia-smi with detailed query
	// utilization.gpu, memory.used, memory.total
	cmd := exec.Command("nvidia-smi", "--query-gpu=name,utilization.gpu,memory.used,memory.total", "--format=csv,noheader,nounits")
	output, err := cmd.Output()
	if err == nil {
		lines := strings.Split(strings.TrimSpace(string(output)), "\n")
		cardCount := len(lines)
		
		// Use the first card for the detailed string, but prefix with count
		if cardCount > 0 {
			parts := strings.Split(lines[0], ",")
			if len(parts) >= 4 {
				name := strings.TrimSpace(parts[0])
				util := strings.TrimSpace(parts[1])
				used := strings.TrimSpace(parts[2])
				total := strings.TrimSpace(parts[3])
				return fmt.Sprintf("%d x %s | Util: %s%% | Mem: %s/%s MB", cardCount, name, util, used, total)
			}
		}
		return fmt.Sprintf("%d x NVIDIA GPU", cardCount)
	}

	// Try npu-smi (Ascend)
	cmd = exec.Command("npu-smi", "info")
	if err := cmd.Run(); err == nil {
		return "Ascend NPU"
	}

	return "None"
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

func getDockerServices() []DockerServiceStatus {

	// Use docker ps with format to get key information

	// Format: ID, Names, Image, Status, State, RunningFor (Uptime)

	cmd := exec.Command("docker", "ps", "--format", "{{.ID}}|{{.Names}}|{{.Image}}|{{.Status}}|{{.State}}|{{.RunningFor}}")

	output, err := cmd.Output()

	if err != nil {

		return nil

	}

	return ParseDockerPsOutput(string(output))

}

// ParseDockerPsOutput parses the output of docker ps --format ...

func ParseDockerPsOutput(output string) []DockerServiceStatus {

	// Target services to monitor

	targets := []string{"vllm", "anysearch", "anyzearch", "anythingllm", "anything-llm", "milvus", "lancedb", "chroma", "pgvector", "mineru"}

	var services []DockerServiceStatus

	lines := strings.Split(strings.TrimSpace(output), "\n")

	for _, line := range lines {

		if line == "" {

			continue

		}

		parts := strings.Split(line, "|")

		if len(parts) < 6 {

			continue

		}

		name := parts[1]

		isTarget := false

		for _, t := range targets {

			if strings.Contains(strings.ToLower(name), t) {

				isTarget = true

				break

			}

		}

		if isTarget {

			services = append(services, DockerServiceStatus{

				ID: parts[0],

				Name: parts[1],

				Image: parts[2],

				Status: parts[3],

				State: parts[4],

				Uptime: parts[5],
			})

		}

	}

	return services

}
