# Aeroflare Documentation Scaffolding Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Scaffold a new Docusaurus site in the `docs-site` directory (so as not to conflict with the existing `docs/` dir yet) and set up the Diátaxis directory structure so it can be run locally.

**Architecture:** Use `npx create-docusaurus@latest` to generate a classic template site. Configure `docusaurus.config.js` with Aeroflare's branding and remove the default blog/pages. Create the four Diátaxis quadrants as sidebar categories in the `docs/` folder.

**Tech Stack:** Docusaurus, React, Markdown, Node.js

## Global Constraints

- Run in `docs-site` directory.
- Docusaurus version must be `latest`.
- Use npm as the package manager.
- No placeholders in code generation.

---

### Task 1: Scaffold Docusaurus

**Files:**
- Create: `docs-site/package.json` (and full docusaurus structure)

**Interfaces:**
- Consumes: None
- Produces: A runnable Docusaurus site in `docs-site`

- [ ] **Step 1: Run Docusaurus init**

```bash
npx create-docusaurus@latest docs-site classic --typescript --package-manager npm
```
Expected: PASS with "Success! Created docs-site"

- [ ] **Step 2: Commit scaffolding**

```bash
git add docs-site
git commit -m "docs: scaffold docusaurus site"
```
Expected: PASS

### Task 2: Configure Docusaurus for Aeroflare

**Files:**
- Modify: `docs-site/docusaurus.config.ts:1-200`

**Interfaces:**
- Consumes: Docusaurus scaffold
- Produces: Aeroflare branded configuration

- [ ] **Step 1: Replace docusaurus.config.ts**

```typescript
import {themes as prismThemes} from 'prism-react-renderer';
import type {Config} from '@docusaurus/types';
import type * as Preset from '@docusaurus/preset-classic';

const config: Config = {
  title: 'Aeroflare',
  tagline: 'High-performance OCI-backed Nix binary cache proxy',
  favicon: 'img/favicon.ico',
  url: 'https://aeroflare.dev',
  baseUrl: '/',
  organizationName: 'aeroflare',
  projectName: 'aeroflare',
  onBrokenLinks: 'throw',
  onBrokenMarkdownLinks: 'warn',
  i18n: {
    defaultLocale: 'en',
    locales: ['en'],
  },
  presets: [
    [
      'classic',
      {
        docs: {
          sidebarPath: './sidebars.ts',
          routeBasePath: '/',
          editUrl: 'https://github.com/aeroflare/aeroflare/tree/main/docs-site/',
        },
        blog: false,
        theme: {
          customCss: './src/css/custom.css',
        },
      } satisfies Preset.Options,
    ],
  ],
  themeConfig: {
    colorMode: {
      defaultMode: 'dark',
      respectPrefersColorScheme: true,
    },
    navbar: {
      title: 'Aeroflare',
      logo: {
        alt: 'Aeroflare Logo',
        src: 'img/logo.svg',
      },
      items: [
        {
          type: 'docSidebar',
          sidebarId: 'tutorialSidebar',
          position: 'left',
          label: 'Documentation',
        },
        {
          href: 'https://github.com/aeroflare/aeroflare',
          label: 'GitHub',
          position: 'right',
        },
      ],
    },
    footer: {
      style: 'dark',
      links: [],
      copyright: `Copyright © ${new Date().getFullYear()} Aeroflare. Built with Docusaurus.`,
    },
    prism: {
      theme: prismThemes.github,
      darkTheme: prismThemes.dracula,
      additionalLanguages: ['bash', 'json', 'yaml', 'go'],
    },
  } satisfies Preset.ThemeConfig,
};

export default config;
```

- [ ] **Step 2: Commit config update**

```bash
git add docs-site/docusaurus.config.ts
git commit -m "docs: configure docusaurus for aeroflare"
```
Expected: PASS

### Task 3: Set up Diátaxis Directory Structure

**Files:**
- Create: `docs-site/docs/tutorials/_category_.json`
- Create: `docs-site/docs/how-to/_category_.json`
- Create: `docs-site/docs/reference/_category_.json`
- Create: `docs-site/docs/explanation/_category_.json`
- Create: `docs-site/docs/index.md`

**Interfaces:**
- Consumes: Docusaurus config
- Produces: Base Diátaxis structure

- [ ] **Step 1: Clean up default docs**

```bash
rm -rf docs-site/docs/*
mkdir -p docs-site/docs/tutorials docs-site/docs/how-to docs-site/docs/reference docs-site/docs/explanation
```
Expected: PASS

- [ ] **Step 2: Create category configurations**

```bash
cat << 'EOF' > docs-site/docs/tutorials/_category_.json
{
  "label": "Tutorials",
  "position": 1,
  "link": {
    "type": "generated-index",
    "description": "Learning-oriented tutorials for getting started with Aeroflare."
  }
}
EOF

cat << 'EOF' > docs-site/docs/how-to/_category_.json
{
  "label": "How-To Guides",
  "position": 2,
  "link": {
    "type": "generated-index",
    "description": "Task-oriented guides for achieving specific goals."
  }
}
EOF

cat << 'EOF' > docs-site/docs/reference/_category_.json
{
  "label": "Reference",
  "position": 3,
  "link": {
    "type": "generated-index",
    "description": "Information-oriented lookup material."
  }
}
EOF

cat << 'EOF' > docs-site/docs/explanation/_category_.json
{
  "label": "Explanation",
  "position": 4,
  "link": {
    "type": "generated-index",
    "description": "Understanding-oriented deep dives into Aeroflare internals."
  }
}
EOF
```
Expected: PASS

- [ ] **Step 3: Create Index Page**

```bash
cat << 'EOF' > docs-site/docs/index.md
---
sidebar_position: 0
title: Welcome to Aeroflare
---

# Aeroflare Documentation

Aeroflare is a high-performance OCI-backed Nix binary cache proxy and toolkit.

It allows you to seamlessly cache Nix binaries into an OCI registry (like GitHub Packages), speeding up your CI/CD pipelines and local builds. Use it as a proxy cache, or push/pull blobs directly to/from the registry.

Please use the sidebar to navigate through the Tutorials, How-To Guides, Reference material, and Explanations.
EOF
```
Expected: PASS

- [ ] **Step 4: Commit Diátaxis structure**

```bash
git add docs-site/docs
git commit -m "docs: scaffold diataxis structure"
```
Expected: PASS
