package network

import (
	"bufio"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"strings"
	"context"
	"net/http"

	v1 "github.com/google/go-containerregistry/pkg/v1"
)

type withArtifactType struct {
	v1.Image
	artifactType string
}

func (w *withArtifactType) Manifest() (*v1.Manifest, error) {
	m, err := w.Image.Manifest()
	if err != nil {
		return nil, err
	}
	mCopy := *m
	mCopy.ArtifactType = w.artifactType
	return &mCopy, nil
}

func (w *withArtifactType) RawManifest() ([]byte, error) {
	m, err := w.Manifest()
	if err != nil {
		return nil, err
	}
	return json.Marshal(m)
}

func (w *withArtifactType) Digest() (v1.Hash, error) {
	b, err := w.RawManifest()
	if err != nil {
		return v1.Hash{}, err
	}
	h := sha256.Sum256(b)
	return v1.Hash{
		Algorithm: "sha256",
		Hex:       fmt.Sprintf("%x", h),
	}, nil
}

func (w *withArtifactType) Size() (int64, error) {
	b, err := w.RawManifest()
	if err != nil {
		return 0, err
	}
	return int64(len(b)), nil
}

func ParseAeroflareMetadata(content string) (map[string]string, error) {
	annotations := make(map[string]string)
	scanner := bufio.NewScanner(strings.NewReader(content))

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid metadata format on line: %s", line)
		}
		rawKey := strings.TrimSpace(parts[0])
		key := fmt.Sprintf("aeroflare.%s", strings.ToLower(rawKey))
		val := strings.TrimSpace(parts[1])
		annotations[key] = val
	}
	
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	
	return annotations, nil
}

func FetchAeroflareAnnotations(ctx context.Context, client *http.Client, registry, repository, tag, token string) (map[string]string, error) {
	if client == nil {
		client = http.DefaultClient
	}

	proto := GetProtocol(registry)
	manifestURL := fmt.Sprintf("%s://%s/v2/%s/manifests/%s", proto, registry, repository, tag)

	req, err := http.NewRequestWithContext(ctx, "GET", manifestURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed creating manifest request: %w", err)
	}
	
	req.Header.Set("Accept", "application/vnd.oci.image.manifest.v1+json, application/vnd.docker.distribution.manifest.v2+json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed fetching manifest from %s: %w", manifestURL, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("manifest HTTP %d for %s", resp.StatusCode, manifestURL)
	}

	var manifest struct {
		Annotations map[string]string `json:"annotations"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&manifest); err != nil {
		return nil, fmt.Errorf("failed decoding manifest JSON: %w", err)
	}

	result := make(map[string]string)
	for key, val := range manifest.Annotations {
		if strings.HasPrefix(key, "aeroflare.") {
			result[key] = val
		}
	}

	return result, nil
}
