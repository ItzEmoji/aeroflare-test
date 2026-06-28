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
