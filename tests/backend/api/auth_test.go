package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"anyadmin-backend/pkg/api"
	"anyadmin-backend/pkg/global"
	"anyadmin-backend/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupAuthRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	// Mock routes
	r.POST("/api/v1/login", api.Login)
	r.GET("/api/v1/public-key", api.GetPublicKey)
	return r
}

func TestAuth(t *testing.T) {
	// Setup keys and data
	// Note: utils.InitData() will be called by main usually, here we mock it
	// But utils package has init() that finds keys.
	// We need to ensure Users are initialized with encrypted passwords.
	
	// Mock users
	utils.ExecuteWrite(func() {
		pass, _ := utils.EncryptPassword("password")
		utils.Users = []global.User{
			{
				Username: "testadmin",
				Password: pass,
				Role:     "admin",
			},
		}
	}, false)

	router := setupAuthRouter()

	t.Run("GetPublicKey", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/public-key", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.Nil(t, err)
		assert.NotEmpty(t, response["publicKey"])
	})

	t.Run("Login Success with Encrypted Password", func(t *testing.T) {
		// Encrypt password
		encryptedPass, err := utils.EncryptPassword("password")
		assert.Nil(t, err)

		loginReq := api.LoginRequest{
			Username: "testadmin",
			Password: encryptedPass,
		}
		body, _ := json.Marshal(loginReq)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.NotEmpty(t, response["token"])
	})

	t.Run("Login Failure", func(t *testing.T) {
		encryptedPass, _ := utils.EncryptPassword("wrongpassword")
		loginReq := api.LoginRequest{
			Username: "testadmin",
			Password: encryptedPass,
		}
		body, _ := json.Marshal(loginReq)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}
