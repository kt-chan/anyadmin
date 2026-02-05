package api

import (
	"log"
	"net/http"
	"strconv"

	"anyadmin-backend/pkg/mockdata"
	"anyadmin-backend/pkg/service"
	"github.com/gin-gonic/gin"
)

type ControlRequest struct {
	Name   string `json:"name" binding:"required"`
	Action string `json:"action" binding:"required"` // start, stop, restart
	NodeIP string `json:"node_ip"`
}

func ControlContainer(c *gin.Context) {
	var req ControlRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Async execution
	go func() {
		if err := service.ControlContainer(req.Name, req.Action, req.NodeIP); err != nil {
			log.Printf("[Container] Async action %s failed for %s: %v", req.Action, req.Name, err)
		}
	}()

	username, _ := c.Get("username")
	service.RecordLog(username.(string), "容器控制", "对节点 "+req.NodeIP+" 的容器 "+req.Name+" 执行了 "+req.Action+" 操作", "Info")

	c.JSON(http.StatusOK, gin.H{"message": "Action triggered successfully"})
}

type UpdateVLLMConfigRequest struct {
	NodeIP string            `json:"node_ip" binding:"required"`
	Config map[string]string `json:"config" binding:"required"`
}

func UpdateVLLMConfig(c *gin.Context) {
	var req UpdateVLLMConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Persist to mockdata (and file)
	mockdata.Mu.Lock()
	updated := false
	for i, cfg := range mockdata.InferenceCfgs {
		if cfg.IP == req.NodeIP { // Assuming one config per IP for now, or use name if available
			if val, ok := req.Config["VLLM_MAX_MODEL_LEN"]; ok {
				if v, err := strconv.Atoi(val); err == nil {
					mockdata.InferenceCfgs[i].MaxModelLen = v
				}
			}
			if val, ok := req.Config["VLLM_GPU_MEMORY_UTILIZATION"]; ok {
				if v, err := strconv.ParseFloat(val, 64); err == nil {
					mockdata.InferenceCfgs[i].GpuMemoryUtilization = v
				}
			}
			if val, ok := req.Config["VLLM_MAX_NUM_SEQS"]; ok {
				if v, err := strconv.Atoi(val); err == nil {
					mockdata.InferenceCfgs[i].MaxNumSeqs = v
				}
			}
			if val, ok := req.Config["VLLM_MAX_NUM_BATCHED_TOKENS"]; ok {
				if v, err := strconv.Atoi(val); err == nil {
					mockdata.InferenceCfgs[i].MaxNumBatchedTokens = v
				}
			}
			updated = true
			break
		}
	}

	if updated {
		mockdata.Mu.Unlock()
		if err := mockdata.SaveToFile(); err != nil {
			log.Printf("[Container] Failed to save config to file: %v", err)
		}
	} else {
		mockdata.Mu.Unlock()
	}

	// Always restart for now as per requirement
	if err := service.UpdateVLLMConfig(req.NodeIP, req.Config, true); err != nil {
		log.Printf("[Container] Config update failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	username, _ := c.Get("username")
	service.RecordLog(username.(string), "配置更新", "更新了节点 "+req.NodeIP+" 的 vLLM 配置并重启服务", "Info")

	c.JSON(http.StatusOK, gin.H{"message": "Configuration updated and restart triggered"})
}
