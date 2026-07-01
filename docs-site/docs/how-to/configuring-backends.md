---
sidebar_position: 1
title: Configuring Storage Backends
---

# Configuring Storage Backends

Aeroflare supports two primary types of storage backends for your Nix binaries: **OCI Registries** (like GitHub Container Registry) and **Cloudflare R2** buckets.

## Option 1: Cloudflare R2

Using Cloudflare R2 requires creating a bucket and deploying a Cloudflare Worker to act as the proxy interface.

### Automated Setup

The easiest way to configure Cloudflare R2 is via the initialization wizard:
```bash
nix run github:ItzEmoji/aeroflare -- init
```
During the setup, select **Cloudflare R2**. The wizard will prompt you for your Cloudflare API token and Account ID, create the bucket, and deploy the Worker automatically.

### Manual Settings

If you prefer to configure this manually without provisioning, use the interactive settings UI:
```bash
nix run github:ItzEmoji/aeroflare -- settings
```
Under **Registry Login & Setup**, select **Cloudflare R2** and provide your credentials.

## Option 2: GitHub Container Registry (GHCR)

Using GHCR (or GitLab Registry) stores your `.nar` blobs as standard Docker images.

### Automated Setup

Run the initialization wizard:
```bash
nix run github:ItzEmoji/aeroflare -- init
```
Select **GitHub Packages (ghcr.io)**. You will need a GitHub Personal Access Token with the `write:packages` and `read:packages` scopes. The wizard will configure your local environment to target your namespace.

### Custom OCI Registries

For any other OCI registry, you can define a custom cache URL:
```bash
nix run github:ItzEmoji/aeroflare -- settings
```
Select **Custom OCI Registry** and provide the full base URL (e.g., `registry.example.com/my-org/my-nix-cache`).
