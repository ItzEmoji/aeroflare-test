package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// API base URLs for the live validation checks. They are vars (not consts) so
// tests can point them at local httptest servers.
var (
	githubAPIBase     = "https://api.github.com"
	gitlabAPIBase     = "https://gitlab.com/api/v4"
	cloudflareAPIBase = "https://api.cloudflare.com/client/v4"
)

// validateClient is the HTTP client used for validation calls. The short
// timeout keeps `auth status` responsive even when a provider is unreachable.
var validateClient = &http.Client{Timeout: 5 * time.Second}

// githubRequiredScopes are the OAuth scopes aeroflare needs a GitHub token to
// carry; validateGithub warns about any that are missing.
var githubRequiredScopes = []string{"write:packages", "workflow"}

// validateGithub verifies a GitHub token by calling GET /user, returning the
// authenticated login and warnings for any missing required OAuth scopes.
func validateGithub(ctx context.Context, vals map[string]string) (*Identity, error) {
	token := vals["token"]
	if token == "" {
		return nil, fmt.Errorf("no github token to validate")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, githubAPIBase+"/user", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := validateClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github token rejected (status %d)", resp.StatusCode)
	}

	granted := make(map[string]bool)
	for _, s := range strings.Split(resp.Header.Get("X-OAuth-Scopes"), ",") {
		if s = strings.TrimSpace(s); s != "" {
			granted[s] = true
		}
	}

	var data struct {
		Login string `json:"login"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	id := &Identity{User: data.Login}
	for _, scope := range githubRequiredScopes {
		if !granted[scope] {
			id.Warnings = append(id.Warnings, fmt.Sprintf("token missing scope: %s", scope))
		}
	}
	return id, nil
}

// validateGitlab verifies a GitLab token by calling GET /user, returning the
// authenticated username.
func validateGitlab(ctx context.Context, vals map[string]string) (*Identity, error) {
	token := vals["token"]
	if token == "" {
		return nil, fmt.Errorf("no gitlab token to validate")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, gitlabAPIBase+"/user", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := validateClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("gitlab token rejected (status %d)", resp.StatusCode)
	}

	var data struct {
		Username string `json:"username"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}
	return &Identity{User: data.Username}, nil
}

// validateCloudflare verifies a Cloudflare API token via the token
// verification endpoint. The returned Identity's User is the configured
// account ID (if any), since the verify endpoint does not name a user.
func validateCloudflare(ctx context.Context, vals map[string]string) (*Identity, error) {
	token := vals["token"]
	if token == "" {
		return nil, fmt.Errorf("no cloudflare token to validate")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, cloudflareAPIBase+"/user/tokens/verify", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := validateClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	var data struct {
		Success bool `json:"success"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK || !data.Success {
		return nil, fmt.Errorf("cloudflare token rejected (status %d)", resp.StatusCode)
	}

	return &Identity{User: vals["account_id"]}, nil
}

func init() {
	// Wire live validation onto the fixed catalog services. Done here (rather
	// than in the struct literals in service.go) to keep credential data and
	// validation logic in separate files.
	githubService.Validate = validateGithub
	gitlabService.Validate = validateGitlab
	cloudflareService.Validate = validateCloudflare
	fixedServices = []Service{githubService, gitlabService, cloudflareService}
}
