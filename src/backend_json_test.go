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

func TestJSONBackend_PushReceipts(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "aeroflare-json-backend-test-*")
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
	var fetchedIndex, pushedIndex bool

	mockRegistry := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v2/" {
			w.WriteHeader(http.StatusOK)
			return
		}

		if r.Method == "GET" && strings.HasPrefix(r.URL.Path, "/v2/test-repo/manifests/cache-index") {
			mu.Lock()
			fetchedIndex = true
			mu.Unlock()
			w.WriteHeader(http.StatusNotFound)
			return
		}

		if r.Method == "POST" && strings.HasPrefix(r.URL.Path, "/v2/test-repo/blobs/uploads/") {
			w.Header().Set("Location", r.URL.Path+"1234")
			w.WriteHeader(http.StatusAccepted)
			return
		}
		if r.Method == "PATCH" && strings.HasPrefix(r.URL.Path, "/v2/test-repo/blobs/uploads/") {
			w.Header().Set("Location", r.URL.Path)
			w.WriteHeader(http.StatusAccepted)
			return
		}
		if r.Method == "PUT" && strings.HasPrefix(r.URL.Path, "/v2/test-repo/blobs/uploads/") {
			w.WriteHeader(http.StatusCreated)
			return
		}
		if r.Method == "HEAD" && strings.HasPrefix(r.URL.Path, "/v2/test-repo/blobs/") {
			w.WriteHeader(http.StatusNotFound) // force upload
			return
		}

		if r.Method == "PUT" && strings.HasPrefix(r.URL.Path, "/v2/test-repo/manifests/cache-index") {
			mu.Lock()
			pushedIndex = true
			mu.Unlock()
			w.WriteHeader(http.StatusCreated)
			return
		}

		w.WriteHeader(http.StatusNotFound)
	}))
	defer mockRegistry.Close()

	u := strings.TrimPrefix(mockRegistry.URL, "http://")

	cfg := BackendConfig{
		Registry:   u,
		Repository: "test-repo",
		Token:      "mock-token",
	}
	backend := &JSONBackend{cfg: cfg}

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

	if !fetchedIndex {
		t.Errorf("PushReceipts did not fetch existing index")
	}

	if !pushedIndex {
		t.Errorf("PushReceipts did not push updated index")
	}
}
