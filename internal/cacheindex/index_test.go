package cacheindex

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestFetchCacheIndex_404ReturnsEmptyIndex(t *testing.T) {
	mockRegistry := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v2/" {
			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer mockRegistry.Close()

	u := strings.TrimPrefix(mockRegistry.URL, "http://")
	index, err := FetchCacheIndex(u, "test-repo", "mock-token")
	if err != nil {
		t.Fatalf("expected no error for 404 (empty cache), got: %v", err)
	}
	if index == nil || index.Entries == nil {
		t.Fatal("expected empty initialized index for 404")
	}
	if len(index.Entries) != 0 {
		t.Fatalf("expected 0 entries, got %d", len(index.Entries))
	}
}

func TestFetchCacheIndex_ServerErrorReturnsError(t *testing.T) {
	mockRegistry := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer mockRegistry.Close()

	u := strings.TrimPrefix(mockRegistry.URL, "http://")
	_, err := FetchCacheIndex(u, "test-repo", "mock-token")
	if err == nil {
		t.Fatal("expected error for HTTP 500 manifest fetch, got nil (index would be silently wiped)")
	}
}

func TestFetchCacheIndex_ConnectionErrorReturnsError(t *testing.T) {
	mockRegistry := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	u := strings.TrimPrefix(mockRegistry.URL, "http://")
	mockRegistry.Close() // connection refused from here on

	_, err := FetchCacheIndex(u, "test-repo", "mock-token")
	if err == nil {
		t.Fatal("expected error for unreachable registry, got nil")
	}
}

func TestFetchCacheIndex_BlobErrorReturnsError(t *testing.T) {
	blobDigest := "sha256:0000000000000000000000000000000000000000000000000000000000000000"
	mockRegistry := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v2/" {
			w.WriteHeader(http.StatusOK)
			return
		}
		if r.Method == "GET" && strings.HasPrefix(r.URL.Path, "/v2/test-repo/manifests/cache-index") {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"schemaVersion":2,"layers":[{"digest":"` + blobDigest + `"}]}`))
			return
		}
		// Blob fetch fails.
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer mockRegistry.Close()

	u := strings.TrimPrefix(mockRegistry.URL, "http://")
	_, err := FetchCacheIndex(u, "test-repo", "mock-token")
	if err == nil {
		t.Fatal("expected error when index blob cannot be fetched, got nil")
	}
}

func TestFetchCacheIndex_InvalidJSONReturnsError(t *testing.T) {
	blobDigest := "sha256:0000000000000000000000000000000000000000000000000000000000000000"
	mockRegistry := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v2/" {
			w.WriteHeader(http.StatusOK)
			return
		}
		if r.Method == "GET" && strings.HasPrefix(r.URL.Path, "/v2/test-repo/manifests/cache-index") {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"schemaVersion":2,"layers":[{"digest":"` + blobDigest + `"}]}`))
			return
		}
		if r.Method == "GET" && r.URL.Path == "/v2/test-repo/blobs/"+blobDigest {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{not json`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer mockRegistry.Close()

	u := strings.TrimPrefix(mockRegistry.URL, "http://")
	_, err := FetchCacheIndex(u, "test-repo", "mock-token")
	if err == nil {
		t.Fatal("expected error for corrupt index JSON, got nil")
	}
}

// newUploadOKRegistry returns a mock registry that accepts blob uploads and
// manifest PUTs, reporting 404 for the existing index.
func newUploadOKRegistry(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/v2/":
			w.WriteHeader(http.StatusOK)
		case r.Method == "POST" && strings.Contains(r.URL.Path, "/blobs/uploads/"):
			w.Header().Set("Location", r.URL.Path+"1234")
			w.WriteHeader(http.StatusAccepted)
		case r.Method == "PATCH" && strings.Contains(r.URL.Path, "/blobs/uploads/"):
			w.Header().Set("Location", r.URL.Path)
			w.WriteHeader(http.StatusAccepted)
		case r.Method == "PUT" && strings.Contains(r.URL.Path, "/blobs/uploads/"):
			w.WriteHeader(http.StatusCreated)
		case r.Method == "HEAD" && strings.Contains(r.URL.Path, "/blobs/"):
			w.WriteHeader(http.StatusNotFound) // force upload
		case r.Method == "PUT" && strings.Contains(r.URL.Path, "/manifests/"):
			w.WriteHeader(http.StatusCreated)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}

func TestUpdateCacheIndex_MissingNarinfoReturnsError(t *testing.T) {
	mockRegistry := newUploadOKRegistry(t)
	defer mockRegistry.Close()

	u := strings.TrimPrefix(mockRegistry.URL, "http://")
	existing := &PushCacheIndex{Entries: make(map[string]PushCacheEntry)}
	receipts := []PushReceipt{
		{
			StorePath:   "/nix/store/11111111111111111111111111111111-test1",
			NarinfoPath: "/nonexistent/path/test1.narinfo",
			NarDigest:   "sha256:fake1",
			NarSize:     13,
			IsRoot:      true,
		},
	}

	err := UpdateCacheIndex(receipts, existing, u, "test-repo", "mock-token", "", nil)
	if err == nil {
		t.Fatal("expected error when a receipt's narinfo cannot be read, got nil (entry silently dropped)")
	}
}
