package network

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"aeroflare/src/prepare/narinfo"
	"github.com/google/go-containerregistry/pkg/v1/types"
	"golang.org/x/sync/errgroup"
)

type NativeBackend struct {
	cfg BackendConfig
}

func (b *NativeBackend) PushReceipts(ctx context.Context, receipts []PushReceipt) error {
	eg, _ := errgroup.WithContext(ctx)
	eg.SetLimit(5)

	for _, r := range receipts {
		r := r
		eg.Go(func() error {
			narinfoData, err := os.ReadFile(r.NarinfoPath)
			if err != nil {
				return fmt.Errorf("failed to read narinfo (%s): %w", r.NarinfoPath, err)
			}
			ni, err := narinfo.Parse(string(narinfoData))
			if err != nil {
				return fmt.Errorf("failed to parse narinfo (%s): %w", r.NarinfoPath, err)
			}

			layer, _, err := NewLayerFast(r.NarPath, types.MediaType("application/vnd.aeroflare.nar.v1+"+ni.Compression), ni)
			if err != nil {
				return fmt.Errorf("failed to create layer for (%s): %w", r.NarPath, err)
			}

			basename := filepath.Base(ni.StorePath)
			tag := strings.SplitN(basename, "-", 2)[0]

			err = PushNarPackage(layer, ni, tag, b.cfg.Registry, b.cfg.Repository, b.cfg.Token)
			if err != nil {
				return fmt.Errorf("failed to push native artifact (%s): %w", r.StorePath, err)
			}
			return nil
		})
	}
	return eg.Wait()
}
