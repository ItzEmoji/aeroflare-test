package ci

import "testing"

func TestResolve_FileBaseWithDefaults(t *testing.T) {
	fc := FileConfig{Builds: []string{".#a"}, Caches: []string{"ghcr.io;o/c"}}
	spec, err := Resolve(fc, Inputs{})
	if err != nil {
		t.Fatal(err)
	}
	if len(spec.Builds) != 1 || spec.Builds[0] != ".#a" {
		t.Errorf("builds = %v", spec.Builds)
	}
	if len(spec.Caches) != 1 || spec.Caches[0].Registry != "ghcr.io" {
		t.Errorf("caches = %+v", spec.Caches)
	}
	if spec.Compression != "zstd" {
		t.Errorf("compression = %q, want zstd", spec.Compression)
	}
	if spec.Workers != 50 {
		t.Errorf("workers = %d, want 50", spec.Workers)
	}
	if spec.UpstreamCache != "https://cache.nixos.org" {
		t.Errorf("upstream = %q", spec.UpstreamCache)
	}
}

func TestResolve_InlineReplacesLists(t *testing.T) {
	fc := FileConfig{Builds: []string{".#a"}, Caches: []string{"ghcr.io;o/c"}, Compression: "xz"}
	in := Inputs{Builds: []string{".#b", ".#c"}, Caches: []string{"docker.io;o/d"}, Compression: "gzip", Workers: 4}
	spec, err := Resolve(fc, in)
	if err != nil {
		t.Fatal(err)
	}
	if len(spec.Builds) != 2 || spec.Builds[0] != ".#b" {
		t.Errorf("builds = %v (inline should replace)", spec.Builds)
	}
	if len(spec.Caches) != 1 || spec.Caches[0].Registry != "docker.io" {
		t.Errorf("caches = %+v (inline should replace)", spec.Caches)
	}
	if spec.Compression != "gzip" {
		t.Errorf("compression = %q, want gzip", spec.Compression)
	}
	if spec.Workers != 4 {
		t.Errorf("workers = %d, want 4", spec.Workers)
	}
}

func TestResolve_MissingBuilds(t *testing.T) {
	_, err := Resolve(FileConfig{Caches: []string{"ghcr.io;o/c"}}, Inputs{})
	if err == nil {
		t.Fatal("expected error for missing builds")
	}
}

func TestResolve_MissingCaches(t *testing.T) {
	_, err := Resolve(FileConfig{Builds: []string{".#a"}}, Inputs{})
	if err == nil {
		t.Fatal("expected error for missing caches")
	}
}

func TestResolve_BadCache(t *testing.T) {
	_, err := Resolve(FileConfig{Builds: []string{".#a"}, Caches: []string{"nope"}}, Inputs{})
	if err == nil {
		t.Fatal("expected error for malformed cache")
	}
}

func TestLoadFile_MissingOptional(t *testing.T) {
	fc, found, err := LoadFile("/nonexistent/aeroflare-ci.yaml", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found {
		t.Error("found should be false")
	}
	if len(fc.Builds) != 0 {
		t.Error("expected empty config")
	}
}
