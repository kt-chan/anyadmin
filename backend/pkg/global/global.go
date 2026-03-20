package global

import (
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/viper"
	"gorm.io/gorm"
)

var ServerPort string

type InferenceConfig struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deletedAt,omitempty"`
	Name      string         `gorm:"uniqueIndex" json:"name"`
	ModelType string         `json:"model_type"` // llm, vlm, asr, omni, embedding, reranker
	IsManaged bool           `json:"is_managed"`
	APIKey    string         `json:"api_key,omitempty"`
	BaseURL   string         `json:"base_url,omitempty"`
	Engine    string         `json:"engine"`
	ModelName string         `json:"model_name"`
	ModelPath string         `json:"model_path"`
	IP        string         `json:"ip"`
	Port      string         `json:"port"`

	// CalculateConfigParams
	Mode           string  `json:"mode"` // max_token, max_concurrency, balanced
	GPUMemoryGB    float64 `json:"gpu_memory_size"`
	GPUUtilization float64 `json:"gpu_utilization"`

	// Unified parameters
	MaxModelLen          int     `json:"max_model_len"`
	MaxNumSeqs           int     `json:"max_num_seqs"`
	MaxNumBatchedTokens  int     `json:"max_num_batched_tokens"`
	GpuMemoryUtilization float64 `json:"gpu_memory_utilization"`
}

type RagAppConfig struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deletedAt,omitempty"`
	Name      string         `json:"name"`
	IsManaged bool           `json:"is_managed"`
	Host      string         `json:"host"`
	Port      string         `json:"port"`

	// AnythingLLM Specifics
	StorageDir                   string `json:"storage_dir"`
	LLMProvider                  string `json:"llm_provider"`
	GenericOpenAIBasePath        string `json:"generic_open_ai_base_path"`
	GenericOpenAIModelPref       string `json:"generic_open_ai_model_pref"`
	GenericOpenAIModelTokenLimit int    `json:"generic_open_ai_model_token_limit"`
	GenericOpenAIMaxTokens       int    `json:"generic_open_ai_max_tokens"`
	GenericOpenAIKey             string `json:"generic_open_ai_api_key"`
	VectorDB                     string `json:"vector_db"`
}

type AgentConfig struct {
	DeploymentTime string `json:"deployment_time"`
	LogFile        string `json:"log_file"`
	MgmtHost       string `json:"mgmt_host"`
	MgmtPort       string `json:"mgmt_port"`
	NodeIP         string `json:"node_ip"`
	NodePort       string `json:"node_port"`
}

type DeploymentNode struct {
	NodeIP        string            `json:"node_ip"`
	Hostname      string            `json:"hostname"`
	AgentConfig   AgentConfig       `json:"agent_config"`
	InferenceCfgs []InferenceConfig `json:"inference_cfgs"`
	RagAppCfgs    []RagAppConfig    `json:"rag_app_cfgs"`
}

type ImportTask struct {
	ID         uint           `gorm:"primarykey" json:"id"`
	CreatedAt  time.Time      `json:"createdAt"`
	UpdatedAt  time.Time      `json:"updatedAt"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"deletedAt,omitempty"`
	Name       string         `json:"name"`
	SourceType string         `json:"sourceType"`
	SourcePath string         `json:"sourcePath"`
	Status     string         `json:"status"`
	Progress   int            `json:"progress"`
	TotalFiles int            `json:"totalFiles"`
	Processed  int            `json:"processed"`
	Message    string         `json:"message"`
}

type BackupRecord struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deletedAt,omitempty"`
	Name      string         `json:"name"`
	Path      string         `json:"path"`
	Size      int64          `json:"size"`
	Type      string         `json:"type"`
	Status    string         `json:"status"`
	Message   string         `json:"message"`
}

type User struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deletedAt,omitempty"`
	Username  string         `gorm:"uniqueIndex" json:"username"`
	Password  string         `json:"password"`
	Role      string         `json:"role"`
}

type OperationLog struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deletedAt,omitempty"`
	Username  string         `json:"username"`
	Action    string         `json:"action"`
	Detail    string         `json:"detail"`
	Level     string         `json:"level"`
}

type DeploymentConfig struct {
	MgmtHost       string `json:"mgmt_host"`
	MgmtPort       string `json:"mgmt_port"`
	TargetNodes    string `json:"target_nodes"`
	Mode           string `json:"mode"`
	Platform       string `json:"platform"`
	ModelType      string `json:"model_type"` // llm, vlm, asr, omni, embedding, reranker
	InferenceHost  string `json:"inference_host"`
	InferencePort  string `json:"inference_port"`
	ModelName      string `json:"model_name"`
	ServiceName    string `json:"service_name,omitempty"` // User defined instance name
	APIKey         string `json:"api_key,omitempty"`
	BaseURL        string `json:"base_url,omitempty"`
	EnableRAG      bool   `json:"enable_rag"`
	RAGHost        string `json:"rag_host,omitempty"`
	RAGPort        string `json:"rag_port,omitempty"`
	EnableVectorDB bool   `json:"enable_vectordb"`
	VectorDBType   string `json:"vector_db,omitempty"`
	VectorDBHost   string `json:"vectordb_host,omitempty"`
	VectorDBPort   string `json:"vectordb_port,omitempty"`
	EnableParser   bool   `json:"enable_parser"`
	ParserHost     string `json:"parser_host,omitempty"`
	ParserPort     string `json:"parser_port,omitempty"`
}

