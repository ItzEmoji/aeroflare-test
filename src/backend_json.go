package network

import (
	"context"
	"fmt"
)

type JSONBackend struct {
	cfg BackendConfig
}

func (b *JSONBackend) PushReceipts(ctx context.Context, receipts []PushReceipt) error {
	existingIndex, err := FetchCacheIndex(b.cfg.Registry, b.cfg.Repository, b.cfg.Token)
	if err != nil {
		return fmt.Errorf("failed to fetch existing index: %w", err)
	}

	return UpdateCacheIndex(receipts, existingIndex, b.cfg.Registry, b.cfg.Repository, b.cfg.Token, b.cfg.PubKeyPath, b.cfg.ConfigAnnotations)
}
