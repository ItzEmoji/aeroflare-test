package cmd

import (
	"errors"
	"fmt"
	"os"

	"aeroflare/internal/auth"
)

// isTerminal reports whether stdin is an interactive character device, used
// to decide whether it's safe to launch an interactive auth prompt or
// whether we should just fail with an actionable error message instead.
func isTerminal() bool {
	stat, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return (stat.Mode() & os.ModeCharDevice) != 0
}

// RequireGithubToken resolves a GitHub token from (in order) the --github-token
// flag, the secrets manager/environment, or an interactive device/manual auth
// flow if stdin is a terminal. Exits the process if no token can be obtained
// non-interactively.
func RequireGithubToken() string {
	if globalGithubToken != "" {
		return globalGithubToken
	}

	token, err := auth.ResolveGithubToken(getSecretsManager())
	if err != nil && !errors.Is(err, auth.ErrTokenNotFound) {
		fmt.Fprintf(os.Stderr, "Warning: failed to read token from keychain: %v\n", err)
	}
	if token != "" {
		return token
	}

	if isTerminal() {
		fmt.Println("GitHub token is required but not found. Launching authentication...")
		return runInteractiveGithubAuth()
	}

	PrintError("GitHub token required. Please set GITHUB_TOKEN or run 'aeroflare auth login'.")
	os.Exit(1)
	return ""
}

// RequireGitlabToken resolves a GitLab token the same way RequireGithubToken
// resolves a GitHub token: flag, then secrets manager/environment, then an
// interactive prompt if possible.
func RequireGitlabToken() string {
	if globalGitlabToken != "" {
		return globalGitlabToken
	}

	token, err := auth.ResolveGitlabToken(getSecretsManager())
	if err != nil && !errors.Is(err, auth.ErrTokenNotFound) {
		fmt.Fprintf(os.Stderr, "Warning: failed to read token from keychain: %v\n", err)
	}
	if token != "" {
		return token
	}

	if isTerminal() {
		fmt.Println("GitLab token is required but not found. Launching authentication...")
		return runInteractiveGitlabAuth()
	}

	PrintError("GitLab token required. Please set GITLAB_TOKEN or run 'aeroflare auth login'.")
	os.Exit(1)
	return ""
}

// RequireCloudflareToken resolves the Cloudflare API token and account ID,
// each independently from a flag, secrets manager, or matching env var
// (CLOUDFLARE_API_TOKEN / CLOUDFLARE_ACCOUNT_ID). If either is still missing
// it falls back to an interactive prompt, or exits if not on a terminal.
func RequireCloudflareToken() (string, string) {
	cf, _ := auth.ServiceByID("cloudflare")

	apiToken := globalCfToken
	if apiToken == "" {
		apiToken = resolveField(cf, "token")
	}

	userID := globalCfUserID
	if userID == "" {
		userID = resolveField(cf, "account_id")
	}

	if apiToken != "" && userID != "" {
		return apiToken, userID
	}

	if isTerminal() {
		fmt.Println("Cloudflare credentials required but incomplete. Launching authentication...")
		return runInteractiveCloudflareAuth()
	}

	PrintError("Cloudflare credentials required. Please set CLOUDFLARE_API_TOKEN and CLOUDFLARE_ACCOUNT_ID, or run 'aeroflare auth login'.")
	os.Exit(1)
	return "", ""
}

// GetOCIToken looks up a saved username/token pair for an arbitrary OCI
// registry, without prompting. Returns empty strings for whichever half is
// not found; use RequireOCIToken if an interactive fallback is wanted.
func GetOCIToken(registry string) (string, string) {
	svc := auth.ServiceForRegistry(registry)
	return resolveField(svc, "username"), resolveField(svc, "token")
}

// resolveField resolves a single named field of a service from the environment
// and the secrets manager, returning "" if it is unset. A read error other
// than "not found" is reported to stderr, matching the tolerant behavior the
// callers below rely on. Fields the service does not declare resolve to "".
func resolveField(svc auth.Service, name string) string {
	field, ok := svc.Field(name)
	if !ok {
		return ""
	}
	val, err := field.Resolve(getSecretsManager())
	if err != nil && !errors.Is(err, auth.ErrTokenNotFound) {
		fmt.Fprintf(os.Stderr, "Warning: failed to read credential from keychain: %v\n", err)
	}
	return val
}

// RequireOCIToken returns a username/token pair for registry, prompting
// interactively (via runInteractiveOCIAuth) if credentials aren't already
// saved. Exits the process if credentials are missing and stdin isn't a
// terminal.
func RequireOCIToken(registry string) (string, string) {
	user, pass := GetOCIToken(registry)
	if user != "" && pass != "" {
		return user, pass
	}

	if isTerminal() {
		fmt.Printf("Credentials for registry %s are required. Launching authentication...\n", registry)
		return runInteractiveOCIAuth(registry)
	}

	PrintError(fmt.Sprintf("Credentials required for registry %s. Run 'aeroflare auth login' to set them.", registry))
	os.Exit(1)
	return "", ""
}

// getTokenForRegistry resolves (prompting interactively if needed) and
// exports the token for registry into the process environment (oci_token,
// plus GITHUB_TOKEN for ghcr.io) so downstream Nix/OCI tooling can pick it up.
func getTokenForRegistry(registry string) string {
	if registry == "ghcr.io" {
		token := RequireGithubToken()
		_ = os.Setenv("oci_token", token)
		_ = os.Setenv("GITHUB_TOKEN", token)
		return token
	} else if registry != "" {
		_, token := RequireOCIToken(registry)
		_ = os.Setenv("oci_token", token)
		return token
	}
	return ""
}

// getOptionalTokenForRegistry is like getTokenForRegistry but never prompts
// or exits: it returns "" if no token is already saved, which lets the proxy
// serve unauthenticated (public) registries without forcing a login.
func getOptionalTokenForRegistry(registry string) string {
	if registry == "" {
		return ""
	}
	token, _ := auth.ResolveRegistryToken(registry, getSecretsManager())
	if token != "" {
		_ = os.Setenv("oci_token", token)
		if registry == "ghcr.io" {
			_ = os.Setenv("GITHUB_TOKEN", token)
		}
	}
	return token
}
