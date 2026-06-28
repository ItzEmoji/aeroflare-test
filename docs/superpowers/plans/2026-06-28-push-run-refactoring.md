# Push and Run Command Refactoring Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Refactor the `push` and `run` commands to decouple their execution logic from the CLI layer into `src/push` and `src/run` packages, displaying a clean summary of operations before auto-proceeding.

**Architecture:** We will create `src/push` and `src/run` packages. We will move the massive `performPush` logic into `src/push.RunPush` and create `ParseConfig`, `Preflight`, and `DisplaySummary` functions. We will similarly refactor `run`'s command execution logic into `src/run`.

**Tech Stack:** Go, Cobra

## Global Constraints

- Must decouple logic from CLI files `cmd/push.go` and `cmd/run.go`.
- Must include a `DisplaySummary` before the main execution.
- Must auto-proceed after summary without blocking prompts (Approach 2).
- Keep existing log messages, but organize them inside the new structures.

---

### Task 1: Create `src/push` configuration and basic parser

**Files:**
- Create: `src/push/push.go`
- Create: `src/push/push_test.go`

**Interfaces:**
- Consumes: Nothing
- Produces: `PushConfig` struct, `ParseConfig` function.

- [ ] **Step 1: Write the failing test for ParseConfig**

```go
package push

import (
	"testing"
)

func TestParseConfig_NoPaths(t *testing.T) {
	_, err := ParseConfig([]string{}, "", "")
	if err == nil {
		t.Fatal("expected error when no paths provided")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./src/push -v`
Expected: FAIL with "no Go files" or "undefined: ParseConfig"

- [ ] **Step 3: Write minimal implementation**

```go
package push

import (
	"bufio"
	"errors"
	"os"
	"strings"

	"aeroflare/src/prepare/prepare"
)

type PushConfig struct {
	TargetPaths     []string
	Compression     string
	CacheURL        string
	Workers         int
	PrepareRefs     bool
	SigningKey      string
	KeepFiles       bool
	ForcePush       bool
}

func ParseConfig(args []string, storePath string, inputFile string) (*PushConfig, error) {
	var targetPaths []string
	if storePath != "" {
		targetPaths = append(targetPaths, storePath)
	}
	if inputFile != "" {
		filePaths, err := prepare.ParseInputFile(inputFile)
		if err != nil {
			return nil, err
		}
		targetPaths = append(targetPaths, filePaths...)
	}

	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line != "" && !strings.HasPrefix(line, "#") {
				targetPaths = append(targetPaths, line)
			}
		}
	}

	if len(targetPaths) == 0 && len(args) == 0 {
		return nil, errors.New("No store paths found. Provide --store-path, --input, or pipe paths via stdin.")
	}
	targetPaths = append(targetPaths, args...)

	return &PushConfig{
		TargetPaths: targetPaths,
	}, nil
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./src/push -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add src/push/push.go src/push/push_test.go
git commit -m "feat: implement push config parser"
```

### Task 2: Implement Preflight and DisplaySummary in `src/push`

**Files:**
- Modify: `src/push/push.go`

**Interfaces:**
- Consumes: `PushConfig`
- Produces: `PushPlan` struct, `Preflight` function, `DisplaySummary` function

- [ ] **Step 1: Write minimal implementation for Preflight and Summary**

```go
// Add to src/push/push.go

import (
    "fmt"
)

type PushPlan struct {
	Config        *PushConfig
	FilteredPaths []string
	SkippedCount  int
}

// Preflight checks which paths actually need to be pushed.
// For now, it simply copies the paths, but will be integrated with cache index logic.
func Preflight(cfg *PushConfig) (*PushPlan, error) {
    // Note: To keep tasks small, we just stub this out first. 
    // The actual cache checking logic from cmd/push.go will be moved here in Task 4.
	return &PushPlan{
		Config:        cfg,
		FilteredPaths: cfg.TargetPaths,
		SkippedCount:  0,
	}, nil
}

// DisplaySummary prints what is about to happen.
func DisplaySummary(plan *PushPlan) {
	fmt.Println("\n=== Push Operation Summary ===")
	fmt.Printf("Total paths provided: %d\n", len(plan.Config.TargetPaths))
	if plan.SkippedCount > 0 {
		fmt.Printf("Already cached (skipped): %d\n", plan.SkippedCount)
	}
	fmt.Printf("Paths to be pushed: %d\n", len(plan.FilteredPaths))
	fmt.Println("==============================")
	fmt.Println("\nStarting push...")
}
```

