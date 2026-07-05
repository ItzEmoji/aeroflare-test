package network

import (
	"bufio"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"strings"
	"context"
	"net/http"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
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
	imageRef := fmt.Sprintf("%s/%s:%s", registry, repository, tag)
	ref, err := name.ParseReference(imageRef)
	if err != nil {
		return nil, fmt.Errorf("failed to parse reference: %w", err)
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
		return nil, fmt.Errorf("failed to fetch image: %w", err)
	}

	raw, err := img.RawManifest()
	if err != nil {
		return nil, fmt.Errorf("failed to read raw manifest: %w", err)
	}

	var rawManifest struct {
		Annotations map[string]string `json:"annotations"`
		Labels      map[string]string `json:"labels"`
	}
	if err := json.Unmarshal(raw, &rawManifest); err != nil {
		return nil, fmt.Errorf("failed to unmarshal raw manifest: %w", err)
	}

	result := make(map[string]string)
	for key, val := range rawManifest.Annotations {
		result[key] = val
	}
	for key, val := range rawManifest.Labels {
		result[key] = val
	}

	return result, nil
}
