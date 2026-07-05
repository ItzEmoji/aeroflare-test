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

// Dummy backend implementations to satisfy interface for now
type JSONBackend struct {
	cfg BackendConfig
}

func (b *JSONBackend) PushReceipts(ctx context.Context, receipts []PushReceipt) error {
	return nil
}

type R2Backend struct {
	cfg BackendConfig
}

func (b *R2Backend) PushReceipts(ctx context.Context, receipts []PushReceipt) error {
	return nil
}

type NativeBackend struct {
	cfg BackendConfig
}

func (b *NativeBackend) PushReceipts(ctx context.Context, receipts []PushReceipt) error {
	return nil
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
	case "native":
		fallthrough
	default:
		return &NativeBackend{cfg: cfg}
	}
}
