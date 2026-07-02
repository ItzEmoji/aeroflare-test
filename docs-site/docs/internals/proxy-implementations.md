---
id: proxy-implementations
title: Proxy Implementations
sidebar_position: 3
---

# Proxy Implementations

Aeroflare provides three distinct proxy variants implemented as Cloudflare Workers. These variants serve as the edge caching and resolution layer for Nix store paths, interfacing directly with OCI registries (like GHCR) and optional auxiliary storage (like Cloudflare R2).

The core responsibility of these proxies is to intercept Nix HTTP requests (`.narinfo` and `/nar/`) and map them to their underlying storage representations. 

The three variants are:
1. `no-webui-json`
2. `no-webui-native`
3. `no-webui-r2`

## Mode Comparison Matrix

Below is a comparison of how the three variants resolve paths, handle indexes, and interface with storage backends:

| Feature / Aspect | `no-webui-json` | `no-webui-native` | `no-webui-r2` |
| :--- | :--- | :--- | :--- |
| **Metadata Source** | Monolithic JSON index in OCI | Dedicated OCI Manifest per tag | Cloudflare R2 Object Store |
| **NAR Binary Source** | OCI Registry Blobs | OCI Registry Blobs | OCI Registry Blobs |
| **Lookup Complexity** | $O(1)$ memory read after load | $O(1)$ dynamic manifest request | $O(1)$ R2 Object HTTP fetch |
| **Setup Overhead** | Simple, standard registries | Simple, standard registries | Medium (Requires R2 Bucket configuration) |
| **Index Syncing** | Push updates central `cache-index` | None. Independent manifests | None. Direct metadata uploads |
| **Recommended for** | Registries with slow manifest API response times | Standard setups requiring zero-state indexing | Production workloads demanding low-latency metadata queries |

## Common Architecture

All three variants are deployed as Cloudflare Workers (defined in `worker.js`) using standard `fetch` event handlers. They share common mechanics for querying OCI registries:
- **`getOciToken(env, ...)`**: Fetches an OAuth2/JWT token for authenticating with the configured OCI registry (`NIXCACHE_REGISTRY`), scoped to `pull` on the specific repository (`NIXCACHE_REPO`).
- **Upstream Fallback**: Requests that cannot be resolved in the primary registry are proxied to the configured `NIXCACHE_UPSTREAM` caches (e.g., `cache.nixos.org`).

## Variant: `no-webui-json`

This variant relies on a monolithic JSON index pushed to the OCI registry as a custom artifact. It uses the `cache-index` tag.

### Mechanics

- **`getIndex(env, ctx)`**: 
  1. Requests `manifests/cache-index` from the OCI registry.
  2. Extracts the `digest` of the first layer in the manifest.
  3. Uses the `digest` to fetch the blob containing the actual JSON index.
  4. Caches the parsed index using Cloudflare's Cache API (`caches.default`) under `https://internal.cache/cache-index.json`.

### HTTP Handlers

- **`/*.narinfo`**:
  - Calls `getIndex` to retrieve the JSON map.
  - Looks up the entry by the store hash (the filename prefix without `.narinfo`).
  - Returns the raw `.narinfo` text stored in `entry.narinfo`.

- **`/nar/*`**:
  - Calls `getIndex` and executes `findNarDigest(index, narBasename)` to parse the `URL: ` line from the `.narinfo` string within the JSON index.
  - Extracts the OCI blob digest corresponding to the `.nar` file.
  - Proxies the `fetch` to `https://${registry}/v2/${repo}/nix-cache/blobs/${narDigest}` and streams the OCI blob back to the Nix client.

## Variant: `no-webui-native`

This variant eschews the central JSON index in favor of mapping Nix store paths directly to OCI image tags (a 1:1 mapping between a Nix derivation and an OCI manifest).

### Mechanics

- **`getOciManifestAndPath(env, ctx, tag)`**:
  - Dynamically fetches the manifest for a specific tag (the store hash).
  - Caches the manifest in `caches.default` under `https://internal.cache/manifest-v3/${tag}`.
  - Implements multi-path resolution by checking both the base repository and the `nix-cache` sub-repository (`${repo}` vs `${repo}/nix-cache`).

- **`getNarLabels(env, ctx, manifest, imagePath)`**:
  - Extracts Narinfo metadata stored as annotations/labels on the OCI image.
  - Checks three locations in order of precedence:
    1. Standard OCI annotations (`manifest.annotations["vnd.aeroflare.nar.*"]`).
    2. Root-level labels (common in older/non-standard tools).
    3. Fetches the image config blob (`manifest.config.digest`) and inspects the inner `config.Labels` or `config.labels`.

### HTTP Handlers

- **`/*.narinfo`**:
  - Uses the requested store hash as the OCI image tag.
  - Invokes `getOciManifestAndPath` and extracts labels using `getNarLabels`.
  - Dynamically reconstructs the Narinfo file contents via `generateNarinfo(labels)`, mapping custom keys like `vnd.aeroflare.nar.storepath` to standard Narinfo fields (`StorePath`).

- **`/nar/*`**:
  - Extracts the store hash (the first 32 characters of the filename).
  - Queries the OCI manifest for the store hash.
  - Reads the digest of the very first layer (`manifest.layers[0].digest`)—which represents the `.nar` file itself.
  - Fetches the blob using the digest and streams it to the client.

## Variant: `no-webui-r2`

This variant is a hybrid. It uses the same `cache-index` JSON artifact as `no-webui-json` for locating `.nar` blobs, but offloads `.narinfo` storage to Cloudflare R2 object storage.

### Mechanics

- **`getIndex(env, ctx)`**: Identical to `no-webui-json`. Fetches and caches the `cache-index.json` from the OCI registry to locate NAR blobs.

### HTTP Handlers

- **`/*.narinfo`**:
  - Requires the `BUCKET` environment variable binding.
  - Maps the path directly to an R2 object key: `narinfo/` + `filename`.
  - Performs a direct `env.BUCKET.get(objectKey)` and returns the body as `text/x-nix-narinfo`. 
  - This avoids pulling or parsing the monolithic JSON index purely to resolve metadata requests.

- **`/nar/*`**:
  - Identical to `no-webui-json`. Uses `findNarDigest(index, narBasename)` against the JSON index to locate the OCI layer digest and fetch the `.nar` blob from the registry.
