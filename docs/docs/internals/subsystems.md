---
id: subsystems
title: Core Subsystems
sidebar_position: 2
---

# Core Subsystems

This document provides a deep technical breakdown of the core subsystems within Aeroflare: Authentication & Secrets, Storage Integration (R2), and the Push Pipeline. It is intended for developers and contributors to understand the mechanics of these systems.

## Authentication & Secrets Management

Aeroflare manages authentication primarily around accessing OCI registries (like GitHub Container Registry or Docker Hub) and securely storing user credentials.

### Token Resolution (`internal/auth/resolver.go`)
The `Resolver` struct is responsible for locating authentication tokens. It follows a strict hierarchy of sources:
1. **Command-line flags** (highest priority).
2. **Environment variables** (e.g., `GITHUB_TOKEN`, `GH_TOKEN`, `oci_token`).
3. **Secrets Manager** (secure local keychain or fallback).

Specific functions like `ResolveGithubToken` and `ResolveRegistryToken` wrap the resolver with the correct environment variables and secret keys depending on the target registry (e.g., mapping `ghcr.io` to the GitHub token).

### Token Exchange (`internal/oci/token.go` & `internal/proxy/token_manager.go`)
OCI registries often require a token exchange—trading a Personal Access Token (PAT) via Basic Auth for a short-lived JWT Bearer token.
- `ExchangeToken()` discovers the authentication realm and service by probing the registry's `/v2/` endpoint and parsing the `Www-Authenticate` header. It then requests a token with `scope=repository:<repo>:pull,push`.
- `TokenManager` (used in the proxy) caches the OCI Bearer token in memory. It uses a `sync.Mutex` for thread safety and caches tokens for 4 minutes before requiring a refresh via `fetchToken()`. It also supports static overrides via the `oci_token` environment variable.

### Device Flow Auth (`internal/auth/github.go`)
For interactive logins, Aeroflare implements the OAuth 2.0 Device Authorization Grant flow for GitHub:
- `RequestDeviceCode` hits `/login/device/code` to obtain a `user_code` for the user to enter in their browser and a `device_code`.
- `PollAccessToken` polls `/login/oauth/access_token` until the user authorizes the app, handling `authorization_pending` and `slow_down` responses appropriately.

### Secrets Manager (`internal/secrets/manager.go`)
The `secrets.Manager` interface abstracts credential storage.
- It uses the `zalando/go-keyring` library to securely store credentials in the OS native keychain (macOS Keychain, Windows Credential Manager, Linux Secret Service).
- **Fallback Mechanism:** If the keychain is unavailable (e.g., in CI environments or headless servers), it gracefully falls back to plain-text JSON storage in `~/.config/aeroflare/secrets.json`.
- An internal `_keys_index` is maintained to keep track of which keys are managed by Aeroflare.

## Storage Integration (Cloudflare R2)

Aeroflare natively supports dual-backend storage, utilizing OCI registries for heavy blobs (NARs) and Cloudflare R2 (or any S3-compatible API) for metadata (`narinfo`).

### R2 Configuration (`internal/r2/r2.go`)
Configuration is resolved via environment variables (`R2_BUCKET`, `R2_ENDPOINT`, `R2_ACCESS_KEY_ID`, `R2_SECRET_ACCESS_KEY`) or dynamically via OCI manifest annotations (e.g., `aeroflare.r2.bucket`).
The system uses the AWS SDK v2 (`aws-sdk-go-v2/service/s3`) with `UsePathStyle = true` to communicate with the endpoint.

### Narinfo Uploads
- The `UploadNarinfo` function reads the `.narinfo` file and uploads it directly to the root of the R2 bucket.
- The object key is constructed using just the hash of the store path (`<hash>.narinfo`).
- It strictly sets the `ContentType` to `text/x-nix-narinfo` to ensure compatibility with standard Nix HTTP binary caches.

## The Push Pipeline (`internal/push/push.go`)

The Push Pipeline is the engine that converts local Nix store paths into OCI-compliant artifacts and pushes them to remote registries.

### Configuration & Preflight
The pipeline parses paths from CLI arguments, an input file, or `stdin`. 
Before processing, a `Preflight` phase (and subsequent cache index checks) filters out paths that are already present in the remote registry to avoid redundant work.

### Batch Preparation
Paths are processed in chunks (default 100) to manage memory and disk I/O.
- The `prepare.PrepareBatch` step runs the local Nix daemon to generate NAR files and sign `.narinfo` files into a temporary directory (`/tmp/aeroflare-push-*`).
- If `KeepFiles` is not specified, this temporary directory is aggressively cleaned up to prevent disk exhaustion.

### Concurrent Upload Strategy
The upload phase uses `golang.org/x/sync/errgroup` to execute uploads concurrently, bounded by the configured number of workers.
To achieve maximum performance, Aeroflare employs a "brutal speed" optimization:
- `network.NewLayerFast` creates the OCI layer representation without reading and hashing the entire NAR file on disk beforehand. It relies on the pre-computed size and hash from the `.narinfo` file.

### Dual-Strategy Pushing
Depending on the cache index annotations (`aeroflare.backend = native`), the push operation selects its strategy:
1. **Native OCI Push**: Uses `network.PushNarPackage` to push the artifact using OCI standard manifest layouts.
2. **Legacy Fallback**: Uses `network.PushLayer` as a fallback or for generic registries to push the blob directly.

After the blobs are pushed, the `.narinfo` is uploaded either to the OCI registry as a blob or to the configured R2 bucket. Finally, the remote cache index is updated with the new receipts.
