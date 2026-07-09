package ci

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// FileConfig is the on-disk .aeroflare-ci.yaml schema.
type FileConfig struct {
	Builds        []string `yaml:"builds"`
	Caches        []string `yaml:"caches"`
	Compression   string   `yaml:"compression"`
	SigningKey    string   `yaml:"signing-key"`
	Workers       int      `yaml:"workers"`
	UpstreamCache string   `yaml:"upstream-cache"`
}

// Inputs are inline overrides (flags/env). Empty/zero fields mean "not set".
type Inputs struct {
	Builds        []string
	Caches        []string
	Compression   string
	SigningKey    string
	Workers       int
	UpstreamCache string
}

// RunSpec is the fully resolved configuration for one aeroflare-ci run.
type RunSpec struct {
	Builds        []string
	Caches        []CacheSpec
	Compression   string
	SigningKey    string
	Workers       int
	UpstreamCache string // "none" disables upstream ref filtering
}

// LoadFile reads and parses a config file. A missing file is only an error when
// required is true (i.e. an explicit --config path); the optional default path
// returns an empty config with found=false.
func LoadFile(path string, required bool) (FileConfig, bool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) && !required {
			return FileConfig{}, false, nil
		}
		return FileConfig{}, false, err
	}
	var fc FileConfig
	if err := yaml.Unmarshal(data, &fc); err != nil {
		return FileConfig{}, false, fmt.Errorf("parsing %s: %w", path, err)
	}
	return fc, true, nil
}

// Resolve merges a file config with inline inputs (inputs override; list fields
// replace rather than append), applies defaults, validates, and parses caches.
func Resolve(fc FileConfig, in Inputs) (RunSpec, error) {
	spec := RunSpec{
		Builds:        fc.Builds,
		Compression:   fc.Compression,
		SigningKey:    fc.SigningKey,
		Workers:       fc.Workers,
		UpstreamCache: fc.UpstreamCache,
	}
	rawCaches := fc.Caches

	if len(in.Builds) > 0 {
		spec.Builds = in.Builds
	}
	if len(in.Caches) > 0 {
		rawCaches = in.Caches
	}
	if in.Compression != "" {
		spec.Compression = in.Compression
	}
	if in.SigningKey != "" {
		spec.SigningKey = in.SigningKey
	}
	if in.Workers != 0 {
		spec.Workers = in.Workers
	}
	if in.UpstreamCache != "" {
		spec.UpstreamCache = in.UpstreamCache
	}

	if spec.Compression == "" {
		spec.Compression = "zstd"
	}
	if spec.Workers == 0 {
		spec.Workers = 50
	}
	if spec.UpstreamCache == "" {
		spec.UpstreamCache = "https://cache.nixos.org"
	}

	if len(spec.Builds) == 0 {
		return RunSpec{}, fmt.Errorf("no builds configured: set 'builds' in the config file or pass --build")
	}
	if len(rawCaches) == 0 {
		return RunSpec{}, fmt.Errorf("no caches configured: set 'caches' in the config file or pass --cache")
	}
	for _, c := range rawCaches {
		cs, err := ParseCacheSpec(c)
		if err != nil {
			return RunSpec{}, err
		}
		spec.Caches = append(spec.Caches, cs)
	}
	return spec, nil
}
