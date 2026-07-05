package proxy

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	network "aeroflare/src"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
)

// CacheIndex handles loading, refreshing, and querying the cache index.
//
// The cache-index OCI manifest is the single source of truth. Its annotations
// determine the mode via the "index-type" label:
//   - "r2": narinfo is served from public-r2-url; only the manifest (not the
//     large index blob) is fetched.
//   - "native": narinfo hashes are used as tags in the registry.
//   - "json" (or absent): the index blob is downloaded and cached locally so
//     narinfo and NAR lookups can be served from memory.
type CacheIndex struct {
	mu                  sync.RWMutex
	refreshMu           sync.Mutex
	Data                *CacheIndexData
	NarLookups          map[string]string // narBasename -> narDigest (json mode only)
	ManifestAnnotations map[string]string
	LastFetch           time.Time
	IndexDir            string
	CacheFileName       string // local cache file name (json mode)
	IndexTTL            time.Duration
	TokenMgr            *TokenManager
	Registry            string
	Repository          string
}

// IndexType returns the cache's backend mode ("r2", "native", or "json"),
// read from the cache-index manifest annotations captured by the last
// refresh. Defaults to "json" when no annotations have been loaded yet or
// none of the recognized keys are present.
func (ci *CacheIndex) IndexType() string {
	ci.mu.RLock()
	defer ci.mu.RUnlock()
	if ci.ManifestAnnotations == nil {
		return "json"
	}
	// Current keys first, then the pre-migration "index-type" keys so caches
	// written by older versions keep working.
	for _, key := range []string{"aeroflare.backend", "backend", "aeroflare.index-type", "index-type"} {
		if t := ci.ManifestAnnotations[key]; t != "" {
			return t
		}
	}
	return "json"
}

// IsR2 reports whether the cache is configured to serve narinfo from R2.
func (ci *CacheIndex) IsR2() bool {
	return ci.IndexType() == "r2"
}

// PublicR2URL returns the configured public R2 URL for narinfo, or "" if none.
func (ci *CacheIndex) PublicR2URL() string {
	ci.mu.RLock()
	defer ci.mu.RUnlock()
	if ci.ManifestAnnotations == nil {
		return ""
	}
	if u := ci.ManifestAnnotations["public-r2-url"]; u != "" {
		return u
	}
	return ci.ManifestAnnotations["aeroflare.r2.public_url"]
}

// indexRefreshFailureCooldown limits how often a failed initial load is
// retried, so an unreachable registry doesn't serialize every request behind
// a synchronous refresh attempt.
var indexRefreshFailureCooldown = 10 * time.Second

// Get returns the current cache index data and NAR lookups map, triggering a TTL-based refresh if needed.
func (ci *CacheIndex) Get(ctx context.Context) (*CacheIndexData, map[string]string) {
	ci.mu.RLock()
	needRefresh := time.Since(ci.LastFetch) > ci.IndexTTL
	isNil := ci.Data == nil && ci.ManifestAnnotations == nil
	inFailureCooldown := !ci.LastFetch.IsZero() && time.Since(ci.LastFetch) <= indexRefreshFailureCooldown
	ci.mu.RUnlock()

	if isNil && !inFailureCooldown {
		ci.refreshMu.Lock()
		ci.mu.RLock()
		isNil = ci.Data == nil && ci.ManifestAnnotations == nil
		ci.mu.RUnlock()
		if isNil {
			if err := ci.refresh(ctx); err != nil {
				slog.Warn("Failed to refresh index. Using cached index.", "error", err)
				if ci.Data == nil {
					_ = ci.loadLocal()
				}
			}
			ci.mu.Lock()
			ci.LastFetch = time.Now()
			ci.mu.Unlock()
		}
		ci.refreshMu.Unlock()
	} else if needRefresh {
		if ci.refreshMu.TryLock() {
			// context.WithoutCancel (Go 1.21+) ensures the background refresh isn't aborted
			// if the HTTP request triggering it is completed/closed by the client.
			bgCtx := context.WithoutCancel(ctx)
			go func() {
				defer ci.refreshMu.Unlock()
				if err := ci.refresh(bgCtx); err != nil {
					slog.Warn("Failed to refresh index in background", "error", err)
				}
				ci.mu.Lock()
				ci.LastFetch = time.Now()
				ci.mu.Unlock()
			}()
		}
	}

	ci.mu.RLock()
	defer ci.mu.RUnlock()

	if ci.Data == nil {
		return &CacheIndexData{Entries: make(map[string]IndexEntry)}, make(map[string]string)
	}
	return ci.Data, ci.NarLookups
}

