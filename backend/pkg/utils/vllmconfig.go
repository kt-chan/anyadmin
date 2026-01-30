package utils

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

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
		params.GPUUtilization = 0.9
	}

	// 如果内存小于8GB且利用率为默认值0.9，则调低至0.85以更保守
	if params.GPUMemoryGB < 8 && params.GPUUtilization == 0.9 {
		params.GPUUtilization = 0.85
	}

	// 尝试加载模型配置
	modelConfig, err := loadModelConfig(params.ModelNameOrPath)
	if err != nil {
		log.Printf("警告: 无法从配置读取模型参数: %v", err)
		log.Println("将尝试从模型名称估算参数...")
		modelConfig = EstimateModelConfigFromName(params.ModelNameOrPath)
	}

	// 确保模型名称正确
	if modelConfig.Name == "" {
		modelConfig.Name = params.ModelNameOrPath
	}

	// 创建GPU配置
	gpuConfig := GPUConfig{
		MemoryGB:    params.GPUMemoryGB,
		Utilization: params.GPUUtilization,
		ReservedGB:  1.0, // 默认预留1GB给系统
	}

	// 根据模式计算优化配置
	var vllmConfig VLLMConfig
	switch strings.ToLower(params.Mode) {
	case "max_token":
		vllmConfig = CalculateMaxTokenConfig(modelConfig, gpuConfig)
	case "max_concurrency":
		vllmConfig = CalculateMaxConcurrencyConfig(modelConfig, gpuConfig)
	case "balanced":
		fallthrough
	default:
		vllmConfig = CalculateBalancedConfig(modelConfig, gpuConfig)
	}

	// 总是启用前缀缓存（对性能有益）
	vllmConfig.EnablePrefixCaching = true

	return vllmConfig, modelConfig, nil
}

func main() {
	// 解析命令行参数
	model := flag.String("model", "", "模型名称或本地路径（格式：author/model-name 或 /path/to/model）")
	gpuMemoryStr := flag.String("gpu_memory", "8G", "GPU内存（如：8G、16G、24G）")
	mode := flag.String("mode", "balanced", "优化模式：max_token（最大长度）、max_concurrency（最大并发）、balanced（平衡）")
	utilization := flag.Float64("utilization", 0.9, "GPU内存利用率（0.0-1.0，小显存<8G时自动调整为0.85）")
	enableSwap := flag.Bool("enable_swap", false, "是否启用交换空间")

	flag.Parse()

	if *model == "" {
		fmt.Println("错误: 必须指定 --model 参数")
		fmt.Println("用法示例:")
		fmt.Println("  vllm-optimizer.exe --model \"D:\\models\\huggingface\\hub\\Qwen3-1.7B\" --gpu_memory 8G")
		fmt.Println("  vllm-optimizer.exe --model Qwen/Qwen3-1.7B --gpu_memory 8G")
		os.Exit(1)
	}

	// 解析GPU内存字符串
	gpuMemoryGB, err := parseMemoryString(*gpuMemoryStr)
	if err != nil {
		log.Fatalf("解析GPU内存错误: %v", err)
	}

	// 使用公共函数计算
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

	// 如果启用交换空间，设置交换空间大小 (CLI特有逻辑)
	if *enableSwap && gpuMemoryGB < 16 {
		vllmConfig.SwapSpaceGB = 8
	}

	// 创建GPU配置用于打印 (Recover GPU Config for printing)
	gpuConfig := GPUConfig{
		MemoryGB:    gpuMemoryGB,
		Utilization: *utilization,
		ReservedGB:  1.0,
	}

	// 输出配置
	printConfig(modelConfig, gpuConfig, vllmConfig, *mode)

	// 输出vLLM命令行
	printVLLMCommand(*model, vllmConfig)
}

// parseMemoryString 解析内存字符串（如 "8G" -> 8.0）
func parseMemoryString(memoryStr string) (float64, error) {
	re := regexp.MustCompile(`^(\d+(?:\.\d+)?)\s*([GT]?B?)$`)
	matches := re.FindStringSubmatch(strings.ToUpper(memoryStr))

	if matches == nil {
		return 0, fmt.Errorf("无效的内存格式: %s，请使用如 '8G', '16GB', '24G' 的格式", memoryStr)
	}

	value, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return 0, err
	}

	unit := matches[2]
	if strings.Contains(unit, "T") {
		value *= 1024 // TB转GB
	}

	return value, nil
}

