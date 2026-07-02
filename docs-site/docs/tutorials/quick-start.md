---
id: quick-start
title: Quick Start
sidebar_position: 1
---
Welcome to Aeroflare! This guide will get you up and running with your own lightning-fast Nix cache infrastructure in just a few minutes. 

We'll cover how to initialize your configuration, authenticate with your cache provider, run the proxy, and push your first cached build.

## 1. Install & Initialize

The fastest way to get started is to use the interactive setup wizard via Nix. This provisions your backend (like Cloudflare R2 or GitHub Container Registry) and configures your local environment.

```bash
nix run github:ItzEmoji/aeroflare -- init
```

The `init` command guides you through backend selection and automatically provisions the required buckets or namespaces.

:::info Authentication

During initialization, the wizard will prompt you for the necessary credentials. If you don't have them defined in your local OS keychain or secrets manager, you'll be asked to provide:

- A **GitHub Personal Access Token** (if using GHCR)
- Or a **Cloudflare API Token** (if using Cloudflare R2)

Aeroflare securely saves these tokens for future use.
:::

## 2. Run the Proxy

Aeroflare operates as a local proxy that intercepts Nix daemon requests. To spin up the proxy server, use:

```bash
nix run github:ItzEmoji/aeroflare -- proxy start
```

> **Note:** This command runs in the foreground and will block your terminal. Please run it in the background or open a new terminal window to proceed with the next steps.

This starts the local proxy server, ready to route requests and handle caching.

## 3. Push to the Cache

With your infrastructure initialized and proxy running, you can execute a cached build. The most efficient way is to use the `run` execution wrapper.

This command configures your Nix daemon to use the proxy as an official substituter, executes your target command, and automatically pushes any resulting output paths back to the cache.

```bash
nix run github:ItzEmoji/aeroflare -- run -- nix build .#default
```

### The Cache Lifecycle:
1. **Pulling**: If the required build outputs already exist in your remote cache, they are pulled immediately, bypassing the local compilation process entirely.
2. **Execution**: If the artifacts are missing, the standard `nix build` command executes locally.
3. **Pushing**: Upon successful build completion, Aeroflare automatically isolates the new Nix store paths and uploads them as compressed blobs directly to your configured backend.

Congratulations! You've successfully configured and used Aeroflare to accelerate your Nix builds.
