package service

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/anyzearch/admin/core/internal/global"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

func GenerateAndStart(config global.InferenceConfig) (string, string, error) {
	path, err := GenerateCompose(config)
	if err != nil {
		return "", "", err
	}

	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return path, "", fmt.Errorf("docker 客户端初始化失败: %v", err)
	}

	image := "vllm/vllm-openai:latest"
	if config.Engine == "MindIE" {
		image = "ascend-mindie:latest"
	}

	fmt.Printf("[Docker] 正在清理旧容器: %s\n", config.Name)
	cli.ContainerRemove(ctx, config.Name, container.RemoveOptions{Force: true})

	// 准备 GPU 配置
	var deviceRequests []container.DeviceRequest
	if config.Engine == "vLLM" {
		deviceRequests = []container.DeviceRequest{
			{
				Driver:       "nvidia",
				Count:        -1, // 所有的 GPU
				Capabilities: [][]string{{"gpu"}},
			},
		}
	}

	gpuUtil := fmt.Sprintf("%.1f", config.GpuMemory)
	if config.GpuMemory <= 0 {
		gpuUtil = "0.9"
	}

	hostPort := config.Port
	if hostPort == "" {
		hostPort = "8000"
	}

	fmt.Printf("[Docker] 正在创建容器并分配 GPU: %s, 映射端口: %s\n", config.Name, hostPort)
	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: image,
		Cmd: []string{
			"--model", "/model",
			"--max-model-len", fmt.Sprintf("%d", config.TokenLimit),
			"--trust-remote-code",
			"--gpu-memory-utilization", gpuUtil,
			"--max-num-seqs", fmt.Sprintf("%d", config.MaxConcurrency),
		},
		ExposedPorts: nat.PortSet{"8000/tcp": struct{}{}},
	}, &container.HostConfig{
		Binds: []string{
			fmt.Sprintf("%s:/model", config.ModelPath),
		},
		PortBindings: nat.PortMap{
			"8000/tcp": []nat.PortBinding{{HostIP: "0.0.0.0", HostPort: hostPort}},
		},
		Resources: container.Resources{
			DeviceRequests: deviceRequests,
		},
	}, nil, nil, config.Name)
	if err != nil {
		return path, "", fmt.Errorf("容器创建失败: %v", err)
	}

	fmt.Printf("[Docker] 启动容器 ID: %s\n", resp.ID)
	if err := cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return path, resp.ID, fmt.Errorf("容器启动失败: %v", err)
	}

	fmt.Printf("[Docker] 部署任务完成，容器已拉起\n")
	return path, resp.ID, nil
}

func GenerateCompose(config global.InferenceConfig) (string, error) {
	var image string
	var gpuConfig string

	if config.Engine == "vLLM" {
		image = "vllm/vllm-openai:latest"
		gpuConfig = `
    deploy:
      resources:
        reservations:
          devices:
            - driver: nvidia
              count: all
              capabilities: [gpu]`
	} else if config.Engine == "MindIE" {
		image = "ascend-mindie:latest"
	}

	composeTemplate := `version: '3.8'
services:
  inference:
    image: %s
    container_name: %s
    ports:
      - "%s:8000"
    volumes:
      - %s:/model
    command: --model /model --max-model-len %d --trust-remote-code
    %s
`
	hostPort := config.Port
	if hostPort == "" {
		hostPort = "8000"
	}
	content := fmt.Sprintf(composeTemplate, image, config.Name, hostPort, config.ModelPath, config.TokenLimit, gpuConfig)
	outputPath := filepath.Join("deployments", config.Name)
	os.MkdirAll(outputPath, 0755)
	filePath := filepath.Join(outputPath, "docker-compose.yml")
	err := os.WriteFile(filePath, []byte(content), 0644)
	return filePath, err
}