// loadModelConfig 尝试从模型配置文件中读取模型配置
func loadModelConfig(modelPath string) (ModelConfig, error) {
	var config ModelConfig

	// 尝试不同的配置路径
	var configPaths []string

	// 首先，直接使用提供的路径作为目录，查找config.json
	configPaths = append(configPaths, filepath.Join(modelPath, "config.json"))

	// 如果路径包含"/"，可能是相对路径，尝试在当前目录下查找
	if strings.Contains(modelPath, "/") || strings.Contains(modelPath, "\\") {
		// 已经是路径格式，保留原样
	} else {
		// 可能是模型名称，尝试在HuggingFace缓存中查找
		home, err := os.UserHomeDir()
		if err == nil {
			// 尝试标准HuggingFace缓存路径
			hfCachePath := filepath.Join(home, ".cache", "huggingface", "hub")
			modelCachePath := strings.ReplaceAll(modelPath, "/", "--")
			configPaths = append(configPaths,
				filepath.Join(hfCachePath, "models--"+modelCachePath, "snapshots", "latest", "config.json"),
				filepath.Join(hfCachePath, "models--"+modelCachePath, "config.json"),
			)
		}

		// 尝试 backend/deployments/models 目录 (For Project Structure)
		configPaths = append(configPaths, filepath.Join("backend", "deployments", "models", modelPath, "config.json"))
		configPaths = append(configPaths, filepath.Join("..", "backend", "deployments", "models", modelPath, "config.json"))
		// Fix for running tests from pkg/utils
		configPaths = append(configPaths, filepath.Join("..", "..", "deployments", "models", modelPath, "config.json"))
		// 如果运行在backend目录
		configPaths = append(configPaths, filepath.Join("deployments", "models", modelPath, "config.json"))
		// 如果运行在根目录
		configPaths = append(configPaths, filepath.Join("models", modelPath, "config.json"))

		// 也尝试在当前目录下查找
		configPaths = append(configPaths, filepath.Join(".", modelPath, "config.json"))
	}

	// 尝试读取配置文件
	var configData []byte
	var configFile string
	var lastError error

	for _, path := range configPaths {
		log.Printf("尝试读取配置文件: %s", path)
		if data, err := os.ReadFile(path); err == nil {
			configData = data
			configFile = path
			log.Printf("成功从 %s 读取配置文件", configFile)
			break
		} else {
			lastError = err
			log.Printf("读取 %s 失败: %v", path, err)
		}
	}

	if configData == nil {
		return config, fmt.Errorf("无法找到模型配置文件，最后错误: %v", lastError)
	}

	// 解析JSON配置
	var hfConfig HuggingFaceConfig
	if err := json.Unmarshal(configData, &hfConfig); err != nil {
		return config, fmt.Errorf("解析配置文件失败: %v", err)
	}

	// 填充模型配置
	config.HiddenSize = hfConfig.HiddenSize
	config.NumHiddenLayers = hfConfig.NumHiddenLayers
	config.NumAttentionHeads = hfConfig.NumAttentionHeads
	config.MaxPositionEmbeddings = hfConfig.MaxPositionEmbeddings
	config.NumKeyValueHeads = hfConfig.NumKeyValueHeads

	// 尝试从路径中提取名称 (e.g. "Qwen/Qwen2.5-7B" from path)
	if modelPath != "" {
		// If modelPath looks like a path (contains slashes), use the last part as name
		baseName := filepath.Base(modelPath)
		if baseName == "latest" || baseName == "." {
			// If path ends in 'latest' or '.', try the parent
			baseName = filepath.Base(filepath.Dir(modelPath))
		}
		// Clean up "models--" prefix from huggingface cache
		baseName = strings.TrimPrefix(baseName, "models--")
		baseName = strings.ReplaceAll(baseName, "--", "/")
		
		config.Name = baseName
	}

	// 计算每个头的维度
	if hfConfig.NumAttentionHeads > 0 {
		config.HeadDim = hfConfig.HiddenSize / hfConfig.NumAttentionHeads
	} else {
		config.HeadDim = 128 // 默认值
	}

	// 估算模型参数大小（十亿）
	config.ParamsBillion = estimateModelParams(hfConfig)

	log.Printf("模型配置: hidden_size=%d, num_hidden_layers=%d, num_attention_heads=%d, max_position_embeddings=%d, params=%.1fB",
		config.HiddenSize, config.NumHiddenLayers, config.NumAttentionHeads,
		config.MaxPositionEmbeddings, config.ParamsBillion)

	return config, nil
}

