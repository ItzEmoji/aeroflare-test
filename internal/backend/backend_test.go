package backend

import (
	"aeroflare/internal/r2"
	"reflect"
	"testing"
)

func TestNewCacheBackend(t *testing.T) {
	tests := []struct {
		name     string
		cfg      BackendConfig
		expected CacheBackend
	}{
		{
			name:     "default is native",
			cfg:      BackendConfig{},
			expected: &NativeBackend{cfg: BackendConfig{}},
		},
		{
			name: "config annotations set json",
			cfg: BackendConfig{
				ConfigAnnotations: map[string]string{
					"aeroflare.backend": "json",
				},
			},
			expected: &JSONBackend{cfg: BackendConfig{
				ConfigAnnotations: map[string]string{
					"aeroflare.backend": "json",
				},
			}},
		},
		{
			name: "r2 config overrides config annotations",
			cfg: BackendConfig{
				ConfigAnnotations: map[string]string{
					"aeroflare.backend": "json",
				},
				R2: &r2.R2Config{
					PublicURL: "https://example.com",
				},
			},
			expected: &R2Backend{cfg: BackendConfig{
				ConfigAnnotations: map[string]string{
					"aeroflare.backend": "json",
				},
				R2: &r2.R2Config{
					PublicURL: "https://example.com",
				},
			}},
		},
		{
			name: "r2 config without public URL uses config annotations",
			cfg: BackendConfig{
				ConfigAnnotations: map[string]string{
					"aeroflare.backend": "json",
				},
				R2: &r2.R2Config{},
			},
			expected: &JSONBackend{cfg: BackendConfig{
				ConfigAnnotations: map[string]string{
					"aeroflare.backend": "json",
				},
				R2: &r2.R2Config{},
			}},
		},
		{
			name: "unknown annotation falls back to native",
			cfg: BackendConfig{
				ConfigAnnotations: map[string]string{
					"aeroflare.backend": "unknown",
				},
			},
			expected: &NativeBackend{cfg: BackendConfig{
				ConfigAnnotations: map[string]string{
					"aeroflare.backend": "unknown",
				},
			}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewCacheBackend(tt.cfg)
			if reflect.TypeOf(result) != reflect.TypeOf(tt.expected) {
				t.Errorf("expected type %T, got %T", tt.expected, result)
			}
		})
	}
}
