package cmd

import (
	"bytes"
	"errors"
	"testing"
	"github.com/spf13/cobra"
)

type mockManager struct {
	data map[string]string
	err  error
}

func (m *mockManager) Set(key, value string) error {
	if m.err != nil {
		return m.err
	}
	m.data[key] = value
	return nil
}

func (m *mockManager) Get(key string) (string, error) {
	if val, ok := m.data[key]; ok {
		return val, nil
	}
	return "", errors.New("not found")
}

func executeCommand(root *cobra.Command, args ...string) (output string, err error) {
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs(args)

	err = root.Execute()

	return buf.String(), err
}

func TestAuthCmdExists(t *testing.T) {
	if authCmd.Use != "auth" {
		t.Errorf("Expected auth command")
	}
}

func TestAuthCmdBehavior(t *testing.T) {
	mock := &mockManager{data: make(map[string]string)}
	SecretsManager = mock
	defer func() { SecretsManager = nil }() // reset

	// Reset variables since they are package level
	githubToken = ""
	cfToken = ""

	_, err := executeCommand(rootCmd, "auth", "--github-token=test-gh", "--cf-token=test-cf")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if mock.data["github-token"] != "test-gh" {
		t.Errorf("Expected github-token to be 'test-gh', got %s", mock.data["github-token"])
	}

	if mock.data["cf-token"] != "test-cf" {
		t.Errorf("Expected cf-token to be 'test-cf', got %s", mock.data["cf-token"])
	}
}

func TestAuthCmdError(t *testing.T) {
	mock := &mockManager{data: make(map[string]string), err: errors.New("mock error")}
	SecretsManager = mock
	defer func() { SecretsManager = nil }()

	githubToken = ""
	cfToken = ""

	_, err := executeCommand(rootCmd, "auth", "--github-token=fail")
	// error swallowed and printed
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func TestAuthSetCmdBehavior(t *testing.T) {
	mock := &mockManager{data: make(map[string]string)}
	SecretsManager = mock
	defer func() { SecretsManager = nil }()

	_, err := executeCommand(rootCmd, "auth", "set", "custom-key", "custom-value")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if mock.data["custom-key"] != "custom-value" {
		t.Errorf("Expected custom-key to be 'custom-value', got %s", mock.data["custom-key"])
	}
}
