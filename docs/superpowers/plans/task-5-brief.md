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
