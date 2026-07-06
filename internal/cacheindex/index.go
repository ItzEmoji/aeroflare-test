package cacheindex

import (
	"aeroflare/internal/oci"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/static"
	"github.com/google/go-containerregistry/pkg/v1/types"
)

// Types for cache-index.json management
type PushCacheIndex struct {
	Version   int                       `json:"version"`
	Repo      string                    `json:"repo"`
	Registry  string                    `json:"registry"`
	Image     string                    `json:"image"`
	Generated string                    `json:"generated"`
	PublicKey string                    `json:"public_key"`
	Entries   map[string]PushCacheEntry `json:"entries"`
	GCRoots   []string                  `json:"gc_roots"`
}

type PushCacheEntry struct {
	Name      string `json:"name"`
	Narinfo   string `json:"narinfo"`
	NarDigest string `json:"nar_digest"`
	NarSize   int64  `json:"nar_size"`
	Added     string `json:"added"`
}

type PushReceipt struct {
	StorePath   string
	NarinfoPath string
	NarDigest   string
	NarSize     int64
	NarPath     string
	Compression string // narinfo Compression value (e.g. "xz", "zstd")
	IsRoot      bool
}

// FetchCacheIndex fetches the current cache-index. A missing index (HTTP 404)
// yields an empty index; any other failure is returned as an error so callers
// never mistake an unreachable index for an empty cache.
func FetchCacheIndex(registry, repository, token string) (*PushCacheIndex, error) {
	index, _, err := FetchCacheIndexWithDigest(registry, repository, token)
	return index, err
}

// FetchCacheIndexWithDigest fetches the cache-index and the digest of its
// manifest (Docker-Content-Digest). The digest is "" when no index exists yet.
func FetchCacheIndexWithDigest(registry, repository, token string) (*PushCacheIndex, string, error) {
	if reg, err := name.NewRegistry(registry); err == nil {
		registry = reg.RegistryStr()
	}
	scheme := oci.GetProtocol(registry)
	manifestURL := fmt.Sprintf("%s://%s/v2/%s/manifests/cache-index", scheme, registry, repository)

	client := &http.Client{Timeout: 30 * time.Second}

	emptyIndex := func() *PushCacheIndex {
		return &PushCacheIndex{Entries: make(map[string]PushCacheEntry)}
	}

	resp, err := oci.DoWithRetry(client, func() (*http.Request, error) {
		req, err := http.NewRequest("GET", manifestURL, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Accept", "application/vnd.oci.image.manifest.v1+json")
		return req, nil
	})
	if err != nil {
		return nil, "", fmt.Errorf("fetch cache-index manifest: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusNotFound {
		// No index yet: a fresh cache.
		return emptyIndex(), "", nil
	}
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return nil, "", fmt.Errorf("fetch cache-index manifest: HTTP %d: %s", resp.StatusCode, string(bodyBytes))
	}

	manifestBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("read cache-index manifest: %w", err)
	}
	// Compute the digest from the body rather than trusting a header; the OCI
	// spec stores manifests byte-exact, so this matches the canonical digest.
	manifestDigest := manifestSHA256(manifestBytes)

	var manifest struct {
		Layers []struct {
			Digest string `json:"digest"`
		} `json:"layers"`
	}
	if err := json.Unmarshal(manifestBytes, &manifest); err != nil {
		return nil, "", fmt.Errorf("decode cache-index manifest: %w", err)
	}
	if len(manifest.Layers) == 0 {
		// Manifest exists but carries no index blob; treat as empty.
		return emptyIndex(), manifestDigest, nil
	}

	blobURL := fmt.Sprintf("%s://%s/v2/%s/blobs/%s", scheme, registry, repository, manifest.Layers[0].Digest)
	blobResp, err := oci.DoWithRetry(client, func() (*http.Request, error) {
		blobReq, err := http.NewRequest("GET", blobURL, nil)
		if err != nil {
			return nil, err
		}
		blobReq.Header.Set("Authorization", "Bearer "+token)
		return blobReq, nil
	})
	if err != nil {
		return nil, "", fmt.Errorf("fetch cache-index blob: %w", err)
	}
	defer func() { _ = blobResp.Body.Close() }()
	if blobResp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("fetch cache-index blob: HTTP %d", blobResp.StatusCode)
	}

	blobBytes, err := io.ReadAll(blobResp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("read cache-index blob: %w", err)
	}

	existingIndex := emptyIndex()
	if err := json.Unmarshal(blobBytes, existingIndex); err != nil {
		return nil, "", fmt.Errorf("decode cache-index blob: %w", err)
	}
	if existingIndex.Entries == nil {
		existingIndex.Entries = make(map[string]PushCacheEntry)
	}

	return existingIndex, manifestDigest, nil
}

