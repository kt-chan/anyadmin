package api

import (
	"net/http"

	"anyadmin-backend/pkg/global"
	"anyadmin-backend/pkg/middleware"
	"anyadmin-backend/pkg/service"
	"anyadmin-backend/pkg/utils"
	"github.com/gin-gonic/gin"
)

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func GetPublicKey(c *gin.Context) {
	key, err := utils.GetPublicKeyContent()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load public key"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"publicKey": key})
}

func Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	var user *global.User
	utils.ExecuteRead(func() {
		for i := range utils.Users {
			// Access by index to avoid loop variable address issues
			u := &utils.Users[i]
			if u.Username == req.Username {
				// Decrypt stored password
				storedPass, err := utils.DecryptPassword(u.Password)
				if err != nil {
					// If stored password is not encrypted (legacy), use as is
					storedPass = u.Password
				}

				// Decrypt incoming password
				incomingPass, err := utils.DecryptPassword(req.Password)
				if err != nil {
					// If incoming password is not encrypted (e.g. from tests or legacy client), use as is
					incomingPass = req.Password
				}

				if storedPass == incomingPass {
					user = u
					break
				}
			}
		}
	})

	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户名或密码错误"})
		return
	}

	token, err := middleware.GenerateToken(user.Username, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "无法生成令牌"})
		return
	}

	service.RecordLog(user.Username, "用户登录", "成功登录管理后台", "Info")

	c.JSON(http.StatusOK, gin.H{
		"token": token,
		"user": gin.H{
			"username": user.Username,
			"role":     user.Role,
		},
	})
}
