package utils

import (
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
)

var (
	// DebugMode controls whether to output detailed logs
	DebugMode = false
)

func debugLog(format string, v ...interface{}) {
	if DebugMode {
		log.Printf("[DEBUG] "+format, v...)
	}
}

// ModelConfig 存储模型的基本配置
type ModelConfig struct {
	Name                  string  // 模型名称
	ParamsBillion         float64 // 模型参数（十亿）
	HiddenSize            int     // 隐藏层维度
	NumHiddenLayers       int     // 隐藏层数
	NumAttentionHeads     int     // 注意力头数
	MaxPositionEmbeddings int     // 最大位置编码（上下文长度）
	HeadDim               int     // 每个头的维度
	NumKeyValueHeads      int     // KV头数（用于GQA）
}

// GPUConfig 存储GPU配置
type GPUConfig struct {
	MemoryGB    float64 // GPU内存（GB）
	Utilization float64 // 内存利用率
	ReservedGB  float64 // 系统预留内存（GB）
}

// VLLMConfig vLLM配置输出
type VLLMConfig struct {
	MaxModelLen         int     `json:"max_model_len"`          // 最大模型长度
	MaxNumSeqs          int     `json:"max_num_seqs"`           // 最大并发序列数
	MaxNumBatchedTokens int     `json:"max_num_batched_tokens"` // 最大批处理tokens数
	GPUMemoryUtil       float64 `json:"gpu_memory_util"`        // GPU内存利用率
	SwapSpaceGB         int     `json:"swap_space_gb"`          // 交换空间（GB）
	EnablePrefixCaching bool    `json:"enable_prefix_caching"`  // 是否启用前缀缓存
	KVBlockSize         int     `json:"kv_block_size"`          // KV缓存块大小
}

// HuggingFaceConfig 从config.json读取的原始配置
type HuggingFaceConfig struct {
	Architectures         []string `json:"architectures,omitempty"`
	HiddenSize            int      `json:"hidden_size,omitempty"`
	NumHiddenLayers       int      `json:"num_hidden_layers,omitempty"`
	NumAttentionHeads     int      `json:"num_attention_heads,omitempty"`
	MaxPositionEmbeddings int      `json:"max_position_embeddings,omitempty"`
	IntermediateSize      int      `json:"intermediate_size,omitempty"`
	VocabSize             int      `json:"vocab_size,omitempty"`
	ModelType             string   `json:"model_type,omitempty"`
	TorchDtype            string   `json:"torch_dtype,omitempty"`
	NumKeyValueHeads      int      `json:"num_key_value_heads,omitempty"` // 对于GQA模型
	RopeTheta             float64  `json:"rope_theta,omitempty"`
	SlidingWindow         int      `json:"sliding_window,omitempty"` // 滑动窗口注意力
}

// CalculateConfigParams 计算配置的输入参数
type CalculateConfigParams struct {
	ModelNameOrPath string
	GPUMemoryGB     float64
	Mode            string
	GPUUtilization  float64
}

// CalculateVLLMConfig 计算vLLM配置 (API friendly version)
func CalculateVLLMConfig(params CalculateConfigParams) (VLLMConfig, ModelConfig, error) {
	// 默认利用率
	if params.GPUUtilization <= 0 {
		params.GPUUtilization = 0.85
	}

	// 如果内存小于8GB且利用率为默认值0.85，则保持一致
	if params.GPUMemoryGB < 8 && params.GPUUtilization == 0.85 {
		params.GPUUtilization = 0.85
	}

	// 尝试加载模型配置
	modelConfig := EstimateModelConfigFromName(params.ModelNameOrPath)

	// 确保模型名称正确
	if modelConfig.Name == "" {
		modelConfig.Name = params.ModelNameOrPath
	}

	// 创建GPU配置
	gpuConfig := GPUConfig{
		MemoryGB:    params.GPUMemoryGB,
		Utilization: params.GPUUtilization,
		ReservedGB:  1.5, // 默认预留1.5GB给系统
	}

	// 根据模式计算优化配置
	var vllmConfig VLLMConfig
	switch strings.ToLower(params.Mode) {
	case "max_token":
		vllmConfig = CalculateMaxTokenConfig(modelConfig, gpuConfig)
	case "max_concurrency":
		vllmConfig = CalculateMaxConcurrencyConfig(modelConfig, gpuConfig)
	case "balanced":
		vllmConfig = CalculateBalancedConfig(modelConfig, gpuConfig)
	default:
		vllmConfig = CalculateBalancedConfig(modelConfig, gpuConfig)
	}

	// 总是启用前缀缓存（对性能有益）
	vllmConfig.EnablePrefixCaching = true

	return vllmConfig, modelConfig, nil
}

