# Secrets Management Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement a unified secrets manager using `zalando/go-keyring` with a plaintext fallback, and integrate it into the Aeroflare CLI.

**Architecture:** A unified `Manager` interface in `src/secrets` handles reading/writing to the OS keychain, automatically falling back to a restrictive `0600` JSON file. The CLI exposes this via a new `auth` command and existing commands check the manager before falling back to environment variables.

**Tech Stack:** Go, `github.com/zalando/go-keyring`, `github.com/spf13/cobra`

## Global Constraints

- Must fall back to `~/.config/aeroflare/secrets.json` with `0600` permissions.
- Service name for the keychain is `aeroflare`.

---

### Task 1: Add dependencies

**Files:**
- Modify: `go.mod`
- Modify: `go.sum`

**Interfaces:**
- Consumes: N/A
- Produces: N/A

- [ ] **Step 1: Download dependency**

```bash
go get github.com/zalando/go-keyring
```

- [ ] **Step 2: Commit**

```bash
git add go.mod go.sum
git commit -m "chore: add go-keyring dependency"
```

### Task 2: Implement SecretManager Interface and Fallback

**Files:**
- Create: `src/secrets/manager.go`
- Create: `src/secrets/manager_test.go`

**Interfaces:**
- Consumes: N/A
- Produces: `func NewManager() Manager`, `type Manager interface { Set(key, value string) error; Get(key string) (string, error) }`

- [ ] **Step 1: Write the failing test**

```go
package secrets_test

import (
	"os"
	"testing"
	"aeroflare/src/secrets"
)

func TestFallbackManager(t *testing.T) {
	// Use a dummy config dir for tests to avoid touching real keychains or config files
	tmpDir := t.TempDir()
	os.Setenv("XDG_CONFIG_HOME", tmpDir)
	defer os.Unsetenv("XDG_CONFIG_HOME")

	manager := secrets.NewManager()
	
	err := manager.Set("test-key", "test-value")
	if err != nil {
		t.Fatalf("Failed to set secret: %v", err)
	}

	val, err := manager.Get("test-key")
	if err != nil {
		t.Fatalf("Failed to get secret: %v", err)
	}

	if val != "test-value" {
		t.Errorf("Expected 'test-value', got '%s'", val)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./src/secrets/...`
Expected: FAIL (no packages)

- [ ] **Step 3: Write minimal implementation**

```go
package secrets

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/zalando/go-keyring"
)

type Manager interface {
	Set(key, value string) error
	Get(key string) (string, error)
}

type defaultManager struct {
	serviceName string
}

func NewManager() Manager {
	return &defaultManager{serviceName: "aeroflare"}
}

func getConfigDir() string {
	configDir := os.Getenv("XDG_CONFIG_HOME")
	if configDir == "" {
		homeDir, err := os.UserHomeDir()
		if err == nil && homeDir != "" {
			configDir = filepath.Join(homeDir, ".config")
		}
	}
	return filepath.Join(configDir, "aeroflare")
}

func getFallbackFile() string {
	return filepath.Join(getConfigDir(), "secrets.json")
}

func (m *defaultManager) Set(key, value string) error {
	err := keyring.Set(m.serviceName, key, value)
	if err == nil {
		return nil
	}
	
	// Fallback
	dir := getConfigDir()
	os.MkdirAll(dir, 0755)
	
	file := getFallbackFile()
	secrets := make(map[string]string)
	
	data, err := os.ReadFile(file)
	if err == nil {
		json.Unmarshal(data, &secrets)
	}
	
	secrets[key] = value
	
	out, err := json.MarshalIndent(secrets, "", "  ")
	if err != nil {
		return err
	}
	
	return os.WriteFile(file, out, 0600)
}

func (m *defaultManager) Get(key string) (string, error) {
	val, err := keyring.Get(m.serviceName, key)
	if err == nil {
		return val, nil
	}
	
	// Fallback
	file := getFallbackFile()
	data, err := os.ReadFile(file)
	if err != nil {
		return "", err
	}
	
	secrets := make(map[string]string)
	if err := json.Unmarshal(data, &secrets); err != nil {
		return "", err
	}
	
	if v, ok := secrets[key]; ok {
		return v, nil
	}
	
	return "", keyring.ErrNotFound
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./src/secrets/... -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add src/secrets/
git commit -m "feat: implement unified secret manager with fallback"
```

### Task 3: Add `aeroflare auth` CLI commands

**Files:**
- Create: `cmd/auth.go`
- Create: `cmd/auth_test.go`

**Interfaces:**
- Consumes: `secrets.NewManager() Manager`
- Produces: `authCmd *cobra.Command`

- [ ] **Step 1: Write the failing test**

```go
package cmd

import (
	"testing"
)

func TestAuthCmdExists(t *testing.T) {
	if authCmd.Use != "auth" {
		t.Errorf("Expected auth command")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./cmd -run TestAuthCmdExists`
Expected: FAIL (undefined: authCmd)

- [ ] **Step 3: Write minimal implementation**

```go
package cmd

import (
	"fmt"
	"aeroflare/src/secrets"
	"github.com/spf13/cobra"
)

var (
	githubToken string
	cfToken     string
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Manage Aeroflare authentication secrets",
	Run: func(cmd *cobra.Command, args []string) {
		manager := secrets.NewManager()
		
		if githubToken != "" {
			manager.Set("github-token", githubToken)
			fmt.Println("Saved github-token")
		}
		
		if cfToken != "" {
			manager.Set("cf-token", cfToken)
			fmt.Println("Saved cf-token")
		}
		
		if githubToken == "" && cfToken == "" {
			fmt.Println("Interactive mode not fully implemented in CLI yet, please use flags.")
		}
	},
}

var authSetCmd = &cobra.Command{
	Use:   "set [key] [value]",
	Short: "Set an arbitrary secret",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		manager := secrets.NewManager()
		err := manager.Set(args[0], args[1])
		if err != nil {
			PrintError(err.Error())
			return
		}
		fmt.Printf("Saved %s\n", args[0])
	},
}

func init() {
	authCmd.Flags().StringVar(&githubToken, "github-token", "", "GitHub Token")
	authCmd.Flags().StringVar(&cfToken, "cf-token", "", "Cloudflare Token")
	
	authCmd.AddCommand(authSetCmd)
	rootCmd.AddCommand(authCmd)
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./cmd -run TestAuthCmdExists`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add cmd/auth.go cmd/auth_test.go
git commit -m "feat: add auth CLI commands"
```

### Task 4: Integrate SecretManager into existing commands

**Files:**
- Modify: `cmd/root.go`

**Interfaces:**
- Consumes: `secrets.NewManager() Manager`

- [ ] **Step 1: Write integration logic for `getGithubToken`**

Update `getGithubToken` in `cmd/root.go` to use the SecretManager first. Since modifying an existing function, we'll write the replacement directly.

```go
import "aeroflare/src/secrets"

// ... inside getGithubToken() ...
func getGithubToken() string {
	manager := secrets.NewManager()
	if val, err := manager.Get("github-token"); err == nil && val != "" {
		return val
	}
	
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		token = os.Getenv("GH_TOKEN")
	}
	return token
}
```

- [ ] **Step 2: Verify compilation**

Run: `go build ./...`
Expected: Successful compilation without errors.

- [ ] **Step 3: Commit**

```bash
git add cmd/root.go
git commit -m "feat: integrate secret manager with github token retrieval"
```
