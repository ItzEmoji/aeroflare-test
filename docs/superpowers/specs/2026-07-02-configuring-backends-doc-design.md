# Configuring Storage Backends Documentation Update Design Spec

**Topic:** Update Configuring Storage Backends documentation to correctly reference CLI command roles.

## Goal
Correct the documentation in `docs/docs/how-to/configuring-backends.md` to properly guide users on changing cache backends via `configure` and managing local client settings via `auth` commands/env vars, rather than incorrect use of `settings` or `init`.

## Document Specifications
- Target file: `docs/docs/how-to/configuring-backends.md`
- Design approach: Command-centric restructuring separating initial setup (`init`), configuration changes (`configure`), and client settings (`auth` / env vars).
