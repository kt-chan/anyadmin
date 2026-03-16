package integration

import (
	"anyadmin-backend/pkg/global"
	"anyadmin-backend/pkg/service"
	"anyadmin-backend/pkg/utils"
	"testing"

	"github.com/spf13/viper"
)

func TestVerifyRemoteDecryptedEnv(t *testing.T) {
	global.InitConfig()
	remoteHost := viper.GetString("REMOTE_HOST")
	remotePort := viper.GetString("REMOTE_SSH_PORT")
	if remotePort == "" {
		remotePort = "22"
	}

	// 1. Load data and trigger sync
	utils.LoadFromFile()
	err := service.SyncLiteLLMConfig(remoteHost)
	if err != nil {
		t.Fatalf("Sync failed: %v", err)
	}

	// 2. Check remote .env-litellm via SSH
	client, err := service.GetSSHClient(remoteHost, remotePort)
	if err != nil {
		t.Fatalf("SSH failed: %v", err)
	}
	defer client.Close()

	output, err := service.ExecuteCommand(client, "cat /home/anyadmin/docker/.env-litellm")
	if err != nil {
		t.Fatalf("Failed to read remote env: %v", err)
	}

	t.Logf("Remote .env content:\n%s", output)
}
