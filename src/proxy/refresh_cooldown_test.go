package proxy

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// TestCacheIndex_FailedInitialRefreshCooldown verifies that when the registry
// is unreachable and there is no local cache file, the proxy does not retry a
// synchronous refresh on every incoming request — that would serialize all
// requests behind registry timeouts.
func TestCacheIndex_FailedInitialRefreshCooldown(t *testing.T) {
	var manifestHits atomic.Int32

	mockRegistry := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v2/" {
			w.WriteHeader(http.StatusOK)
			return
		}
		if strings.HasPrefix(r.URL.Path, "/token") {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"token":"mock-token"}`))
			return
		}
		if strings.Contains(r.URL.Path, "/manifests/") {
			manifestHits.Add(1)
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer mockRegistry.Close()

	registryHost := strings.TrimPrefix(mockRegistry.URL, "http://")
	ci := &CacheIndex{
		IndexDir:   t.TempDir(),
		IndexTTL:   5 * time.Minute,
		TokenMgr:   NewTokenManager(registryHost, "test-repo", ""),
		Registry:   registryHost,
		Repository: "test-repo",
	}

	ci.Get(context.Background())
	firstHits := manifestHits.Load()
	if firstHits == 0 {
		t.Fatal("expected an initial refresh attempt")
	}

	// Immediately after a failed refresh, further requests must not trigger
	// another synchronous refresh.
	ci.Get(context.Background())
	ci.Get(context.Background())
	if manifestHits.Load() != firstHits {
		t.Fatalf("requests during the failure cooldown re-attempted refresh: %d hits, want %d", manifestHits.Load(), firstHits)
	}

	// Once the cooldown has passed, a new attempt is allowed.
	ci.mu.Lock()
	ci.LastFetch = time.Now().Add(-2 * indexRefreshFailureCooldown)
	ci.mu.Unlock()
	ci.Get(context.Background())
	if manifestHits.Load() == firstHits {
		t.Fatal("refresh was never re-attempted after the cooldown elapsed")
	}
}
