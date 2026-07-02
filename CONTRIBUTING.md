# Contributing to Aeroflare

Thank you for your interest in contributing! Here is a guide to getting started.

## Local Development Setup
This project uses Nix flakes for its build and development environment.

1. **Enter the Development Environment**:
   ```bash
   nix develop
   ```
   This automatically installs the compatible Go compiler, linter (`golangci-lint`), and vendor tools.

2. **Code Style & Formatting**:
   Ensure code is formatted:
   ```bash
   go fmt ./...
   ```

3. **Running the Linter**:
   Verify code linting:
   ```bash
   golangci-lint run
   ```

4. **Running Tests**:
   Run tests before submitting changes:
   ```bash
   go test ./... -v
   ```

## Dependency Management
The project uses `govendor` to manage Nix dependency locks in `govendor.toml`.
- After modifying `go.mod`, run `govendor` in the root of the project to update `govendor.toml`.
- Run `govendor --check` to check if manifests have drifted and need updating.

## Pull Request Guidelines
- Follow Conventional Commits format (e.g., `feat: ...`, `fix: ...`, `docs: ...`).
- Keep pull requests focused on a single change.
- Ensure all tests and lint checks pass.
