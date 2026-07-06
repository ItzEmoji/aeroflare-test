package cmd

import (
	"aeroflare/internal/secrets"
	"bytes"
	"encoding/json"
	"errors"
	"strings"
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
	return "", secrets.ErrNotFound
}

func (m *mockManager) List() ([]string, error) {
	var keys []string
	for k := range m.data {
		keys = append(keys, k)
	}
	return keys, nil
}

func (m *mockManager) Delete(key string) error {
	delete(m.data, key)
	return nil
}

// resetStatusFlags clears the persistent `auth status` flag variables. Cobra
// does not reset flag values between Execute calls in the same process, so
// tests that share the process must clear them to avoid one test's --json or
// --no-verify leaking into the next.
func resetStatusFlags() {
	authStatusJSON = false
	authStatusNoVerify = false
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
	defer func() {
		globalGithubToken = ""
		globalGitlabToken = ""
		globalCfToken = ""
		globalCfUserID = ""
	}()

	_, err := executeCommand(rootCmd, "auth", "login", "--github-token=test-gh", "--gitlab-token=test-gl", "--cf-token=test-cf", "--cf-user-id=test-id")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if mock.data["github-token"] != "test-gh" {
		t.Errorf("Expected github-token to be 'test-gh', got %s", mock.data["github-token"])
	}

	if mock.data["gitlab-token"] != "test-gl" {
		t.Errorf("Expected gitlab-token to be 'test-gl', got %s", mock.data["gitlab-token"])
	}

	if mock.data["cf-token"] != "test-cf" {
		t.Errorf("Expected cf-token to be 'test-cf', got %s", mock.data["cf-token"])
	}

	if mock.data["cf-user-id"] != "test-id" {
		t.Errorf("Expected cf-user-id to be 'test-id', got %s", mock.data["cf-user-id"])
	}
}

func TestAuthCmdError(t *testing.T) {
	mock := &mockManager{data: make(map[string]string), err: errors.New("mock error")}
	SecretsManager = mock
	defer func() { SecretsManager = nil }()

	defer func() {
		globalGithubToken = ""
		globalGitlabToken = ""
		globalCfToken = ""
		globalCfUserID = ""
	}()

	_, err := executeCommand(rootCmd, "auth", "login", "--github-token=fail")
	if err == nil {
		t.Fatalf("Expected an error, got nil")
	}
}

func TestAuthSetCmdBehavior(t *testing.T) {
	mock := &mockManager{data: make(map[string]string)}
	SecretsManager = mock
	defer func() { SecretsManager = nil }()

	_, err := executeCommand(rootCmd, "auth", "set", "github", "gh-tok")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if mock.data["github-token"] != "gh-tok" {
		t.Errorf("Expected github-token to be 'gh-tok', got %s", mock.data["github-token"])
	}
}

func TestAuthSetCmd_UnknownService(t *testing.T) {
	mock := &mockManager{data: make(map[string]string)}
	SecretsManager = mock
	defer func() { SecretsManager = nil }()

	_, err := executeCommand(rootCmd, "auth", "set", "notaservice", "x")
	if err == nil {
		t.Fatalf("Expected an error for an unknown service")
	}
}

func TestAuthSetCmd_MultiFieldAndOCI(t *testing.T) {
	mock := &mockManager{data: make(map[string]string)}
	SecretsManager = mock
	defer func() { SecretsManager = nil }()

	if _, err := executeCommand(rootCmd, "auth", "set", "cloudflare", "cf-tok", "acct-1"); err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if mock.data["cf-token"] != "cf-tok" || mock.data["cf-user-id"] != "acct-1" {
		t.Errorf("cloudflare fields not stored: %+v", mock.data)
	}

	if _, err := executeCommand(rootCmd, "auth", "set", "oci", "docker.io", "bob", "dckr-tok"); err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if mock.data["oci-docker.io-username"] != "bob" || mock.data["oci-docker.io-token"] != "dckr-tok" {
		t.Errorf("oci fields not stored: %+v", mock.data)
	}
}

func TestAuthGetCmd(t *testing.T) {
	mock := &mockManager{data: map[string]string{"github-token": "secret-gh"}}
	SecretsManager = mock
	defer func() { SecretsManager = nil }()
	t.Setenv("GITHUB_TOKEN", "")
	t.Setenv("GH_TOKEN", "")

	out, err := executeCommand(rootCmd, "auth", "get", "github")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if got := strings.TrimSpace(out); got != "secret-gh" {
		t.Errorf("expected raw token 'secret-gh', got %q", got)
	}
}

func TestAuthGetCmd_Missing(t *testing.T) {
	mock := &mockManager{data: map[string]string{}}
	SecretsManager = mock
	defer func() { SecretsManager = nil }()
	t.Setenv("GITLAB_TOKEN", "")

	if _, err := executeCommand(rootCmd, "auth", "get", "gitlab"); err == nil {
		t.Fatalf("expected error when credential is missing")
	}
}

func TestAuthRemoveCmd_Service(t *testing.T) {
	mock := &mockManager{data: map[string]string{"cf-token": "a", "cf-user-id": "b"}}
	SecretsManager = mock
	defer func() { SecretsManager = nil }()

	if _, err := executeCommand(rootCmd, "auth", "remove", "cloudflare"); err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if _, ok := mock.data["cf-token"]; ok {
		t.Errorf("cf-token should have been removed")
	}
	if _, ok := mock.data["cf-user-id"]; ok {
		t.Errorf("cf-user-id should have been removed")
	}
}

func TestAuthStatusCmd_JSON(t *testing.T) {
	mock := &mockManager{data: map[string]string{
		"github-token": "gh",
		"cf-token":     "cf",
	}}
	SecretsManager = mock
	defer func() { SecretsManager = nil }()

	resetStatusFlags()
	out, err := executeCommand(rootCmd, "auth", "status", "--json", "--no-verify")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	var entries []struct {
		ID     string `json:"id"`
		Fields []struct {
			Name string `json:"name"`
			Set  bool   `json:"set"`
		} `json:"fields"`
	}
	if err := json.Unmarshal([]byte(out), &entries); err != nil {
		t.Fatalf("status --json did not emit valid JSON: %v\n%s", err, out)
	}
	ids := map[string]bool{}
	for _, e := range entries {
		ids[e.ID] = true
	}
	if !ids["github"] || !ids["cloudflare"] {
		t.Errorf("expected github and cloudflare in status, got %v", ids)
	}
}

func TestAuthStatusCmd_TableColumns(t *testing.T) {
	mock := &mockManager{data: map[string]string{
		"github-token": "ghtokenvalue",
		"cf-token":     "cfval",
		"cf-user-id":   "acct-1",
	}}
	SecretsManager = mock
	defer func() { SecretsManager = nil }()

	resetStatusFlags()
	out, err := executeCommand(rootCmd, "auth", "status", "--no-verify")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	for _, want := range []string{"Service", "ID", "Field", "github", "cloudflare", "token", "account_id", "acct-1"} {
		if !strings.Contains(out, want) {
			t.Errorf("expected status table to contain %q, output:\n%s", want, out)
		}
	}
}

func TestAuthStatusCmd_RedactsSecrets(t *testing.T) {
	mock := &mockManager{data: map[string]string{"github-token": "supersecretvalue"}}
	SecretsManager = mock
	defer func() { SecretsManager = nil }()

	resetStatusFlags()
	out, err := executeCommand(rootCmd, "auth", "status", "--no-verify")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if strings.Contains(out, "supersecretvalue") {
		t.Errorf("status leaked a secret value:\n%s", out)
	}
}
