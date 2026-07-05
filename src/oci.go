package network

import (
	"bufio"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"aeroflare/src/proxy"
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

func FetchAeroflareAnnotations(registry, repository, tag, token string) (map[string]string, error) {
	opts := []name.Option{}
	if proxy.GetProtocol(registry) == "http" {
		opts = append(opts, name.Insecure)
	}

	refStr := fmt.Sprintf("%s/%s:%s", registry, repository, tag)
	ref, err := name.ParseReference(refStr, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to parse reference: %w", err)
	}

	remoteOpts := []remote.Option{
		remote.WithTransport(optimizedTransport),
	}
	if token != "" {
		remoteOpts = append(remoteOpts, remote.WithAuth(&authn.Bearer{Token: token}))
	}

	img, err := remote.Image(ref, remoteOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch image: %w", err)
	}

	manifest, err := img.Manifest()
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest: %w", err)
	}

	result := make(map[string]string)
	for key, val := range manifest.Annotations {
		if strings.HasPrefix(key, "aeroflare.") {
			result[key] = val
		}
	}

	return result, nil
}
