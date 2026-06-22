package network

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	narhash "aeroflare/src/prepare/hash"
)

// GetProtocol determines http vs https protocol (useful for local mock registry tests)
func GetProtocol(registry string) string {
	if strings.HasPrefix(registry, "127.0.0.1") || strings.HasPrefix(registry, "localhost") {
		return "http"
	}
	return "https"
}

// IndexManifest represents the OCI image manifest for the cache index or configuration.
type IndexManifest struct {
	Layers      []IndexLayer      `json:"layers"`
	Annotations map[string]string `json:"annotations,omitempty"`
}

// IndexLayer represents a layer in the OCI manifest.
type IndexLayer struct {
	Digest string `json:"digest"`
	Size   int64  `json:"size"`
}

// IndexEntry represents a single cached store path entry.
type IndexEntry struct {
	NarInfo   string `json:"narinfo"`
	NarDigest string `json:"nar_digest"`
}

// CacheIndexData represents the parsed cache-index JSON payload.
type CacheIndexData struct {
	Entries   map[string]IndexEntry `json:"entries"`
	PublicKey string                `json:"public_key"`
	Generated string                `json:"generated"`
}

// RemoteConfig represents the optional dynamic configuration JSON loaded from GHCR.
// Used by the push/configure pipeline, not by the proxy.
type RemoteConfig struct {
	WorkerURL      string   `json:"worker_url"`
	PublicKey      string   `json:"public_key"`
	UpstreamCaches []string `json:"upstream_caches"`
}

// TokenManager handles retrieving and caching the OCI Bearer token.
type TokenManager struct {
	registry    string
	repository  string
	githubToken string
	mu          sync.Mutex
	token       string
	expiry      time.Time
}

// NewTokenManager creates a new OCI token manager.
func NewTokenManager(registry, repository, githubToken string) *TokenManager {
	return &TokenManager{
		registry:    registry,
		repository:  repository,
		githubToken: githubToken,
	}
}

// GetToken returns a valid OCI Bearer token, performing token exchange if necessary.
func (tm *TokenManager) GetToken() (string, error) {
	if t := os.Getenv("oci_token"); t != "" && !strings.HasPrefix(t, "ghp_") && !strings.HasPrefix(t, "github_pat_") {
		return t, nil
	}
	if t := os.Getenv("NIXCACHE_TOKEN"); t != "" && !strings.HasPrefix(t, "ghp_") && !strings.HasPrefix(t, "github_pat_") {
		return t, nil
	}

	tm.mu.Lock()
	defer tm.mu.Unlock()

	if tm.token != "" && time.Now().Before(tm.expiry) {
		return tm.token, nil
	}

	token, err := tm.fetchToken()
	if err != nil {
		return "", err
	}

	tm.token = token
	tm.expiry = time.Now().Add(4 * time.Minute) // Cache for 4 minutes
	return tm.token, nil
}

func (tm *TokenManager) fetchToken() (string, error) {
	scope := fmt.Sprintf("repository:%s:pull", tm.repository)
	proto := GetProtocol(tm.registry)
	tokenURL := fmt.Sprintf("%s://%s/token?scope=%s&service=%s", proto, tm.registry, scope, tm.registry)

	client := &http.Client{Timeout: 10 * time.Second}

	if tm.githubToken != "" {
		req, err := http.NewRequest("GET", tokenURL, nil)
		if err == nil {
			req.Header.Set("User-Agent", "aeroflare/1.0")
			req.SetBasicAuth("token", tm.githubToken)

			resp, err := client.Do(req)
			if err == nil {
				defer func() { _ = resp.Body.Close() }()
				if resp.StatusCode == http.StatusOK {
					var result struct {
						Token string `json:"token"`
					}
					if err := json.NewDecoder(resp.Body).Decode(&result); err == nil && result.Token != "" {
						return result.Token, nil
					}
				}
			}
		}
	}

	req, err := http.NewRequest("GET", tokenURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "aeroflare/1.0")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to fetch token (HTTP %d): %s", resp.StatusCode, string(bodyBytes))
	}

	var result struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	if result.Token == "" {
		return "", fmt.Errorf("empty token returned from registry")
	}

	return result.Token, nil
}