// estimateModelParams 根据模型配置估算参数数量（十亿）
func estimateModelParams(hfConfig HuggingFaceConfig) float64 {
	// 如果没有足够信息，尝试从模型名称中提取
	if hfConfig.HiddenSize == 0 || hfConfig.NumHiddenLayers == 0 {
		return extractParamsFromName(hfConfig.ModelType)
	}

	// 使用公式估算：参数 ≈ vocab_size * hidden_size + num_layers * (12 * hidden_size^2)
	vocabSize := hfConfig.VocabSize
	if vocabSize == 0 {
		vocabSize = 100000 // 默认词汇表大小
	}

	hiddenSize := float64(hfConfig.HiddenSize)
	numLayers := float64(hfConfig.NumHiddenLayers)

	// 估算嵌入层参数
	embeddingParams := float64(vocabSize) * hiddenSize

	// 估算Transformer层参数（每层约12*hidden_size^2）
	// 这是近似公式：每层有自注意力层（4*h^2）和前馈层（8*h^2）
	transformerParamsPerLayer := 12.0 * hiddenSize * hiddenSize
	transformerParams := numLayers * transformerParamsPerLayer

	// 总参数（转换为十亿）
	totalParams := (embeddingParams + transformerParams) / 1e9

	// 四舍五入到一位小数
	return totalParams
}

// extractParamsFromName 从模型名称中提取参数大小（十亿）
func extractParamsFromName(modelName string) float64 {
	modelNameLower := strings.ToLower(modelName)

	// 常见模型名称模式
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`(\d+(?:\.\d+)?)[Bb]`),    // 匹配 "1.7B", "7B", "13B"
		regexp.MustCompile(`-(\d+)b`),                // 匹配 "-7b", "-13b"
		regexp.MustCompile(`(\d+)[Bb]`),              // 匹配 "7B", "13B"
		regexp.MustCompile(`(\d+(?:\.\d+)?)b`),       // 匹配 "1.7b", "7b"
		regexp.MustCompile(`qwen3-(\d+(?:\.\d+)?)b`), // 匹配 "qwen3-1.7b"
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

	// 根据常见模型名称猜测
	if strings.Contains(modelNameLower, "tiny") {
		return 0.1
	} else if strings.Contains(modelNameLower, "small") {
		return 0.3
	} else if strings.Contains(modelNameLower, "medium") {
		return 1.5
	} else if strings.Contains(modelNameLower, "large") {
		return 7.0
	} else if strings.Contains(modelNameLower, "xlarge") {
		return 13.0
	} else if strings.Contains(modelNameLower, "2b") || strings.Contains(modelNameLower, "2.7b") {
		return 2.7
	} else if strings.Contains(modelNameLower, "3b") {
		return 3.0
	} else if strings.Contains(modelNameLower, "6b") || strings.Contains(modelNameLower, "6.7b") {
		return 6.7
	} else if strings.Contains(modelNameLower, "8b") || strings.Contains(modelNameLower, "7b") {
		return 7.0
	} else if strings.Contains(modelNameLower, "13b") {
		return 13.0
	} else if strings.Contains(modelNameLower, "34b") || strings.Contains(modelNameLower, "32b") {
		return 34.0
	} else if strings.Contains(modelNameLower, "70b") {
		return 70.0
	}

	// Qwen系列的特殊处理
	if strings.Contains(modelNameLower, "qwen3") {
		if strings.Contains(modelNameLower, "0.5") {
			return 0.5
		} else if strings.Contains(modelNameLower, "1.5") {
			return 1.5
		} else if strings.Contains(modelNameLower, "1.7") {
			return 1.7
		} else if strings.Contains(modelNameLower, "4") {
			return 4.0
		} else if strings.Contains(modelNameLower, "7") {
			return 7.0
		} else if strings.Contains(modelNameLower, "14") {
			return 14.0
		} else if strings.Contains(modelNameLower, "32") {
			return 32.0
		} else if strings.Contains(modelNameLower, "72") {
			return 72.0
		}
	}

	return 7.0 // 默认值
}

