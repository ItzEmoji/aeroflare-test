# Aeroflare Documentation Design Spec

## Overview
This spec outlines the structure and design for the Aeroflare documentation, which will be built using Docusaurus in the `docs-site/` directory. The goal is to provide a detailed, technical, and fluff-free explanation of every single module in the project.

## Structure
The documentation is split into two primary sections to separate user-facing operations from internal developer architecture.

### 1. CLI Reference (User Manual)
Covers the `cmd/` package, explaining command usage, flags, and configuration.
- **Getting Started / Core**: Root `aeroflare`, `aeroflare init`, `aeroflare configure`, `aeroflare settings`
- **Authentication**: `aeroflare auth` and its subcommands (`import`, `list`, `resolve`, `wizard`)
- **Cache Operations**: `aeroflare push`, `aeroflare run`, `aeroflare proxy`
- **Maintenance & Utils**: `aeroflare clean`, `aeroflare gc`, `aeroflare blob`, `aeroflare prepare`, `aeroflare scaffold`

### 2. Architecture & Internals (Developer Docs)
Covers the `src/` and `proxy/` packages, diving into the core implementation.
- **Core Architecture**: Data flow, `network.go`, `index.go`
- **Core Subsystems**:
  - Auth & Secrets: `src/auth/`, `src/secrets/`, `token.go`
  - Storage Engine: `r2.go`
  - Push Pipeline: `src/push/`, `src/prepare/`
  - Proxy Engine: `src/proxy/`
- **Proxy Implementations**: Details on `proxy/no-webui-json`, `proxy/no-webui-native`, `proxy/no-webui-r2`
- **Background Tasks & UI**: `gc.go` and `src/ui/`

## Implementation Details
- Markdown/MDX files will be placed in `docs-site/docs/`
- `docs-site/sidebars.ts` will be updated to reflect this two-section structure
- Content must be highly technical, explaining the actual code mechanics without unnecessary fluff.
