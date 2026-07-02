---
sidebar_position: 2
title: Authentication & Authorization
---

# Authentication & Authorization

Aeroflare manages credentials securely by integrating directly with your operating system's native keychain or secrets manager. This avoids storing highly sensitive tokens in plain text configuration files.

## Logging In

To authenticate interactively, use the `auth login` command:

```bash
nix run github:ItzEmoji/aeroflare -- auth login
```

This command will prompt you to enter your tokens for various providers (GitHub, GitLab, Cloudflare) and save them securely.

## Importing Credentials

If you already use tools like the GitHub CLI (`gh`), GitLab CLI (`glab`), or Docker (`docker login`), Aeroflare can import your existing sessions:

```bash
nix run github:ItzEmoji/aeroflare -- auth import
```

## Managing Saved Secrets

You can view which keys are currently stored in your keychain (values will be hidden):

```bash
nix run github:ItzEmoji/aeroflare -- auth list
```

To remove a compromised or expired token:

```bash
nix run github:ItzEmoji/aeroflare -- auth remove github-token
```

To manually set an arbitrary secret:

```bash
nix run github:ItzEmoji/aeroflare -- auth set my-custom-key "my-secret-value"
```

## CI/CD Environments

In headless environments like GitHub Actions, interactive login is not possible. Instead, you can provide authentication tokens directly via environment variables:

```bash
export AEROFLARE_GITHUB_TOKEN="${{ secrets.GITHUB_TOKEN }}"
```

Aeroflare will prioritize these environment variables over the local secrets manager.
