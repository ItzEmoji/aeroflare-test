package backend

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"aeroflare/internal/cacheindex"
)

// TestJSONBackend_ConcurrentWriterRetries simulates a lost-update race: right
// after our first index PUT, another writer overwrites the cache-index tag.
// The backend must detect its write was clobbered, re-merge, and PUT again so
// the final index still contains our entries.
func TestJSONBackend_ConcurrentWriterRetries(t *testing.T) {
	tmpDir := t.TempDir()
	narinfoPath := filepath.Join(tmpDir, "test1.narinfo")
	if err := os.WriteFile(narinfoPath, []byte("StorePath: /nix/store/11111111111111111111111111111111-test1\nURL: test1.nar\n"), 0644); err != nil {
		t.Fatal(err)
	}

	interloperBlobDigest := "sha256:1111111111111111111111111111111111111111111111111111111111111111"
	interloperManifest := []byte(`{"schemaVersion":2,"layers":[{"mediaType":"application/vnd.nix.cache.index.v1+json","digest":"` + interloperBlobDigest + `","size":10}]}`)
	interloperIndex := `{"version":1,"entries":{"99999999999999999999999999999999":{"name":"other","narinfo":"x","nar_digest":"sha256:other","nar_size":1,"added":""}}}`

	var mu sync.Mutex
	var storedManifest []byte
	blobs := map[string][]byte{}
	uploads := map[string][]byte{}
	indexPutCount := 0
	uploadCount := 0
	sabotaged := false

	mockRegistry := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		switch {
		case r.URL.Path == "/v2/":
			w.WriteHeader(http.StatusOK)
		case r.Method == "GET" && strings.HasSuffix(r.URL.Path, "/manifests/cache-index"):
			if storedManifest == nil {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(storedManifest)
		case r.Method == "PUT" && strings.HasSuffix(r.URL.Path, "/manifests/cache-index"):
			body, _ := io.ReadAll(r.Body)
			storedManifest = body
			indexPutCount++
			if !sabotaged {
				// Another writer wins immediately after our first write.
				storedManifest = interloperManifest
				sabotaged = true
			}
			w.WriteHeader(http.StatusCreated)
		case r.Method == "GET" && strings.Contains(r.URL.Path, "/blobs/"+interloperBlobDigest):
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(interloperIndex))
		case r.Method == "GET" && strings.Contains(r.URL.Path, "/blobs/"):
			parts := strings.Split(r.URL.Path, "/")
			digest := parts[len(parts)-1]
			if b, ok := blobs[digest]; ok {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write(b)
				return
			}
			w.WriteHeader(http.StatusNotFound)
		case r.Method == "POST" && strings.Contains(r.URL.Path, "/blobs/uploads/"):
			uploadCount++
			w.Header().Set("Location", fmt.Sprintf("%ssession-%d", r.URL.Path, uploadCount))
			w.WriteHeader(http.StatusAccepted)
		case r.Method == "PATCH" && strings.Contains(r.URL.Path, "/blobs/uploads/"):
			body, _ := io.ReadAll(r.Body)
			uploads[r.URL.Path] = append(uploads[r.URL.Path], body...)
			w.Header().Set("Location", r.URL.Path)
			w.WriteHeader(http.StatusAccepted)
		case r.Method == "PUT" && strings.Contains(r.URL.Path, "/blobs/uploads/"):
			digest := r.URL.Query().Get("digest")
			body, _ := io.ReadAll(r.Body)
			if len(body) == 0 {
				body = uploads[r.URL.Path]
			}
			blobs[digest] = body
			w.WriteHeader(http.StatusCreated)
		case r.Method == "HEAD" && strings.Contains(r.URL.Path, "/blobs/"):
			w.WriteHeader(http.StatusNotFound)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer mockRegistry.Close()

	u := strings.TrimPrefix(mockRegistry.URL, "http://")
	backend := &JSONBackend{cfg: BackendConfig{Registry: u, Repository: "test-repo", Token: "mock-token"}}
	receipts := []cacheindex.PushReceipt{
		{StorePath: "/nix/store/11111111111111111111111111111111-test1", NarinfoPath: narinfoPath, NarDigest: "sha256:fake1", NarSize: 13, IsRoot: true},
	}

	if err := backend.PushReceipts(context.Background(), receipts); err != nil {
		t.Fatalf("PushReceipts failed: %v", err)
	}

	mu.Lock()
	defer mu.Unlock()
	if indexPutCount < 2 {
		t.Fatalf("expected a retry after concurrent overwrite (>=2 index PUTs), got %d", indexPutCount)
	}

	// The final stored index must contain BOTH our entry and the interloper's.
	var finalManifest struct {
		Layers []struct {
			Digest string `json:"digest"`
		} `json:"layers"`
	}
	if err := json.Unmarshal(storedManifest, &finalManifest); err != nil || len(finalManifest.Layers) == 0 {
		t.Fatalf("bad final manifest: %v", err)
	}
	finalIndexBytes, ok := blobs[finalManifest.Layers[0].Digest]
	if !ok {
		t.Fatalf("final index blob %s not uploaded", finalManifest.Layers[0].Digest)
	}
	var finalIndex cacheindex.PushCacheIndex
	if err := json.Unmarshal(finalIndexBytes, &finalIndex); err != nil {
		t.Fatal(err)
	}
	if _, ok := finalIndex.Entries["11111111111111111111111111111111"]; !ok {
		t.Error("final index lost our entry")
	}
	if _, ok := finalIndex.Entries["99999999999999999999999999999999"]; !ok {
		t.Error("final index lost the concurrent writer's entry")
	}
}
