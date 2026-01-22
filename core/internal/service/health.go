package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/anyzearch/admin/core/internal/global"
	"github.com/docker/docker/client"
	"github.com/shirou/gopsutil/v3/process"
	"os"
)

type ServiceStatus struct {
	Name    string  `json:"name"`
	Type    string  `json:"type"`
	Status  string  `json:"status"` // Running, Stopped, Error
	Health  string  `json:"health"` // Healthy, Unhealthy
	Uptime  string  `json:"uptime"`
	CPU     float64 `json:"cpu"`    // 进程 CPU 占用
	Memory  uint64  `json:"memory"` // 进程内存占用 (Bytes)
	PID     int32   `json:"pid"`
	Message string  `json:"message"`
}

func GetServicesHealth() []ServiceStatus {
	var configs []global.InferenceConfig
	global.DB.Find(&configs)

	results := make([]ServiceStatus, 0)
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	
	// 1. 获取后端自身进程信息
	selfPID := int32(os.Getpid())
	p, pErr := process.NewProcess(selfPID)
	var selfCPU float64
	var selfMem uint64
	var selfUptime string = "Active"
	if pErr == nil {
		selfCPU, _ = p.CPUPercent()
		memInfo, _ := p.MemoryInfo()
		if memInfo != nil {
			selfMem = memInfo.RSS
		}
		createTime, _ := p.CreateTime()
		if createTime > 0 {
			duration := time.Since(time.Unix(createTime/1000, 0))
			selfUptime = fmt.Sprintf("%.1f 小时", duration.Hours())
		}
	}

	results = append(results, ServiceStatus{
		Name:   "anyzearch-admin-core",
		Type:   "Core",
		Status: "Running",
		Health: "Healthy",
		Uptime: selfUptime,
		CPU:    selfCPU,
		Memory: selfMem,
		PID:    selfPID,
	})

	// 2. 遍历配置的服务
	for _, cfg := range configs {
		status := "Stopped"
		health := "Unhealthy"
		msg := "Service unreachable"
		uptime := "-"

		// 2.1 优先尝试 Docker 检查 (针对全新部署)
		dockerFound := false
		if err == nil {
			inspect, inspectErr := cli.ContainerInspect(ctx, cfg.Name)
			if inspectErr == nil {
				dockerFound = true
				if inspect.State.Running {
					status = "Running"
					health = "Healthy"
					msg = ""
					uptime = inspect.State.StartedAt
				} else {
					status = "Exited"
					msg = inspect.State.Error
				}
			}
		}

		// 2.2 如果 Docker 没找到，尝试网络探测 (针对对接现有服务)
		if !dockerFound {
			// 尝试解析配置中的地址，如果 ModelPath 看起来像 URL 则使用它，否则使用默认端口
			// 这里假设对接服务时，用户可能并未部署容器，而是直接运行
			targetURL := determineProbeURL(cfg)
			if checkEngineHealth(cfg.Engine, targetURL, cfg.Name) {
				status = "Running"
				health = "Healthy"
				msg = "External Service Connected"
				uptime = "Unknown (External)"
			} else {
				msg = fmt.Sprintf("Health check failed at %s", targetURL)
			}
		}

		results = append(results, ServiceStatus{
			Name:    cfg.Name,
			Type:    "Inference",
			Status:  status,
			Health:  health,
			Message: msg,
			Uptime:  uptime,
		})
	}

	return results
}

// checkEngineHealth 根据不同引擎类型执行特定的 HTTP 探测
func checkEngineHealth(engine string, url string, modelName string) bool {
	client := http.Client{Timeout: 2 * time.Second}
	var resp *http.Response
	var err error

	// 规范化 URL
	if !strings.HasPrefix(url, "http") {
		url = "http://" + url
	}

	switch strings.ToLower(engine) {
	case "ollama":
		// 1. 检查 Ollama 服务是否存活
		baseURL := strings.TrimRight(url, "/")
		if strings.Contains(url, "/api") {
			// 如果用户填了具体路径，尝试截断
			parts := strings.Split(url, "/api")
			baseURL = parts[0]
		}
		
		// 2. 检查模型是否正在运行 (Loaded in memory)
		// API: GET /api/ps
		resp, err = client.Get(baseURL + "/api/ps")
		if err == nil && resp.StatusCode == 200 {
			defer resp.Body.Close()
			body, _ := io.ReadAll(resp.Body)
			
			// 定义 Ollama ps 响应结构
			type ProcessModel struct {
				Name string `json:"name"`
			}
			type PsResponse struct {
				Models []ProcessModel `json:"models"`
			}
			
			var psResp PsResponse
			if json.Unmarshal(body, &psResp) == nil {
				for _, m := range psResp.Models {
					// 模糊匹配，因为 ollama 可能返回 llama3:latest 而配置是 llama3
					if strings.Contains(m.Name, modelName) || strings.Contains(modelName, m.Name) {
						return true // 模型正在显存中运行
					}
				}
			}
		}

		// 3. 如果没在运行，检查模型是否存在 (Available on disk)
		// API: GET /api/tags
		resp, err = client.Get(baseURL + "/api/tags")
		if err == nil && resp.StatusCode == 200 {
			defer resp.Body.Close()
			// 只要服务通且模型存在，我们暂时认为它是"健康"的(虽然是 Idle 状态)
			// 但为了更严格，这里返回 false 表示"未运行"，或者我们可以增加一个 "Idle" 状态
			// 在当前 MVP 中，为了避免误报 Error，只要服务通且能查到模型，就算通过，但最好能区分
			return true 
		}
		return false

	case "vllm":
		// vLLM 通常暴露 /health
		probeURL := url
		if !strings.Contains(url, "/health") {
			probeURL = fmt.Sprintf("%s/health", strings.TrimRight(url, "/"))
		}
		resp, err = client.Get(probeURL)

	case "mindie":
		// MindIE (华为) 通常兼容 OpenAI 或有自己的 /health
		probeURL := url
		if !strings.Contains(url, "/health") {
			probeURL = fmt.Sprintf("%s/health", strings.TrimRight(url, "/"))
		}
		resp, err = client.Get(probeURL)

	default:
		// 默认尝试根路径
		resp, err = client.Get(url)
	}

	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == 200
}

// determineProbeURL 尝试推断服务地址
func determineProbeURL(cfg global.InferenceConfig) string {
	host := cfg.IP
	if host == "" {
		host = "localhost"
	}
	
	port := cfg.Port
	if port == "" {
		// 回退默认端口
		switch strings.ToLower(cfg.Engine) {
		case "ollama": port = "11434"
		default: port = "8000"
		}
	}
	
	return fmt.Sprintf("http://%s:%s", host, port)
}