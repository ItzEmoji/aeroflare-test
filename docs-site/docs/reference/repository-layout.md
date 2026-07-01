---
sidebar_position: 3
title: Repository Layout & Codebase
---

# Repository Layout & Codebase Walkthrough

For contributors and advanced users, understanding exactly how Aeroflare is structured makes it much easier to debug issues or extend functionality. 

Aeroflare is written in Go and structured into two primary domains: the **CLI boundary** (`cmd/`) and the **Core library** (`src/`).

## The `cmd/` Directory (CLI Boundary)

This directory contains the [Cobra](https://github.com/spf13/cobra) command definitions. Each file here maps directly to a subcommand you can run in your terminal. 

These files are deliberately kept thin. They are responsible for parsing flags, reading configuration via [Viper](https://github.com/spf13/viper), and passing arguments into the `src/` core logic.

* `root.go`: The entrypoint of the CLI. Initializes Viper and loads `aeroflare.yaml`.
* `proxy.go`: Maps to `aeroflare proxy`. Invokes the `src/proxy` package.
* `run.go`: Maps to `aeroflare run`. Starts the ephemeral proxy and wraps the Nix command.
* `init.go` / `configure.go` / `settings.go`: Maps to the interactive wizards (`aeroflare init` and `aeroflare settings`). These invoke the `huh` TUI forms.
* `auth*.go`: Handles all `aeroflare auth` subcommands, routing credential resolution to the `src/secrets` package.
* `push.go` / `gc.go`: CLI bindings for cache mutation commands.

## The `src/` Directory (Core Logic)

This is where the actual magic happens. The logic is heavily decoupled so that network operations, local file preparations, and proxying do not depend on CLI state.

### 1. Networking & OCI Integration
* `src/network.go`: **The most critical file in the project.** This file implements the OCI Network Layer. It uses `google/go-containerregistry` to stream `.nar` blobs as OCI layers and maps Nix `.narinfo` metadata onto OCI Manifest Annotations (`vnd.aeroflare.nar.*`).
* `src/r2.go`: Specific logic for interacting with Cloudflare R2's S3-compatible endpoints when an OCI registry is not being used.
* `src/index.go`: Handles remote cache indexing (maintaining a manifest of known hashes to avoid repeated network lookups).

### 2. The Proxy Server
* `src/proxy/`: Contains the HTTP server logic that tricks the Nix daemon into thinking it's talking to a standard binary cache.
  * `proxy_server.go`: Intercepts `/*.narinfo` and `/*.nar` HTTP requests. It parses the incoming Nix Hash, and dynamically translates the request into an OCI manifest fetch using `network.go`. It reconstructs the `.narinfo` plain text format on the fly.
  * `token_manager.go`: Handles authorization headers if the upstream registry requires authentication.

### 3. Execution Wrapper
* `src/run/`: Implements the `aeroflare run` logic.
  * It spawns the proxy server on an ephemeral port.
  * It instruments the `nix` subprocess with `--option extra-substituters`.
  * Post-execution, it computes the newly generated Nix derivations and triggers a concurrent push operation.

### 4. Nix Artifact Preparation
* `src/prepare/`: Before a local Nix store path can be pushed, it must be packaged.
  * This package generates the `.nar` (Nix Archive) format from local store paths.
  * It compresses the archive (e.g., using `zstd`).
  * It extracts dependencies and computes cryptographic hashes to generate the `.narinfo` map.

### 5. Security & Authentication
* `src/secrets/`: Abstracts token storage. It integrates directly with the OS Native Keychain (using `zalando/go-keyring`) to ensure Cloudflare and GitHub tokens are encrypted at rest.
* `src/auth/`: Implements OAuth flows and token validation with remote providers (e.g., verifying a GitHub PAT has `write:packages` scopes).

### 6. Wizards & UI
* `src/init/`: Contains the complex infrastructure provisioning logic. It talks directly to the Cloudflare API to provision R2 buckets and Edge Workers based on user selections.
* `src/ui/`: Shared terminal UI components (progress bars, spinners, and customized `huh` themes) used across the CLI.

---

## How Data Flows Through the Codebase

To truly understand Aeroflare, let's trace exactly what happens when you run `aeroflare push --store-path /nix/store/abc-package`:

1. **Entry**: `cmd/push.go` parses the flags and passes `/nix/store/abc-package` to the core.
2. **Prepare**: `src/prepare` serializes the directory into a `.nar` file and compresses it via `zstd`. It generates a `.narinfo` struct in memory.
3. **Upload Blob**: `src/network.go` opens a stream to the OCI Registry (e.g., GHCR). It uploads the compressed `.nar` as a raw `v1.Layer`.
4. **Map Metadata**: `src/network.go` creates a new OCI Manifest. It takes the `.narinfo` struct and maps every field (FileHash, StorePath, Sig) into `vnd.aeroflare.nar.*` annotations on the manifest.
5. **Tagging**: The manifest is tagged with `abc` (the 32-character Nix hash) and pushed to the registry. The operation completes with O(1) fetch capability guaranteed for the future.
