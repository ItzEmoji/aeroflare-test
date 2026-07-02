---
id: proxy-modes
title: Proxy Modes
sidebar_position: 2
---

Aeroflare's proxy supports three different operational modes for storing and serving Nix cache data (narinfos). Understanding these modes will help you choose the best configuration for your performance, storage, and compatibility needs.

## `native-OCI-Tags`

The `native-OCI-Tags` mode leverages OCI (Open Container Initiative) tags as an "index" to store the `narinfos`. 

This is a clever approach where the registry's tagging system itself is used to keep track of the cached artifacts. It takes advantage of standard OCI registry features, ensuring broad compatibility with registries that fully support OCI tags, without requiring extra storage layers just for an index.

## `json`

The `json` mode stores the index in a JSON format. This approach is very similar to the `cache-index.json` method used by the `nixcache-oci` project.

By maintaining a dedicated JSON index, this mode can be easier to inspect and parse manually or with external tools. It's a great choice if you need the cache index to be accessible as a simple, human-readable file, or if you are migrating from or interoperating with systems familiar with `nixcache-oci`.

## `r2`

The `r2` mode saves the `narinfos` directly into a Cloudflare R2 bucket.

If your infrastructure relies heavily on Cloudflare or you want to take advantage of R2's cost-effective, globally distributed object storage, this mode is ideal. Instead of storing the index data inside the OCI registry itself, Aeroflare offloads the `narinfos` to R2, potentially improving scalability and reducing registry bloat.
