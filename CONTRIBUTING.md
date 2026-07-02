# Contributing to Aeroflare

Thank you for your interest in contributing! Here is a guide to getting started.

## Local Development Setup
This project uses Nix flakes for its build and development environment.

1. **Enter the Development Environment**:
   ```bash
   nix develop
   # or nix-shell
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

## Pull Request Guidelines
- Follow Conventional Commits format (e.g., `feat: ...`, `fix: ...`, `docs: ...`).
- Keep pull requests focused on a single change.
- Ensure all tests and lint checks pass.
