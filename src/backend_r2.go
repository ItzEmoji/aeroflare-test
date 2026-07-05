package network

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"golang.org/x/sync/errgroup"
)

type R2Backend struct {
	cfg BackendConfig
}

func (b *R2Backend) PushReceipts(ctx context.Context, receipts []PushReceipt) error {
	client, err := b.cfg.R2.NewClient(ctx)
	if err != nil {
		return err
	}

	eg, egCtx := errgroup.WithContext(ctx)
	eg.SetLimit(10)

	var layers []map[string]interface{}

	for _, r := range receipts {
		r := r
		// Add layer reference to prevent orphaning
		layers = append(layers, map[string]interface{}{
			"mediaType": "application/vnd.aeroflare.nar.v1+zstd", // or generic
			"digest":    r.NarDigest,
			"size":      r.NarSize,
		})

		eg.Go(func() error {
			return b.cfg.R2.UploadNarinfo(egCtx, client, r.StorePath, r.NarinfoPath)
		})
	}

	if err := eg.Wait(); err != nil {
		return err
	}

	// Push chunked manifest to prevent orphaning
	if len(layers) > 0 {
		manifest := map[string]interface{}{
			"schemaVersion": 2,
			"mediaType":     "application/vnd.oci.image.manifest.v1+json",
			"config": map[string]interface{}{
				"mediaType": "application/vnd.oci.empty.v1+json",
				"digest":    "sha256:44136fa355b3678a1146ad16f7e8649e94fb4fc21fe77e8310c060f61caaff8a",
				"size":      2,
			},
			"layers": layers,
		}

		manifestBytes, _ := json.Marshal(manifest)
		chunkTag := fmt.Sprintf("chunk-%d", time.Now().Unix())
		protocol := GetProtocol(b.cfg.Registry)
		manifestURL := fmt.Sprintf("%s://%s/v2/%s/manifests/%s", protocol, b.cfg.Registry, b.cfg.Repository, chunkTag)

		req, _ := http.NewRequest("PUT", manifestURL, bytes.NewReader(manifestBytes))
		req.Header.Set("Authorization", "Bearer "+b.cfg.Token)
		req.Header.Set("Content-Type", "application/vnd.oci.image.manifest.v1+json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return fmt.Errorf("failed to push chunk manifest: %w", err)
		}
		defer resp.Body.Close()
		
		if resp.StatusCode >= 300 {
		    return fmt.Errorf("unexpected status pushing chunk manifest: %s", resp.Status)
		}
	}

	return nil
}
