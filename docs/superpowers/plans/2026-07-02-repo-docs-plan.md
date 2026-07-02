# Repository Documentation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add standard GPL-3.0 LICENSE, SECURITY.md, and CONTRIBUTING.md files to the project root.

**Architecture:** Create plain Markdown files in the root folder. LICENSE will contain the verbatim GPL-3.0 license text already fetched from GitHub.

**Tech Stack:** Markdown, git, jq

## Global Constraints
- Target files: `LICENSE`, `SECURITY.md`, `CONTRIBUTING.md`

---

### Task 1: Create SECURITY.md

**Files:**
- Create: `SECURITY.md`

- [ ] **Step 1: Write SECURITY.md file content**
  Write standard security policy outlining GitHub vulnerability reporting.

- [ ] **Step 2: Verify file existence**
  Run: `test -f SECURITY.md`
  Expected: Success

- [ ] **Step 3: Commit SECURITY.md**
  Run: `git add SECURITY.md && git commit -m "docs: add SECURITY.md"`

### Task 2: Create CONTRIBUTING.md

**Files:**
- Create: `CONTRIBUTING.md`

- [ ] **Step 1: Write CONTRIBUTING.md file content**
  Write contribution guidelines covering `nix develop`, formatting, linting, testing, and Conventional Commits.

- [ ] **Step 2: Verify file existence**
  Run: `test -f CONTRIBUTING.md`
  Expected: Success

- [ ] **Step 3: Commit CONTRIBUTING.md**
  Run: `git add CONTRIBUTING.md && git commit -m "docs: add CONTRIBUTING.md"`

### Task 3: Finalize and Commit LICENSE

**Files:**
- Modify: `LICENSE`

- [ ] **Step 1: Verify LICENSE exists and contains GPL-3.0 preamble**
  Run: `head -n 5 LICENSE`
  Expected: Contains "GNU GENERAL PUBLIC LICENSE"

- [ ] **Step 2: Commit LICENSE**
  Run: `git add LICENSE && git commit -m "docs: add GPL-3.0 LICENSE"`
