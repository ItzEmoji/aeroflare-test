---
id: maintenance
title: Maintenance & Utils
sidebar_position: 4
---

# Maintenance & Utils

Aeroflare provides a suite of maintenance and utility commands designed for managing remote indexes, garbage collecting orphaned assets, interacting directly with blobs, and scaffolding local environments. These commands are critical for operators directly manipulating the cache states and for developers extending Aeroflare's core workflows.

## `clean-index`

The `clean-index` command completely wipes the remote cache index from the registry.

**Usage:**
```bash
aeroflare clean-index
```

### Technical Mechanics

When executed, this command requests user confirmation before proceeding. Once confirmed, it resets the upstream cache index:

1. **Authentication:** Fetches credentials via `network.GetToken()` leveraging the `oci_token`, `GITHUB_TOKEN`, or `GH_TOKEN` environment variables.
2. **Index Reset:** Constructs an empty `PushCacheIndex` containing an empty `Entries` map and an empty `GCRoots` slice.
3. **Registry Update:** Uses `proxy.BootstrapConfigWithAnnotations` to resolve network parameters, configures R2 via `network.GetR2Config`, and executes `network.UpdateCacheIndex()` to push the empty state back to the OCI registry backend.


## `push-blob` and `pull-blob`

These lower-level utilities offer direct OCI blob interactions. They are heavily utilized for manual state debugging or bespoke cache layer manipulations without relying on Nix commands.

**Usage:**
```bash
aeroflare push-blob <file-path>
aeroflare pull-blob <digest> <output-file>
```

### Technical Mechanics

- **`push-blob`**: Resolves the target registry and repository configuration, grabs the appropriate token, and wraps `network.PushBlob(filePath, ...)`. It streams the given file payload directly to the OCI registry as a single blob and returns its final computed digest.
- **`pull-blob`**: Wraps `network.PullBlob(digest, outFile, ...)`. It queries the registry for the specified blob digest and streams the bytes down, writing them locally to `outFile`.

## `prepare`

The `prepare` command is essential for processing Nix store paths into valid NAR archives and corresponding `narinfo` metadata prior to registry submission.

**Usage:**
```bash
aeroflare prepare --store-path <path> [flags]
aeroflare prepare --input <file> [flags]
```

**Flags:**
- `--store-path` / `--input`: Pass a single store path or a batch input file.
- `--output-dir` (default: `./output`): Directory for generated `.nar` and `.narinfo` assets.
- `--compression` (default: `zstd`): Formats supported via `compress.ParseType()` (zstd, xz, gzip, none).
- `--workers` (default: `50`): Parallelization limit for batch workloads.
- `--prepare-refs`: Enable recursive preparation of missing references one-level deep.
- `--signing-key`: Provide an ed25519 signing key generated via `nix key-gen-secret`.
- `--upstream-cache` (default: `https://cache.nixos.org`): The upstream cache for evaluating missing store references.

### Technical Mechanics

1. **Input & Config:** Constructs a `prepare.Config` populated with parsed compression routines, the parsed `signing.PrivateKey`, and concurrency parameters.
2. **Processing Pipelines:** 
   - Single targets evaluate through `prepare.Prepare()`.
   - Batch targets utilize `prepare.PrepareBatch()`, dispatching to a worker pool bound by the `--workers` flag.
3. **Reference Tracking:** It evaluates `.narinfo` files against the configured `--upstream-cache`. If `--prepare-refs` is true, the resolver fetches missing downstream dependency boundaries (one level deep) and prepares them concurrently. The results compile into a `prepare.Result` struct, detailing `MissingRefs`, `MissingRefResults`, and boolean signature status (`Signed`).

## `scaffold`

The `scaffold` command pulls a specified release from GitHub and establishes a local worker environment pre-configured with Cloudflare bindings.

**Usage:**
```bash
aeroflare scaffold [flags]
```

**Flags:**
- `--release`: Optional target GitHub release tag.
- `--output-dir` (default: `./aeroflare-proxy`): Local path to extract proxy configurations.

### Technical Mechanics

1. **Release Resolution:** If `--release` is not explicitly declared, the CLI invokes an interactive prompt (using `huh` UI components) after fetching available releases from the GitHub API.
2. **Extraction:** Downloads the corresponding tarball from `github.com/ItzEmoji/aeroflare/archive/refs/tags/...` and pipes it via `exec.Command` directly into `tar -xz -C <targetDir> --strip-components=1`.
3. **Backend Injection:** The user selects a backend mode (e.g., `Cloudflare R2`, `Native OCI Tags`, or `JSON index in OCI`). This choice dictates which subdirectory (e.g. `proxy/no-webui-r2`) is utilized.
4. **Configuration Patching:** Executes `patchWranglerToml()`, reading `AEROFLARE_REGISTRY` / `NIXCACHE_REGISTRY` variables. It performs string substitutions on the template `wrangler.toml` file to inject runtime parameters (e.g. replacing `# NIXCACHE_REPO = "<NIXCACHE_REPO>"` and appending `[[r2_buckets]]` definitions when R2 is the chosen backend adapter).
