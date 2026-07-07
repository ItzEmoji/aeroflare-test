---
id: proxy-modes
title: Storage Model
sidebar_position: 2
---

Aeroflare stores and serves Nix cache data (narinfos) using native OCI tags.

## Native OCI Tags

Aeroflare leverages OCI (Open Container Initiative) tags as the "index" for
storing the `narinfos`.

This is a clever approach where the registry's tagging system itself is used to
keep track of the cached artifacts: each Nix package becomes an OCI image tagged
with its store hash, and the `.narinfo` metadata rides along in the image's
manifest annotations. It takes advantage of standard OCI registry features,
ensuring broad compatibility with registries that fully support OCI tags,
without requiring any extra storage layer or a central index to keep in sync.