// CacheIndex handles loading, refreshing, and querying the cache index.
//
// The cache-index OCI manifest is the single source of truth. Its annotations
// determine the mode via the "index-type" label:
//   - "r2": narinfo is served from public-r2-url; only the manifest (not the
//     large index blob) is fetched.
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

// IndexType returns the current manifest's index-type annotation, defaulting to "json".
func (ci *CacheIndex) IndexType() string {
	ci.mu.RLock()
	defer ci.mu.RUnlock()
	if ci.ManifestAnnotations == nil {
		return "json"
	}
	if t := ci.ManifestAnnotations["index-type"]; t != "" {
		return t
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

// Get returns the current cache index data and NAR lookups map, triggering a TTL-based refresh if needed.
func (ci *CacheIndex) Get() (*CacheIndexData, map[string]string) {
	ci.mu.RLock()
	needRefresh := time.Since(ci.LastFetch) > ci.IndexTTL
	isNil := ci.Data == nil && ci.ManifestAnnotations == nil
	ci.mu.RUnlock()

	if isNil {
		ci.refreshMu.Lock()
		ci.mu.RLock()
		isNil = ci.Data == nil && ci.ManifestAnnotations == nil
		ci.mu.RUnlock()
		if isNil {
			if err := ci.refresh(); err != nil {
				fmt.Fprintf(os.Stderr, "[aeroflare proxy] Warning: failed to refresh index: %v. Using cached index.\n", err)
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
			go func() {
				defer ci.refreshMu.Unlock()
				if err := ci.refresh(); err != nil {
					fmt.Fprintf(os.Stderr, "[aeroflare proxy] Warning: failed to refresh index: %v.\n", err)
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
func (ci *CacheIndex) ForceRefresh() (int, error) {
	ci.refreshMu.Lock()
	defer ci.refreshMu.Unlock()

	err := ci.refresh()
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

func (ci *CacheIndex) refresh() error {
	token, err := ci.TokenMgr.GetToken()
	if err != nil {
		return fmt.Errorf("failed to get token: %w", err)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	proto := GetProtocol(ci.Registry)

	// 1. Fetch manifest — this is the single source of truth for annotations.
	manifestURL := fmt.Sprintf("%s://%s/v2/%s/manifests/cache-index", proto, ci.Registry, ci.Repository)
	req, err := http.NewRequest("GET", manifestURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/vnd.oci.image.manifest.v1+json")
	req.Header.Set("User-Agent", "aeroflare/1.0")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("manifest HTTP %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var manifest IndexManifest
	if err := json.NewDecoder(resp.Body).Decode(&manifest); err != nil {
		return err
	}

	ci.mu.Lock()
	ci.ManifestAnnotations = manifest.Annotations
	ci.mu.Unlock()

	// In r2 mode the index blob is not needed: narinfo is served from R2 and
	// NAR digests are parsed from the R2 narinfo on demand.
	if manifest.Annotations != nil && manifest.Annotations["index-type"] == "r2" {
		ci.mu.Lock()
		// Clear any stale in-memory index from a previous json-mode run.
		ci.Data = &CacheIndexData{Entries: make(map[string]IndexEntry)}
		ci.NarLookups = make(map[string]string)
		ci.mu.Unlock()
		fmt.Fprintf(os.Stderr, "[aeroflare proxy] Manifest refreshed (r2 mode): public-r2-url=%s\n", ci.PublicR2URL())
		return nil
	}

	if len(manifest.Layers) == 0 {
		return fmt.Errorf("no layers in cache-index manifest")
	}

	digest := manifest.Layers[0].Digest

	// 2. Fetch blob (json mode)
	blobURL := fmt.Sprintf("%s://%s/v2/%s/blobs/%s", proto, ci.Registry, ci.Repository, digest)
	req, err = http.NewRequest("GET", blobURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "aeroflare/1.0")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	blobClient := &http.Client{Timeout: 120 * time.Second}
	resp, err = blobClient.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("blob HTTP %d: %s", resp.StatusCode, string(bodyBytes))
	}

	dataBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var indexData CacheIndexData
	if err := json.Unmarshal(dataBytes, &indexData); err != nil {
		return err
	}

	if err := os.MkdirAll(ci.IndexDir, 0755); err == nil {
		if ci.CacheFileName == "" {
			ci.CacheFileName = "cache-index.json"
		}
		indexFile := filepath.Join(ci.IndexDir, ci.CacheFileName)
		_ = os.WriteFile(indexFile, dataBytes, 0644)
	}

	ci.updateInMemory(&indexData)
	fmt.Fprintf(os.Stderr, "[aeroflare proxy] Index refreshed: %d entries\n", len(indexData.Entries))
	return nil
}

// ProxyServer bridges the Nix binary cache protocol to GHCR and upstream caches.
type ProxyServer struct {
	Port            int
	ListenAddr      string
	Registry        string
	Repository      string
	UpstreamCaches  []string
	TokenMgr        *TokenManager
	CacheIndex      *CacheIndex
	HttpClient      *http.Client
	HttpShortClient *http.Client
}

// Handler handles all incoming HTTP requests for the Nix binary cache proxy.
func (ps *ProxyServer) Handler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimSuffix(r.URL.Path, "/")

	switch r.Method {
	case http.MethodGet, http.MethodHead:
		switch {
		case path == "/nix-cache-info":
			ps.serveNixCacheInfo(w)
		case path == "/public-key":
			ps.servePublicKey(w)
		case path == "/api/public-key":
			ps.serveApiPublicKey(w)
		case path == "/_status":
			ps.serveStatus(w)
		case strings.HasSuffix(path, ".narinfo"):
			ps.serveNarInfo(w, r, path)
		case strings.HasPrefix(path, "/nar/"):
			ps.serveNar(w, r, path, r.Method)
		default:
			http.Error(w, "Not Found", http.StatusNotFound)
		}
	case http.MethodPost:
		switch path {
		case "/_refresh":
			ps.handleRefresh(w)
		default:
			http.Error(w, "Not Found", http.StatusNotFound)
		}
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

func (ps *ProxyServer) serveNixCacheInfo(w http.ResponseWriter) {
	data := []byte("StoreDir: /nix/store\nWantMassQuery: 1\nPriority: 40\n")
	w.Header().Set("Content-Type", "text/x-nix-cache-info")
	w.Header().Set("Content-Length", strconv.Itoa(len(data)))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

func (ps *ProxyServer) servePublicKey(w http.ResponseWriter) {
	// Primary source: the cache-index manifest annotation.
	ps.CacheIndex.mu.RLock()
	pubKey := ""
	if ps.CacheIndex.ManifestAnnotations != nil {
		pubKey = ps.CacheIndex.ManifestAnnotations["public-key"]
	}
	ps.CacheIndex.mu.RUnlock()

	// Fallback (json mode): the public_key field of the cached index blob.
	if pubKey == "" && !ps.CacheIndex.IsR2() {
		indexData, _ := ps.CacheIndex.Get()
		pubKey = indexData.PublicKey
	}

	if pubKey != "" {
		pubKey = strings.TrimSpace(pubKey) + "\n"
		data := []byte(pubKey)
		w.Header().Set("Content-Type", "text/plain")
		w.Header().Set("Content-Length", strconv.Itoa(len(data)))
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(data)
	} else {
		http.Error(w, "No public key configured", http.StatusNotFound)
	}
}

func (ps *ProxyServer) serveApiPublicKey(w http.ResponseWriter) {
	ps.CacheIndex.mu.RLock()
	annotations := ps.CacheIndex.ManifestAnnotations
	ps.CacheIndex.mu.RUnlock()

	pubKey := ""
	if annotations != nil {
		pubKey = annotations["public-key"]
	}
	if pubKey != "" {
		pubKey = strings.TrimSpace(pubKey) + "\n"
		data := []byte(pubKey)
		w.Header().Set("Content-Type", "text/plain")
		w.Header().Set("Content-Length", strconv.Itoa(len(data)))
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(data)
	} else {
		http.Error(w, "No public key configured in manifest", http.StatusNotFound)
	}
}

func (ps *ProxyServer) serveStatus(w http.ResponseWriter) {
	indexData, _ := ps.CacheIndex.Get()

	status := map[string]interface{}{
		"index_entries":   len(indexData.Entries),
		"index_generated": indexData.Generated,
		"index_ttl":       int(ps.CacheIndex.IndexTTL.Seconds()),
		"repo":            ps.Repository,
		"upstream":        ps.UpstreamCaches,
	}

	ps.CacheIndex.mu.RLock()
	annotations := ps.CacheIndex.ManifestAnnotations
	ps.CacheIndex.mu.RUnlock()

	if annotations["index-type"] == "r2" || annotations["public-r2-url"] != "" || annotations["aeroflare.r2.public_url"] != "" {
		publicURL := annotations["public-r2-url"]
		if publicURL == "" {
			publicURL = annotations["aeroflare.r2.public_url"]
		}
		status["r2_enabled"] = true
		status["r2_public_url"] = publicURL
	} else {
		status["r2_enabled"] = false
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(status)
}

func (ps *ProxyServer) handleRefresh(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")

	count, err := ps.CacheIndex.ForceRefresh()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"refreshed": false,
			"error":     err.Error(),
		})
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"refreshed": true,
		"entries":   count,
	})
}

func (ps *ProxyServer) serveNarInfo(w http.ResponseWriter, r *http.Request, path string) {
	storeHash := strings.TrimPrefix(path, "/")
	storeHash = strings.TrimSuffix(storeHash, ".narinfo")

	// r2 mode: redirect the client to the public R2 URL.
	if ps.CacheIndex.IsR2() {
		publicURL := ps.CacheIndex.PublicR2URL()
		if publicURL == "" {
			http.Error(w, "Error: The cache uses R2 but no public-url is configured. It is not possible to access this cache.", http.StatusForbidden)
			return
		}
		redirectURL := fmt.Sprintf("%s/%s.narinfo", strings.TrimRight(publicURL, "/"), storeHash)
		http.Redirect(w, r, redirectURL, http.StatusFound)
		return
	}

	// json mode: serve from the cached index.
	indexData, _ := ps.CacheIndex.Get()
	if entry, ok := indexData.Entries[storeHash]; ok && entry.NarInfo != "" {
		body := []byte(entry.NarInfo)
		w.Header().Set("Content-Type", "text/x-nix-narinfo")
		w.Header().Set("Content-Length", strconv.Itoa(len(body)))
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(body)
		return
	}

	// Fallback to upstream cache.
	upstreamPath := fmt.Sprintf("/%s.narinfo", storeHash)
	if ps.proxyUpstream(w, r, upstreamPath) {
		return
	}

	http.Error(w, "Narinfo Not Found", http.StatusNotFound)
}

func (ps *ProxyServer) serveNar(w http.ResponseWriter, r *http.Request, path string, method string) {
	narBasename := strings.TrimPrefix(path, "/nar/")
	contentType := "application/x-nix-nar"
	if strings.HasSuffix(narBasename, ".xz") {
		contentType = "application/x-xz"
	}

	var digest string

	// r2 mode: derive the blob digest from the narinfo's FileHash served by R2.
	if ps.CacheIndex.IsR2() {
		if d, err := ps.digestFromR2Narinfo(narBasename); err == nil && d != "" {
			digest = d
		} else if err != nil {
			fmt.Fprintf(os.Stderr, "[aeroflare proxy] Warning: failed to derive digest from R2 narinfo for %s: %v. Trying upstream...\n", narBasename, err)
		}
	} else {
		// json mode: use the NAR lookup built from the cached index.
		_, narLookups := ps.CacheIndex.Get()
		if d, ok := narLookups[narBasename]; ok && d != "" {
			digest = d
		}
	}

	if digest != "" && strings.HasPrefix(digest, "sha256:") {
		err := ps.streamBlob(w, digest, contentType, method)
		if err == nil {
			return
		}
		fmt.Fprintf(os.Stderr, "[aeroflare proxy] Warning: failed to stream blob %s from GHCR: %v. Trying upstream...\n", digest, err)
	}

	if ps.proxyUpstream(w, r, path) {
		return
	}

	http.Error(w, "NAR Not Found", http.StatusNotFound)
}

// digestFromR2Narinfo fetches the narinfo for a NAR from the public R2 URL and
// converts its FileHash (sha256:<nix-base32> of the compressed NAR) into the
// GHCR blob digest form (sha256:<hex>).
func (ps *ProxyServer) digestFromR2Narinfo(narBasename string) (string, error) {
	publicURL := ps.CacheIndex.PublicR2URL()
	if publicURL == "" {
		return "", fmt.Errorf("no public R2 URL configured")
	}

	// The narinfo key is the store hash: strip the compression extension from
	// the NAR basename (e.g. <hash>.nar.xz -> <hash>).
	storeHash := narBasename
	if idx := strings.Index(storeHash, ".nar"); idx != -1 {
		storeHash = storeHash[:idx]
	}

	narinfoURL := fmt.Sprintf("%s/%s.narinfo", strings.TrimRight(publicURL, "/"), storeHash)
	req, err := http.NewRequest("GET", narinfoURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "aeroflare/1.0")

	resp, err := ps.HttpShortClient.Do(req)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("R2 narinfo HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return fileHashToBlobDigest(string(body))
}

// fileHashToBlobDigest extracts the "FileHash: sha256:<nix-base32>" line from a
// narinfo and returns the equivalent GHCR blob digest "sha256:<hex>".
func fileHashToBlobDigest(narinfo string) (string, error) {
	var fileHash string
	for _, line := range strings.Split(narinfo, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "FileHash:") {
			fileHash = strings.TrimSpace(strings.TrimPrefix(line, "FileHash:"))
			break
		}
	}
	if fileHash == "" {
		return "", fmt.Errorf("narinfo has no FileHash line")
	}
	if !strings.HasPrefix(fileHash, "sha256:") {
		return "", fmt.Errorf("unsupported FileHash algorithm: %s", fileHash)
	}
	encoded := strings.TrimPrefix(fileHash, "sha256:")
	raw, err := narhash.DecodeBase32(encoded)
	if err != nil {
		return "", fmt.Errorf("decode nix-base32 FileHash: %w", err)
	}
	return "sha256:" + hex.EncodeToString(raw), nil
}

func (ps *ProxyServer) streamBlob(w http.ResponseWriter, digest string, contentType string, method string) error {
	token, err := ps.TokenMgr.GetToken()
	if err != nil {
		return err
	}

	proto := GetProtocol(ps.Registry)
	url := fmt.Sprintf("%s://%s/v2/%s/blobs/%s", proto, ps.Registry, ps.Repository, digest)
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "aeroflare/1.0")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := ps.HttpClient.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("GHCR blob download HTTP %d: %s", resp.StatusCode, string(bodyBytes))
	}

	w.Header().Set("Content-Type", contentType)
	if resp.ContentLength > 0 {
		w.Header().Set("Content-Length", strconv.FormatInt(resp.ContentLength, 10))
	}
	w.WriteHeader(http.StatusOK)

	_, err = io.Copy(w, resp.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[aeroflare proxy] Warning: stream interrupted for blob %s: %v\n", digest, err)
	}
	return nil
}

func (ps *ProxyServer) proxyUpstream(w http.ResponseWriter, r *http.Request, upstreamPath string) bool {
	if len(ps.UpstreamCaches) == 0 {
		return false
	}
	upstreamURL := fmt.Sprintf("%s%s", strings.TrimSuffix(ps.UpstreamCaches[0], "/"), upstreamPath)
	req, err := http.NewRequest(r.Method, upstreamURL, nil)
	if err != nil {
		return false
	}
	req.Header.Set("User-Agent", "aeroflare/1.0")
	resp, err := ps.HttpClient.Do(req)
	if err != nil {
		return false
	}
	defer func() { _ = resp.Body.Close() }()

	for k, vv := range resp.Header {
		for _, v := range vv {
			w.Header().Add(k, v)
		}
	}
	w.WriteHeader(resp.StatusCode)
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[aeroflare proxy] Warning: stream interrupted for upstream path %s: %v\n", upstreamPath, err)
	}
	return true
}

// BootstrapConfig fetches configuration dynamically from the GHCR 'cache-config' OCI image/blob.
// Used by the push/configure pipeline; the proxy itself reads from the cache-index manifest.
func BootstrapConfig(registry, repository string, tokenMgr *TokenManager) (*RemoteConfig, error) {
	config, _, err := BootstrapConfigWithAnnotations(registry, repository, tokenMgr)
	return config, err
}

func BootstrapConfigWithAnnotations(registry, repository string, tokenMgr *TokenManager) (*RemoteConfig, map[string]string, error) {
	token, err := tokenMgr.GetToken()
	if err != nil {
		return nil, nil, err
	}

	client := &http.Client{Timeout: 10 * time.Second}
	proto := GetProtocol(registry)

	manifestURL := fmt.Sprintf("%s://%s/v2/%s/manifests/cache-config", proto, registry, repository)
	req, err := http.NewRequest("GET", manifestURL, nil)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Accept", "application/vnd.oci.image.manifest.v1+json")
	req.Header.Set("User-Agent", "aeroflare/1.0")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, nil, fmt.Errorf("config manifest HTTP %d", resp.StatusCode)
	}

	var manifest IndexManifest
	if err := json.NewDecoder(resp.Body).Decode(&manifest); err != nil {
		return nil, nil, err
	}

	if len(manifest.Layers) == 0 {
		return nil, manifest.Annotations, fmt.Errorf("no layers in cache-config manifest")
	}

	digest := manifest.Layers[0].Digest
	blobURL := fmt.Sprintf("%s://%s/v2/%s/blobs/%s", proto, registry, repository, digest)
	req, err = http.NewRequest("GET", blobURL, nil)
	if err != nil {
		return nil, manifest.Annotations, err
	}
	req.Header.Set("User-Agent", "aeroflare/1.0")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err = client.Do(req)
	if err != nil {
		return nil, manifest.Annotations, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, manifest.Annotations, fmt.Errorf("config blob HTTP %d", resp.StatusCode)
	}

	var config RemoteConfig
	if err := json.NewDecoder(resp.Body).Decode(&config); err != nil {
		return nil, manifest.Annotations, err
	}

	if pk, ok := manifest.Annotations["public-key"]; ok && pk != "" && config.PublicKey == "" {
		config.PublicKey = pk
	}

	return &config, manifest.Annotations, nil
}

// StartProxy starts the proxy HTTP server on the configured address.
func StartProxy(ctx context.Context, port int, listenAddr string, registry string, repository string, indexDir string, cacheFileName string, indexTTLSeconds int, upstreams []string, githubToken string) (int, error) {
	tokenMgr := NewTokenManager(registry, repository, githubToken)

	if cacheFileName == "" {
		cacheFileName = "cache-index.json"
	}

	ttl := time.Duration(indexTTLSeconds) * time.Second
	cacheIndex := &CacheIndex{
		IndexDir:      indexDir,
		CacheFileName: cacheFileName,
		IndexTTL:      ttl,
		TokenMgr:      tokenMgr,
		Registry:      registry,
		Repository:    repository,
	}

	// Seed the local cache file (if present) so the very first request doesn't
	// block on a registry round-trip, then refresh in the background.
	_ = cacheIndex.loadLocal()
	go func() {
		_, _ = cacheIndex.Get()
	}()

	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.MaxIdleConns = 100
	transport.MaxIdleConnsPerHost = 100
	transport.IdleConnTimeout = 90 * time.Second

	ps := &ProxyServer{
		Port:           port,
		ListenAddr:     listenAddr,
		Registry:       registry,
		Repository:     repository,
		UpstreamCaches: upstreams,
		TokenMgr:       tokenMgr,
		CacheIndex:     cacheIndex,
		HttpClient: &http.Client{
			Transport: transport,
			Timeout:   30 * time.Minute,
		},
		HttpShortClient: &http.Client{
			Transport: transport,
			Timeout:   10 * time.Second,
		},
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", ps.Handler)

	addr := fmt.Sprintf("%s:%d", listenAddr, port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return 0, err
	}
	actualPort := listener.Addr().(*net.TCPAddr).Port

	server := &http.Server{
		Handler: mux,
	}

	go func() {
		fmt.Fprintf(os.Stderr, "[aeroflare proxy] Starting proxy server on http://%s:%d\n", listenAddr, actualPort)
		fmt.Fprintf(os.Stderr, "  Repo: %s\n", repository)
		fmt.Fprintf(os.Stderr, "  Upstream: %s\n", strings.Join(upstreams, ", "))
		fmt.Fprintf(os.Stderr, "  Index TTL: %s\n", ttl)

		if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		}
	}()

	go func() {
		<-ctx.Done()
		fmt.Fprintf(os.Stderr, "\n[aeroflare proxy] Shutting down...\n")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
	}()

	return actualPort, nil
}
