package api

import (
	"net/http"

	"anyadmin-backend/pkg/global"
	"anyadmin-backend/pkg/utils"
	"github.com/gin-gonic/gin"
)

func GetUsers(c *gin.Context) {
	utils.Mu.Lock()
	defer utils.Mu.Unlock()
	c.JSON(http.StatusOK, utils.Users)
}

func CreateUser(c *gin.Context) {
	var user global.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	utils.Mu.Lock()
	utils.Users = append(utils.Users, user)
	utils.Mu.Unlock()
	c.JSON(http.StatusOK, user)
}

func DeleteUser(c *gin.Context) {
	// Mock delete
	c.JSON(http.StatusOK, gin.H{"message": "用户已删除"})
}
