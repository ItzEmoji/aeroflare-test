# Aeroflare Documentation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a comprehensive, highly technical Docusaurus documentation site covering all CLI commands and internal modules of Aeroflare.

**Architecture:** Create Markdown/MDX files categorized into CLI Reference and Architecture & Internals, wired into the Docusaurus sidebar.

**Tech Stack:** Docusaurus, Markdown, MDX, TypeScript

## Global Constraints

- Content must be highly technical, explaining the actual code mechanics without unnecessary fluff.
- All docs must be placed in `docs-site/docs/`.
- Markdown files must use proper Docusaurus frontmatter (e.g., `id`, `title`, `sidebar_position`).

---

### Task 1: Update Docusaurus Sidebars

**Files:**
- Modify: `docs-site/sidebars.ts`

**Interfaces:**
- Consumes: N/A
- Produces: Updated sidebar navigation

- [ ] **Step 1: Update sidebars.ts**

Replace the existing `sidebars` object in `docs-site/sidebars.ts` to include the new structure:

```typescript
import type {SidebarsConfig} from '@docusaurus/plugin-content-docs';

const sidebars: SidebarsConfig = {
  tutorialSidebar: [
    {
      type: 'category',
      label: 'CLI Reference',
      items: [
        'cli/core',
        'cli/auth',
        'cli/cache',
        'cli/maintenance',
      ],
    },
    {
      type: 'category',
      label: 'Architecture & Internals',
      items: [
        'internals/architecture',
        'internals/subsystems',
        'internals/proxy-implementations',
        'internals/tasks-ui',
      ],
    },
  ],
};

export default sidebars;
```

- [ ] **Step 2: Commit**

```bash
git add docs-site/sidebars.ts
git commit -m "docs: restructure docusaurus sidebars for new doc layout"
```

### Task 2: Create CLI Core Docs

**Files:**
- Create: `docs-site/docs/cli/core.md`

**Interfaces:**
- Consumes: `cmd/root.go`, `cmd/init.go`, `cmd/configure.go`, `cmd/settings.go`
- Produces: `docs-site/docs/cli/core.md`

- [ ] **Step 1: Read the source files**
Run `cat cmd/root.go cmd/init.go cmd/configure.go cmd/settings.go` to understand the code mechanics.

- [ ] **Step 2: Write core.md**
Write the documentation file covering the global flags, initialization logic, and configuration management. Use Docusaurus frontmatter:
```markdown
---
id: core
title: Core Commands
sidebar_position: 1
---
# Core Commands
...
```

- [ ] **Step 3: Commit**
```bash
git add docs-site/docs/cli/core.md
git commit -m "docs: add cli core reference"
```

### Task 3: Create CLI Auth Docs

**Files:**
- Create: `docs-site/docs/cli/auth.md`

**Interfaces:**
- Consumes: `cmd/auth.go`, `cmd/auth_import.go`, `cmd/auth_list.go`, `cmd/auth_resolve.go`, `cmd/auth_wizard.go`
- Produces: `docs-site/docs/cli/auth.md`

- [ ] **Step 1: Read the source files**
Run `cat cmd/auth*.go` to understand the authentication CLI workflows.

- [ ] **Step 2: Write auth.md**
Write the technical documentation for all auth subcommands.
```markdown
---
id: auth
title: Authentication Commands
sidebar_position: 2
---
# Authentication Commands
...
```

- [ ] **Step 3: Commit**
```bash
git add docs-site/docs/cli/auth.md
git commit -m "docs: add cli auth reference"
```

### Task 4: Create CLI Cache Docs

**Files:**
- Create: `docs-site/docs/cli/cache.md`

**Interfaces:**
- Consumes: `cmd/push.go`, `cmd/run.go`, `cmd/proxy.go`
- Produces: `docs-site/docs/cli/cache.md`

- [ ] **Step 1: Read the source files**
Run `cat cmd/push.go cmd/run.go cmd/proxy.go` to analyze cache operations.

- [ ] **Step 2: Write cache.md**
Write the detailed documentation explaining how paths are pushed and the server is run.
```markdown
---
id: cache
title: Cache Operations
sidebar_position: 3
---
# Cache Operations
...
```

- [ ] **Step 3: Commit**
```bash
git add docs-site/docs/cli/cache.md
git commit -m "docs: add cli cache reference"
```

### Task 5: Create CLI Maintenance Docs

