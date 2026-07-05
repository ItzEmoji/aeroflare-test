package setup

import (
	"fmt"
	"os"
	"strings"
)

// BackendType represents the index storage backend.
type BackendType string

const (
	BackendR2     BackendType = "r2"
	BackendNative BackendType = "native"
	BackendOCI    BackendType = "oci"
)

// String returns a human-readable label.
func (b BackendType) String() string {
	switch b {
	case BackendR2:
		return "Cloudflare R2"
	case BackendNative:
		return "Native OCI Tags"
	case BackendOCI:
		return "JSON index stored in OCI"
	default:
		return string(b)
	}
}

// GitProvider represents the Git hosting provider for CI/CD integration.
type GitProvider string

const (
	GitNone   GitProvider = "none"
	GitGitHub GitProvider = "github"
	GitGitLab GitProvider = "gitlab"
)

// String returns a human-readable label.
func (g GitProvider) String() string {
	switch g {
	case GitNone:
		return "None"
	case GitGitHub:
		return "GitHub"
	case GitGitLab:
		return "GitLab"
	default:
		return string(g)
	}
}

// InitConfig holds all parameters collected by the setup wizard.
type InitConfig struct {
	CacheName  string
	Registry   string
	Repository string

	Backend     BackendType
	GitProvider GitProvider

	CloudflareAccountID string
	CloudflareToken     string

	GitToken    string
	GitUsername string

	WorkerName string
	R2Bucket   string

	// Internal fields populated during provisioning.
	OCIToken    string
	ScriptTag   string // Worker script tag returned by the Cloudflare deploy API; reserved for a future Workers Builds integration, not yet read elsewhere.
	CfTokenID   string // Cloudflare API token ID; reserved for a future Workers Builds integration, not yet read elsewhere.
	GitCloneURL string // Clone URL (with embedded credentials) for pushing to the created git repository.
}

// DeriveDefaults populates computed fields (Repository, WorkerName, R2Bucket)
// from the cache name, registry and backend the user already supplied.
// Must be called after CacheName/Registry/Backend are set and before the
// values are used elsewhere (e.g. by the credentials prompt or provisioning).
func (c *InitConfig) DeriveDefaults() {
	c.CacheName = strings.ToLower(c.CacheName)

	// Docker Hub allows "user/repo" style names directly; other registries
	// (e.g. ghcr.io) get an explicit "/nix-cache" suffix appended.
	if (c.Registry == "docker.io" || c.Registry == "index.docker.io" || c.Registry == "registry-1.docker.io") && strings.Contains(c.CacheName, "/") {
		c.Repository = c.CacheName
	} else {
		c.Repository = fmt.Sprintf("%s/nix-cache", c.CacheName)
	}

	// Worker and R2 bucket names can't contain "/", so slashes in the cache
	// name (e.g. "user/repo") are flattened to hyphens.
	sanitized := strings.ReplaceAll(c.CacheName, "/", "-")
	c.WorkerName = fmt.Sprintf("aeroflare-%s", sanitized)

	if c.Backend == BackendR2 && c.R2Bucket == "" {
		c.R2Bucket = fmt.Sprintf("%s-index", sanitized)
	}
}

// Print helpers below give the setup wizard and provisioning pipeline a
// consistent terminal output style (icon + indented message).

func printError(msg string) {
	fmt.Fprintf(os.Stderr, "  \u2717 %s\n", msg)
}

func printSuccess(msg string) {
	fmt.Printf("  \u2713 %s\n", msg)
}

func printInfo(msg string) {
	fmt.Printf("  \u2022 %s\n", msg)
}

func printStep(step, total int, msg string) {
	fmt.Printf("\n  [%d/%d] %s\n", step, total, msg)
}

func printWarning(msg string) {
	fmt.Printf("  \u26a0 %s\n", msg)
}
