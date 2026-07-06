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

// R2Backend uploads .narinfo files directly to R2 (nar blobs themselves are
// assumed already in the R2 bucket) and records each nar as a layer in a
// chunked series of OCI manifests, so GC tooling that walks manifest layers
// can still see the blobs and won't treat them as orphaned.
type R2Backend struct {
	cfg BackendConfig
}

func (b *R2Backend) PushReceipts(ctx context.Context, receipts []PushReceipt) error {
	s3Client, err := b.cfg.R2.NewClient(ctx)
	if err != nil {
		return err
	}

	eg, egCtx := errgroup.WithContext(ctx)
	eg.SetLimit(workerLimit(b.cfg.Workers, 10))

	var layers []map[string]interface{}

	for _, r := range receipts {
		r := r
		mediaType := "application/vnd.aeroflare.nar.v1+zstd"
		if r.Compression != "" {
			mediaType = "application/vnd.aeroflare.nar.v1+" + r.Compression
		}
		// Add layer reference to prevent orphaning
		layers = append(layers, map[string]interface{}{
			"mediaType": mediaType,
			"digest":    r.NarDigest,
			"size":      r.NarSize,
		})

		eg.Go(func() error {
			return b.cfg.R2.UploadNarinfo(egCtx, s3Client, r.StorePath, r.NarinfoPath)
		})
	}

	if err := eg.Wait(); err != nil {
		return err
	}

	// Push chunked manifests to prevent orphaning
	chunkSize := 5000
	for i := 0; i < len(layers); i += chunkSize {
		end := i + chunkSize
		if end > len(layers) {
			end = len(layers)
		}
		chunk := layers[i:end]

		manifest := map[string]interface{}{
			"schemaVersion": 2,
			"mediaType":     "application/vnd.oci.image.manifest.v1+json",
			"config": map[string]interface{}{
				"mediaType": "application/vnd.oci.empty.v1+json",
				"digest":    "sha256:44136fa355b3678a1146ad16f7e8649e94fb4fc21fe77e8310c060f61caaff8a",
				"size":      2,
			},
			"layers": chunk,
		}

		manifestBytes, err := json.Marshal(manifest)
		if err != nil {
			return fmt.Errorf("failed to marshal chunk manifest: %w", err)
		}

		chunkTag := fmt.Sprintf("chunk-%d-%d", time.Now().UnixNano(), i/chunkSize)
		protocol := GetProtocol(b.cfg.Registry)
		manifestURL := fmt.Sprintf("%s://%s/v2/%s/manifests/%s", protocol, b.cfg.Registry, b.cfg.Repository, chunkTag)

		httpClient := &http.Client{Timeout: 30 * time.Second}
		resp, err := DoWithRetry(httpClient, func() (*http.Request, error) {
			req, err := http.NewRequestWithContext(ctx, "PUT", manifestURL, bytes.NewReader(manifestBytes))
			if err != nil {
				return nil, err
			}
			req.Header.Set("Authorization", "Bearer "+b.cfg.Token)
			req.Header.Set("Content-Type", "application/vnd.oci.image.manifest.v1+json")
			return req, nil
		})
		if err != nil {
			return fmt.Errorf("failed to push chunk manifest: %w", err)
		}

		if resp.StatusCode >= 300 {
			resp.Body.Close()
			return fmt.Errorf("unexpected status pushing chunk manifest: %s", resp.Status)
		}
		resp.Body.Close()
	}

	return nil
}
