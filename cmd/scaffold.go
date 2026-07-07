package cmd

import (
	"aeroflare/internal/oci"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

var scaffoldRelease string
var scaffoldOutputDir string

var scaffoldCmd = &cobra.Command{
	Use:   "scaffold",
	Short: "Generate local project files for an Aeroflare worker",
	Long: `Download an Aeroflare release and scaffold a local project directory
containing the worker source, wrangler.toml, and configuration files.

This command generates local files only — it does not create or modify
any remote infrastructure. Use 'aeroflare init' for that.`,
	Run: func(cmd *cobra.Command, args []string) {
		runScaffold()
	},
}

func init() {
	scaffoldCmd.Flags().StringVar(&scaffoldRelease, "release", "", "Release tag to download (default: prompt)")
	scaffoldCmd.Flags().StringVar(&scaffoldOutputDir, "output-dir", "./aeroflare-proxy", "Directory to scaffold into")
	rootCmd.AddCommand(scaffoldCmd)
}

func runScaffold() {
	// Fetch available releases.
	PrintInfo("Fetching available releases...")
	resp, err := http.Get("https://api.github.com/repos/ItzEmoji/aeroflare/releases")
	if err != nil {
		PrintError(fmt.Sprintf("Failed to fetch releases: %v", err))
		os.Exit(1)
	}
	defer func() { _ = resp.Body.Close() }()

	var releases []struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		PrintError(fmt.Sprintf("Failed to decode releases: %v", err))
		os.Exit(1)
	}
	if len(releases) == 0 {
		PrintError("No releases found.")
		os.Exit(1)
	}

	releaseTag := scaffoldRelease
	if releaseTag == "" {
		var options []huh.Option[string]
		for _, r := range releases {
			options = append(options, huh.NewOption(r.TagName, r.TagName))
		}

		err = huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Which release of Aeroflare do you want to use?").
					Options(options...).
					Value(&releaseTag),
			),
		).Run()
		if err != nil {
			os.Exit(0)
		}
	}

	targetDir := scaffoldOutputDir
	if targetDir == "" {
		targetDir = "./aeroflare-proxy"
	}

	// Prompt for extraction directory.
	err = huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Output directory").
				Description("Where should the project files be extracted?").
				Value(&targetDir),
		),
	).Run()
	if err != nil {
		os.Exit(0)
	}

	if targetDir == "" {
		targetDir = "./aeroflare-proxy"
	}
	_ = os.MkdirAll(targetDir, 0755)

	// Download and extract.
	PrintInfo(fmt.Sprintf("Downloading source for release %s...", releaseTag))
	tarURL := fmt.Sprintf("https://github.com/ItzEmoji/aeroflare/archive/refs/tags/%s.tar.gz", releaseTag)

	tarResp, err := http.Get(tarURL)
	if err != nil {
		PrintError(fmt.Sprintf("Failed to download source: %v", err))
		os.Exit(1)
	}
	defer func() { _ = tarResp.Body.Close() }()
	if tarResp.StatusCode < 200 || tarResp.StatusCode >= 300 {
		PrintError(fmt.Sprintf("Failed to download source: GitHub returned %s", tarResp.Status))
		os.Exit(1)
	}

	downloadCmd := exec.Command("tar", "-xz", "-C", targetDir, "--strip-components=1")
	downloadCmd.Stdin = tarResp.Body
	downloadCmd.Stdout = os.Stdout
	downloadCmd.Stderr = os.Stderr
	if err := downloadCmd.Run(); err != nil {
		PrintError(fmt.Sprintf("Failed to download or extract source: %v", err))
		os.Exit(1)
	}

	proxyDir := fmt.Sprintf("%s/proxy/no-webui-native", targetDir)
	if _, err := os.Stat(proxyDir); os.IsNotExist(err) {
		PrintError(fmt.Sprintf("Proxy directory %s not found in the release", proxyDir))
		os.Exit(1)
	}

	// Patch wrangler.toml with environment values if available.
	patchWranglerToml(proxyDir)

	PrintSuccess(fmt.Sprintf("Project scaffolded at %s", proxyDir))
	PrintInfo("You can now customize the worker and deploy with 'aeroflare init' or 'npx wrangler deploy'.")
}

// patchWranglerToml applies configuration values from environment variables
// to the scaffolded wrangler.toml template.
func patchWranglerToml(proxyDir string) {
	registry, repository := "", ""
	if r := os.Getenv("AEROFLARE_REGISTRY"); r != "" {
		registry = r
	} else if r := os.Getenv("NIXCACHE_REGISTRY"); r != "" {
		registry = r
	}

	if c := os.Getenv("AEROFLARE_CACHE"); c != "" {
		repository = c
	} else if c := os.Getenv("NIXCACHE_REPO"); c != "" {
		repository = c
	}

	// Try from the network package if env vars are set.
	if registry == "" || repository == "" {
		r, repo := oci.GetRegistryAndRepository()
		if registry == "" {
			registry = r
		}
		if repository == "" {
			repository = repo
		}
	}

	wranglerPath := fmt.Sprintf("%s/wrangler.toml", proxyDir)
	content, err := os.ReadFile(wranglerPath)
	if err != nil {
		PrintWarning(fmt.Sprintf("Could not read wrangler.toml: %v", err))
		return
	}

	// Set the two vars the worker reads, replacing either a commented placeholder
	// or a concrete default line so it works whatever the template ships with.
	s := string(content)
	if repository != "" {
		s = setWranglerVar(s, "NIXCACHE_REPO", repository)
	}
	if registry != "" {
		s = setWranglerVar(s, "NIXCACHE_REGISTRY_URL", workerRegistryURL(registry))
	}

	_ = os.WriteFile(wranglerPath, []byte(s), 0644)
}

// setWranglerVar sets `key = "value"` in a wrangler.toml body, replacing an
// existing (possibly commented) assignment for key if present, else appending.
func setWranglerVar(body, key, value string) string {
	line := fmt.Sprintf(`%s = "%s"`, key, value)
	re := regexp.MustCompile(`(?m)^#?\s*` + regexp.QuoteMeta(key) + `\s*=.*$`)
	if re.MatchString(body) {
		return re.ReplaceAllString(body, line)
	}
	return strings.TrimRight(body, "\n") + "\n" + line + "\n"
}

// workerRegistryURL turns a registry host (e.g. "ghcr.io") into the base URL the
// worker expects for NIXCACHE_REGISTRY_URL (e.g. "https://ghcr.io"); the worker
// appends the spec's "/v2" prefix itself.
func workerRegistryURL(registry string) string {
	if strings.HasPrefix(registry, "http://") || strings.HasPrefix(registry, "https://") {
		return strings.TrimRight(registry, "/")
	}
	return "https://" + strings.TrimRight(registry, "/")
}
