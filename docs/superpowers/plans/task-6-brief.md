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

Note: In an earlier task, `cmd/run.go` might have been partially updated. Please overwrite its `Run` function body to match the one exactly specified here, and make sure to remove any leftover proxy code that is now inside `run.ExecuteCommand`.

- [ ] **Step 2: Build and format**

Run: `go build ./... && go fmt ./...`
Expected: Successful compilation

- [ ] **Step 3: Commit**

```bash
git add cmd/run.go
git commit -m "refactor: update cmd/run to use src/run and src/push packages"
```
