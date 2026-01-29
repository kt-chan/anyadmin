package global

import (
	"log"
	"time"

	"github.com/spf13/viper"
	"gorm.io/gorm"
)

var ServerPort string

type InferenceConfig struct {
	ID        uint `gorm:"primarykey" json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deletedAt,omitempty"`
	Name           string  `gorm:"uniqueIndex" json:"name"`
	Engine         string  `json:"engine"`
	ModelPath      string  `json:"modelPath"`
	IP             string  `json:"ip"`
	Port           string  `json:"port"`
	MaxConcurrency int     `json:"maxConcurrency"`
	TokenLimit     int     `json:"tokenLimit"`
	BatchSize      int     `json:"batchSize"`
	GpuMemory      float64 `json:"gpuMemory"`
}

type ImportTask struct {
	ID        uint `gorm:"primarykey" json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deletedAt,omitempty"`
	Name       string `json:"name"`
	SourceType string `json:"sourceType"`
	SourcePath string `json:"sourcePath"`
	Status     string `json:"status"`
	Progress   int    `json:"progress"`
	TotalFiles int    `json:"totalFiles"`
	Processed  int    `json:"processed"`
	Message    string `json:"message"`
}

type BackupRecord struct {
	ID        uint `gorm:"primarykey" json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deletedAt,omitempty"`
	Name    string `json:"name"`
	Path    string `json:"path"`
	Size    int64  `json:"size"`
	Type    string `json:"type"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

type User struct {
	ID        uint `gorm:"primarykey" json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deletedAt,omitempty"`
	Username string `gorm:"uniqueIndex" json:"username"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

type OperationLog struct {
	ID        uint `gorm:"primarykey" json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deletedAt,omitempty"`
	Username string `json:"username"`
	Action   string `json:"action"`
	Detail   string `json:"detail"`
	Level    string `json:"level"`
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
