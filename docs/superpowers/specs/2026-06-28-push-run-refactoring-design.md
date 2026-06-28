# Push and Run Command Refactoring Design

## Overview
Refactor the `push` and `run` commands in `aeroflare` to follow the clean structural design of the `init` command. The logic will be decoupled from the CLI layer and split into dedicated `src/push` and `src/run` packages. The commands will display a clean summary of what is happening, but will auto-proceed without a blocking confirmation prompt.

## Architecture

### 1. `src/push` Package
Move the core logic from `cmd/push.go` to a new package.
- **`PushConfig`**: Struct holding all configuration and parsed paths.
- **`PushPlan`**: Struct holding the result of preflight checks (e.g. which paths need uploading vs already cached).
- **`ParseConfig(cmd *cobra.Command, args []string) (*PushConfig, error)`**: Gathers paths from flags, arguments, stdin, and parsing the input file.
- **`Preflight(cfg *PushConfig) (*PushPlan, error)`**: Analyzes the registry and upstream cache to determine exactly which paths need to be prepared and pushed.
- **`DisplaySummary(plan *PushPlan)`**: Prints a clean summary of the push operation (how many packages are found, skipped, and to be pushed). It will auto-proceed.
- **`RunPush(plan *PushPlan) error`**: Executes the multi-threaded preparation and upload, retaining the existing progress output but wrapped cleanly.

### 2. `src/run` Package
Move the core logic from `cmd/run.go` to a new package.
- **`RunConfig`**: Struct holding the proxy config and the command to run.
- **`ParseConfig(cmd *cobra.Command, args []string) (*RunConfig, error)`**: Parses flags and arguments.
- **`DisplaySummary(cfg *RunConfig)`**: Prints a summary stating that the background proxy will start and the specific command will be executed.
- **`ExecuteCommand(cfg *RunConfig) ([]string, error)`**: Starts the proxy, runs the command, and returns the harvested `/nix/store/` paths from the output.

### 3. CLI Layer Updates (`cmd/push.go` & `cmd/run.go`)
- The code within `cmd/push.go` and `cmd/run.go` will be drastically reduced in size.
- They will only instantiate configs, call the functions from the respective `src/` packages, and handle errors/exits gracefully, matching the brevity of `cmd/init.go`.

## Data Flow
- **Push**: `ParseConfig` -> `Preflight` -> `DisplaySummary` -> `RunPush`
- **Run**: `ParseConfig` -> `DisplaySummary` -> `ExecuteCommand` -> (passes paths to `PushConfig` initialization) -> Push flow.

## Error Handling
Errors returned by the `src` packages will be caught in the CLI commands and displayed cleanly using `PrintError`.
