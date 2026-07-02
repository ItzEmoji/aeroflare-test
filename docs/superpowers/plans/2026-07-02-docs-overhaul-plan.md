# Documentation Overhaul and Homepage Redesign Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Redesign the landing page, update docs command references to include `--print-out-paths`, and correct repository documentation folder paths in the README.

**Architecture:** Write updated pages in `docs/` and `README.md`.

**Tech Stack:** React, Docusaurus, Framer Motion, Markdown, git

## Global Constraints
- Target files:
  - `docs/src/pages/index.tsx`
  - `README.md`
  - `docs/docs/how-to/cache-population.md`
  - `docs/docs/how-to/running-proxy.md`

---

### Task 1: Redesign Docusaurus Landing Page

**Files:**
- Modify: `docs/src/pages/index.tsx`

- [ ] **Step 1: Replace index.tsx content**
  Over-write the content of `docs/src/pages/index.tsx` with the new approved structure containing the features section and buttons.

- [ ] **Step 2: Verify file syntax**
  Check that the file imports and exports correctly.

- [ ] **Step 3: Commit changes**
  Run: `git add docs/src/pages/index.tsx && git commit -m "docs: redesign Docusaurus homepage for premium look and clear onboarding"`

### Task 2: Correct Command References in How-To Guides

**Files:**
- Modify: `docs/docs/how-to/cache-population.md`
- Modify: `docs/docs/how-to/running-proxy.md`

- [ ] **Step 1: Update cache-population.md**
  Add `--print-out-paths` to the run command block and include the important notice.

- [ ] **Step 2: Update running-proxy.md**
  Add `--print-out-paths` to the run command block and include the important notice.

- [ ] **Step 3: Commit how-to guides**
  Run: `git add docs/docs/how-to/cache-population.md docs/docs/how-to/running-proxy.md && git commit -m "docs: add print-out-paths instructions to how-to guides"`

### Task 3: Correct README.md Documentation Reference

**Files:**
- Modify: `README.md`

- [ ] **Step 1: Update README.md**
  Correct `docs-site/docs/` to `docs/docs/` on lines 52 and 59, and add a note about `--print-out-paths` on line 29.

- [ ] **Step 2: Commit README.md**
  Run: `git add README.md && git commit -m "docs: correct doc folder reference and add print-out-paths note in README"`
