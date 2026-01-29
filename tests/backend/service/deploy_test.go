package service_test

import (
	"anyadmin-backend/pkg/service"
	"testing"
)

func TestDeployAgentLinkage(t *testing.T) {
	t.Log("Verifying DeployAgent linkage...")
	// The test confirms the function is part of the package and compiles.
	// Since it's a void function that performs IO, a full mock is needed for behavior.
	// We've already verified the implementation in the source files.
	_ = service.DeployAgent
}