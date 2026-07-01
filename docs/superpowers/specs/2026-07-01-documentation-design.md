# Aeroflare Documentation Design Specification

## Overview
This document outlines the design and structure for the complete, authoritative documentation of the Aeroflare project. The documentation will be built using Docusaurus and structured according to the Diátaxis framework, prioritizing a dry, technical, and precise style similar to Playwright and Kubernetes.

## 1. Documentation Structure (Diátaxis)

### Tutorials (Learning-Oriented)
* **Introduction**: What is Aeroflare?
* **Quick Start**: Getting a basic proxy cache running in under 5 minutes.
* **Installation**: Comprehensive steps for Building from Source, Nix Flakes, OCI/Docker images, and GitHub Releases.

### How-To Guides (Task-Oriented)
* **Configuring Storage Backends**: Setting up Cloudflare Integration and GHCR Integration.
* **Authentication and Authorization**: Managing tokens, user IDs, and secrets.
* **Running the Proxy Server**: Operating the proxy daemon in production.
* **Cache Population**: Using the CLI to push builds and blobs to the cache.
* **Using the CLI Toolkit**: Running commands with proxy substituters and generating NARs.
* **Cache Maintenance**: Managing garbage collection and cleaning remote indexes.

### Reference (Information-Oriented)
* **CLI Reference**: Exhaustive documentation of all commands, subcommands, and flags.
* **Configuration**: `aeroflare.yaml`, environment variables, and defaults.
* **API Reference**: Public APIs, network endpoints, and data formats.
* **Repository Layout**: Where everything lives in the codebase.

### Explanation (Understanding-Oriented)
* **Concepts & Glossary**: Core terminology.
* **Architecture & Design Decisions**: Subsystems, invariants, tradeoffs, and failure modes.
* **Nix Binary Cache Protocol**: Deep dive into NAR files and `narinfo` files.
* **OCI Integration**: How Nix artifacts map to OCI registry storage layouts.
* **Signatures and Trust**: The security model and trust assumptions.
* **Performance Characteristics & Limitations**: Expected throughput, constraints, and known bounds.
* **Contributing**: Development setup, CI workflows, and release processes.
* **Troubleshooting & FAQ**: Common errors and diagnostic steps.

## 2. Source Code Mapping

To build the documentation, the following repository areas serve as the source of truth:
* **CLI Reference & Configuration**: Derived directly from `cmd/*.go` and Viper bindings.
* **Architecture, API & Protocol**: Derived from `src/` subdirectories (`proxy/`, `push/`, `prepare/`, `auth/`, `secrets/`) and `src/network.go` / `src/r2.go`.
* **Installation & Deployment**: Derived from `flake.nix`, `go.mod`, and root structure.
* **Release & Contributing**: Derived from `.github/workflows/` and `.github/release-please-config.json`.

## 3. Unclear Areas (Pending Input)
* **Performance Benchmarks**: Latency bounds, throughput numbers, or scale limits to document.
* **Failure Modes**: Expected behavior when rate limits are hit on Cloudflare R2 or GHCR.
* **Trust Assumptions**: Clarification on whether trust relies solely on Nix's Ed25519 signatures or includes an OCI-layer signature verification mechanism.

## 4. Missing Documentation to Create
* **Architecture Diagrams**: Request flow diagrams for the proxy server and storage diagrams mapping Nix NARs to OCI blobs.
* **Troubleshooting Guide**: A matrix of common failure modes and resolutions.
* **Complete `aeroflare.yaml` Reference**: Documenting every configurable key, default value, and environment variable override.
* **Nix to OCI Mapping**: Explicit explanation of how a `narinfo` file translates to an OCI manifest.
