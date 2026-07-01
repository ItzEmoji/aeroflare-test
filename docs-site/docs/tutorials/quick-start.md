---
sidebar_position: 2
title: Quick Start
---

# Quick Start

This guide covers setting up Aeroflare to cache a standard Nix build against GitHub Container Registry (GHCR).

## Prerequisites

1. A GitHub Personal Access Token (PAT) with `read:packages` and `write:packages` scopes.
2. A working Nix installation with flake support enabled.

## 1. Authentication

Initialize the authentication configuration. Aeroflare uses Viper for configuration, typically storing credentials in `~/.config/aeroflare/aeroflare.yaml`.

```bash
export AEROFLARE_GITHUB_TOKEN="ghp_your_token_here"
```

Alternatively, use the interactive setup wizard:

```bash
aeroflare init
```

## 2. Configure the Cache Backend

Specify the target OCI registry URL. For GitHub Packages, this follows the format `ghcr.io/<namespace>/<cache-name>`.

```bash
export AEROFLARE_CACHE_URL="ghcr.io/my-org/my-nix-cache"
```

## 3. Run a Cached Build

The most efficient way to use Aeroflare is via the `run` wrapper. This command starts an ephemeral proxy server, configures Nix to use it as a substituter, executes your build, and automatically pushes the resulting outputs to the registry.

```bash
aeroflare run --cache-url ghcr.io/my-org/my-nix-cache -- nix build .#default
```

If the build outputs already exist in the OCI registry, they will be pulled immediately, bypassing the local build process. If they do not exist, the build will execute locally and the results will be uploaded to GHCR.
