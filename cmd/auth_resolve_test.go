package cmd

import (
	"testing"
)

func TestRequireGithubToken_Global(t *testing.T) {
	globalGithubToken = "global-gh-token"
	defer func() { globalGithubToken = "" }()

	if token := RequireGithubToken(); token != "global-gh-token" {
		t.Errorf("Expected global token, got %s", token)
	}
}

func TestRequireGithubToken_Env(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "env-gh-token")

	if token := RequireGithubToken(); token != "env-gh-token" {
		t.Errorf("Expected env token, got %s", token)
	}
}

func TestRequireGithubToken_SecretsManager(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "")
	t.Setenv("GH_TOKEN", "")
	mock := &mockManager{data: map[string]string{"github-token": "secret-gh-token"}}
	SecretsManager = mock
	defer func() { SecretsManager = nil }()

	if token := RequireGithubToken(); token != "secret-gh-token" {
		t.Errorf("Expected secret token, got %s", token)
	}
}

func TestRequireGitlabToken_Global(t *testing.T) {
	globalGitlabToken = "global-gl-token"
	defer func() { globalGitlabToken = "" }()

	if token := RequireGitlabToken(); token != "global-gl-token" {
		t.Errorf("Expected global token, got %s", token)
	}
}

func TestRequireGitlabToken_Env(t *testing.T) {
	t.Setenv("GITLAB_TOKEN", "env-gl-token")

	if token := RequireGitlabToken(); token != "env-gl-token" {
		t.Errorf("Expected env token, got %s", token)
	}
}

func TestRequireGitlabToken_SecretsManager(t *testing.T) {
	t.Setenv("GITLAB_TOKEN", "")
	mock := &mockManager{data: map[string]string{"gitlab-token": "secret-gl-token"}}
	SecretsManager = mock
	defer func() { SecretsManager = nil }()

	if token := RequireGitlabToken(); token != "secret-gl-token" {
		t.Errorf("Expected secret token, got %s", token)
	}
}

func TestRequireCloudflareToken_Global(t *testing.T) {
	globalCfToken = "global-cf-token"
	globalCfUserID = "global-cf-user"
	defer func() {
		globalCfToken = ""
		globalCfUserID = ""
	}()

	token, user := RequireCloudflareToken()
	if token != "global-cf-token" || user != "global-cf-user" {
		t.Errorf("Expected global token and user, got %s and %s", token, user)
	}
}

func TestRequireCloudflareToken_SecretsManager(t *testing.T) {
	t.Setenv("CLOUDFLARE_API_TOKEN", "")
	t.Setenv("CLOUDFLARE_ACCOUNT_ID", "")
	mock := &mockManager{data: map[string]string{
		"cf-token":   "secret-cf-token",
		"cf-user-id": "secret-cf-user",
	}}
	SecretsManager = mock
	defer func() { SecretsManager = nil }()

	token, user := RequireCloudflareToken()
	if token != "secret-cf-token" || user != "secret-cf-user" {
		t.Errorf("Expected secret token and user, got %s and %s", token, user)
	}
}

func TestRequireCloudflareToken_Env(t *testing.T) {
	t.Setenv("CLOUDFLARE_API_TOKEN", "env-cf-token")
	t.Setenv("CLOUDFLARE_ACCOUNT_ID", "env-cf-user")

	token, user := RequireCloudflareToken()
	if token != "env-cf-token" || user != "env-cf-user" {
		t.Errorf("Expected env token and user, got %s and %s", token, user)
	}
}

func TestRequireOCIToken_SecretsManager(t *testing.T) {
	mock := &mockManager{data: map[string]string{
		"oci-docker.io-username": "docker-user",
		"oci-docker.io-token":    "docker-pass",
	}}
	SecretsManager = mock
	defer func() { SecretsManager = nil }()

	user, pass := RequireOCIToken("docker.io")
	if user != "docker-user" || pass != "docker-pass" {
		t.Errorf("Expected secret user and pass, got %s and %s", user, pass)
	}
}
