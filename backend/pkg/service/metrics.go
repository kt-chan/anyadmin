package service

type DeviceStats struct {
	Index    int     `json:"index"`
	Model    string  `json:"model"`
	Usage    float64 `json:"usage"`    // 算力利用率
	MemUsed  uint64  `json:"memUsed"`
	MemTotal uint64  `json:"memTotal"`
}

type SystemStats struct {
	CPUUsage    float64       `json:"cpuUsage"`
	MemoryTotal uint64        `json:"memoryTotal"`
	MemoryUsed  uint64        `json:"memoryUsed"`
	MemoryFree  uint64        `json:"memoryFree"`
	DiskTotal   uint64        `json:"diskTotal"`
	DiskUsed    uint64        `json:"diskUsed"`
	OS          string        `json:"os"`
	Arch        string        `json:"arch"`
	GPUUsage    float64       `json:"gpuUsage"`    // 总平均算力利用率
	GPUMemUsed  uint64        `json:"gpuMemUsed"`  // 总已用显存
	GPUMemTotal uint64        `json:"gpuMemTotal"` // 总显存
	GPUDevices  []DeviceStats `json:"gpuDevices"`  // NVIDIA 设备列表
	NPUUsage    float64       `json:"npuUsage"`    // 总平均 NPU 利用率
	NPUMemUsed  uint64        `json:"npuMemUsed"`
	NPUMemTotal uint64        `json:"npuMemTotal"`
	NPUDevices  []DeviceStats `json:"npuDevices"`  // 昇腾设备列表
}

func GetSystemStats() (*SystemStats, error) {
	// Mock System Stats
	return &SystemStats{
		CPUUsage:    15.5,
		MemoryTotal: 32 * 1024 * 1024 * 1024,
		MemoryUsed:  14 * 1024 * 1024 * 1024,
		MemoryFree:  18 * 1024 * 1024 * 1024,
		DiskTotal:   512 * 1024 * 1024 * 1024,
		DiskUsed:    128 * 1024 * 1024 * 1024,
		OS:          "linux",
		Arch:        "amd64",
		GPUUsage:    0,
		GPUMemUsed:  0,
		GPUMemTotal: 0,
		GPUDevices:  nil,
		NPUUsage:    45.0,
		NPUMemUsed:  16 * 1024 * 1024 * 1024,
		NPUMemTotal: 32 * 1024 * 1024 * 1024,
		NPUDevices: []DeviceStats{
			{
				Index:    0,
				Model:    "Ascend 910",
				Usage:    45.0,
				MemUsed:  16 * 1024 * 1024 * 1024,
				MemTotal: 32 * 1024 * 1024 * 1024,
			},
		},
	}, nil
}
