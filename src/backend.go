package network

import "context"

type CacheBackend interface {
	PushReceipts(ctx context.Context, receipts []PushReceipt) error
}

type BackendConfig struct {
	Registry          string
	Repository        string
	Token             string
	PubKeyPath        string
	ConfigAnnotations map[string]string
	R2                *R2Config
}

func NewCacheBackend(cfg BackendConfig) CacheBackend {
	backendType := "native"
	if cfg.ConfigAnnotations != nil && cfg.ConfigAnnotations["aeroflare.index-type"] != "" {
		backendType = cfg.ConfigAnnotations["aeroflare.index-type"]
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
