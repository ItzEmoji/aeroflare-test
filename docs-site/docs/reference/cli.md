---
sidebar_position: 1
title: CLI Reference
---

# CLI Reference

Aeroflare's CLI provides tools for proxying, provisioning, authenticating, and manipulating OCI-backed Nix caches.

## Global Flags

The following flags can be passed to almost any `aeroflare` command:

* `--cache-url string`: OCI registry URL for the cache.
* `--cf-token string`: Cloudflare API Token.
* `--cf-user-id string`: Cloudflare Account ID.
* `--github-token string`: GitHub Token.
* `--gitlab-token string`: GitLab Token.
* `-v, --verbose count`: Enable verbose output (`-v` for packages, `-vv` for requests).

---

## Core Commands

### `aeroflare init`

Initializes Aeroflare infrastructure via an interactive wizard. It provisions the OCI repository, Cloudflare R2 buckets, Cloudflare Worker deployments, and Git repository integrations.

```bash
aeroflare init
```

### `aeroflare settings`

Provides an interactive terminal UI (TUI) to configure themes, registry logins, and custom caching URLs. The selections are persisted to `aeroflare.yaml`.

```bash
aeroflare settings
```

### `aeroflare proxy`

Starts the local cache proxy server that intercepts `.narinfo` and `.nar` requests from the Nix daemon.

```bash
aeroflare proxy
```

### `aeroflare run`

Wraps a command (typically a `nix build`), starts an ephemeral proxy substituter, executes the command, and automatically pushes any new output paths to the cache.

```bash
aeroflare run [--] <command>... [flags]
```

**Flags:**
* `--compression string`: Compression type (`zstd`, `xz`, `gzip`, `none`). Default is `zstd`.
* `--force`: Force push files even if they appear to exist.
* `--keep`: Keep generated `.nar` and `.narinfo` temporary files.
* `--prepare-refs`: Prepare references that are not on the upstream cache. Default is `true`.
* `--signing-key string`: Path to Nix signing private key file.
* `--upstream-cache string`: Upstream binary cache URL. Default is `https://cache.nixos.org`.
* `--workers int`: Number of concurrent push workers. Default is `50`.

---

## Cache Management

### `aeroflare push`

Pushes a build or specific Nix store paths to the cache.

```bash
aeroflare push [flags]
```

**Flags:**
* Same flags as `run`, plus:
* `--input string`: File containing store paths (one per line, `#` for comments).
* `--store-path string`: A single Nix store path to prepare and push.

### `aeroflare gc`

Garbage collects the remote cache, removing old or unreferenced blobs.

```bash
aeroflare gc
```

### `aeroflare clean-index`

Completely wipes the remote index on the registry. Use with caution.

```bash
aeroflare clean-index
```

### `aeroflare push-blob` / `aeroflare pull-blob`

Low-level commands to directly interact with the OCI registry to push or pull a single blob.

---

## Authentication

### `aeroflare auth`

Manage Aeroflare authentication secrets.

**Subcommands:**
* `aeroflare auth login`: Authenticate interactively and save tokens to the OS keychain.
* `aeroflare auth list`: List saved authentication credentials.
* `aeroflare auth remove [key]`: Remove a saved credential.
* `aeroflare auth set [key] [value]`: Set an arbitrary secret manually.
* `aeroflare auth import`: Import credentials from other CLIs (e.g., `gh`, `glab`, `docker`).
