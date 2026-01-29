package api

import (
	"net/http"
	"strconv"
	"strings"

	"anyadmin-backend/pkg/service"
	"github.com/gin-gonic/gin"
)

// GetSystemSSHKey returns the public key for passwordless login setup
func GetSystemSSHKey(c *gin.Context) {
	pubKey, err := service.GetPublicKey()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get system key: " + err.Error()})
		return
	}
	// Return as text/plain so it can be downloaded or displayed easily
	c.Data(http.StatusOK, "text/plain", []byte(pubKey))
}

type VerifySSHRequest struct {
	Type string `json:"type"` // "ssh", "inference", etc.
	Host string `json:"host"` // For SSH: newline separated list of IP:Port
	Port int    `json:"port"` // Optional, default 22
}

// VerifyNodeConnection checks if the system can SSH into the target nodes
func VerifyNodeConnection(c *gin.Context) {
	var req VerifySSHRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": err.Error()})
		return
	}

	if req.Type != "ssh" {
		// Pass through to other verification logic if needed, but for now we only handle SSH here
		// or maybe this endpoint is shared? 
		// The frontend calls `/deployment/api/test-connection`.
		// We should probably check if we need to merge this with existing logic.
		// Since I am creating a NEW handler, I will assume I need to wire it to the route.
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Invalid verification type for this endpoint"})
		return
	}

	nodesRaw := req.Host
	// Split by newline
	lines := strings.Split(nodesRaw, "\n")
	
	successCount := 0
	failCount := 0
	errors := []string{}

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.Split(line, ":")
		host := parts[0]
		port := 22
		if len(parts) > 1 {
			if p, err := strconv.Atoi(parts[1]); err == nil {
				port = p
			}
		}

		if err := service.CheckSSHConnection(host, port); err != nil {
			failCount++
			errors = append(errors, host+": "+err.Error())
		} else {
			successCount++
		}
	}

	if failCount > 0 {
		msg := "Connectivity failed for some nodes."
		if successCount == 0 {
			msg = "Connectivity failed for ALL nodes."
		}
		c.JSON(http.StatusOK, gin.H{
			"status": "error", 
			"message": msg,
			"details": errors,
			"success_count": successCount,
			"fail_count": failCount,
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"status": "success",
			"message": "All nodes verified successfully.",
			"count": successCount,
		})
	}
}