func main() {
	// 解析命令行参数
	model := flag.String("model", "", "模型名称或本地路径")
	gpuMemoryStr := flag.String("gpu_memory", "8G", "GPU内存")
	mode := flag.String("mode", "balanced", "优化模式")
	utilization := flag.Float64("utilization", 0.85, "GPU内存利用率")
	enableSwap := flag.Bool("enable_swap", false, "是否启用交换空间")

	flag.Parse()

	if *model == "" {
		os.Exit(1)
	}

	gpuMemoryGB, _ := parseMemoryString(*gpuMemoryStr)

	params := CalculateConfigParams{
		ModelNameOrPath: *model,
		GPUMemoryGB:     gpuMemoryGB,
		Mode:            *mode,
		GPUUtilization:  *utilization,
	}

	vllmConfig, modelConfig, err := CalculateVLLMConfig(params)
	if err != nil {
		log.Fatalf("计算配置失败: %v", err)
	}

	if *enableSwap && gpuMemoryGB < 16 {
		vllmConfig.SwapSpaceGB = 8
	}

	gpuConfig := GPUConfig{
		MemoryGB:    gpuMemoryGB,
		Utilization: *utilization,
		ReservedGB:  1.5,
	}

	printConfig(modelConfig, gpuConfig, vllmConfig, *mode)
	printVLLMCommand(*model, vllmConfig)
}

// parseMemoryString 解析内存字符串
func parseMemoryString(memoryStr string) (float64, error) {
	re := regexp.MustCompile(`^(\d+(?:\.\d+)?)\s*([GT]?B?)$`)
	matches := re.FindStringSubmatch(strings.ToUpper(memoryStr))

	if matches == nil {
		return 0, fmt.Errorf("无效的内存格式")
	}

	value, _ := strconv.ParseFloat(matches[1], 64)
	unit := matches[2]
	if strings.Contains(unit, "T") {
		value *= 1024
	}

	return value, nil
}

func estimateModelParams(hfConfig HuggingFaceConfig) float64 {
	if hfConfig.HiddenSize == 0 || hfConfig.NumHiddenLayers == 0 {
		return extractParamsFromName(hfConfig.ModelType)
	}
	vocabSize := hfConfig.VocabSize
	if vocabSize == 0 {
		vocabSize = 100000
	}
	hiddenSize := float64(hfConfig.HiddenSize)
	numLayers := float64(hfConfig.NumHiddenLayers)
	embeddingParams := float64(vocabSize) * hiddenSize
	transformerParamsPerLayer := 12.0 * hiddenSize * hiddenSize
	transformerParams := numLayers * transformerParamsPerLayer
	return (embeddingParams + transformerParams) / 1e9
}

func extractParamsFromName(modelName string) float64 {
	modelNameLower := strings.ToLower(modelName)
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`(\d+(?:\.\d+)?)[Bb]`),
		regexp.MustCompile(`-(\d+)b`),
		regexp.MustCompile(`(\d+)[Bb]`),
		regexp.MustCompile(`(\d+(?:\.\d+)?)b`),
		regexp.MustCompile(`qwen3-(\d+(?:\.\d+)?)b`),
	}
	for _, pattern := range patterns {
		matches := pattern.FindStringSubmatch(modelNameLower)
		if matches != nil {
			params, err := strconv.ParseFloat(matches[1], 64)
			if err == nil {
				return params
			}
		}
	}
	return 7.0
}

