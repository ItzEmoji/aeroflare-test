package backend

import (
	"aeroflare/internal/cacheindex"
	"aeroflare/internal/r2"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

func TestR2Backend_PushReceipts(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "aeroflare-r2-backend-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	narinfoPath := filepath.Join(tmpDir, "test.narinfo")
	narinfoData := `StorePath: /nix/store/11111111111111111111111111111111-test
URL: test.nar
Compression: none
FileHash: sha256:0j56d7hwksf9zckz58ps8q969h7aak13d9s6n33sjslyfkkb3swx
FileSize: 13
NarHash: sha256:0j56d7hwksf9zckz58ps8q969h7aak13d9s6n33sjslyfkkb3swx
NarSize: 13
References:
Deriver: 22222222222222222222222222222222-test.drv
System: x86_64-linux
Sig: `
	err = os.WriteFile(narinfoPath, []byte(narinfoData), 0644)
	if err != nil {
		t.Fatalf("Failed to create narinfo file: %v", err)
	}

	receipts := []cacheindex.PushReceipt{
		{
			StorePath:   "/nix/store/11111111111111111111111111111111-test",
			NarinfoPath: narinfoPath,
			NarDigest:   "sha256:fake",
			NarSize:     13,
			IsRoot:      true,
		},
	}

	tests := []struct {
		name                 string
		receipts             []cacheindex.PushReceipt
		s3UploadStatus       int
		manifestStatus       int
		expectError          bool
		expectUploaded       bool
		expectPushedManifest bool
	}{
		{
			name:                 "Happy path",
			receipts:             receipts,
			s3UploadStatus:       http.StatusOK,
			manifestStatus:       http.StatusCreated,
			expectError:          false,
			expectUploaded:       true,
			expectPushedManifest: true,
		},
		{
			name:                 "Empty receipts",
			receipts:             []cacheindex.PushReceipt{},
			s3UploadStatus:       http.StatusOK,
			manifestStatus:       http.StatusCreated,
			expectError:          false,
			expectUploaded:       false,
			expectPushedManifest: false,
		},
		{
			name:                 "S3 upload fails",
			receipts:             receipts,
			s3UploadStatus:       http.StatusInternalServerError,
			manifestStatus:       http.StatusCreated,
			expectError:          true,
			expectUploaded:       true,
			expectPushedManifest: false, // Since S3 fails, it shouldn't proceed
		},
		{
			name:                 "Registry push fails",
			receipts:             receipts,
			s3UploadStatus:       http.StatusOK,
			manifestStatus:       http.StatusInternalServerError,
			expectError:          true,
			expectUploaded:       true,
			expectPushedManifest: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var mu sync.Mutex
			var uploadedNarinfo, pushedManifest bool
			var manifestBody []byte

			mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				mu.Lock()
				defer mu.Unlock()

				if r.Method == "PUT" && strings.HasPrefix(r.URL.Path, "/test-bucket/11111111111111111111111111111111.narinfo") {
					uploadedNarinfo = true
					w.WriteHeader(tt.s3UploadStatus)
					return
				}

				if r.Method == "PUT" && strings.HasPrefix(r.URL.Path, "/v2/test-repo/manifests/chunk-") {
					pushedManifest = true
					body, _ := io.ReadAll(r.Body)
					manifestBody = body
					w.WriteHeader(tt.manifestStatus)
					return
				}

				w.WriteHeader(http.StatusNotFound)
			}))
			defer mockServer.Close()

			u := strings.TrimPrefix(mockServer.URL, "http://")

			cfg := BackendConfig{
				Registry:   u,
				Repository: "test-repo",
				Token:      "mock-token",
				R2: &r2.R2Config{
					Bucket:    "test-bucket",
					Endpoint:  mockServer.URL,
					AccessKey: "fake-access",
					SecretKey: "fake-secret",
					PublicURL: "https://example.com",
				},
			}
			backend := &R2Backend{cfg: cfg}

			err := backend.PushReceipts(context.Background(), tt.receipts)
			if (err != nil) != tt.expectError {
				t.Errorf("PushReceipts expected error: %v, got: %v", tt.expectError, err)
			}

			mu.Lock()
			defer mu.Unlock()

			if uploadedNarinfo != tt.expectUploaded {
				t.Errorf("expected uploadedNarinfo=%v, got %v", tt.expectUploaded, uploadedNarinfo)
			}

			if pushedManifest != tt.expectPushedManifest {
				t.Errorf("expected pushedManifest=%v, got %v", tt.expectPushedManifest, pushedManifest)
			}

			// Verify the request body contains expected NarDigest for happy path
			if tt.expectPushedManifest && tt.manifestStatus == http.StatusCreated && len(tt.receipts) > 0 {
				var manifest map[string]interface{}
				if err := json.Unmarshal(manifestBody, &manifest); err != nil {
					t.Fatalf("failed to parse chunk manifest body: %v", err)
				}
				layers, ok := manifest["layers"].([]interface{})
				if !ok || len(layers) != len(tt.receipts) {
					t.Fatalf("expected %d layers, got %v", len(tt.receipts), len(layers))
				}
				layerMap := layers[0].(map[string]interface{})
				if layerMap["digest"] != "sha256:fake" {
					t.Errorf("expected digest sha256:fake, got %v", layerMap["digest"])
				}
			}
		})
	}
}
