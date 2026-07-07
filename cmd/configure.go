package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/itzemoji/aeroflare/internal/oci"
	"github.com/itzemoji/aeroflare/internal/proxy"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

var configureCmd = &cobra.Command{
	Use:   "configure",
	Short: "Interactively configure cache settings",
	Run: func(cmd *cobra.Command, args []string) {
		registry, repository := oci.GetRegistryAndRepository()
		ociToken := oci.GetToken(registry, repository, "")
		if ociToken == "" {
			PrintError("Authentication token missing (oci_token, GITHUB_TOKEN or GH_TOKEN)")
			os.Exit(1)
		}

		// Fetch existing config manifest so we can prefill the public key.
		var existingPublicKey = ""

		token := getTokenForRegistry(registry)
		tokenMgr := proxy.NewTokenManager(registry, repository, token)
		remoteConf, existingAnnotations, _ := proxy.BootstrapConfigWithAnnotations(context.Background(), nil, registry, repository, tokenMgr)
		if existingAnnotations != nil {
			if pk := existingAnnotations["aeroflare.public-key"]; pk != "" {
				existingPublicKey = pk
			} else if pk := existingAnnotations["public-key"]; pk != "" {
				existingPublicKey = pk
			}
		} else if remoteConf != nil && remoteConf.PublicKey != "" {
			existingPublicKey = remoteConf.PublicKey
		}

		publicKey := existingPublicKey

		form := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("Public Key").
					Description("Enter your nix cache public key (optional)").
					Value(&publicKey),
			),
		)

		err := form.Run()
		if err != nil {
			if err.Error() != "user aborted" {
				PrintError(fmt.Sprintf("Form error: %v", err))
			}
			os.Exit(1)
		}

		annotations := map[string]string{
			"aeroflare.public-key": publicKey,
		}

		PrintInfo("Saving configuration to OCI manifest annotations...")

		err = oci.PushConfigManifest(registry, repository, ociToken, annotations)
		if err != nil {
			PrintError(fmt.Sprintf("Failed to save config: %v", err))
			os.Exit(1)
		}

		PrintSuccess("Configuration successfully saved to cache-config manifest!")
	},
}

func init() {
	rootCmd.AddCommand(configureCmd)
}