func EstimateModelConfigFromName(modelName string) ModelConfig {
	paramsBillion := extractParamsFromName(modelName)
	var hiddenSize, numLayers, numHeads, maxPosEmbeddings int
	switch {
	case paramsBillion <= 1.0:
		hiddenSize, numLayers, numHeads, maxPosEmbeddings = 1024, 12, 12, 8192
	case paramsBillion <= 1.7:
		hiddenSize, numLayers, numHeads, maxPosEmbeddings = 1536, 16, 12, 131072
	case paramsBillion <= 3.0:
		hiddenSize, numLayers, numHeads, maxPosEmbeddings = 2048, 24, 16, 32768
	case paramsBillion <= 7.0:
		hiddenSize, numLayers, numHeads, maxPosEmbeddings = 4096, 32, 32, 32768
	case paramsBillion <= 13.0:
		hiddenSize, numLayers, numHeads, maxPosEmbeddings = 5120, 40, 40, 32768
	case paramsBillion <= 34.0:
		hiddenSize, numLayers, numHeads, maxPosEmbeddings = 8192, 60, 64, 131072
	default:
		hiddenSize, numLayers, numHeads, maxPosEmbeddings = 8192, 80, 64, 131072
	}
	return ModelConfig{
		Name:                  modelName,
		ParamsBillion:         paramsBillion,
		HiddenSize:            hiddenSize,
		NumHiddenLayers:       numLayers,
		NumAttentionHeads:     numHeads,
		MaxPositionEmbeddings: maxPosEmbeddings,
		HeadDim:               hiddenSize / numHeads,
		NumKeyValueHeads:      numHeads,
	}
}

func calculateModelWeightMemory(model ModelConfig) float64 {
	return model.ParamsBillion * 2.0
}

func calculateKVCachePerToken(model ModelConfig) float64 {
	kvChannels := model.NumKeyValueHeads
	if kvChannels == 0 {
		kvChannels = model.NumAttentionHeads
	}
	safetyFactor := 1.20
	return 2.0 * float64(model.NumHiddenLayers) * float64(kvChannels) * float64(model.HeadDim) * 2.0 * safetyFactor
}

// CalculateMaxTokenConfig 计算最大化序列长度的配置
func CalculateMaxTokenConfig(model ModelConfig, gpu GPUConfig) VLLMConfig {
	availableMemoryGB := gpu.MemoryGB * gpu.Utilization
	weightMemoryGB := calculateModelWeightMemory(model)
	systemReservedGB := 1.5 // Standard reservation

	// Fix 1: Add a safety buffer (0.85) to account for fragmentation/activation overhead
	kvCacheMemoryGB := (availableMemoryGB - weightMemoryGB - systemReservedGB) * 0.85

	if kvCacheMemoryGB < 0.5 {
		kvCacheMemoryGB = 0.5
	}

	kvCachePerTokenBytes := calculateKVCachePerToken(model)
	kvCachePerTokenGB := kvCachePerTokenBytes / (1024 * 1024 * 1024)

	// Total physical tokens we can fit in RAM
	totalTokenCapacity := int(kvCacheMemoryGB / kvCachePerTokenGB)

	// Calculate MaxModelLen
	maxTokens := totalTokenCapacity
	if model.MaxPositionEmbeddings > 0 && maxTokens > model.MaxPositionEmbeddings {
		maxTokens = model.MaxPositionEmbeddings
	}

	// Hard floor
	if maxTokens < 2048 {
		maxTokens = 2048
	}

	// Align to 128 for hardware efficiency
	maxTokens = (maxTokens / 128) * 128

	// Fix 2: Strict check against capacity again after alignment
	if maxTokens > totalTokenCapacity {
		maxTokens = (totalTokenCapacity / 128) * 128
	}

	maxConcurrency := 1
	if kvCacheMemoryGB > 4.0 {
		maxConcurrency = 2
	}

	// Fix 3: MaxNumBatchedTokens logic
	// It must be at least equal to MaxModelLen to process a full context prompt
	batchTokens := maxTokens

	// Ensure batch tokens doesn't exceed physical limit
	if batchTokens > totalTokenCapacity {
		batchTokens = totalTokenCapacity
	}

	return VLLMConfig{
		MaxModelLen:         maxTokens,
		MaxNumSeqs:          maxConcurrency,
		MaxNumBatchedTokens: batchTokens,
		GPUMemoryUtil:       gpu.Utilization,
		KVBlockSize:         16,
	}
}

