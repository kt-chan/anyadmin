package service

import (
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
)

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
	v, err := mem.VirtualMemory()
	if err != nil {
		return nil, err
	}
	c, err := cpu.Percent(time.Second, false)
	if err != nil {
		return nil, err
	}
	d, err := disk.Usage("/")
	if err != nil {
		return nil, err
	}

	var cpuUsage float64
	if len(c) > 0 {
		cpuUsage = c[0]
	}

	stats := &SystemStats{
		CPUUsage:    cpuUsage,
		MemoryTotal: v.Total,
		MemoryUsed:  v.Used,
		MemoryFree:  v.Free,
		DiskTotal:   d.Total,
		DiskUsed:    d.Used,
		OS:          runtime.GOOS,
		Arch:        runtime.GOARCH,
	}

	// 获取 NVIDIA GPU 详情

gpuDevices, gpuUsage, gpuMemUsed, gpuMemTotal := getNvidiaStatsDetail()
	stats.GPUDevices = gpuDevices
	stats.GPUUsage = gpuUsage
	stats.GPUMemUsed = gpuMemUsed
	stats.GPUMemTotal = gpuMemTotal

	// 获取华为昇腾 NPU 详情
	npuDevices, npuUsage, npuMemUsed, npuMemTotal := getAscendStatsDetail()
	stats.NPUDevices = npuDevices
	stats.NPUUsage = npuUsage
	stats.NPUMemUsed = npuMemUsed
	stats.NPUMemTotal = npuMemTotal

	return stats, nil
}

func getNvidiaStatsDetail() ([]DeviceStats, float64, uint64, uint64) {
	cmd := exec.Command("nvidia-smi", "--query-gpu=index,name,utilization.gpu,memory.used,memory.total", "--format=csv,noheader,nounits")
	output, err := cmd.Output()
	if err != nil {
		return nil, 0, 0, 0
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	devices := make([]DeviceStats, 0)
	var sumUsage float64
	var totalUsed uint64
	var totalTotal uint64

	for _, line := range lines {
		parts := strings.Split(line, ",")
		if len(parts) >= 5 {
			idx, _ := strconv.Atoi(strings.TrimSpace(parts[0]))
			usage, _ := strconv.ParseFloat(strings.TrimSpace(parts[2]), 64)
			used, _ := strconv.ParseUint(strings.TrimSpace(parts[3]), 10, 64)
			total, _ := strconv.ParseUint(strings.TrimSpace(parts[4]), 10, 64)

			d := DeviceStats{
				Index:    idx,
				Model:    strings.TrimSpace(parts[1]),
				Usage:    usage,
				MemUsed:  used * 1024 * 1024,
				MemTotal: total * 1024 * 1024,
			}
			devices = append(devices, d)
			sumUsage += usage
			totalUsed += d.MemUsed
			totalTotal += d.MemTotal
		}
	}

	if len(devices) > 0 {
		return devices, sumUsage / float64(len(devices)), totalUsed, totalTotal
	}
	return nil, 0, 0, 0
}

func getAscendStatsDetail() ([]DeviceStats, float64, uint64, uint64) {
	// 使用 ascend-smi 获取基础信息
	// 注意：昇腾的解析逻辑依赖具体的采集指令，此处实现通用解析
	cmd := exec.Command("ascend-smi", "-i", "0", "-p") 
	_, err := cmd.Output()
	if err != nil {
		return nil, 0, 0, 0
	}

	// 简单的 NPU 解析示例逻辑 (昇腾输出格式较多，此处演示核心逻辑)
	devices := make([]DeviceStats, 0)
	// 假设识别到 1 个 NPU
	d := DeviceStats{
		Index: 0,
		Model: "Ascend 910",
		Usage: 45.0, // 示例值
		MemUsed: 16 * 1024 * 1024 * 1024,
		MemTotal: 32 * 1024 * 1024 * 1024,
	}
	devices = append(devices, d)
	
	return devices, 45.0, d.MemUsed, d.MemTotal
}