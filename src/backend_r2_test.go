package network

import (
	"context"
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

	var mu sync.Mutex
	var uploadedNarinfo, pushedManifest bool

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()

		// Mock R2 S3 Upload
		if r.Method == "PUT" && strings.HasPrefix(r.URL.Path, "/test-bucket/11111111111111111111111111111111.narinfo") {
			uploadedNarinfo = true
			w.WriteHeader(http.StatusOK)
			return
		}

		// Mock Registry Chunk Manifest Push
		if r.Method == "PUT" && strings.HasPrefix(r.URL.Path, "/v2/test-repo/manifests/chunk-") {
			pushedManifest = true
			w.WriteHeader(http.StatusCreated)
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
		R2: &R2Config{
			Bucket:    "test-bucket",
			Endpoint:  mockServer.URL,
			AccessKey: "fake-access",
			SecretKey: "fake-secret",
			PublicURL: "https://example.com",
		},
	}
	backend := &R2Backend{cfg: cfg}

	receipts := []PushReceipt{
		{
			StorePath:   "/nix/store/11111111111111111111111111111111-test",
			NarinfoPath: narinfoPath,
			NarDigest:   "sha256:fake",
			NarSize:     13,
			IsRoot:      true,
		},
	}

	err = backend.PushReceipts(context.Background(), receipts)
	if err != nil {
		t.Fatalf("PushReceipts failed: %v", err)
	}

	mu.Lock()
	defer mu.Unlock()

	if !uploadedNarinfo {
		t.Errorf("PushReceipts did not upload narinfo to R2")
	}

	if !pushedManifest {
		t.Errorf("PushReceipts did not push chunk manifest to registry")
	}
}
