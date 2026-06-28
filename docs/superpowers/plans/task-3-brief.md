### Task 3: Move `performPush` logic to `RunPush`

**Files:**
- Modify: `src/push/push.go`
- Modify: `cmd/push.go` (if you need to copy from it, though technically we only extract the logic)

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
