package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func hasWarning(warnings []string, substr string) bool {
	for _, w := range warnings {
		if contains(w, substr) {
			return true
		}
	}
	return false
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || indexOf(s, substr) >= 0)
}

func indexOf(s, substr string) int {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func TestValidateGithub_ReturnsUserAndScopeWarnings(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer tok" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.Header().Set("X-OAuth-Scopes", "repo, read:packages")
		_, _ = w.Write([]byte(`{"login":"octocat"}`))
	}))
	defer srv.Close()

	old := githubAPIBase
	githubAPIBase = srv.URL
	defer func() { githubAPIBase = old }()

	id, err := validateGithub(context.Background(), map[string]string{"token": "tok"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id.User != "octocat" {
		t.Errorf("expected user octocat, got %q", id.User)
	}
	if !hasWarning(id.Warnings, "write:packages") {
		t.Errorf("expected a write:packages warning, got %v", id.Warnings)
	}
	if !hasWarning(id.Warnings, "workflow") {
		t.Errorf("expected a workflow warning, got %v", id.Warnings)
	}
}

func TestValidateGithub_AllScopesNoWarnings(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-OAuth-Scopes", "repo, workflow, write:packages, read:packages")
		_, _ = w.Write([]byte(`{"login":"octocat"}`))
	}))
	defer srv.Close()

	old := githubAPIBase
	githubAPIBase = srv.URL
	defer func() { githubAPIBase = old }()

	id, err := validateGithub(context.Background(), map[string]string{"token": "tok"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(id.Warnings) != 0 {
		t.Errorf("expected no warnings, got %v", id.Warnings)
	}
}

func TestValidateGithub_InvalidToken(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()

	old := githubAPIBase
	githubAPIBase = srv.URL
	defer func() { githubAPIBase = old }()

	if _, err := validateGithub(context.Background(), map[string]string{"token": "bad"}); err == nil {
		t.Fatalf("expected error for invalid token")
	}
}

func TestValidateGitlab_ReturnsUser(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer gltok" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		_, _ = w.Write([]byte(`{"username":"gluser"}`))
	}))
	defer srv.Close()

	old := gitlabAPIBase
	gitlabAPIBase = srv.URL
	defer func() { gitlabAPIBase = old }()

	id, err := validateGitlab(context.Background(), map[string]string{"token": "gltok"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id.User != "gluser" {
		t.Errorf("expected user gluser, got %q", id.User)
	}
}

func TestValidateCloudflare_Valid(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer cftok" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		_, _ = w.Write([]byte(`{"success":true,"result":{"status":"active"}}`))
	}))
	defer srv.Close()

	old := cloudflareAPIBase
	cloudflareAPIBase = srv.URL
	defer func() { cloudflareAPIBase = old }()

	id, err := validateCloudflare(context.Background(), map[string]string{"token": "cftok", "account_id": "acct123"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id.User != "acct123" {
		t.Errorf("expected account acct123, got %q", id.User)
	}
}

func TestValidateCloudflare_Invalid(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"success":false}`))
	}))
	defer srv.Close()

	old := cloudflareAPIBase
	cloudflareAPIBase = srv.URL
	defer func() { cloudflareAPIBase = old }()

	if _, err := validateCloudflare(context.Background(), map[string]string{"token": "bad"}); err == nil {
		t.Fatalf("expected error for invalid cloudflare token")
	}
}

func TestServiceValidateWiredUp(t *testing.T) {
	for _, id := range []string{"github", "gitlab", "cloudflare"} {
		svc, _ := ServiceByID(id)
		if svc.Validate == nil {
			t.Errorf("service %q has no Validate wired up", id)
		}
	}
}
