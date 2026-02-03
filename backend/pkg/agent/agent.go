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

	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("error marshalling heartbeat: %w", err)
	}

	url := fmt.Sprintf("%s/api/v1/agent/heartbeat", mgmtURL)

	client := &http.Client{
		Transport: &http.Transport{
			Proxy: nil,
		},
		Timeout: 10 * time.Second,
	}

	resp, err := client.Post(url, "application/json", bytes.NewBuffer(data))
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
	return calculateCPUPercent(idle0, total0, idle1, total1), nil
}

func readProcStat() (uint64, uint64, error) {
	data, err := os.ReadFile("/proc/stat")
	if err != nil {
		return 0, 0, err
	}
	lines := strings.Split(string(data), "\n")
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

func calculateCPUPercent(idle0, total0, idle1, total1 uint64) float64 {
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
	var total, available float64
	lines := strings.Split(string(data), "\n")
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
	
	targets := []string{"vllm", "anysearch", "anyzearch", "anythingllm", "anything-llm", "milvus", "lancedb", "chroma", "pgvector", "mineru"}
	var services []DockerServiceStatus
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	
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

	dockerComposeFile := "/home/anyadmin/docker/docker-compose.yaml"

	// Validate docker-compose file existence
	if _, err := os.Stat(dockerComposeFile); os.IsNotExist(err) {
		http.Error(w, fmt.Sprintf("docker-compose file not found at %s", dockerComposeFile), http.StatusInternalServerError)
		return
	}

	var cmd *exec.Cmd

	switch req.Action {
	case "start":
		cmd = exec.Command("docker", "compose", "-f", dockerComposeFile, "start", req.ContainerName)
	case "stop":
		cmd = exec.Command("docker", "compose", "-f", dockerComposeFile, "stop", req.ContainerName)
	case "restart":
		cmd = exec.Command("docker", "compose", "-f", dockerComposeFile, "restart", req.ContainerName)
	default:
		http.Error(w, "Unknown action. Supported: start, stop, restart", http.StatusBadRequest)
		return
	}

	output, err := cmd.CombinedOutput()
	
	resp := ContainerControlResponse{
		Success: err == nil,
		Output:  string(output),
	}
	if err != nil {
		resp.Message = err.Error()
		log.Printf("Command failed: %v, Output: %s", err, output)
	} else {
		resp.Message = "Success"
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
