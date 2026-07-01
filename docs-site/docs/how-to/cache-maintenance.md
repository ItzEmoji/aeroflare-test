---
sidebar_position: 5
title: Cache Maintenance
---

# Cache Maintenance

Over time, your binary cache will accumulate blobs that are no longer referenced by active deployments. 

## Garbage Collection

The `gc` command analyzes the remote cache and deletes blobs that are stale or unreferenced.

```bash
nix run github:ItzEmoji/aeroflare -- gc
```

*Note: Garbage collection relies on resolving the root derivations. Depending on the backend (especially OCI registries), this operation can take a significant amount of time.*

## Cleaning the Index

Aeroflare sometimes maintains a remote index to speed up cache hit resolution without repeatedly querying the OCI manifest layer. 

If you suspect the index is corrupted or out of sync with the actual storage blobs, you can completely wipe the index.

```bash
nix run github:ItzEmoji/aeroflare -- clean-index
```

The index will be naturally rebuilt as new `push` operations occur.
