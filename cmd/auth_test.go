package cmd

import (
	"testing"
)

func TestAuthCmdExists(t *testing.T) {
	if authCmd.Use != "auth" {
		t.Errorf("Expected auth command")
	}
}
