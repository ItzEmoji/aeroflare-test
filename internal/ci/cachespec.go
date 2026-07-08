// Package ci implements aeroflare-ci: a lightweight, non-interactive CI runner
// that builds Nix flake installables and pushes them to one or more OCI caches.
package ci

import (
	"fmt"
	"os"
	"strings"
)

// CacheSpec is a single push destination parsed from "<registry>;<repository>".
type CacheSpec struct {
	Registry   string // e.g. "ghcr.io"
	Repository string // e.g. "itzemoji/nix-cache"
	Raw        string // original entry, for messages
}

// ParseCacheSpec parses a "<registry>;<repository>" entry.
func ParseCacheSpec(s string) (CacheSpec, error) {
	raw := strings.TrimSpace(s)
	parts := strings.SplitN(raw, ";", 2)
	if len(parts) != 2 {
		return CacheSpec{}, fmt.Errorf("invalid cache %q: expected <registry>;<repository>", s)
	}
	reg := strings.TrimSpace(parts[0])
	repo := strings.TrimSpace(parts[1])
	if reg == "" || repo == "" {
		return CacheSpec{}, fmt.Errorf("invalid cache %q: registry and repository must be non-empty", s)
	}
	return CacheSpec{Registry: reg, Repository: repo, Raw: reg + ";" + repo}, nil
}

// TokenEnvVar returns the generic override env var name for a registry host,
// e.g. "ghcr.io" -> "AEROFLARE_TOKEN_GHCR_IO".
func TokenEnvVar(registry string) string {
	up := strings.ToUpper(registry)
	up = strings.NewReplacer(".", "_", ":", "_").Replace(up)
	return "AEROFLARE_TOKEN_" + up
}

// ResolveToken resolves the push token for a registry host from the environment.
// The generic AEROFLARE_TOKEN_<HOST> override wins; otherwise a per-host default
// applies (ghcr.io -> GITHUB_TOKEN). Returns "" if nothing is set.
func ResolveToken(registry string) string {
	if v := os.Getenv(TokenEnvVar(registry)); v != "" {
		return v
	}
	switch registry {
	case "ghcr.io":
		return os.Getenv("GITHUB_TOKEN")
	}
	return ""
}
