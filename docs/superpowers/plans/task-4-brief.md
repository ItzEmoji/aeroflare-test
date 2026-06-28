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
