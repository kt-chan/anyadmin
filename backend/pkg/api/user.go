package api

import (
	"net/http"

	"anyadmin-backend/pkg/global"
	"anyadmin-backend/pkg/mockdata"
	"github.com/gin-gonic/gin"
)

func GetUsers(c *gin.Context) {
	mockdata.Mu.Lock()
	defer mockdata.Mu.Unlock()
	c.JSON(http.StatusOK, mockdata.Users)
}

func CreateUser(c *gin.Context) {
	var user global.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	mockdata.Mu.Lock()
	mockdata.Users = append(mockdata.Users, user)
	mockdata.Mu.Unlock()
	c.JSON(http.StatusOK, user)
}

func DeleteUser(c *gin.Context) {
	// Mock delete
	c.JSON(http.StatusOK, gin.H{"message": "用户已删除"})
}
