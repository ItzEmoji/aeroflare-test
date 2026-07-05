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
		fmt.Printf("Warning: failed to fetch existing index: %v\n", err)
	}

	return UpdateCacheIndex(receipts, existingIndex, b.cfg.Registry, b.cfg.Repository, b.cfg.Token, b.cfg.PubKeyPath, nil, b.cfg.ConfigAnnotations)
}