**Files:**
- Create: `docs-site/docs/cli/maintenance.md`

**Interfaces:**
- Consumes: `cmd/clean.go`, `cmd/gc.go`, `cmd/blob.go`, `cmd/prepare.go`, `cmd/scaffold.go`
- Produces: `docs-site/docs/cli/maintenance.md`

- [ ] **Step 1: Read the source files**
Run `cat cmd/clean.go cmd/gc.go cmd/blob.go cmd/prepare.go cmd/scaffold.go` to understand maintenance utilities.

- [ ] **Step 2: Write maintenance.md**
Document the garbage collection, blob inspection, and scaffolding commands.
```markdown
---
id: maintenance
title: Maintenance & Utils
sidebar_position: 4
---
# Maintenance & Utils
...
```

- [ ] **Step 3: Commit**
```bash
git add docs-site/docs/cli/maintenance.md
git commit -m "docs: add cli maintenance reference"
```

### Task 6: Create Architecture & Internals Docs - Overview

**Files:**
- Create: `docs-site/docs/internals/architecture.md`

**Interfaces:**
- Consumes: `src/network.go`, `src/index.go`
- Produces: `docs-site/docs/internals/architecture.md`

- [ ] **Step 1: Read the source files**
Run `cat src/network.go src/index.go`

- [ ] **Step 2: Write architecture.md**
Document the high-level routing, HTTP layer, and cache indexing logic.
```markdown
---
id: architecture
title: Core Architecture
sidebar_position: 1
---
# Core Architecture
...
```

- [ ] **Step 3: Commit**
```bash
git add docs-site/docs/internals/architecture.md
git commit -m "docs: add internals architecture overview"
```

### Task 7: Create Architecture & Internals Docs - Subsystems

**Files:**
- Create: `docs-site/docs/internals/subsystems.md`

**Interfaces:**
- Consumes: `src/auth/*`, `src/secrets/*`, `src/token.go`, `src/r2.go`, `src/push/*`, `src/proxy/*`
- Produces: `docs-site/docs/internals/subsystems.md`

- [ ] **Step 1: Read the source files**
Run `cat src/token.go src/r2.go` and use `grep` or `cat` for the subdirectories.

- [ ] **Step 2: Write subsystems.md**
Provide a deep technical breakdown of how Auth/Secrets validate tokens, how R2 storage integration operates, and how the Push Pipeline parses derivations.
```markdown
---
id: subsystems
title: Core Subsystems
sidebar_position: 2
---
# Core Subsystems
...
```

- [ ] **Step 3: Commit**
```bash
git add docs-site/docs/internals/subsystems.md
git commit -m "docs: add internals subsystems breakdown"
```

### Task 8: Create Architecture & Internals Docs - Proxy Implementations

**Files:**
- Create: `docs-site/docs/internals/proxy-implementations.md`

**Interfaces:**
- Consumes: `proxy/no-webui-json/*`, `proxy/no-webui-native/*`, `proxy/no-webui-r2/*`
- Produces: `docs-site/docs/internals/proxy-implementations.md`

- [ ] **Step 1: Read the proxy implementations**
Inspect the code inside the `proxy/` subdirectories.

- [ ] **Step 2: Write proxy-implementations.md**
Document the differences and mechanics of the three proxy variants.
```markdown
---
id: proxy-implementations
title: Proxy Implementations
sidebar_position: 3
---
# Proxy Implementations
...
```

- [ ] **Step 3: Commit**
```bash
git add docs-site/docs/internals/proxy-implementations.md
git commit -m "docs: add proxy implementations docs"
```

### Task 9: Create Architecture & Internals Docs - Tasks & UI

**Files:**
- Create: `docs-site/docs/internals/tasks-ui.md`

**Interfaces:**
- Consumes: `src/gc.go`, `src/ui/*`
- Produces: `docs-site/docs/internals/tasks-ui.md`

- [ ] **Step 1: Read the source files**
Run `cat src/gc.go` and inspect `src/ui/`.

- [ ] **Step 2: Write tasks-ui.md**
Explain the garbage collection algorithm and the terminal UI rendering engine.
```markdown
---
id: tasks-ui
title: Background Tasks & UI
sidebar_position: 4
---
# Background Tasks & UI
...
```

- [ ] **Step 3: Commit**
```bash
git add docs-site/docs/internals/tasks-ui.md
git commit -m "docs: add tasks and ui internals"
```
