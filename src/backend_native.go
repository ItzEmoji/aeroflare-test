package network

import "context"

type NativeBackend struct {
	cfg BackendConfig
}

func (b *NativeBackend) PushReceipts(ctx context.Context, receipts []PushReceipt) error {
	return nil
}
