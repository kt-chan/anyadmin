package agent

import (
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

	"anyadmin-backend/pkg/utils"
)

// --- Client / Heartbeat Logic ---

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
	CPUCapacity    string                `json:"cpu_capacity"`
	MemoryUsage    float64               `json:"memory_usage"`
	MemoryCapacity string                `json:"memory_capacity"`
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

	url := fmt.Sprintf("%s/api/v1/agent/heartbeat", mgmtURL)

	resp, err := utils.PostJSON(url, payload, 10*time.Second)
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

// --- System Info Helpers ---

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

func getMemoryCapacity() string {
	if runtime.GOOS != "linux" {
		return "Unknown"
	}
	cmd := exec.Command("bash", "-c", "free -h | grep Mem: | awk '{print $2}'")
	output, err := cmd.Output()
	if err != nil {
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
	data, err := os.ReadFile("/etc/os-release")
	if err == nil {
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "PRETTY_NAME=") {
				return strings.Trim(strings.TrimPrefix(line, "PRETTY_NAME="), "\"")
			}
		}
	}
	cmd := exec.Command("lsb_release", "-ds")
	output, err := cmd.Output()
	if err == nil {
		return strings.TrimSpace(string(output))
	}
	return fmt.Sprintf("%s %s", runtime.GOOS, runtime.GOARCH)
}

func getGPUStatus() string {
	cmd := exec.Command("nvidia-smi", "--query-gpu=name,utilization.gpu,memory.used,memory.total", "--format=csv,noheader,nounits")
	output, err := cmd.Output()
	if err == nil {
		lines := strings.Split(strings.TrimSpace(string(output)), "\n")
		cardCount := len(lines)
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
	cmd = exec.Command("npu-smi", "info")
	if err := cmd.Run(); err == nil {
		return "Ascend NPU"
	}
	return "None"
}

func getCPUUsage() (float64, error) {
	if runtime.GOOS != "linux" {
		return 0, nil
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

// ParseProcStat parses /proc/stat content
func ParseProcStat(content string) (uint64, uint64, error) {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) > 0 && fields[0] == "cpu" {
			var total uint64
			var idle uint64
			for i, field := range fields[1:] {
				val, err := strconv.ParseUint(field, 10, 64)
				if err != nil {
					return 0, 0, err
				}
				total += val
				if i == 3 {
					idle = val
				}
			}
			return idle, total, nil
		}
	}
	return 0, 0, fmt.Errorf("cpu line not found")
}

// CalculateCPUPercent calculates CPU usage percent
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
	return ((total - available) / total) * 100.0, nil
}

func checkDocker() string {
	cmd := exec.Command("docker", "info")
	if err := cmd.Run(); err != nil {
		return "inactive"
	}
	return "active"
}

func getDockerServices() []DockerServiceStatus {
	cmd := exec.Command("docker", "ps", "--format", "{{.ID}}|{{.Names}}|{{.Image}}|{{.Status}}|{{.State}}|{{.RunningFor}}")
	output, err := cmd.Output()
	if err != nil {
		return nil
	}
	return ParseDockerPsOutput(string(output))
}

// ParseDockerPsOutput parses docker ps output
func ParseDockerPsOutput(output string) []DockerServiceStatus {
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
				ID:     parts[0],
				Name:   parts[1],
				Image:  parts[2],
				Status: parts[3],
				State:  parts[4],
				Uptime: parts[5],
			})
		}
	}
	return services
}

// --- Server / Control Logic ---

type ContainerControlRequest struct {
	ContainerName string `json:"container_name"`
	Action        string `json:"action"` // start, stop, restart
}

type ContainerControlResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Output  string `json:"output"`
}

func StartServer(port string) {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", handleHealth)
	mux.HandleFunc("/container/control", handleContainerControl)

	addr := ":" + port
	log.Printf("Agent server listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("Failed to start agent server: %v", err)
	}
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func handleContainerControl(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ContainerControlRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	log.Printf("Received control request: %s %s", req.Action, req.ContainerName)

	if req.ContainerName == "" || req.Action == "" {
		http.Error(w, "Missing container_name or action", http.StatusBadRequest)
		return
	}

	if strings.Contains(req.ContainerName, " ") || strings.Contains(req.ContainerName, ";") {
		http.Error(w, "Invalid container name", http.StatusBadRequest)
		return
	}

	workDir := "/home/anyadmin/docker/"

	// Validate working directory existence (optional, but good for debugging)
	// We assume the directory exists as per the environment setup description.

	var args []string
	containerEnv := "/home/anyadmin/docker/.env-" + req.ContainerName

	args = append(args, "compose")

	// 1. Always load the base .env
	args = append(args, "--env-file", "/home/anyadmin/docker/.env")

	// 2. Load service-specific env files
	args = append(args, "--env-file", containerEnv)

	// 3. Determine command based on action
	switch req.Action {
	case "start":
		// Use 'up -d' to ensure env vars and config are applied
		args = append(args, "up", "-d", req.ContainerName)
	case "stop":
		args = append(args, "stop", req.ContainerName)
	case "restart":
		// Use 'up -d --force-recreate' to force config reload/env application
		args = append(args, "up", "-d", "--force-recreate", req.ContainerName)
	default:
		http.Error(w, "Unknown action. Supported: start, stop, restart", http.StatusBadRequest)
		return
	}

	cmdStr := "docker " + strings.Join(args, " ")
	log.Printf("Executing command in background: %s in %s", cmdStr, workDir)

	// Start command in background
	go func() {
		// Use bash -c to execute the full command string for better compatibility and clear logging
		cmd := exec.Command("bash", "-c", "cd "+workDir+" && "+cmdStr)
		output, err := cmd.CombinedOutput()
		if err != nil {
			log.Printf("Background command failed: %v, Output: %s", err, output)
		} else {
			log.Printf("Background command succeeded: %s", output)
		}
	}()

	resp := ContainerControlResponse{
		Success: true,
		Message: "Action triggered in background",
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("Error encoding response: %v", err)
	} else {
		log.Printf("Response sent successfully.")
	}
}
