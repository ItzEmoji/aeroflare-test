package r2

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type R2Config struct {
	Bucket    string
	Endpoint  string
	AccessKey string
	SecretKey string
	PublicURL string
}

// GetR2Config builds an R2Config from environment variables, falling back to
// the cache-config annotations (e.g. aeroflare.r2.bucket) when the
// corresponding env var is unset. It returns nil when no bucket is
// configured, signaling that the R2 backend is not in use.
func GetR2Config(annotations map[string]string) *R2Config {
	bucket := os.Getenv("R2_BUCKET")
	if bucket == "" && annotations != nil {
		bucket = annotations["aeroflare.r2.bucket"]
	}

	if bucket == "" {
		return nil
	}

	endpoint := os.Getenv("R2_ENDPOINT")
	if endpoint == "" && annotations != nil {
		endpoint = annotations["aeroflare.r2.endpoint"]
	}

	accessKey := os.Getenv("R2_ACCESS_KEY_ID")
	secretKey := os.Getenv("R2_SECRET_ACCESS_KEY")

	publicURL := os.Getenv("R2_PUBLIC_URL")
	if publicURL == "" && annotations != nil {
		if u := annotations["public-r2-url"]; u != "" {
			publicURL = u
		} else if u := annotations["aeroflare.r2.public_url"]; u != "" {
			publicURL = u
		}
	}

	return &R2Config{
		Bucket:    bucket,
		Endpoint:  endpoint,
		AccessKey: accessKey,
		SecretKey: secretKey,
		PublicURL: publicURL,
	}
}

// NewClient builds an S3 client for R2 using path-style addressing (R2
// requires this rather than virtual-hosted-style bucket URLs).
func (r *R2Config) NewClient(ctx context.Context) (*s3.Client, error) {
	if r.AccessKey == "" || r.SecretKey == "" {
		return nil, fmt.Errorf("R2 credentials missing: set R2_ACCESS_KEY_ID and R2_SECRET_ACCESS_KEY")
	}
	if r.Endpoint == "" {
		return nil, fmt.Errorf("R2 endpoint missing: set R2_ENDPOINT or the aeroflare.r2.endpoint annotation")
	}
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(r.AccessKey, r.SecretKey, "")),
		config.WithRegion("auto"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load R2 config: %w", err)
	}

	return s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(r.Endpoint)
		o.UsePathStyle = true
	}), nil
}

// UploadNarinfo uploads the .narinfo file for storePath to R2, keyed by the
// store hash (e.g. "<hash>.narinfo") so it can be looked up directly.
func (r *R2Config) UploadNarinfo(ctx context.Context, client *s3.Client, storePath, narinfoPath string) error {
	b, err := os.ReadFile(narinfoPath)
	if err != nil {
		return fmt.Errorf("read narinfo: %w", err)
	}

	basename := filepath.Base(storePath)
	parts := strings.SplitN(basename, "-", 2)
	if len(parts) < 2 {
		return fmt.Errorf("invalid store path format: %s", basename)
	}
	hash := parts[0]
	objectKey := fmt.Sprintf("%s.narinfo", hash)

	_, err = client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(r.Bucket),
		Key:         aws.String(objectKey),
		Body:        bytes.NewReader(b),
		ContentType: aws.String("text/x-nix-narinfo"),
	})

	if err != nil {
		return fmt.Errorf("upload to R2: %w", err)
	}

	return nil
}
