package ci

import (
	"strings"
	"testing"
)

func TestSummaryLine_Success(t *testing.T) {
	got := summaryLine(2, 2, 1, 1, 417)
	if !strings.Contains(got, "builds 2/2") || !strings.Contains(got, "pushes 1/1") {
		t.Errorf("got %q", got)
	}
	if !strings.Contains(got, "paths 417") {
		t.Errorf("got %q", got)
	}
	if !strings.Contains(got, "OK") {
		t.Errorf("expected OK marker, got %q", got)
	}
}

func TestSummaryLine_Failure(t *testing.T) {
	got := summaryLine(2, 1, 1, 0, 5)
	if !strings.Contains(got, "FAILED") {
		t.Errorf("expected FAILED marker, got %q", got)
	}
}

// The prepare step filters the closure against this URL: an empty string means
// "prepare everything", so "none" must map to "" and a real URL must survive.
func TestUpstreamCacheURL(t *testing.T) {
	for _, tc := range []struct {
		name string
		spec string
		want string
	}{
		{"configured url is used for filtering", "https://cache.nixos.org", "https://cache.nixos.org"},
		{"none disables filtering", "none", ""},
		{"empty disables filtering", "", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := upstreamCacheURL(RunSpec{UpstreamCache: tc.spec}); got != tc.want {
				t.Errorf("upstreamCacheURL(%q) = %q, want %q", tc.spec, got, tc.want)
			}
		})
	}
}
