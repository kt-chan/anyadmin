package integration

import (
	"anyadmin-backend/pkg/service"
	"anyadmin-backend/pkg/utils"
	"testing"
)

func TestVerifyRemoteDecryptedEnv(t *testing.T) {
	// 1. Load data and trigger sync
	utils.LoadFromFile()
	err := service.SyncLiteLLMConfig("172.25.208.100")
	if err != nil {
		t.Fatalf("Sync failed: %v", err)
	}

	// 2. Check remote .env-litellm via SSH
	client, err := service.GetSSHClient("172.25.208.100", "22")
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
