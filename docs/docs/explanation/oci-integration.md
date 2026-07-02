---
sidebar_position: 2
title: OCI Integration Protocol
---

# OCI Integration Protocol

Aeroflare bridges the Nix ecosystem and standard container registries by mapping Nix archive (NAR) artifacts and their metadata directly into Open Container Initiative (OCI) primitives.

## The `.narinfo` to Manifest Mapping

In a traditional Nix binary cache, `.narinfo` files serve as plain-text metadata describing a store path, its dependencies, its compression format, and its cryptographic signatures. 

Aeroflare eliminates the need for a separate `.narinfo` storage mechanism by embedding this metadata directly into the **OCI Image Manifest Annotations**.

When pushing a package:
1. The compressed `.nar` blob is uploaded to the registry as a standard `v1.Layer` with the media type `application/vnd.aeroflare.nar.v1+<compression>`.
2. An OCI Manifest (`types.OCIManifestSchema1`) is created to reference this layer.
3. The contents of the `.narinfo` are serialized into the manifest's annotations using the `vnd.aeroflare.nar.*` namespace.

### Annotation Schema

| Nix `.narinfo` Key | OCI Annotation Key |
|--------------------|--------------------|
| `StorePath` | `vnd.aeroflare.nar.storepath` |
| `URL` | `vnd.aeroflare.nar.url` |
| `Compression` | `vnd.aeroflare.nar.compression` |
| `FileHash` | `vnd.aeroflare.nar.filehash` |
| `FileSize` | `vnd.aeroflare.nar.filesize` |
| `NarHash` | `vnd.aeroflare.nar.narhash` |
| `NarSize` | `vnd.aeroflare.nar.narsize` |
| `References` | `vnd.aeroflare.nar.references` (space-separated) |
| `Deriver` | `vnd.aeroflare.nar.deriver` |
| `System` | `vnd.aeroflare.nar.system` |
| `Sig` | `vnd.aeroflare.nar.sig` |

## On-the-Fly Reconstruction

When the Nix daemon requests a `<hash>.narinfo` file:
1. The Aeroflare proxy translates this into an OCI manifest fetch for `registry/repository:<hash>`.
2. The proxy reads the `vnd.aeroflare.nar.*` annotations from the manifest.
3. It reconstructs the plain-text `.narinfo` file format in memory and serves it to the Nix daemon with the `text/x-nix-narinfo` content type.

This allows standard OCI registries (like GHCR) to act as fully functional Nix binary caches without requiring any custom server-side logic or database indexing.
