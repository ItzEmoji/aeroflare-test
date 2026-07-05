package network

import "context"

type JSONBackend struct {
	cfg BackendConfig
}

func (b *JSONBackend) PushReceipts(ctx context.Context, receipts []PushReceipt) error {
	return nil
}
