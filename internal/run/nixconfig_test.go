package run

import (
	"strings"
	"testing"
)

func TestBuildNixConfig_DoesNotAcceptFlakeConfig(t *testing.T) {
	cfg := buildNixConfig("", 8080)
	if strings.Contains(cfg, "accept-flake-config") {
		t.Fatal("run must not inject accept-flake-config; it silently trusts arbitrary flake-defined substituters and keys")
	}
	if !strings.Contains(cfg, "extra-substituters = http://127.0.0.1:8080") {
		t.Fatalf("missing substituter entry: %q", cfg)
	}
}

func TestBuildNixConfig_PreservesExistingConfig(t *testing.T) {
	cfg := buildNixConfig("keep-outputs = true", 9999)
	if !strings.Contains(cfg, "keep-outputs = true") {
		t.Fatalf("existing NIX_CONFIG lost: %q", cfg)
	}
	if !strings.Contains(cfg, "extra-substituters = http://127.0.0.1:9999") {
		t.Fatalf("missing substituter entry: %q", cfg)
	}
}