// ForceRefresh triggers an unconditional refresh of the cache index.
func (ci *CacheIndex) ForceRefresh(ctx context.Context) (int, error) {
	ci.refreshMu.Lock()
	defer ci.refreshMu.Unlock()

	err := ci.refresh(ctx)
	ci.mu.Lock()
	ci.LastFetch = time.Now()
	entries := 0
	if ci.Data != nil {
		entries = len(ci.Data.Entries)
	}
	ci.mu.Unlock()

	if err != nil {
		return 0, err
	}
	return entries, nil
}

// loadLocal reads the last-persisted cache-index JSON blob from disk (json
// mode) and populates the in-memory index and NAR lookup map from it. Used
// to seed the cache at startup and as a fallback when a refresh fails.
func (ci *CacheIndex) loadLocal() error {
	if ci.CacheFileName == "" {
		ci.CacheFileName = "cache-index.json"
	}
	indexFile := filepath.Join(ci.IndexDir, ci.CacheFileName)
	dataBytes, err := os.ReadFile(indexFile)
	if err != nil {
		return err
	}
	var indexData CacheIndexData
	if err := json.Unmarshal(dataBytes, &indexData); err != nil {
		return err
	}
	ci.updateInMemory(&indexData)
	return nil
}

// updateInMemory installs data as the current index and rebuilds NarLookups
// (narBasename -> narDigest) by extracting the "URL: " field's basename out
// of each entry's narinfo text (json mode only).
func (ci *CacheIndex) updateInMemory(data *CacheIndexData) {
	narLookups := make(map[string]string)
	if data.Entries != nil {
		for _, entry := range data.Entries {
			if entry.NarInfo == "" || entry.NarDigest == "" {
				continue
			}
			idx := strings.Index(entry.NarInfo, "URL: ")
			if idx != -1 {
				start := idx + 5
				end := strings.IndexByte(entry.NarInfo[start:], '\n')
				var urlVal string
				if end == -1 {
					urlVal = entry.NarInfo[start:]
				} else {
					urlVal = entry.NarInfo[start : start+end]
				}
				urlVal = strings.TrimSpace(urlVal)
				parts := strings.Split(urlVal, "/")
				if len(parts) > 0 {
					basename := parts[len(parts)-1]
					narLookups[basename] = entry.NarDigest
				}
			}
		}
	}
	ci.mu.Lock()
	ci.Data = data
	ci.NarLookups = narLookups
	ci.mu.Unlock()
}