// EstimateModelConfigFromName 从模型名称估算模型配置
func EstimateModelConfigFromName(modelName string) ModelConfig {
	paramsBillion := extractParamsFromName(modelName)

	// 根据参数大小估算配置
	var hiddenSize, numLayers, numHeads, maxPosEmbeddings int

	switch {
	case paramsBillion <= 1.0:
		hiddenSize, numLayers, numHeads, maxPosEmbeddings = 1024, 12, 12, 8192
	case paramsBillion <= 1.7:
		hiddenSize, numLayers, numHeads, maxPosEmbeddings = 1536, 16, 12, 131072 // Qwen3-1.7B的配置
	case paramsBillion <= 3.0:
		hiddenSize, numLayers, numHeads, maxPosEmbeddings = 2048, 24, 16, 32768
	case paramsBillion <= 7.0:
		hiddenSize, numLayers, numHeads, maxPosEmbeddings = 4096, 32, 32, 32768
	case paramsBillion <= 13.0:
		hiddenSize, numLayers, numHeads, maxPosEmbeddings = 5120, 40, 40, 32768
	case paramsBillion <= 34.0:
		hiddenSize, numLayers, numHeads, maxPosEmbeddings = 8192, 60, 64, 131072
	default: // 70B+
		hiddenSize, numLayers, numHeads, maxPosEmbeddings = 8192, 80, 64, 131072
	}

	// Qwen系列的特殊处理
	modelNameLower := strings.ToLower(modelName)
	if strings.Contains(modelNameLower, "qwen") {
		if strings.Contains(modelNameLower, "qwen3") {
			// Qwen3系列通常支持长上下文
			maxPosEmbeddings = 131072
		} else {
			maxPosEmbeddings = 32768
		}
	}

	return ModelConfig{
		ParamsBillion:         paramsBillion,
		HiddenSize:            hiddenSize,
		NumHiddenLayers:       numLayers,
		NumAttentionHeads:     numHeads,
		MaxPositionEmbeddings: maxPosEmbeddings,
		HeadDim:               hiddenSize / numHeads,
		NumKeyValueHeads:      numHeads, // 默认与注意力头数相同
	}
}

// calculateModelWeightMemory 计算模型权重内存（GB，FP16精度）
func calculateModelWeightMemory(model ModelConfig) float64 {
	// FP16精度，每个参数2字节
	return model.ParamsBillion * 2.0
}

// calculateKVCachePerToken 计算每个token的KV缓存大小（字节）
func calculateKVCachePerToken(model ModelConfig) float64 {
	// KV缓存大小 = 2 * num_layers * kv_channels * head_dim * 2 (bytes, for float16)
	// 注意：对于GQA模型，kv_channels可能小于num_attention_heads
	kvChannels := model.NumKeyValueHeads
	if kvChannels == 0 {
		kvChannels = model.NumAttentionHeads
	}

	return 2.0 * float64(model.NumHiddenLayers) * float64(kvChannels) * float64(model.HeadDim) * 2.0
}

