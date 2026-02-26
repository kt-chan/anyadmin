package api

import (
	"net/http"

	"anyadmin-backend/pkg/utils"
	"github.com/gin-gonic/gin"
)

func GetModelTypes(c *gin.Context) {
	utils.ExecuteRead(func() {
		c.JSON(http.StatusOK, utils.ModelTypes)
	})
}

func AddModelType(c *gin.Context) {
	var req struct {
		Type string `json:"type" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := utils.ExecuteWrite(func() {
		// Check if already exists
		for _, t := range utils.ModelTypes {
			if t == req.Type {
				return
			}
		}
		utils.ModelTypes = append(utils.ModelTypes, req.Type)
	}, true)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save model types"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Model type added", "types": utils.ModelTypes})
}

func DeleteModelType(c *gin.Context) {
	modelType := c.Param("type")
	if modelType == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Type is required"})
		return
	}

	err := utils.ExecuteWrite(func() {
		for i, t := range utils.ModelTypes {
			if t == modelType {
				utils.ModelTypes = append(utils.ModelTypes[:i], utils.ModelTypes[i+1:]...)
				break
			}
		}
	}, true)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete model type"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Model type deleted", "types": utils.ModelTypes})
}