// refresh re-fetches the cache-config (or, on failure, cache-index) manifest
// annotations to determine the current IndexType, then, only in json mode,
// downloads the full index blob and persists it locally via updateInMemory.
// In r2/native mode the in-memory index is cleared instead, since narinfo is
// served without it. Callers must hold refreshMu for the duration of a call.
func (ci *CacheIndex) refresh(ctx context.Context) error {
	token, err := ci.TokenMgr.GetToken(ctx)
	if err != nil {
		return fmt.Errorf("failed to get token: %w", err)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	proto := network.GetProtocol(ci.Registry)

	// 1. Fetch manifest — this is the single source of truth for annotations.
	anns, err := network.FetchAeroflareAnnotations(ctx, client, ci.Registry, ci.Repository, "cache-config", token)
	if err != nil {
		slog.Error("cache-config failed", "err", err)
		anns, err = network.FetchAeroflareAnnotations(ctx, client, ci.Registry, ci.Repository, "cache-index", token)
		if err != nil {
			slog.Error("cache-index failed", "err", err)
			return fmt.Errorf("failed to fetch config annotations: %w", err)
		}
	}

	ci.mu.Lock()
	ci.ManifestAnnotations = anns
	ci.mu.Unlock()

	indexType := ci.IndexType()

	// In r2 and native modes the index blob is not needed.
	if indexType == "r2" || indexType == "native" {
		ci.mu.Lock()
		// Clear any stale in-memory index from a previous json-mode run.
		ci.Data = &CacheIndexData{Entries: make(map[string]IndexEntry)}
		ci.NarLookups = make(map[string]string)
		ci.mu.Unlock()
		if indexType == "r2" {
			slog.Info("Manifest refreshed", "mode", indexType, "public_r2_url", ci.PublicR2URL())
		} else {
			slog.Info("Manifest refreshed", "mode", indexType)
		}
		return nil
	}

	// 2. Fetch the cache-index image manifest to locate the json index blob's digest.
	imageRef := fmt.Sprintf("%s/%s:cache-index", ci.Registry, ci.Repository)
	ref, err := name.ParseReference(imageRef)
	if err != nil {
		return fmt.Errorf("failed to parse reference: %w", err)
	}

	opts := []remote.Option{
		remote.WithContext(ctx),
	}
	if client != nil && client.Transport != nil {
		opts = append(opts, remote.WithTransport(client.Transport))
	} else if client != nil {
		opts = append(opts, remote.WithTransport(http.DefaultTransport))
	}
	if token != "" {
		opts = append(opts, remote.WithAuth(&authn.Bearer{Token: token}))
	} else {
		opts = append(opts, remote.WithAuthFromKeychain(authn.DefaultKeychain))
	}

	img, err := remote.Image(ref, opts...)
	if err != nil {
		return fmt.Errorf("failed to fetch image: %w", err)
	}

	manifest, err := img.Manifest()
	if err != nil {
		return fmt.Errorf("failed to read manifest: %w", err)
	}

	if len(manifest.Layers) == 0 {
		return fmt.Errorf("no layers in cache-index manifest")
	}

	digest := manifest.Layers[0].Digest.String()

	// 3. Fetch the blob itself (json mode) and cache it locally.
	blobURL := fmt.Sprintf("%s://%s/v2/%s/blobs/%s", proto, ci.Registry, ci.Repository, digest)
	blobClient := &http.Client{Timeout: 120 * time.Second}
	resp, err := network.DoWithRetry(blobClient, func() (*http.Request, error) {
		req, err := http.NewRequestWithContext(ctx, "GET", blobURL, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("User-Agent", userAgent)
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		return req, nil
	})
	if err != nil {
		return fmt.Errorf("failed fetching blob from %s: %w", blobURL, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("blob HTTP %d: %s", resp.StatusCode, string(bodyBytes))
	}

	dataBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed reading blob response body: %w", err)
	}

	var indexData CacheIndexData
	if err := json.Unmarshal(dataBytes, &indexData); err != nil {
		return fmt.Errorf("failed unmarshaling blob JSON: %w", err)
	}

	if err := os.MkdirAll(ci.IndexDir, 0755); err == nil {
		if ci.CacheFileName == "" {
			ci.CacheFileName = "cache-index.json"
		}
		indexFile := filepath.Join(ci.IndexDir, ci.CacheFileName)
		_ = os.WriteFile(indexFile, dataBytes, 0644)
	}

	ci.updateInMemory(&indexData)
	slog.Info("Index refreshed", "entries", len(indexData.Entries))
	return nil
}
