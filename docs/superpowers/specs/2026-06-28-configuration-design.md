# Aeroflare Configuration & Viper Integration Design

## Overview
Aeroflare will support a unified configuration system utilizing `github.com/spf13/viper`. This will allow users to configure global settings and command-specific parameters (like the `init` wizard) via CLI flags, environment variables, and a central YAML configuration file.

## Precedence
Configurations will be resolved in the following order of precedence:
1. CLI Flags (e.g., `--cache-url`)
2. Environment Variables (e.g., `AEROFLARE_CACHE_URL`)
3. Configuration File (`aeroflare.yaml`)
4. Default Values

## File Location & Auto-Generation
- **Path**: `$XDG_CONFIG_HOME/aeroflare/aeroflare.yaml` (fallback to `~/.config/aeroflare/aeroflare.yaml`).
- **Initialization**: When Aeroflare starts, if this file does not exist, it will be automatically generated with commented-out defaults to serve as documentation.

## Variables & Environment Handling
- **Prefix**: All environment variables will use the `AEROFLARE_` prefix (`viper.SetEnvPrefix("AEROFLARE")`).
- **Cache vs. Cache URL**:
  - `AEROFLARE_CACHE_URL` (or `--cache-url` flag): Used to specify a full OCI registry URL (e.g., `oci://docker.io/my-org/my-cache`).
  - `AEROFLARE_CACHE`: A convenience variable. If set and `AEROFLARE_CACHE_URL` is empty, it resolves to `ghcr.io/<AEROFLARE_CACHE>`.
- **Init Wizard Support**: The `aeroflare init` command will check Viper for all its required steps (e.g., `AEROFLARE_BACKEND`, `AEROFLARE_GIT_PROVIDER`). If a value is provided via config, env, or flag, the interactive prompt for that step will be skipped.

## Themes
- **Configuration Key**: `theme` (Environment: `AEROFLARE_THEME`).
- **Supported Options**:
  - `default` (Cyan accents)
  - `catppuccin`
  - `gruvbox-dark`
  - `gruvbox-light`
- The `AeroflareTheme()` function in `src/init/theme.go` will be updated to return the corresponding `huh.Theme` based on the configuration.

## Dependencies
- Add `github.com/spf13/viper` to `go.mod`.
