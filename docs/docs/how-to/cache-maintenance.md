---
sidebar_position: 5
title: Cache Maintenance
---

# Cache Maintenance

Aeroflare sometimes maintains a remote index to speed up cache hit resolution without repeatedly querying the OCI manifest layer. 


## Cleaning the Index

Aeroflare sometimes maintains a remote index to speed up cache hit resolution without repeatedly querying the OCI manifest layer. 

If you suspect the index is corrupted or out of sync with the actual storage blobs, you can completely wipe the index.

```bash
nix run github:ItzEmoji/aeroflare -- clean-index
```

The index will be naturally rebuilt as new `push` operations occur.
