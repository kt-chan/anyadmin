package service_test

import (
	"anyadmin-backend/pkg/service"
	"testing"
)

func TestRedeployAgent(t *testing.T) {
	nodeIP := "172.20.0.10"
	mgmtHost := "172.20.0.1"
	mgmtPort := "8080"
	mode := "integrate_existing"

	t.Logf("Testing deployment to %s", nodeIP)
	service.DeployAgent(nodeIP, mgmtHost, mgmtPort, mode)
}
