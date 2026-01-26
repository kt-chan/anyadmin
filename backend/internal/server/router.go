package server

import (
	"anyadmin-backend/internal/api"
	"anyadmin-backend/internal/middleware"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func NewRouter() *gin.Engine {
	r := gin.Default()

	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization", "Accept", "X-Requested-With"}
	config.ExposeHeaders = []string{"Content-Length"}
	config.MaxAge = 12 * 3600
	r.Use(cors.New(config))

	v1 := r.Group("/api/v1")
	{
		v1.POST("/login", api.Login)

		// 受 JWT 保护的路由
		auth := v1.Group("/")
		auth.Use(middleware.JWTAuth())
		{
			auth.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "UP"}) })
			auth.GET("/system/stats", api.GetSystemStats)
			auth.GET("/dashboard/stats", api.GetDashboardStats)
			auth.POST("/container/control", api.ControlContainer)
			auth.GET("/container/logs/:name", api.StreamLogs)
			auth.GET("/configs/inference", api.GetInferenceConfigs)
			auth.POST("/configs/inference", api.SaveInferenceConfig)
			auth.DELETE("/configs/inference/:id", api.DeleteInferenceConfig)
			auth.GET("/deploy/ssh-key", api.GetSystemSSHKey)
			auth.POST("/deploy/verify-ssh", api.VerifyNodeConnection)
			auth.POST("/deploy/generate", api.DeployService)
			auth.GET("/import/tasks", api.GetImportTasks)
			auth.POST("/import/tasks", api.CreateImportTask)
			auth.GET("/backups", api.GetBackups)
			auth.POST("/backups", api.CreateBackup)
			auth.POST("/backups/restore/:id", api.RestoreBackup)

			// 用户管理
			auth.GET("/users", api.GetUsers)
			auth.POST("/users", api.CreateUser)
			auth.DELETE("/users/:id", api.DeleteUser)
		}
	}

	return r
}
