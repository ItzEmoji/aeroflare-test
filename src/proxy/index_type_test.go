package proxy

import "testing"

func TestIndexType(t *testing.T) {
	cases := []struct {
		name        string
		annotations map[string]string
		want        string
	}{
		{"nil annotations defaults to json", nil, "json"},
		{"empty annotations defaults to json", map[string]string{}, "json"},
		{"aeroflare.backend wins", map[string]string{"aeroflare.backend": "r2"}, "r2"},
		{"bare backend key", map[string]string{"backend": "native"}, "native"},
		{"legacy aeroflare.index-type still honored", map[string]string{"aeroflare.index-type": "r2"}, "r2"},
		{"legacy index-type still honored", map[string]string{"index-type": "native"}, "native"},
		{"new key preferred over legacy", map[string]string{"aeroflare.backend": "json", "aeroflare.index-type": "r2"}, "json"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ci := &CacheIndex{ManifestAnnotations: tc.annotations}
			if got := ci.IndexType(); got != tc.want {
				t.Errorf("IndexType() = %q, want %q", got, tc.want)
			}
		})
	}
}
