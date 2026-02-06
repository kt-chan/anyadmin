package global

import (
	"log"
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
	Host      string         `json:"host"`
	Port      string         `json:"port"`

	// AnythingLLM Specifics
	StorageDir                   string `json:"storage_dir"`
	LLMProvider                  string `json:"llm_provider"`
	GenericOpenAIBasePath        string `json:"generic_openai_base_path"`
	GenericOpenAIModelPref       string `json:"generic_openai_model_pref"`
	GenericOpenAIModelTokenLimit int    `json:"generic_openai_model_token_limit"`
	GenericOpenAIMaxTokens       int    `json:"generic_openai_max_tokens"`
	GenericOpenAIKey             string `json:"generic_openai_api_key"`
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
	InferenceHost  string `json:"inference_host"`
	InferencePort  string `json:"inference_port"`
	ModelName      string `json:"model_name"`
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
	ID     string `json:"id"`
	Name   string `json:"name"`
	Image  string `json:"image"`
	Status string `json:"status"`
	State  string `json:"state"`
	Uptime string `json:"uptime"`
}

func InitConfig() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.SetEnvPrefix("")
	viper.AutomaticEnv()
	viper.BindEnv("server.port", "SERVER_PORT")
	viper.SetDefault("server.port", "8080")

	if err := viper.ReadInConfig(); err != nil {
		log.Println("No config file found, using defaults")
	}
	ServerPort = viper.GetString("server.port")
	log.Printf("[Config] 端口配置: %s, 绑定地址: 0.0.0.0", ServerPort)
}
