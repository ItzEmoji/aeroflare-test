package network

import "context"

type R2Backend struct {
	cfg BackendConfig
}

func (b *R2Backend) PushReceipts(ctx context.Context, receipts []PushReceipt) error {
	return nil
}
