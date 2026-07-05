package network

import "context"

// CacheBackend publishes completed pushes (nar + narinfo pairs) to wherever
// the cache stores its index and blobs. Implementations: JSONBackend (a
// single cache-index.json manifest), NativeBackend (one OCI image per
// package), and R2Backend (blobs in R2 plus chunked manifests).
type CacheBackend interface {
	PushReceipts(ctx context.Context, receipts []PushReceipt) error
}

// BackendConfig holds the destination and credentials shared by all
// CacheBackend implementations.
type BackendConfig struct {
	Registry          string
	Repository        string
	Token             string
	PubKeyPath        string
	ConfigAnnotations map[string]string
	R2                *R2Config
}

// NewCacheBackend selects a CacheBackend based on cfg: an R2 config with a
// public URL forces the "r2" backend; otherwise the aeroflare.backend
// annotation picks the backend, defaulting to "native".
func NewCacheBackend(cfg BackendConfig) CacheBackend {
	backendType := "native"
	if cfg.ConfigAnnotations != nil {
		if t := cfg.ConfigAnnotations["aeroflare.backend"]; t != "" {
			backendType = t
		}
	}
	if cfg.R2 != nil && cfg.R2.PublicURL != "" {
		backendType = "r2"
	}

	switch backendType {
	case "json":
		return &JSONBackend{cfg: cfg}
	case "r2":
		return &R2Backend{cfg: cfg}
	default:
		return &NativeBackend{cfg: cfg}
	}
}
