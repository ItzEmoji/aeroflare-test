---
sidebar_position: 2
title: Quick Start
---

# Quick Start

This guide covers provisioning your cache infrastructure and running your first cached Nix build using Aeroflare. 

## 1. Initialize Infrastructure

The most direct way to get started is to use the interactive setup wizard directly via Nix. This tool provisions your backend (like Cloudflare R2 or GitHub Container Registry) and configures your local environment automatically.

```bash
nix run github:ItzEmoji/aeroflare -- init
```

The `init` command performs several critical actions:

1. **Interactive Authentication**: If you do not have credentials defined in your local OS keychain or secrets manager, the wizard will immediately prompt you for the required tokens (e.g., a GitHub Personal Access Token or Cloudflare API token) and save them securely for future use.
2. **Backend Selection**: You will be asked to choose whether to back your cache with a Cloudflare R2 bucket or an OCI-compliant registry (such as GHCR).
3. **Provisioning**: After summarizing your choices, Aeroflare securely provisions the required buckets, worker deployments, or OCI namespaces, and writes the required routing to your local configuration.

## 2. Run a Cached Build

Once your infrastructure is initialized, the most efficient way to utilize the cache is via the `run` execution wrapper. 

This command spins up an ephemeral Aeroflare proxy server locally, temporarily configures your Nix daemon to use it as an official substituter, executes your target command, and subsequently pushes any resulting output paths back to the cache.

```bash
nix run github:ItzEmoji/aeroflare -- run -- nix build .#default
```

### The `run` Lifecycle

1. **Substitution**: If the required build outputs already exist in your remote cache, they are pulled immediately, bypassing the local compilation process entirely.
2. **Execution**: If the artifacts are missing, the standard `nix build` command executes locally.
3. **Pushing**: Upon successful build completion, Aeroflare automatically isolates the new Nix store paths and uploads them as compressed blobs directly to your configured backend.
