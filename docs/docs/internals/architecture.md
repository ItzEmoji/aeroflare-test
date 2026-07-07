---
id: architecture
title: Core Architecture
sidebar_position: 1
---

# Core Architecture

Aeroflare's core architecture relies on an optimized networking layer for OCI interactions and a fully stateless, native OCI storage model. This document outlines the inner workings of `network.go` and `index.go`, targeting contributors and developers who need to understand the exact code mechanics and struct shapes utilized within the system.

## HTTP Layer and Routing (`network.go`)

Aeroflare operates primarily as an HTTP client interacting with OCI registries. Rather than a traditional HTTP server router, its "routing" and HTTP request handling are built around customized transports and registry clients.

### Optimized Transport and Logging
All outgoing HTTP requests to OCI registries utilize a tailored `http.Transport` wrapped in a custom `loggingTransport`.

*   **`optimizedTransport`**: Configured to force HTTP/2 (`ForceAttemptHTTP2: true`) and tuned for high concurrency.
    *   `MaxIdleConns`: 1000
    *   `MaxIdleConnsPerHost`: 100
    *   `MaxConnsPerHost`: 100
    *   Timeout settings include a 30-second dial timeout, 90-second idle connection timeout, and 10-second TLS handshake timeout.
*   **`loggingTransport`**: Implements `http.RoundTripper` to intercept requests. When `DebugLogger` is enabled, it logs the HTTP method and URL for every outgoing request before delegating to the underlying transport.

### Core OCI Request Handlers (Client-Side)
Interactions with the registry are encapsulated in specific functions that act as the primary operational handlers:
*   **`PushBlob` & `PushLayer`**: Compute the SHA256 digest of a local file to create a `fileLayer` (which implements `v1.Layer`). `PushLayer` uses `remote.WriteLayer` to stream this layer to the registry.
*   **`PushNarPackage`**: Wraps a layer into an OCI image (`types.OCIManifestSchema1`). It annotates the manifest with `vnd.aeroflare.nar.*` metadata (e.g., `storepath`, `filehash`, `narhash`). Crucially, it enforces an **O(1) Lookup Tagging Rule**: the image tag is strictly set to the 32-character Nix hash (e.g., `xn2nlmvng2im9mgrq46y3wkbz4ll1hnp`) to allow direct pulls without traversing an index.
*   **`PullOCINativeManifest`**: Fetches the image manifest by the 32-character tag and reconstructs the `narinfo.Narinfo` struct entirely from the manifest annotations.

## Storage Model and Metadata (`index.go`)

Aeroflare is stateless: there is no central index to maintain. Each Nix package
is published as its own OCI image tagged with its store hash, and its `.narinfo`
metadata is carried directly in the image's manifest annotations. The only
shared object is a small `cache-config` manifest that holds cache-wide settings
(currently just the public signing key).

### Metadata Struct Shapes

The push pipeline threads results through a single struct:

```go
type PushReceipt struct {
	StorePath   string
	NarinfoPath string
	NarDigest   string
	NarSize     int64
	NarPath     string
	Compression string
	IsRoot      bool
}
```

### Config Manifest Logic

`PushConfigManifest` writes the `cache-config` manifest: an empty OCI image
(artifact type `application/vnd.aeroflare.cache-config.v1`) whose annotations
carry cache-wide settings such as `aeroflare.public-key`. The proxy reads these
annotations to resolve the public key. Package metadata itself never touches
this manifest — it lives on each package's own image (see `PushNarPackage`).
