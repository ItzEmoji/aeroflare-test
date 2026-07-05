---
id: architecture
title: Core Architecture
sidebar_position: 1
---

# Core Architecture

Aeroflare's core architecture relies on an optimized networking layer for OCI interactions and a structured approach to cache indexing. This document outlines the inner workings of `network.go` and `index.go`, targeting contributors and developers who need to understand the exact code mechanics and struct shapes utilized within the system.

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

## Cache Indexing and Metadata (`index.go`)

To maintain the state of the Nix cache, Aeroflare stores an index directly within the OCI registry as a custom manifest named `cache-index`.

### Metadata Struct Shapes

The cache index is marshaled from and to the following specific Go structs:

```go
type PushCacheIndex struct {
	Version   int                       `json:"version"`
	Repo      string                    `json:"repo"`
	Registry  string                    `json:"registry"`
	Image     string                    `json:"image"`
	Generated string                    `json:"generated"`
	PublicKey string                    `json:"public_key"`
	Entries   map[string]PushCacheEntry `json:"entries"`
	GCRoots   []string                  `json:"gc_roots"`
}

type PushCacheEntry struct {
	Name      string `json:"name"`
	Narinfo   string `json:"narinfo"`
	NarDigest string `json:"nar_digest"`
	NarSize   int64  `json:"nar_size"`
	Added     string `json:"added"`
}

type PushReceipt struct {
	StorePath   string
	NarinfoPath string
	NarDigest   string
	NarSize     int64
	IsRoot      bool
}
```

### Cache Index Management Logic

1.  **Fetching the Index (`FetchCacheIndex`)**:
    *   Constructs a `GET` request to `[scheme]://[registry]/v2/[repository]/manifests/cache-index`.
    *   Expects the `Accept` header to be `application/vnd.oci.image.manifest.v1+json`.
    *   If successful, it extracts the first layer digest from the manifest, downloads the blob via `PullBlob`, and unmarshals it into a `PushCacheIndex` struct.

2.  **Updating the Index (`UpdateCacheIndex`)**:
    *   Takes an array of `PushReceipt` objects and merges them into the `existingIndex`.
    *   For each receipt, it reads the `.narinfo` file, splits the basename to extract the hash and name, and inserts a `PushCacheEntry` into the `Entries` map.
    *   If a receipt is marked as `IsRoot`, its hash is appended to `GCRoots` (ensuring uniqueness via a set and then sorting alphabetically).
    *   **Manifest Construction**:
        *   The updated `PushCacheIndex` is marshaled into JSON and pushed as a blob (`application/vnd.nix.cache.index.v1+json`).
        *   An empty configuration JSON blob containing a `created` timestamp is pushed (`application/vnd.oci.image.config.v1+json`).
        *   An OCI manifest is constructed linking these two digests. It includes annotations such as `aeroflare.backend` (defaulting to `"json"`, or `"r2"` if an R2 config is provided) and `aeroflare.public-key`.
    *   Finally, the assembled manifest is pushed to the registry at the `cache-index` tag using a `PUT` request.
