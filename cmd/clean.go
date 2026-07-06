package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"aeroflare/internal/cacheindex"
	"aeroflare/internal/oci"
	"aeroflare/internal/proxy"

	"github.com/spf13/cobra"
)

var cleanIndexCmd = &cobra.Command{
	Use:   "clean-index",
	Short: "Completely wipe the remote index on the registry",
	Run: func(cmd *cobra.Command, args []string) {
		registry, repository := oci.GetRegistryAndRepository()

		fmt.Print("Are you sure you want to completely wipe the remote index on the registry? [y/N]: ")
		var response string
		_, _ = fmt.Scanln(&response)
		response = strings.ToLower(strings.TrimSpace(response))
		if response != "y" && response != "yes" {
			PrintInfo("Aborted.")
			return
		}

		ociToken := oci.GetToken(registry, repository, "")
		if ociToken == "" {
			PrintError("oci_token, GITHUB_TOKEN or GH_TOKEN environment variable is required")
			os.Exit(1)
		}

		emptyIndex := &cacheindex.PushCacheIndex{
			Entries: make(map[string]cacheindex.PushCacheEntry),
			GCRoots: []string{},
		}

		tokenMgr := proxy.NewTokenManager(registry, repository, "")
		_, configAnnotations, _ := proxy.BootstrapConfigWithAnnotations(context.Background(), nil, registry, repository, tokenMgr)

		PrintInfo("Wiping remote index...")
		err := cacheindex.UpdateCacheIndex(nil, emptyIndex, registry, repository, ociToken, "", configAnnotations)
		if err != nil {
			PrintError(fmt.Sprintf("Failed to wipe remote index: %v", err))
			os.Exit(1)
		}
		PrintSuccess("Successfully wiped remote index.")
	},
}

func init() {
	rootCmd.AddCommand(cleanIndexCmd)
}
