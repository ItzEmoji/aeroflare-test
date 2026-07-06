package backend

import (
	"aeroflare/internal/cacheindex"
	"aeroflare/internal/oci"
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
	"time"
)

func TestR2Config_NewClient_MissingCredsFails(t *testing.T) {
	r2 := &r2.R2Config{Bucket: "b", Endpoint: "https://example.com"}
	if _, err := r2.NewClient(context.Background()); err == nil {
		t.Fatal("expected error when R2 credentials are missing; a client with empty creds fails much later with confusing errors")
	}
}

func TestR2Backend_UsesReceiptCompressionMediaType(t *testing.T) {
	restore := oci.RetryBaseDelay
	oci.RetryBaseDelay = time.Millisecond
	defer func() { oci.RetryBaseDelay = restore }()

	tmpDir := t.TempDir()
	narinfoPath := filepath.Join(tmpDir, "test.narinfo")
	if err := os.WriteFile(narinfoPath, []byte("StorePath: /nix/store/11111111111111111111111111111111-test\nURL: test.nar.xz\n"), 0644); err != nil {
		t.Fatal(err)
	}

	var mu sync.Mutex
	var manifestBody []byte
	transientFailures := 1

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		switch {
		case r.Method == "PUT" && strings.Contains(r.URL.Path, ".narinfo"):
			w.WriteHeader(http.StatusOK)
		case r.Method == "PUT" && strings.Contains(r.URL.Path, "/manifests/chunk-"):
			if transientFailures > 0 {
				transientFailures--
				w.WriteHeader(http.StatusServiceUnavailable)
				return
			}
			manifestBody, _ = io.ReadAll(r.Body)
			w.WriteHeader(http.StatusCreated)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer mockServer.Close()

	u := strings.TrimPrefix(mockServer.URL, "http://")
	backend := &R2Backend{cfg: BackendConfig{
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
	}}

	receipts := []cacheindex.PushReceipt{
		{
			StorePath:   "/nix/store/11111111111111111111111111111111-test",
			NarinfoPath: narinfoPath,
			NarDigest:   "sha256:fake",
			NarSize:     13,
			Compression: "xz",
			IsRoot:      true,
		},
	}

	if err := backend.PushReceipts(context.Background(), receipts); err != nil {
		t.Fatalf("PushReceipts failed (transient 503 should have been retried): %v", err)
	}

	mu.Lock()
	defer mu.Unlock()
	var manifest struct {
		Layers []struct {
			MediaType string `json:"mediaType"`
		} `json:"layers"`
	}
	if err := json.Unmarshal(manifestBody, &manifest); err != nil || len(manifest.Layers) == 0 {
		t.Fatalf("bad chunk manifest: %v", err)
	}
	if got := manifest.Layers[0].MediaType; got != "application/vnd.aeroflare.nar.v1+xz" {
		t.Errorf("layer mediaType = %q, want compression from receipt (+xz)", got)
	}
}
