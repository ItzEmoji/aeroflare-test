# Auth Token Resolver Design

## Overview
A centralized authentication token resolver inside the `auth` module that enforces a unified fallback precedence for token resolution across all Aeroflare CLI commands.

## Requirements
- Tokens must be resolved in strict order:
  1. Explicitly provided CLI flag (if any)
  2. Environment variables
  3. The `secrets.Manager` (keychain)
- The architecture must be applied universally to all commands (e.g., GitHub token, Cloudflare token, OCI registry tokens).
- Must decouple the raw fallback logic from the `cmd` package (Cobra/Viper) to the core `auth` package.

## Architecture

### 1. The Builder Pattern
A new component `Resolver` will be introduced in `src/auth/resolver.go`. It utilizes the builder pattern to elegantly construct token requirements.

```go
package auth

import "aeroflare/src/secrets"
import "errors"

var ErrTokenNotFound = errors.New("token not found in flags, environment, or secrets manager")

type Resolver struct {
	secretKey string
	flagValue string
	envVars   []string
}

func NewResolver(secretKey string) *Resolver {
	return &Resolver{secretKey: secretKey}
}

func (r *Resolver) WithFlag(val string) *Resolver {
	r.flagValue = val
	return r
}

func (r *Resolver) WithEnv(keys ...string) *Resolver {
	r.envVars = append(r.envVars, keys...)
	return r
}

func (r *Resolver) Resolve() (string, error)
```

### 2. Resolution Logic
The `Resolve()` method will execute the fallback rules:
1. Check `r.flagValue`. If `!= ""`, return it.
2. Iterate `r.envVars`. Use `os.Getenv()`. If a value is found, return it.
3. Instantiate `secrets.NewManager()`. Call `manager.Get(r.secretKey)`. If no error and value `!= ""`, return it.
4. If all fail, return `"", ErrTokenNotFound`.

### 3. Updating the Command Layer
Legacy ad-hoc fetchers (like `getGithubToken` in `cmd/root.go`) will be refactored to consume the resolver:

```go
func getGithubToken() string {
	token, _ := auth.NewResolver("github-token").
		WithEnv("GITHUB_TOKEN", "GH_TOKEN").
		Resolve()
	return token
}
```
All commands interacting with credentials (e.g. `push`, `run`) will adopt this pattern.

## Testing Strategy
- Unit tests in `src/auth/resolver_test.go` will test all combinations of flag, environment, and keychain state to assert strict precedence.
- Mocking or stubbing environment variables and the `secrets.Manager` interface may be required to prevent test flakiness.
