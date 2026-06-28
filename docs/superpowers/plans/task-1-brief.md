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
