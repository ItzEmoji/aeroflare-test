---
id: core
title: Core Commands
sidebar_position: 1
---
# Core Commands

This document covers the internal implementation mechanics of the core commands and global state management for the Aeroflare CLI, implemented across `cmd/root.go`, `cmd/init.go`, `cmd/configure.go`, and `cmd/settings.go`.

## Global Flags and Environment State (`cmd/root.go`)

The `aeroflare` root command uses Viper to manage global state and flag bindings:

- `--verbose` (`-v`, `-vv`): Controls the `network.DebugLogger` boolean. `-vv` triggers request-level logging (`VerboseCount >= 2`).
- `--cache-url`, `--github-token`, `--gitlab-token`, `--cf-token`, `--cf-user-id`: Core configuration flags.

Viper automatically maps all flags to environment variables by applying the `AEROFLARE_` prefix and replacing dashes with underscores (e.g., `--cache-url` maps to `AEROFLARE_CACHE_URL`).

Additionally, `cmd/root.go` binds a specific alias, `AEROFLARE_CACHE`. When this environment variable is used instead of `AEROFLARE_CACHE_URL`, the `GetCacheURL()` function automatically prefixes the provided cache name with `ghcr.io/` (e.g., `AEROFLARE_CACHE=my-org/my-cache` becomes `ghcr.io/my-org/my-cache`).

## aeroflare init (`cmd/init.go`)

The `init` command orchestrates the `setup.RunWizard()` and `setup.RunProvision(cfg)` flow. 

Crucially, before invoking the provisioner, `init.go` dynamically mutates the process environment by exporting required tokens so that underlying Terraform/Pulumi or SDK calls can authenticate:
- If the R2 backend is selected, it exports `CLOUDFLARE_API_TOKEN` and `CLOUDFLARE_ACCOUNT_ID`.
- If GitHub Registry (`ghcr.io`) or GitHub Git Provider is selected, it exports `GITHUB_TOKEN`.

## aeroflare configure (`cmd/configure.go`)

The `configure` command reads and writes cache backend configuration directly to OCI manifest annotations using `network.PushConfigManifest`.

It manipulates the following specific OCI annotation keys:
- `aeroflare.backend`: The storage backend (`r2`, `native`, or `json`).
- `aeroflare.public-key`: The nix cache public key.
- `aeroflare.r2.bucket`: The R2 bucket name.
- `public-r2-url`: The R2 public URL (e.g., `https://pub-xxx.r2.dev`).
- `aeroflare.r2.endpoint`: The S3 API endpoint for R2.

As a cleanup mechanic, if the user migrates away from the `json` backend (e.g., changing to `r2` or `native`), the code actively issues a deletion request for the `cache-index` image tag via `network.DeleteTag()` to prevent stale JSON indexes from lingering in the registry.

## aeroflare settings (`cmd/settings.go`)

The `settings` command provides an interactive UI, but under the hood, it directly manipulates the Viper configuration store and flushes the changes to disk (`viper.WriteConfig()`).

Depending on the user's choices in the `huh` forms, it mutates the following internal Viper configuration keys:
- `theme`: UI color scheme (e.g., `catppuccin`, `gruvbox-dark`).
- `git-provider`: Set to `github`, `gitlab`, or `none`.
- `git-token`: Stores the corresponding GitHub or GitLab token.
- `cloudflare-api-token`: Stores the Cloudflare token if the R2 registry is selected.
- `cache-url`: Stores the custom OCI registry URL.
