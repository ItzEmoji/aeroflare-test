---
id: core
title: Core Commands
sidebar_position: 1
---
# Core Commands

This document covers the core commands and global settings for the Aeroflare CLI, focusing on initialization, configuration, and settings management.

## Global Flags

The `aeroflare` root command supports the following global flags that apply to all subcommands:

- `-v`, `--verbose`: Enable verbose output. Use `-v` for package logs, and `-vv` for request logs.
- `--cache-url`: OCI registry URL for the cache.
- `--github-token`: GitHub API Token.
- `--gitlab-token`: GitLab API Token.
- `--cf-token`: Cloudflare API Token.
- `--cf-user-id`: Cloudflare Account ID.

These flags can also be provided via environment variables with the `AEROFLARE_` prefix (e.g., `AEROFLARE_CACHE_URL`).

## aeroflare init

Initialize Aeroflare infrastructure via an interactive wizard.

```bash
aeroflare init
```

The setup wizard will provision all required infrastructure:
- OCI repository for storing cache data
- Cloudflare R2 bucket (if selected)
- Cloudflare Worker deployment
- Git repository and CI/CD integration (if selected)

The wizard asks all questions up front and displays a summary before making any changes. No infrastructure is created until you confirm.

## aeroflare configure

Interactively configure the cache backend and related settings.

```bash
aeroflare configure
```

This command allows you to change the underlying storage mechanism and its settings, saving the configuration to OCI manifest annotations.
Supported backends include:
- Cloudflare R2 (Recommended)
- Native OCI Tags (Experimental)
- `cache-index.json` (Not Recommended)

## aeroflare settings

Configure Aeroflare interactively through a user-friendly terminal UI.

```bash
aeroflare settings
```

This command opens an interactive menu to adjust:
- **Appearance Theme:** Choose between Catppuccin, Gruvbox Dark, Gruvbox Light, or the Default Terminal theme.
- **Registry Login & Setup:** Configure authentication for GitHub Packages (ghcr.io), GitLab Registry, Cloudflare R2, or a custom OCI registry.

Settings are saved locally to your `aeroflare.yaml` configuration file, typically located at `~/.config/aeroflare/aeroflare.yaml` or `$XDG_CONFIG_HOME/aeroflare/aeroflare.yaml`.