- [ ] **Step 2: Compile to verify**

Run: `go build ./src/push`
Expected: Successful build

- [ ] **Step 3: Commit**

```bash
git add src/push/push.go
git commit -m "feat: add push preflight and summary display"
```

### Task 3: Move `performPush` logic to `RunPush`

**Files:**
- Modify: `src/push/push.go`

**Interfaces:**
- Consumes: `PushPlan`
- Produces: `RunPush` function.

- [ ] **Step 1: Move execution code**

Open `src/push/push.go` and copy the execution body of `performPush` from `cmd/push.go` into a new function:
```go
func RunPush(plan *PushPlan) error {
    // Copy the entire performPush function body here, starting from:
    // startTime := time.Now()
    // ...
    // Replace `targetPaths` with `plan.FilteredPaths`
    // Ensure all references to `pushStorePath`, etc are mapped from `plan.Config`.
    // Instead of `os.Exit(1)`, return errors.
    return nil
}
```
*(The subagent executing this step will handle the exact variable rewrites for `pushKeepFiles` -> `plan.Config.KeepFiles`, etc.)*

- [ ] **Step 2: Build verification**

Run: `go build ./src/push`
Expected: Might fail with import errors, fix the imports (like `network "aeroflare/src"`).

- [ ] **Step 3: Commit**

```bash
git add src/push/push.go
git commit -m "refactor: extract RunPush execution logic"
```

### Task 4: Refactor `cmd/push.go` to use the new package

**Files:**
- Modify: `cmd/push.go`

**Interfaces:**
- Consumes: `src/push` package

- [ ] **Step 1: Update `cmd/push.go` Run function**

Replace the `Run` body of `pushCmd` with:
```go
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := push.ParseConfig(args, pushStorePath, pushInputFile)
		if err != nil {
			PrintError(err.Error())
			os.Exit(1)
		}
		
		cfg.Compression = pushCompression
		cfg.CacheURL = pushCacheURL
		cfg.Workers = pushWorkers
		cfg.PrepareRefs = pushPrepareRefs
		cfg.SigningKey = pushSigningKey
		cfg.KeepFiles = pushKeepFiles
		cfg.ForcePush = pushForcePush

		plan, err := push.Preflight(cfg)
		if err != nil {
			PrintError(err.Error())
			os.Exit(1)
		}

		push.DisplaySummary(plan)

		if err := push.RunPush(plan); err != nil {
			PrintError(err.Error())
			os.Exit(1)
		}
	},
```
Remove the old `performPush` function from `cmd/push.go`.

- [ ] **Step 2: Build and Test**

Run: `go build ./cmd/...`
Expected: Successful build.

- [ ] **Step 3: Commit**

```bash
git add cmd/push.go
git commit -m "refactor: update cmd/push to use src/push package"
```

### Task 5: Extract `src/run` logic

**Files:**
- Create: `src/run/run.go`

**Interfaces:**
- Produces: `RunConfig`, `DisplaySummary`, `ExecuteCommand`

- [ ] **Step 1: Implement `src/run/run.go`**