// CalculateMaxConcurrencyConfig 计算最大化并发数的配置
func CalculateMaxConcurrencyConfig(model ModelConfig, gpu GPUConfig) VLLMConfig {
	availableMemoryGB := gpu.MemoryGB * gpu.Utilization
	weightMemoryGB := calculateModelWeightMemory(model)
	systemReservedGB := 1.5

	// Safety buffer
	kvCacheMemoryGB := (availableMemoryGB - weightMemoryGB - systemReservedGB) * 0.85
	if kvCacheMemoryGB < 1.0 {
		kvCacheMemoryGB = 1.0
	}

	kvCachePerTokenBytes := calculateKVCachePerToken(model)
	kvCachePerTokenGB := kvCachePerTokenBytes / (1024 * 1024 * 1024)
	totalTokenCapacity := int(kvCacheMemoryGB / kvCachePerTokenGB)

	seqLen := 4096
	if gpu.MemoryGB < 8 {
		seqLen = 2048
	} else if gpu.MemoryGB > 16 {
		seqLen = 8192
	}

	// Clamp seqLen to model limits
	if model.MaxPositionEmbeddings > 0 && seqLen > model.MaxPositionEmbeddings {
		seqLen = model.MaxPositionEmbeddings
	}

	// Fix: Clamp seqLen to physical capacity
	if seqLen > totalTokenCapacity {
		seqLen = totalTokenCapacity
	}

	// Alignment
	seqLen = (seqLen / 128) * 128

	maxConcurrency := totalTokenCapacity / seqLen
	if maxConcurrency < 1 {
		maxConcurrency = 1
	} else if maxConcurrency > 256 {
		maxConcurrency = 256
	}

	// Fix: Batched tokens calculation
	batchTokens := seqLen * maxConcurrency

	// Strict clamp: Batch tokens cannot exceed total physical tokens
	if batchTokens > totalTokenCapacity {
		batchTokens = totalTokenCapacity
	}

	// vLLM recommendation: batch tokens should be at least max_model_len
	if batchTokens < seqLen {
		batchTokens = seqLen
	}

	return VLLMConfig{
		MaxModelLen:         seqLen,
		MaxNumSeqs:          maxConcurrency,
		MaxNumBatchedTokens: batchTokens,
		GPUMemoryUtil:       gpu.Utilization,
		KVBlockSize:         16,
	}
}

