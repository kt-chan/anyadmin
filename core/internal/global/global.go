package global

import (
	"log"

	"github.com/spf13/viper"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB
var ServerPort string

type InferenceConfig struct {
	gorm.Model
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
	gorm.Model
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
	gorm.Model
	Name    string `json:"name"`
	Path    string `json:"path"`
	Size    int64  `json:"size"`
	Type    string `json:"type"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

type User struct {
	gorm.Model
	Username string `gorm:"uniqueIndex" json:"username"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

type OperationLog struct {
	gorm.Model
	Username string `json:"username"`
	Action   string `json:"action"`
	Detail   string `json:"detail"`
	Level    string `json:"level"`
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

func InitDB() {
	var err error
	DB, err = gorm.Open(sqlite.Open("anyzearch.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	DB.AutoMigrate(&InferenceConfig{}, &ImportTask{}, &BackupRecord{}, &User{}, &OperationLog{})

	var count int64
	DB.Model(&User{}).Count(&count)
	if count == 0 {
		DB.Create(&User{Username: "admin", Password: "admin", Role: "admin"})
	}
}