// CalculateMaxTokenConfig 计算最大化序列长度的配置
func CalculateMaxTokenConfig(model ModelConfig, gpu GPUConfig) VLLMConfig {
	// 可用GPU内存（GB）
	availableMemoryGB := gpu.MemoryGB * gpu.Utilization

	// 模型权重内存（GB）
	weightMemoryGB := calculateModelWeightMemory(model)

	// 预留系统内存（GB）
	systemReservedGB := gpu.ReservedGB

	// 可用于KV缓存的内存（GB）
	kvCacheMemoryGB := availableMemoryGB - weightMemoryGB - systemReservedGB

	if kvCacheMemoryGB < 0.5 { // 至少需要0.5GB用于KV缓存
		log.Printf("警告: GPU内存不足，KV缓存可用内存仅 %.2fGB", kvCacheMemoryGB)
		kvCacheMemoryGB = 0.5
	}

	// 计算每个token的KV缓存（字节）
	kvCachePerTokenBytes := calculateKVCachePerToken(model)

	// 转换为GB
	kvCachePerTokenGB := kvCachePerTokenBytes / (1024 * 1024 * 1024)

	// 计算最大支持的token数
	maxTokens := int(kvCacheMemoryGB / kvCachePerTokenGB)

	// 限制不超过模型原生最大上下文长度
	if model.MaxPositionEmbeddings > 0 && maxTokens > model.MaxPositionEmbeddings {
		maxTokens = model.MaxPositionEmbeddings
	}

	// 确保至少有最小长度
	minTokens := 2048
	if maxTokens < minTokens {
		maxTokens = minTokens
	}

	// 设置并发数（最大化长度时，并发数较低）
	maxConcurrency := 1
	if kvCacheMemoryGB > 4.0 { // 如果有足够内存，可以稍微增加并发
		maxConcurrency = 2
	}

	// 计算批处理大小（基于序列长度）
	batchTokens := maxTokens
	if batchTokens > 32768 {
		batchTokens = 32768 // 限制批处理大小
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
	// 可用GPU内存（GB）
	availableMemoryGB := gpu.MemoryGB * gpu.Utilization

	// 模型权重内存（GB）
	weightMemoryGB := calculateModelWeightMemory(model)

	// 预留系统内存（GB）
	systemReservedGB := gpu.ReservedGB

	// 可用于KV缓存的内存（GB）
	kvCacheMemoryGB := availableMemoryGB - weightMemoryGB - systemReservedGB

	if kvCacheMemoryGB < 1.0 { // 至少需要1GB用于KV缓存
		log.Printf("警告: GPU内存不足，KV缓存可用内存仅 %.2fGB", kvCacheMemoryGB)
		kvCacheMemoryGB = 1.0
	}

	// 计算每个token的KV缓存（GB）
	kvCachePerTokenBytes := calculateKVCachePerToken(model)
	kvCachePerTokenGB := kvCachePerTokenBytes / (1024 * 1024 * 1024)

	// 设置合理的序列长度（针对并发优化）
	seqLen := 4096
	if gpu.MemoryGB < 8 {
		seqLen = 2048
	} else if gpu.MemoryGB > 16 {
		seqLen = 8192
	}

	// 限制不超过模型原生最大长度
	if model.MaxPositionEmbeddings > 0 && seqLen > model.MaxPositionEmbeddings {
		seqLen = model.MaxPositionEmbeddings
	}

	// 计算最大并发数
	kvCachePerSeqGB := kvCachePerTokenGB * float64(seqLen)
	maxConcurrency := int(kvCacheMemoryGB / kvCachePerSeqGB)

	// 确保最小和最大并发数
	if maxConcurrency < 1 {
		maxConcurrency = 1
	} else if maxConcurrency > 256 {
		maxConcurrency = 256 // vLLM默认最大值
	}

	// 计算批处理大小（基于并发数）
	batchTokens := seqLen * maxConcurrency
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

// CalculateBalancedConfig 计算平衡配置
func CalculateBalancedConfig(model ModelConfig, gpu GPUConfig) VLLMConfig {
	// 可用GPU内存（GB）
	availableMemoryGB := gpu.MemoryGB * gpu.Utilization

	// 模型权重内存（GB）
	weightMemoryGB := calculateModelWeightMemory(model)

	// 预留系统内存（GB）
	systemReservedGB := gpu.ReservedGB

	// 可用于KV缓存的内存（GB）
	kvCacheMemoryGB := availableMemoryGB - weightMemoryGB - systemReservedGB

	if kvCacheMemoryGB < 1.0 {
		log.Printf("警告: GPU内存不足，KV缓存可用内存仅 %.2fGB", kvCacheMemoryGB)
		kvCacheMemoryGB = 1.0
	}

	// 计算每个token的KV缓存（GB）
	kvCachePerTokenBytes := calculateKVCachePerToken(model)
	kvCachePerTokenGB := kvCachePerTokenBytes / (1024 * 1024 * 1024)

	// 根据GPU内存确定序列长度
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

	// 限制不超过模型原生最大长度
	if model.MaxPositionEmbeddings > 0 && seqLen > model.MaxPositionEmbeddings {
		seqLen = model.MaxPositionEmbeddings
	}

	// 计算可以支持的并发数
	kvCachePerSeqGB := kvCachePerTokenGB * float64(seqLen)
	maxConcurrency := int(kvCacheMemoryGB / kvCachePerSeqGB)

	// 调整并发数以获得平衡
	if maxConcurrency > 8 {
		maxConcurrency = 8
	} else if maxConcurrency < 2 {
		maxConcurrency = 2
	}

	// 计算批处理大小
	batchTokens := seqLen * 2 // 平衡模式下，批处理大小为2个序列
	if batchTokens > 16384 {
		batchTokens = 16384
	}

	return VLLMConfig{
		MaxModelLen:         seqLen,
		MaxNumSeqs:          maxConcurrency,
		MaxNumBatchedTokens: batchTokens,
		GPUMemoryUtil:       gpu.Utilization,
		KVBlockSize:         16,
	}
}

// printConfig 打印配置详情
func printConfig(model ModelConfig, gpu GPUConfig, vllm VLLMConfig, mode string) {
	fmt.Println("╔══════════════════════════════════════════════════════════╗")
	fmt.Println("║                 vLLM 配置优化工具                        ║")
	fmt.Println("╠══════════════════════════════════════════════════════════╣")
	fmt.Printf("║ 模型: %-50s ║\n", model.Name)
	fmt.Printf("║ GPU内存: %.1fGB | 模式: %-15s | 利用率: %.2f ║\n",
		gpu.MemoryGB, mode, gpu.Utilization)
	fmt.Println("╠══════════════════════════════════════════════════════════╣")

	// 模型参数信息
	fmt.Println("║                        模型信息                          ║")
	fmt.Println("╠══════════════════════════════════════════════════════════╣")
	fmt.Printf("║ • 参数大小:        %6.1f B                        ║\n", model.ParamsBillion)
	fmt.Printf("║ • 隐藏层维度:      %6d                          ║\n", model.HiddenSize)
	fmt.Printf("║ • 层数:            %6d                          ║\n", model.NumHiddenLayers)
	fmt.Printf("║ • 注意力头数:      %6d                          ║\n", model.NumAttentionHeads)
	if model.NumKeyValueHeads > 0 {
		fmt.Printf("║ • KV头数:          %6d                          ║\n", model.NumKeyValueHeads)
	}
	if model.MaxPositionEmbeddings > 0 {
		fmt.Printf("║ • 原生最大长度:    %6d                          ║\n", model.MaxPositionEmbeddings)
	}

	// 计算内存使用详情
	weightMemoryGB := calculateModelWeightMemory(model)
	kvCachePerTokenBytes := calculateKVCachePerToken(model)
	kvCachePerTokenGB := kvCachePerTokenBytes / (1024 * 1024 * 1024)
	totalKVCacheGB := kvCachePerTokenGB * float64(vllm.MaxModelLen) * float64(vllm.MaxNumSeqs)
	availableMemoryGB := gpu.MemoryGB * gpu.Utilization
	otherUsageGB := availableMemoryGB - weightMemoryGB - totalKVCacheGB

	// 确保otherUsageGB不为负数
	if otherUsageGB < 0 {
		otherUsageGB = 0
	}

	fmt.Println("╠══════════════════════════════════════════════════════════╣")
	fmt.Println("║                        内存分配                          ║")
	fmt.Println("╠══════════════════════════════════════════════════════════╣")
	fmt.Printf("║ • 模型权重:         %6.2f GB (FP16)                 ║\n", weightMemoryGB)
	fmt.Printf("║ • KV缓存/Token:     %6.2f KB                        ║\n", kvCachePerTokenBytes/1024)
	fmt.Printf("║ • 总KV缓存:         %6.2f GB                        ║\n", totalKVCacheGB)
	fmt.Printf("║ • 系统及其他:       %6.2f GB                        ║\n", otherUsageGB)

	totalUsedGB := weightMemoryGB + totalKVCacheGB + otherUsageGB
	usagePercent := (totalUsedGB / gpu.MemoryGB) * 100
	fmt.Printf("║ • 总使用:           %6.2f GB / %5.1f GB (%.0f%%)      ║\n",
		totalUsedGB, gpu.MemoryGB, usagePercent)

	fmt.Println("╠══════════════════════════════════════════════════════════╣")
	fmt.Println("║                   推荐 vLLM 参数                         ║")
	fmt.Println("╠══════════════════════════════════════════════════════════╣")
	fmt.Printf("║ --max-model-len          %-10d (最大上下文长度)       ║\n", vllm.MaxModelLen)
	fmt.Printf("║ --max-num-seqs           %-10d (最大并发请求数)       ║\n", vllm.MaxNumSeqs)
	fmt.Printf("║ --max-num-batched-tokens %-10d (批处理tokens数)      ║\n", vllm.MaxNumBatchedTokens)
	fmt.Printf("║ --gpu-memory-utilization %-10.2f (GPU内存利用率)      ║\n", vllm.GPUMemoryUtil)

	if vllm.SwapSpaceGB > 0 {
		fmt.Printf("║ --swap-space            %-10d (交换空间GB)         ║\n", vllm.SwapSpaceGB)
	}

	if vllm.EnablePrefixCaching {
		fmt.Printf("║ --enable-prefix-caching                (启用前缀缓存) ║\n")
	}

	fmt.Println("╚══════════════════════════════════════════════════════════╝")
}

// printVLLMCommand 输出完整的vLLM命令
func printVLLMCommand(model string, config VLLMConfig) {
	fmt.Println("\n🚀 生成的 vLLM 启动命令:")
	fmt.Printf("vllm serve %s \\\n", model)
	fmt.Printf("    --max-model-len %d \\\n", config.MaxModelLen)
	fmt.Printf("    --max-num-seqs %d \\\n", config.MaxNumSeqs)
	fmt.Printf("    --max-num-batched-tokens %d \\\n", config.MaxNumBatchedTokens)
	fmt.Printf("    --gpu-memory-utilization %.2f", config.GPUMemoryUtil)

	if config.SwapSpaceGB > 0 {
		fmt.Printf(" \\\n    --swap-space %d", config.SwapSpaceGB)
	}

	if config.EnablePrefixCaching {
		fmt.Printf(" \\\n    --enable-prefix-caching")
	}

	fmt.Println()

	// 使用建议
	fmt.Println("💡 使用建议:")
	if config.MaxModelLen > 32768 {
		fmt.Println("• 你配置了超长上下文(>32K)，建议使用流式响应避免超时")
		fmt.Println("• 考虑启用 --enable-chunked-prefill 参数以更好地处理长序列")
	}
	if config.MaxNumSeqs < 4 {
		fmt.Println("• 并发数较低，适合处理少量长文档任务")
		fmt.Println("• 对于批量处理，考虑增加 --swap-space 或减少序列长度")
	} else if config.MaxNumSeqs > 16 {
		fmt.Println("• 高并发配置，适合聊天API服务")
		fmt.Println("• 监控GPU内存使用，避免OOM错误")
	} else {
		fmt.Println("• 并发数适中，适合通用API服务")
	}
	if config.SwapSpaceGB > 0 {
		fmt.Println("• 已启用交换空间，当GPU内存不足时会使用系统内存")
		fmt.Println("  注意：这会显著降低性能，仅作为备用方案")
	}

	// 性能优化建议
	fmt.Println("\n🔧 性能优化建议:")
	if config.MaxNumBatchedTokens > 32768 {
		fmt.Println("• 考虑降低 --max-num-batched-tokens 以改善TTFT（首次token时间）")
	}
	if config.MaxNumSeqs > 8 && config.MaxModelLen > 8192 {
		fmt.Println("• 长上下文+高并发可能压力较大，考虑使用 --quantization awq 量化")
	}
}
