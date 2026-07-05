package network

import (
	"context"
	"io"
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

	narinfoPath1 := filepath.Join(tmpDir, "test1.narinfo")
	narinfoData1 := `StorePath: /nix/store/11111111111111111111111111111111-test1
URL: test1.nar
Compression: none
FileHash: sha256:fake1
FileSize: 13
NarHash: sha256:fake1
NarSize: 13
References:
Deriver: 22222222222222222222222222222222-test.drv
System: x86_64-linux
Sig: `
	err = os.WriteFile(narinfoPath1, []byte(narinfoData1), 0644)
	if err != nil {
		t.Fatalf("Failed to create narinfo file: %v", err)
	}

	narinfoPath2 := filepath.Join(tmpDir, "test2.narinfo")
	narinfoData2 := `StorePath: /nix/store/33333333333333333333333333333333-test2
URL: test2.nar
Compression: none
FileHash: sha256:fake2
FileSize: 13
NarHash: sha256:fake2
NarSize: 13
References:
Deriver: 22222222222222222222222222222222-test.drv
System: x86_64-linux
Sig: `
	err = os.WriteFile(narinfoPath2, []byte(narinfoData2), 0644)
	if err != nil {
		t.Fatalf("Failed to create narinfo file: %v", err)
	}

	setupMock := func() (*httptest.Server, *bool, *bool) {
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
		return mockRegistry, &fetchedIndex, &pushedIndex
	}

	t.Run("SingleReceipt", func(t *testing.T) {
		mockRegistry, fetchedIndex, pushedIndex := setupMock()
		defer mockRegistry.Close()

		u := strings.TrimPrefix(mockRegistry.URL, "http://")
		cfg := BackendConfig{Registry: u, Repository: "test-repo", Token: "mock-token"}
		backend := &JSONBackend{cfg: cfg}

		receipts := []PushReceipt{
			{StorePath: "/nix/store/11111111111111111111111111111111-test1", NarinfoPath: narinfoPath1, NarDigest: "sha256:fake1", NarSize: 13, IsRoot: true},
		}

		err = backend.PushReceipts(context.Background(), receipts)
		if err != nil {
			t.Fatalf("PushReceipts failed: %v", err)
		}

		if !*fetchedIndex {
			t.Errorf("PushReceipts did not fetch existing index")
		}
		if !*pushedIndex {
			t.Errorf("PushReceipts did not push updated index")
		}
	})

	t.Run("EmptyReceipts", func(t *testing.T) {
		mockRegistry, fetchedIndex, pushedIndex := setupMock()
		defer mockRegistry.Close()

		u := strings.TrimPrefix(mockRegistry.URL, "http://")
		cfg := BackendConfig{Registry: u, Repository: "test-repo", Token: "mock-token"}
		backend := &JSONBackend{cfg: cfg}

		receipts := []PushReceipt{}

		err = backend.PushReceipts(context.Background(), receipts)
		if err != nil {
			t.Fatalf("PushReceipts failed: %v", err)
		}

		if !*fetchedIndex {
			t.Errorf("PushReceipts did not fetch existing index")
		}
		if !*pushedIndex {
			t.Errorf("PushReceipts did not push updated index")
		}
	})

	t.Run("MultipleReceipts", func(t *testing.T) {
		mockRegistry, fetchedIndex, pushedIndex := setupMock()
		defer mockRegistry.Close()

		u := strings.TrimPrefix(mockRegistry.URL, "http://")
		cfg := BackendConfig{Registry: u, Repository: "test-repo", Token: "mock-token"}
		backend := &JSONBackend{cfg: cfg}

		receipts := []PushReceipt{
			{StorePath: "/nix/store/11111111111111111111111111111111-test1", NarinfoPath: narinfoPath1, NarDigest: "sha256:fake1", NarSize: 13, IsRoot: true},
			{StorePath: "/nix/store/33333333333333333333333333333333-test2", NarinfoPath: narinfoPath2, NarDigest: "sha256:fake2", NarSize: 13, IsRoot: false},
		}

		err = backend.PushReceipts(context.Background(), receipts)
		if err != nil {
			t.Fatalf("PushReceipts failed: %v", err)
		}

		if !*fetchedIndex {
			t.Errorf("PushReceipts did not fetch existing index")
		}
		if !*pushedIndex {
			t.Errorf("PushReceipts did not push updated index")
		}
	})

	t.Run("ExistingIndex", func(t *testing.T) {
		var mu sync.Mutex
		var fetchedIndex, pushedIndex, fetchedBlob bool

		existingIndexBlob := `{"version":1,"entries":{"00000000000000000000000000000000":{"name":"test0","narinfo":"","nar_size":10,"added":""}}}`
		blobDigest := "sha256:0000000000000000000000000000000000000000000000000000000000000000"
		var storedManifest []byte

		mockRegistry := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/v2/" {
				w.WriteHeader(http.StatusOK)
				return
			}

			if r.Method == "GET" && strings.HasPrefix(r.URL.Path, "/v2/test-repo/manifests/cache-index") {
				mu.Lock()
				fetchedIndex = true
				stored := storedManifest
				mu.Unlock()
				w.Header().Set("Content-Type", "application/vnd.oci.image.manifest.v1+json")
				w.WriteHeader(http.StatusOK)
				if stored != nil {
					_, _ = w.Write(stored)
					return
				}
				_, _ = w.Write([]byte(`{"schemaVersion":2,"layers":[{"digest":"` + blobDigest + `"}]}`))
				return
			}

			if r.Method == "GET" && r.URL.Path == "/v2/test-repo/blobs/"+blobDigest {
				mu.Lock()
				fetchedBlob = true
				mu.Unlock()
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(existingIndexBlob))
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
				body, _ := io.ReadAll(r.Body)
				mu.Lock()
				pushedIndex = true
				storedManifest = body
				mu.Unlock()
				w.WriteHeader(http.StatusCreated)
				return
			}

			w.WriteHeader(http.StatusNotFound)
		}))
		defer mockRegistry.Close()

		u := strings.TrimPrefix(mockRegistry.URL, "http://")
		cfg := BackendConfig{Registry: u, Repository: "test-repo", Token: "mock-token"}
		backend := &JSONBackend{cfg: cfg}

		receipts := []PushReceipt{
			{StorePath: "/nix/store/11111111111111111111111111111111-test1", NarinfoPath: narinfoPath1, NarDigest: "sha256:fake1", NarSize: 13, IsRoot: true},
		}

		err = backend.PushReceipts(context.Background(), receipts)
		if err != nil {
			t.Fatalf("PushReceipts failed: %v", err)
		}

		if !fetchedIndex {
			t.Errorf("PushReceipts did not fetch existing index")
		}
		if !fetchedBlob {
			t.Errorf("PushReceipts did not fetch existing blob")
		}
		if !pushedIndex {
			t.Errorf("PushReceipts did not push updated index")
		}
	})
}
