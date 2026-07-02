# Aeroflare

> High-performance OCI-backed Nix binary cache proxy written in Go.

Aeroflare bridges the Nix ecosystem and standard container registries (such as GitHub Container Registry or Docker Hub) to act as a stateless, zero-infrastructure binary substituter.

---

## Key Features

- **Stateless Proxying**: Retains zero local binary state. Streams `.nar` blobs directly from OCI.
- **O(1) Manifest Lookups**: Tags artifacts directly with the 32-character Nix store path hash, enabling instantaneous lookups.
- **Interactive Provisioning**: A built-in setup wizard for GitHub, GitLab, and Cloudflare R2 bucket configuration.
- **Execution Wrapper**: Run builds transparently with the `run` wrapper (`aeroflare run -- nix build`).
- **Dual-Backend Support**: Use OCI registries for heavy NAR blobs and Cloudflare R2 for fast metadata (`narinfo`).

---

## Quick Start

### 1. Initialize
Run the interactive onboarding wizard to configure credentials and provision resources:
```bash
nix run github:ItzEmoji/aeroflare -- init
```
### 3. Build & Cache
Execute a build and automatically push the outputs:
```bash
nix run github:ItzEmoji/aeroflare -- run -- nix build .#default --print-out-paths
```

---

## Codebase Directory Map

```
.
├── cmd/                # Cobra CLI commands (flag parsing, Viper configurations)
│   ├── root.go         # Entry point, environment bindings
│   ├── proxy.go        # Proxy CLI command definition
│   ├── run.go          # CLI command for build wrapper
│   └── ...             # Settings, auth, gc, and clean-index CLI commands
├── src/                # Core logic packages (decoupled from Cobra/CLI)
│   ├── auth/           # OAuth token flows, Device authorization
│   ├── secrets/        # Credentials manager (keyring + plaintext fallback)
│   ├── proxy/          # HTTP server, proxy handlers, token management
│   ├── prepare/        # Local NAR serialization, compression, and signing
│   ├── network.go      # OCI network layer (uses go-containerregistry)
│   ├── index.go        # JSON cache-index schema and update logic
│   ├── r2.go           # Cloudflare R2 / S3 client integration
│   └── gc.go           # BFS-based garbage collection algorithm
└── docs-site/          # Docusaurus documentation website
```

---

## Documentation

For full guides, reference manuals, and architecture explanations, check out the [documentation site](https://aeroflare.pages.dev) or browse [docs-site/docs/](docs-site/docs/).
