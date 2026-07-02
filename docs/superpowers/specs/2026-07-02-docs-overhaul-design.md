# Documentation Overhaul and Homepage Redesign Spec

**Topic:** Redesign landing page, correct doc folder references, and add `--print-out-paths` instructions to documentation.

## Goal
Improve user-facing documentation and landing page aesthetics, ensuring they are polished, complete, and correct.

## Specifications

### 1. Landing Page (`docs/src/pages/index.tsx`)
- Overhaul with a premium hero section, "Get Started" / "GitHub" buttons, and an interactive bento/grid features section displaying stateless proxying, O(1) manifest lookups, dual-backend support, and interactive provisioning.
- Use `framer-motion` for reveal animations.

### 2. General Documentation Links & References (`README.md`)
- Correct references from `docs-site/docs/` to `docs/docs/`.
- Add a note about the necessity of `--print-out-paths` for cache population.

### 3. Command Usage updates in Docs (`docs/docs/how-to/cache-population.md` & `docs/docs/how-to/running-proxy.md`)
- Add the `--print-out-paths` flag to example nix build commands.
- Include a callout warning about why `--print-out-paths` is needed for Aeroflare.
