package network

import (
	"context"
	"fmt"
	"math/rand"
	"time"
)

type JSONBackend struct {
	cfg BackendConfig
}

// PushReceipts merges receipts into the remote cache index with optimistic
// concurrency: after writing, it re-reads the index manifest digest and, if
// another writer overwrote our update in the meantime, re-fetches, re-merges
// the receipts, and writes again.
func (b *JSONBackend) PushReceipts(ctx context.Context, receipts []PushReceipt) error {
	const maxAttempts = 4

	for attempt := 1; ; attempt++ {
		if err := ctx.Err(); err != nil {
			return err
		}

		existingIndex, _, err := fetchCacheIndexWithDigest(b.cfg.Registry, b.cfg.Repository, b.cfg.Token)
		if err != nil {
			return fmt.Errorf("failed to fetch existing index: %w", err)
		}

		writtenDigest, err := updateCacheIndexWithDigest(receipts, existingIndex, b.cfg.Registry, b.cfg.Repository, b.cfg.Token, b.cfg.PubKeyPath, b.cfg.ConfigAnnotations, b.cfg.Workers)
		if err != nil {
			return err
		}

		// Verify our write is still the current index; if not, a concurrent
		// writer clobbered it and our entries were lost from the remote index.
		_, currentDigest, err := fetchCacheIndexWithDigest(b.cfg.Registry, b.cfg.Repository, b.cfg.Token)
		if err != nil || currentDigest == "" {
			// The write itself succeeded; verification is best-effort. An
			// empty digest (manifest not yet visible) is read-after-write
			// lag, not evidence of a concurrent overwrite.
			return nil
		}
		if currentDigest == writtenDigest {
			return nil
		}

		if attempt >= maxAttempts {
			return fmt.Errorf("cache index overwritten by concurrent writers %d times; giving up", maxAttempts)
		}
		// Brief jittered pause so competing writers interleave instead of
		// repeatedly colliding.
		time.Sleep(time.Duration(100+rand.Intn(400)) * time.Millisecond)
	}
}