// CalculateBalancedConfig 计算平衡配置
func CalculateBalancedConfig(model ModelConfig, gpu GPUConfig) VLLMConfig {
	availableMemoryGB := gpu.MemoryGB * gpu.Utilization
	weightMemoryGB := calculateModelWeightMemory(model)
	systemReservedGB := 1.5

	// Fix 1: Safety buffer of 0.85 (15% headroom for activation overhead/fragmentation)
	kvCacheMemoryGB := (availableMemoryGB - weightMemoryGB - systemReservedGB) * 0.85

	if kvCacheMemoryGB < 1.0 {
		kvCacheMemoryGB = 1.0
	}

	kvCachePerTokenBytes := calculateKVCachePerToken(model)
	kvCachePerTokenGB := kvCachePerTokenBytes / (1024 * 1024 * 1024)

	// This is the hard physical limit of tokens
	totalTokenCapacity := int(kvCacheMemoryGB / kvCachePerTokenGB)

	var seqLen int
	if gpu.MemoryGB < 6 {
		seqLen = 2048
	} else if gpu.MemoryGB < 12 {
		seqLen = 4096
	} else if gpu.MemoryGB < 24 {
		seqLen = 8192
	} else {
		seqLen = 16384
	}

	if model.MaxPositionEmbeddings > 0 && seqLen > model.MaxPositionEmbeddings {
		seqLen = model.MaxPositionEmbeddings
	}

	// Fix 2: STRICT Clamping. If capacity is 17680, seqLen cannot be 32768
	if seqLen > totalTokenCapacity {
		seqLen = totalTokenCapacity
	}

	// Align to 256
	seqLen = (seqLen / 256) * 256

	// Determine concurrency
	maxConcurrency := totalTokenCapacity / seqLen
	if maxConcurrency > 8 {
		maxConcurrency = 8
	} else if maxConcurrency < 2 {
		maxConcurrency = 2
	}

	// Fix 3: Calculate batch tokens
	batchTokens := seqLen * 2

	// CRITICAL FIX for your specific error:
	// "max-num-batched-tokens is still far too large"
	// We must ensure batchTokens never exceeds totalTokenCapacity
	if batchTokens > totalTokenCapacity {
		batchTokens = totalTokenCapacity
	}

	// Ensure we can at least process one full sequence
	if batchTokens < seqLen {
		batchTokens = seqLen
	}

	// Cap at 32k or 16k if needed by model, but hardware limit (totalTokenCapacity) takes precedence
	if batchTokens > 32768 {
		batchTokens = 32768
	}

	return VLLMConfig{
		MaxModelLen:         seqLen,
		MaxNumSeqs:          maxConcurrency,
		MaxNumBatchedTokens: batchTokens,
		GPUMemoryUtil:       gpu.Utilization,
		KVBlockSize:         16,
	}
}

func printConfig(model ModelConfig, gpu GPUConfig, vllm VLLMConfig, mode string) {
	fmt.Println("╔══════════════════════════════════════════════════════════╗")
	fmt.Println("║                 vLLM 配置优化工具                        ║")
	fmt.Println("╠══════════════════════════════════════════════════════════╣")
	fmt.Printf("║ 模型: %-50s ║\n", model.Name)
	fmt.Printf("║ GPU内存: %.1fGB | 模式: %-15s | 利用率: %.2f ║\n",
		gpu.MemoryGB, mode, gpu.Utilization)
	fmt.Println("╠══════════════════════════════════════════════════════════╣")
	fmt.Printf("║ --max-model-len          %-10d (最大上下文长度)       ║\n", vllm.MaxModelLen)
	fmt.Printf("║ --max-num-seqs           %-10d (最大并发请求数)       ║\n", vllm.MaxNumSeqs)
	fmt.Printf("║ --max-num-batched-tokens %-10d (批处理tokens数)      ║\n", vllm.MaxNumBatchedTokens)
	fmt.Printf("║ --gpu-memory-utilization %-10.2f (GPU内存利用率)      ║\n", vllm.GPUMemoryUtil)
	fmt.Println("╚══════════════════════════════════════════════════════════╝")
}

func printVLLMCommand(model string, config VLLMConfig) {
	fmt.Println("\n🚀 生成的 vLLM 启动命令:")
	fmt.Printf("vllm serve %s \\\n", model)
	fmt.Printf("    --max-model-len %d \\\n", config.MaxModelLen)
	fmt.Printf("    --max-num-seqs %d \\\n", config.MaxNumSeqs)
	fmt.Printf("    --max-num-batched-tokens %d \\\n", config.MaxNumBatchedTokens)
	fmt.Printf("    --gpu-memory-utilization %.2f\n", config.GPUMemoryUtil)
}