type DockerServiceStatus struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Image     string `json:"image"`
	Status    string `json:"status"`
	State     string `json:"state"`
	Uptime    string `json:"uptime"`
	ModelType string `json:"model_type,omitempty"`
	IsManaged bool   `json:"is_managed"`
}

type Model struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deletedAt,omitempty"`
	Name      string         `gorm:"uniqueIndex" json:"name"`
	ModelType string         `json:"model_type"`
	Size      int64          `json:"size"`
}

func InitConfig() {
	// 1. Try to load .env file
	// We prefer the project root .env first, then fallback to relative locations
	cwd, _ := os.Getwd()

	// List of potential .env paths in order of preference
	// 1. Project Root (assuming we are in anyadmin/backend or anyadmin/backend/cmd/server)
	// 2. Current directory
	envPaths := []string{
		filepath.Join(cwd, "..", ".env"),       // if in backend/
		filepath.Join(cwd, "..", "..", ".env"), // if in backend/cmd/server/
		filepath.Join(cwd, ".env"),             // if in root/
	}

	envLoaded := false
	for _, p := range envPaths {
		if _, err := os.Stat(p); err == nil {
			viper.SetConfigFile(p)
			if err := viper.MergeInConfig(); err == nil {
				log.Printf("[Config] Loaded .env from: %s", p)
				envLoaded = true
				break
			}
		}
	}

	if !envLoaded {
		log.Println("[Config] No .env file found, relying on system environment and defaults")
	}

	// 2. Load config.yaml if exists
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	if err := viper.MergeInConfig(); err != nil {
		// Ignore if config.yaml missing
	}

	viper.SetEnvPrefix("")
	viper.AutomaticEnv()

	// 3. Bind all environment variables from .env
	envVars := []string{
		"MgmtHost", "MgmtPort", "ADMIN_USERNAME", "ADMIN_PASSWORD",
		"REMOTE_USER", "REMOTE_HOST", "REMOTE_SSH_PORT", "REMOTE_BIN_DIR", "REMOTE_SRC_DIR",
		"VLLM_MODEL_PATH", "OPEN_API_KEY", "DEEPSEEK_API_KEY", "ZHIPU_API_KEY",
		"UID", "GID", "ANYTHINGLLM_STORAGE_DIR", "LLM_PROVIDER",
		"GENERIC_OPEN_AI_BASE_PATH", "GENERIC_OPEN_AI_MODEL_PREF",
		"GENERIC_OPEN_AI_MODEL_TOKEN_LIMIT", "GENERIC_OPEN_AI_MAX_TOKENS",
		"GENERIC_OPEN_AI_API_KEY", "VECTOR_DB", "DISABLE_TELEMETRY",
		"VLLM_LLM_PORT", "LLM_GPU_DEVICE_ID",
		"VLLM_GPU_MEMORY_UTILIZATION", "VLLM_MAX_MODEL_LEN",
		"VLLM_MAX_NUM_SEQS", "VLLM_MAX_NUM_BATCHED_TOKENS",
	}
	for _, v := range envVars {
		viper.BindEnv(v)
	}

	// Legacy and cross-service bindings
	viper.BindEnv("server.port", "MgmtPort")
	viper.BindEnv("admin.username", "ADMIN_USERNAME")
	viper.BindEnv("admin.password", "ADMIN_PASSWORD")
	viper.BindEnv("mgmt.host", "MgmtHost")
	viper.BindEnv("mgmt.port", "MgmtPort")

	// Set defaults
	viper.SetDefault("MgmtHost", "0.0.0.0")
	viper.SetDefault("MgmtPort", "8080")
	viper.SetDefault("ADMIN_USERNAME", "admin")
	viper.SetDefault("ADMIN_PASSWORD", "password")
	viper.SetDefault("VLLM_MAX_MODEL_LEN", 4096)
	viper.SetDefault("VLLM_MAX_NUM_SEQS", 8)
	viper.SetDefault("VLLM_MAX_NUM_BATCHED_TOKENS", 8192)
	viper.SetDefault("VLLM_GPU_MEMORY_UTILIZATION", 0.85)

	// Final settings
	// SERVER_PORT can still be used as an override if set in environment
	if os.Getenv("SERVER_PORT") != "" {
		ServerPort = os.Getenv("SERVER_PORT")
	} else {
		ServerPort = viper.GetString("MgmtPort")
	}
	log.Printf("[Config] 端口配置: %s, 绑定地址: 0.0.0.0", ServerPort)
}