// manifestSHA256 computes the "sha256:<hex>" digest string for raw manifest
// bytes, matching the OCI content-addressable digest format.
func manifestSHA256(data []byte) string {
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:])
}

// UpdateCacheIndex merges new entries into existingIndex and pushes the updated manifest.
func UpdateCacheIndex(receipts []PushReceipt, existingIndex *PushCacheIndex, registry, repository, token, pubKeyPath string, configAnnotations map[string]string) error {
	_, err := UpdateCacheIndexWithDigest(receipts, existingIndex, registry, repository, token, pubKeyPath, configAnnotations, 0)
	return err
}

// UpdateCacheIndexWithDigest does the merge+push and returns the digest of the
// manifest it wrote, so callers can detect concurrent writers afterwards.
// workers bounds how many receipts are merged concurrently; if <= 0 it falls
// back to a default limit.
func UpdateCacheIndexWithDigest(receipts []PushReceipt, existingIndex *PushCacheIndex, registry, repository, token, pubKeyPath string, configAnnotations map[string]string, workers int) (string, error) {

	if reg, err := name.NewRegistry(registry); err == nil {
		registry = reg.RegistryStr()
	}
	scheme := oci.GetProtocol(registry)
	manifestURL := fmt.Sprintf("%s://%s/v2/%s/manifests/cache-index", scheme, registry, repository)

	// Determine public key
	pubKey := existingIndex.PublicKey
	if pubKeyPath != "" {
		if b, err := os.ReadFile(pubKeyPath + ".pub"); err == nil {
			pubKey = strings.TrimSpace(string(b))
		}
	}

	imageName := os.Getenv("NIXCACHE_IMAGE")
	if imageName == "" {
		imageName = repository
	}

	// Build new index
	newIndex := PushCacheIndex{
		Version:   1,
		Repo:      repository,
		Registry:  registry,
		Image:     imageName,
		Generated: time.Now().UTC().Format("2006-01-02T15:04:05Z"),
		PublicKey: pubKey,
		Entries:   existingIndex.Entries,
		GCRoots:   existingIndex.GCRoots,
	}

	if newIndex.Entries == nil {
		newIndex.Entries = make(map[string]PushCacheEntry)
	}

	gcRootsSet := make(map[string]bool)
	for _, root := range newIndex.GCRoots {
		gcRootsSet[root] = true
	}

	newEntriesCount := 0
	var mu sync.Mutex
	var eg errgroup.Group
	eg.SetLimit(WorkerLimit(workers, 20))

	for _, r := range receipts {
		r := r
		eg.Go(func() error {
			narinfoBytes, err := os.ReadFile(r.NarinfoPath)
			if err != nil {
				return fmt.Errorf("failed to read narinfo (%s): %w", r.NarinfoPath, err)
			}

			basename := filepath.Base(r.StorePath)
			parts := strings.SplitN(basename, "-", 2)
			if len(parts) < 2 {
				return nil
			}
			hash := parts[0]
			storeName := parts[1]

			mu.Lock()
			newIndex.Entries[hash] = PushCacheEntry{
				Name:      storeName,
				Narinfo:   string(narinfoBytes),
				NarDigest: r.NarDigest,
				NarSize:   r.NarSize,
				Added:     newIndex.Generated,
			}
			newEntriesCount++

			if r.IsRoot {
				if !gcRootsSet[hash] {
					newIndex.GCRoots = append(newIndex.GCRoots, hash)
					gcRootsSet[hash] = true
				}
			}
			mu.Unlock()

			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		return "", fmt.Errorf("collect index entries: %w", err)
	}

	sort.Strings(newIndex.GCRoots)

	// Push new index JSON as blob
	indexBytes, err := json.MarshalIndent(newIndex, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal cache index: %w", err)
	}

	indexDigest, err := oci.PushBlobBytes(indexBytes, registry, repository, token)
	if err != nil {
		return "", fmt.Errorf("push cache index blob: %w", err)
	}

	// Push empty config JSON as blob
	configBytes := []byte(fmt.Sprintf(`{"created":"%s"}`, time.Now().UTC().Format(time.RFC3339Nano)))
	configDigest, err := oci.PushBlobBytes(configBytes, registry, repository, token)
	if err != nil {
		return "", fmt.Errorf("push config blob: %w", err)
	}

	// Push Cache-Index Manifest
	manifest := map[string]interface{}{
		"schemaVersion": 2,
		"mediaType":     "application/vnd.oci.image.manifest.v1+json",
		"config": map[string]interface{}{
			"mediaType": "application/vnd.oci.image.config.v1+json",
			"digest":    configDigest,
			"size":      len(configBytes),
		},
		"layers": []map[string]interface{}{
			{
				"mediaType": "application/vnd.nix.cache.index.v1+json",
				"digest":    indexDigest,
				"size":      len(indexBytes),
			},
		},
	}

	manifestAnnotations := map[string]string{}
	if configAnnotations != nil {
		if pk := configAnnotations["aeroflare.public-key"]; pk != "" {
			manifestAnnotations["aeroflare.public-key"] = pk
		} else if pk := configAnnotations["public-key"]; pk != "" {
			manifestAnnotations["aeroflare.public-key"] = pk
		}
		if backend := configAnnotations["aeroflare.backend"]; backend != "" {
			manifestAnnotations["aeroflare.backend"] = backend
		}
	}

	if manifestAnnotations["aeroflare.backend"] == "" {
		manifestAnnotations["aeroflare.backend"] = "json"
	}

	if len(manifestAnnotations) > 0 {
		manifest["annotations"] = manifestAnnotations
	}

	manifestBytes, err := json.Marshal(manifest)
	if err != nil {
		return "", fmt.Errorf("marshal cache-index manifest: %w", err)
	}
	client := &http.Client{Timeout: 30 * time.Second}
	putResp, err := oci.DoWithRetry(client, func() (*http.Request, error) {
		putReq, err := http.NewRequest("PUT", manifestURL, bytes.NewReader(manifestBytes))
		if err != nil {
			return nil, err
		}
		putReq.Header.Set("Authorization", "Bearer "+token)
		putReq.Header.Set("Content-Type", "application/vnd.oci.image.manifest.v1+json")
		return putReq, nil
	})
	if err != nil {
		return "", fmt.Errorf("push manifest: %w", err)
	}
	defer func() { _ = putResp.Body.Close() }()

	if putResp.StatusCode != http.StatusCreated && putResp.StatusCode != http.StatusOK {
		respBytes, _ := io.ReadAll(putResp.Body)
		return "", fmt.Errorf("push manifest failed with HTTP %d: %s", putResp.StatusCode, string(respBytes))
	}

	return manifestSHA256(manifestBytes), nil
}

// PushConfigManifest pushes the cache-config manifest with the provided annotations.
func PushConfigManifest(registry, repository, token string, annotations map[string]string) error {
	opts := []name.Option{}
	if oci.GetProtocol(registry) == "http" {
		opts = append(opts, name.Insecure)
	}

	refStr := fmt.Sprintf("%s/%s:cache-config", registry, repository)
	tagRef, err := name.NewTag(refStr, opts...)
	if err != nil {
		return fmt.Errorf("failed to create tag: %w", err)
	}

	cacheImg := empty.Image
	cacheImg = mutate.MediaType(cacheImg, types.OCIManifestSchema1)
	cacheImg = mutate.ConfigMediaType(cacheImg, "application/vnd.oci.empty.v1+json")

	emptyLayer := static.NewLayer([]byte("{}"), "application/vnd.oci.empty.v1+json")
	cacheImg, err = mutate.AppendLayers(cacheImg, emptyLayer)
	if err != nil {
		return fmt.Errorf("failed to append empty layer: %w", err)
	}

	cacheImg = mutate.Annotations(cacheImg, annotations).(v1.Image)

	finalCacheImg := oci.NewArtifactTypeImage(cacheImg, "application/vnd.aeroflare.cache-config.v1")

	remoteOpts := []remote.Option{
		remote.WithTransport(oci.WriteTransport()),
	}
	if token != "" {
		remoteOpts = append(remoteOpts, remote.WithAuth(&authn.Bearer{Token: token}))
	}

	if err := remote.Write(tagRef, finalCacheImg, remoteOpts...); err != nil {
		return fmt.Errorf("failed to push config manifest: %w", err)
	}

	return nil
}