```go
package run

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	network "aeroflare/src"
	"aeroflare/src/proxy"
)

type RunConfig struct {
	Command []string
}

func DisplaySummary(cfg *RunConfig) {
	fmt.Println("\n=== Run Operation Summary ===")
	fmt.Printf("Command: %s\n", strings.Join(cfg.Command, " "))
	fmt.Println("Action: Starts local background proxy and executes command with proxy substituter.")
	fmt.Println("=============================")
}

// ExecuteCommand starts proxy, runs cmd, and returns harvested store paths
func ExecuteCommand(cfg *RunConfig, registry, repository, indexDir, githubToken string) ([]string, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	port, err := proxy.StartProxy(ctx, 0, "127.0.0.1", registry, repository, indexDir, "", 300, []string{"https://cache.nixos.org"}, githubToken)
	if err != nil {
		return nil, fmt.Errorf("failed to start proxy: %w", err)
	}

	fmt.Printf("Started background proxy on 127.0.0.1:%d\n", port)

	var stdoutBuf bytes.Buffer
	cmdToRun := exec.Command(cfg.Command[0], cfg.Command[1:]...)
	cmdToRun.Stdout = io.MultiWriter(os.Stdout, &stdoutBuf)
	cmdToRun.Stderr = os.Stderr
	cmdToRun.Stdin = os.Stdin

	env := os.Environ()
	nixConfig := os.Getenv("NIX_CONFIG")
	if nixConfig != "" {
		nixConfig += "\n"
	}
	nixConfig += fmt.Sprintf("extra-substituters = http://127.0.0.1:%d", port)
	env = append(env, "NIX_CONFIG="+nixConfig)
	cmdToRun.Env = env

	err = cmdToRun.Run()
	if err != nil {
		return nil, fmt.Errorf("command failed: %w", err)
	}

	var targetPaths []string
	lines := strings.Split(stdoutBuf.String(), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "#") && strings.HasPrefix(line, "/nix/store/") {
			targetPaths = append(targetPaths, line)
		}
	}

	return targetPaths, nil
}
```

- [ ] **Step 2: Build verification**

Run: `go build ./src/run`
Expected: Successful build

- [ ] **Step 3: Commit**

```bash
git add src/run/run.go
git commit -m "feat: implement src/run package"
```

### Task 6: Refactor `cmd/run.go` to use the new package

**Files:**
- Modify: `cmd/run.go`

**Interfaces:**
- Consumes: `src/run`, `src/push`

- [ ] **Step 1: Update `cmd/run.go` Run function**

Replace the `Run` body of `runCmd` with:
```go
	Run: func(cmd *cobra.Command, args []string) {
		registry, repository := network.GetRegistryAndRepository()
		indexDir := getIndexDir(repository)

		cfg := &run.RunConfig{
			Command: args,
		}

		run.DisplaySummary(cfg)

		targetPaths, err := run.ExecuteCommand(cfg, registry, repository, indexDir, getGithubToken())
		if err != nil {
			PrintError(err.Error())
			os.Exit(1)
		}

		if len(targetPaths) == 0 {
			PrintWarning("No nix store paths found in command stdout. Nothing to push.")
			return
		}

		fmt.Printf("\nFound %d store paths to push from run command output.\n", len(targetPaths))

        // Trigger Push
		pushCfg := &push.PushConfig{
			TargetPaths: targetPaths,
			Compression: pushCompression,
			CacheURL:    pushCacheURL,
			Workers:     pushWorkers,
			PrepareRefs: pushPrepareRefs,
			SigningKey:  pushSigningKey,
			KeepFiles:   pushKeepFiles,
			ForcePush:   pushForcePush,
		}

		plan, err := push.Preflight(pushCfg)
		if err != nil {
			PrintError(err.Error())
			os.Exit(1)
		}

		push.DisplaySummary(plan)

		if err := push.RunPush(plan); err != nil {
			PrintError(err.Error())
			os.Exit(1)
		}
	},
```

- [ ] **Step 2: Build and format**

Run: `go build ./... && go fmt ./...`
Expected: Successful compilation

- [ ] **Step 3: Commit**

```bash
git add cmd/run.go
git commit -m "refactor: update cmd/run to use src/run and src/push packages"
```
