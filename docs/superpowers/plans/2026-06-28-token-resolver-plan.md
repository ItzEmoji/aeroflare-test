# Token Resolver Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Centralize authentication token fallback resolution across all Aeroflare CLI commands.

**Architecture:** A unified builder `auth.Resolver` that prioritizes explicit flags, then environment variables, then the `secrets.Manager`, dropping the ad-hoc resolutions currently scattered in `cmd`.

**Tech Stack:** Go standard library.

## Global Constraints

- Must resolve tokens in strict order: explicit flag -> env vars -> secrets manager.
- Error returned on complete miss must be `auth.ErrTokenNotFound`.
- Must preserve existing functionality for all commands.

---

### Task 1: The Token Resolver Builder

**Files:**
- Create: `src/auth/resolver.go`
- Create: `src/auth/resolver_test.go`

**Interfaces:**
- Consumes: `aeroflare/src/secrets`
- Produces: `auth.NewResolver(string) *Resolver`, `auth.Resolver.WithFlag(string) *Resolver`, `auth.Resolver.WithEnv(...string) *Resolver`, `auth.Resolver.Resolve() (string, error)`

- [ ] **Step 1: Write the failing test for resolving rules**

```go
package auth_test

import (
	"os"
	"testing"
	"aeroflare/src/auth"
	"aeroflare/src/secrets"
)

func TestResolver_FlagPriority(t *testing.T) {
	os.Setenv("TEST_ENV_VAR", "env-value")
	defer os.Unsetenv("TEST_ENV_VAR")

	val, err := auth.NewResolver("test-secret").
		WithFlag("flag-value").
		WithEnv("TEST_ENV_VAR").
		Resolve()

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if val != "flag-value" {
		t.Errorf("expected flag-value, got %s", val)
	}
}

func TestResolver_EnvPriority(t *testing.T) {
	os.Setenv("TEST_ENV_VAR", "env-value")
	defer os.Unsetenv("TEST_ENV_VAR")
	
	val, err := auth.NewResolver("test-secret").
		WithFlag("").
		WithEnv("TEST_ENV_VAR").
		Resolve()

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if val != "env-value" {
		t.Errorf("expected env-value, got %s", val)
	}
}

func TestResolver_NotFound(t *testing.T) {
	_, err := auth.NewResolver("test-missing-secret").
		WithFlag("").
		WithEnv("NONEXISTENT_VAR").
		Resolve()

	if err != auth.ErrTokenNotFound {
		t.Fatalf("expected ErrTokenNotFound, got %v", err)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./src/auth/... -run TestResolver -v`
Expected: FAIL with compilation error (auth.NewResolver not defined)

- [ ] **Step 3: Write minimal implementation**

```go
package auth

import (
	"errors"
	"os"

	"aeroflare/src/secrets"
)

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

func (r *Resolver) Resolve() (string, error) {
	if r.flagValue != "" {
		return r.flagValue, nil
	}

	for _, key := range r.envVars {
		if val := os.Getenv(key); val != "" {
			return val, nil
		}
	}

	manager := secrets.NewManager()
	if val, err := manager.Get(r.secretKey); err == nil && val != "" {
		return val, nil
	}

	return "", ErrTokenNotFound
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./src/auth/... -run TestResolver -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add src/auth/resolver.go src/auth/resolver_test.go
git commit -m "feat(auth): implement token resolver builder"
```

---

### Task 2: Refactor Command Layer to use Resolver

**Files:**
- Modify: `cmd/root.go:102-117` (replace `getGithubToken` contents)

**Interfaces:**
- Consumes: `auth.NewResolver`
- Produces: Command layer correctly routes tokens.

- [ ] **Step 1: Write minimal implementation in `cmd/root.go`**

```go
func getGithubToken() string {
	token, _ := auth.NewResolver("github-token").
		WithEnv("GITHUB_TOKEN", "GH_TOKEN").
		Resolve()
	return token
}
```

Make sure you import `aeroflare/src/auth` in `cmd/root.go` if it's not already there.

- [ ] **Step 2: Run test to verify it passes**

Run: `go test ./cmd/... -v`
Expected: PASS

- [ ] **Step 3: Commit**

```bash
git add cmd/root.go
git commit -m "refactor(cmd): utilize auth resolver for github token fallback"
```
