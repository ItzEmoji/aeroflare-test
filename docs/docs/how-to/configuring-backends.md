---
sidebar_position: 1
title: Configuring Storage Backends
---

# Configuring Storage Backends

Aeroflare supports two primary types of storage backends for your Nix binaries: **OCI Registries** (like GitHub Container Registry) and **Cloudflare R2** buckets.

## Initial Setup
To provision resources and configure a cache for the first time, run the interactive onboarding wizard:
```bash
nix run github:ItzEmoji/aeroflare -- init
```

---

## Configuring or Changing the Backend

To change the storage backend (index type) or the Nix public signing key for an existing cache, use the `configure` command:
```bash
nix run github:ItzEmoji/aeroflare -- configure
```
This updates the cache metadata (OCI manifest annotations) directly in the registry. You will be prompted to choose:
- **Cloudflare R2 (Recommended)**: Best performance, dual-backend support.
- **Native OCI Tags (Experimental)**: Fully stateless OCI indexing.
- **cache-index.json (Not Recommended)**: Traditional single-file index.

---

## Client Settings and Credentials

To configure registry logins and R2 tokens on your local machine, you can use the interactive authentication commands or environment variables.

### Interactive CLI Authentication
Run the login command to configure credentials interactively:
```bash
nix run github:ItzEmoji/aeroflare -- auth login
```
This wizard securely stores tokens (GitHub, GitLab, or Cloudflare API tokens) in your local credential store.

You can also manage credentials manually:
- **Set a token**: `nix run github:ItzEmoji/aeroflare -- auth set [key] [value]` (e.g. `cf-token`, `github-token`, `gitlab-token`, `cf-user-id`)
- **List stored keys**: `nix run github:ItzEmoji/aeroflare -- auth list`
- **Remove a credential**: `nix run github:ItzEmoji/aeroflare -- auth remove [key]`

### Environment Variables
Alternatively, you can configure credentials directly via environment variables:

| Variable | Description |
| :--- | :--- |
| `GITHUB_TOKEN` / `GH_TOKEN` | Authentication token for GitHub Packages (ghcr.io) |
| `R2_BUCKET` | Cloudflare R2 bucket name |
| `R2_ENDPOINT` | Cloudflare R2 S3 API Endpoint |
| `R2_ACCESS_KEY_ID` | Cloudflare R2 Access Key ID |
| `R2_SECRET_ACCESS_KEY` | Cloudflare R2 Secret Access Key |
| `R2_PUBLIC_URL` | Cloudflare R2 Public bucket URL (e.g., `https://pub-xxx.r2.dev`) |
