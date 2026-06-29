package auth

import (
	"errors"
	"fmt"
	"os"

	"aeroflare/src/secrets"
)

var ErrTokenNotFound = errors.New("token not found in flags, environment, or secrets manager")

type Resolver struct {
	secretKey      string
	flagValue      string
	envVars        []string
	secretsManager secrets.Manager
}

func NewResolver(secretKey string) *Resolver {
	return &Resolver{secretKey: secretKey}
}

func (r *Resolver) WithSecretsManager(manager secrets.Manager) *Resolver {
	r.secretsManager = manager
	return r
}

func (r *Resolver) WithFlag(val string) *Resolver {
	r.flagValue = val
	return r
}

func (r *Resolver) WithEnv(keys ...string) *Resolver {
	r.envVars = append(r.envVars, keys...)
	return r
}

func (r *Resolver) Resolve() (string, error) {
	if r.flagValue != "" {
		return r.flagValue, nil
	}

	for _, key := range r.envVars {
		if val := os.Getenv(key); val != "" {
			return val, nil
		}
	}

	manager := r.secretsManager
	if manager == nil {
		manager = secrets.NewManager()
	}
	val, err := manager.Get(r.secretKey)
	if err != nil && err != secrets.ErrNotFound {
		return "", err
	}
	if err == nil && val != "" {
		return val, nil
	}

	return "", ErrTokenNotFound
}

func ResolveGithubToken(mgrs ...secrets.Manager) (string, error) {
	resolver := NewResolver("github-token").
		WithEnv("GITHUB_TOKEN", "GH_TOKEN")
	if len(mgrs) > 0 {
		resolver = resolver.WithSecretsManager(mgrs[0])
	}
	return resolver.Resolve()
}

func ResolveGitlabToken(mgrs ...secrets.Manager) (string, error) {
	resolver := NewResolver("gitlab-token").
		WithEnv("GITLAB_TOKEN")
	if len(mgrs) > 0 {
		resolver = resolver.WithSecretsManager(mgrs[0])
	}
	return resolver.Resolve()
}

func ResolveRegistryToken(registry string, mgrs ...secrets.Manager) (string, error) {
	if registry == "ghcr.io" {
		return ResolveGithubToken(mgrs...)
	} else if registry == "registry.gitlab.com" {
		return ResolveGitlabToken(mgrs...)
	}
	// Note: We use WithEnv here in case an explicit oci_token is provided for generic registries
	resolver := NewResolver(fmt.Sprintf("oci-%s-token", registry)).
		WithEnv("oci_token")
	if len(mgrs) > 0 {
		resolver = resolver.WithSecretsManager(mgrs[0])
	}
	return resolver.Resolve()
}
