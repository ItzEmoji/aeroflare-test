package cmd

import (
	"github.com/itzemoji/aeroflare/internal/oci"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var VerboseCount int
var cacheURL string
var IsNewConfig bool

var globalGithubToken string
var globalGitlabToken string
var globalCfToken string
var globalCfUserID string

var rootCmd = &cobra.Command{
	Use:   "aeroflare",
	Short: "A high-performance OCI-backed Nix binary cache proxy and toolkit",
	Long: `A high-performance OCI-backed Nix binary cache proxy and toolkit.

Aeroflare allows you to seamlessly cache Nix binaries into an OCI registry
(like GitHub Packages), speeding up your CI/CD pipelines and local builds.
Use it as a proxy cache, or push/pull blobs directly to/from the registry.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		oci.DebugLogger = (VerboseCount >= 2)
		cacheURL = GetCacheURL()
	},
}

// initConfig locates (or creates) aeroflare.yaml under XDG_CONFIG_HOME (or
// ~/.config), wires it up as the Viper config file, and sets up the
// AEROFLARE_* environment variable prefix. Registered as a Cobra
// OnInitialize hook, so it runs once before any command's RunE/Run.
func initConfig() {
	configDir := os.Getenv("XDG_CONFIG_HOME")
	if configDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			homeDir = os.Getenv("HOME")
		}
		if homeDir == "" {
			PrintError("Could not determine home directory")
			os.Exit(1)
		}
		configDir = filepath.Join(homeDir, ".config")
	}
	aeroDir := filepath.Join(configDir, "aeroflare")

	if err := os.MkdirAll(aeroDir, 0755); err != nil {
		PrintError("Could not create config directory: " + err.Error())
	} else {
		configFile := filepath.Join(aeroDir, "aeroflare.yaml")
		if _, err := os.Stat(configFile); os.IsNotExist(err) {
			IsNewConfig = true
			defaultConfig := []byte(`# Aeroflare Configuration
# theme: catppuccin
# cache-url: oci://docker.io/my-org/my-cache
`)
			if err := os.WriteFile(configFile, defaultConfig, 0644); err != nil {
				PrintError("Could not write default config file: " + err.Error())
			}
		}
		viper.SetConfigFile(configFile)
	}
	viper.SetEnvPrefix("AEROFLARE")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()
	_ = viper.BindEnv("cache", "AEROFLARE_CACHE")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			PrintError("Error reading config file: " + err.Error())
		}
	}
}

// GetCacheURL resolves the effective OCI cache URL: an explicit --cache-url
// flag wins, otherwise a shorthand --cache "org/repo" value is expanded to
// a ghcr.io URL. Returns "" if neither is set.
func GetCacheURL() string {
	if url := viper.GetString("cache-url"); url != "" {
		return url
	}
	if cache := viper.GetString("cache"); cache != "" {
		return "ghcr.io/" + cache
	}
	return ""
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		PrintError(err.Error())
		os.Exit(1)
	}
}

// GetRootCmd returns the root Cobra command, mainly for use by the docs
// generator (cmd/gen_docs).
func GetRootCmd() *cobra.Command {
	return rootCmd
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().CountVarP(&VerboseCount, "verbose", "v", "Enable verbose output (-v for packages, -vv for requests)")
	rootCmd.PersistentFlags().StringVar(&cacheURL, "cache-url", "", "OCI registry URL for the cache")

	rootCmd.PersistentFlags().StringVar(&globalGithubToken, "github-token", "", "GitHub Token")
	rootCmd.PersistentFlags().StringVar(&globalGitlabToken, "gitlab-token", "", "GitLab Token")
	rootCmd.PersistentFlags().StringVar(&globalCfToken, "cf-token", "", "Cloudflare API Token")
	rootCmd.PersistentFlags().StringVar(&globalCfUserID, "cf-user-id", "", "Cloudflare Account ID")

	_ = viper.BindPFlag("cache-url", rootCmd.PersistentFlags().Lookup("cache-url"))
}
