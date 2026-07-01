---
sidebar_position: 1
title: Introduction
---

# Introduction

Aeroflare is a high-performance proxy cache and toolkit for Nix binaries, backed by OCI (Open Container Initiative) registries.

The tool bridges the Nix ecosystem and standard container registries (such as GitHub Container Registry or Docker Hub). By treating Nix archive (NAR) files and `narinfo` metadata as OCI blobs and manifests, Aeroflare leverages existing registry infrastructure to store and distribute Nix build artifacts.

## Core Capabilities

1. **Proxy Server**: Aeroflare runs a local HTTP server that acts as a standard Nix substituter. It handles requests from the Nix daemon, fetches the corresponding OCI artifacts from the remote registry, and serves them seamlessly.
2. **Direct Interaction**: The CLI toolkit permits direct interaction with the OCI registry to push blobs, pull blobs, and manage the cache lifecycle without invoking the proxy server.
3. **Execution Wrapper**: The `aeroflare run` command wraps standard Nix build commands, configures the local Nix daemon to use the Aeroflare proxy, and automatically pushes the resulting output paths to the registry upon completion.

## Architecture Overview

Aeroflare operates without local state storage for cached binaries. It relies entirely on the remote OCI registry as the source of truth.

When a client requests a path:
1. The proxy intercepts the `.narinfo` request.
2. Aeroflare resolves the request to an OCI manifest using the configured credentials.
3. If found, the proxy serves the metadata and streams the corresponding NAR blob directly from the registry to the client.
