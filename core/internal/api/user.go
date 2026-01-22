package api

import (
	"net/http"

	"github.com/anyzearch/admin/core/internal/global"
	"github.com/gin-gonic/gin"
)

func GetUsers(c *gin.Context) {
	var users []global.User
	global.DB.Find(&users)
	c.JSON(http.StatusOK, users)
}

func CreateUser(c *gin.Context) {
	var user global.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	global.DB.Create(&user)
	c.JSON(http.StatusOK, user)
}

func DeleteUser(c *gin.Context) {
	id := c.Param("id")
	global.DB.Delete(&global.User{}, id)
	c.JSON(http.StatusOK, gin.H{"message": "用户已删除"})
}
