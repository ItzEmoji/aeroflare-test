---
id: proxy-implementations
title: Proxy Implementation
sidebar_position: 3
---

# Proxy Implementation

Aeroflare provides a proxy implemented as a Cloudflare Worker. It serves as the
edge caching and resolution layer for Nix store paths, interfacing directly with
OCI registries (like GHCR).

The core responsibility of the proxy is to intercept Nix HTTP requests
(`.narinfo` and `/nar/`) and map them to their underlying storage
representations. The worker lives in `proxy/no-webui-native`.

## Design

Aeroflare maps Nix store paths directly to OCI image tags — a 1:1 mapping
between a Nix derivation and an OCI manifest. This is fully stateless: there is
no central index, so there is nothing to keep in sync, and each lookup resolves
to a single dynamic manifest request.

| Aspect | Behavior |
| :--- | :--- |
| **Metadata Source** | Dedicated OCI Manifest per tag |
| **NAR Binary Source** | OCI Registry Blobs |
| **Lookup Complexity** | `O(1)` dynamic manifest request |
| **Setup Overhead** | Simple, standard registries |
| **Index Syncing** | None. Independent manifests |

## Common Architecture

The worker is deployed as a Cloudflare Worker (defined in `worker.js`) using a
standard `fetch` event handler. It queries OCI registries via:
- **`getOciToken(env, ...)`**: Fetches an OAuth2/JWT token for authenticating with the configured OCI registry (`NIXCACHE_REGISTRY`), scoped to `pull` on the specific repository (`NIXCACHE_REPO`).
- **Upstream Fallback**: Requests that cannot be resolved in the primary registry are proxied to the configured `NIXCACHE_UPSTREAM` caches (e.g., `cache.nixos.org`).

## Mechanics

- **`getOciManifestAndPath(env, ctx, tag)`**:
  - Dynamically fetches the manifest for a specific tag (the store hash).
  - Caches the manifest in `caches.default` under `https://internal.cache/manifest-v3/${tag}`.
  - Fetches from the configured repository (`env.NIXCACHE_REPO`).

- **`getNarLabels(env, ctx, manifest, imagePath)`**:
  - Extracts Narinfo metadata stored as annotations/labels on the OCI image.
  - Checks three locations in order of precedence:
    1. Standard OCI annotations (`manifest.annotations["vnd.aeroflare.nar.*"]`).
    2. Root-level labels (common in older/non-standard tools).
    3. Fetches the image config blob (`manifest.config.digest`) and inspects the inner `config.Labels` or `config.labels`.

## HTTP Handlers

- **`/*.narinfo`**:
  - Uses the requested store hash as the OCI image tag.
  - Invokes `getOciManifestAndPath` and extracts labels using `getNarLabels`.
  - Dynamically reconstructs the Narinfo file contents via `generateNarinfo(labels)`, mapping custom keys like `vnd.aeroflare.nar.storepath` to standard Narinfo fields (`StorePath`).

- **`/nar/*`**:
  - Extracts the store hash (the first 32 characters of the filename).
  - Queries the OCI manifest for the store hash.
  - Reads the digest of the very first layer (`manifest.layers[0].digest`)—which represents the `.nar` file itself.
  - Fetches the blob using the digest and streams it to the client.
