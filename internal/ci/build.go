package ci

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

// scrapeStorePaths extracts /nix/store/... paths from command stdout: trimmed,
// non-empty, non-comment lines beginning with /nix/store/.
func scrapeStorePaths(stdout string) []string {
	var paths []string
	for _, line := range strings.Split(stdout, "\n") {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "#") && strings.HasPrefix(line, "/nix/store/") {
			paths = append(paths, line)
		}
	}
	return paths
}

// BuildInstallable runs `nix build <installable> --print-out-paths` and returns
// the resulting store paths. Build logs stream to the process's stderr/stdout.
func BuildInstallable(installable string) ([]string, error) {
	var stdout bytes.Buffer
	cmd := exec.Command("nix", "build", installable, "--print-out-paths")
	cmd.Stdout = io.MultiWriter(os.Stdout, &stdout)
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("nix build %s: %w", installable, err)
	}
	paths := scrapeStorePaths(stdout.String())
	if len(paths) == 0 {
		return nil, fmt.Errorf("nix build %s produced no store paths", installable)
	}
	return paths, nil
}
