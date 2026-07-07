package cmd

import (
	"fmt"
	"os"

	setup "github.com/itzemoji/aeroflare/internal/init"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// settingsCmd represents the `aeroflare settings` command.
// It provides a user-friendly, interactive terminal UI (using huh)
// for users to configure their themes, registry logins, and custom caching URLs.
var settingsCmd = &cobra.Command{
	Use:   "settings",
	Short: "Configure Aeroflare interactively",
	Run: func(cmd *cobra.Command, args []string) {
		// Define variables to hold the form state.
		var theme string
		var registryAction string
		var cloudflareToken string
		var gitProvider string
		var githubToken string
		var gitlabToken string
		var customRegistryURL string

		// Preload existing configuration from Viper to populate the form defaults.
		// If a user has previously configured Aeroflare, we want the interactive menu
		// to reflect their current choices.
		theme = viper.GetString("theme")
		if theme == "" {
			theme = "default"
		}

		cloudflareToken = viper.GetString("cloudflare-api-token")
		gitProvider = viper.GetString("git-provider")
		if gitProvider == "" {
			gitProvider = "none"
		}

		// Determine the currently active registry action based on the saved config.
		// This sets the default selection in the "Registry Login & Setup" dropdown.
		if gitProvider == "github" {
			githubToken = viper.GetString("git-token")
			registryAction = "github"
		} else if gitProvider == "gitlab" {
			gitlabToken = viper.GetString("git-token")
			registryAction = "gitlab"
		} else if viper.GetString("cache-url") != "" {
			registryAction = "custom"
			customRegistryURL = viper.GetString("cache-url")
		} else if cloudflareToken != "" {
			registryAction = "cloudflare"
		} else {
			registryAction = "none"
		}

		// Build and run the primary configuration form.
		// This form captures general appearance settings and determines
		// which authentication flow (if any) the user wants to proceed with.
		err := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Appearance Theme").
					Options(
						huh.NewOption("Catppuccin", "catppuccin"),
						huh.NewOption("Gruvbox Dark", "gruvbox-dark"),
						huh.NewOption("Gruvbox Light", "gruvbox-light"),
						huh.NewOption("Default Terminal", "default"),
					).
					Value(&theme),
			),
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Registry Login & Setup").
					Description("Configure authentication for a registry").
					Options(
						huh.NewOption("GitHub Packages (ghcr.io)", "github"),
						huh.NewOption("GitLab Registry", "gitlab"),
						huh.NewOption("Cloudflare (Workers)", "cloudflare"),
						huh.NewOption("Custom OCI Registry", "custom"),
						huh.NewOption("None", "none"),
					).
					Value(&registryAction),
			),
		).WithTheme(setup.AeroflareTheme()).Run()

		// If the user aborts the form (e.g. by pressing Ctrl+C), exit gracefully.
		if err != nil {
			PrintError("Settings cancelled")
			os.Exit(1)
		}

		// Based on the user's choice in the primary form, we dynamically construct
		// and present a secondary form to collect the necessary authentication credentials.
		var authGroups []*huh.Group
		switch registryAction {
		case "github":
			authGroups = append(authGroups, huh.NewGroup(
				huh.NewInput().
					Title("GitHub Personal Access Token").
					EchoMode(huh.EchoModePassword).
					Value(&githubToken),
			))
		case "gitlab":
			authGroups = append(authGroups, huh.NewGroup(
				huh.NewInput().
					Title("GitLab Personal Access Token").
					EchoMode(huh.EchoModePassword).
					Value(&gitlabToken),
			))
		case "cloudflare":
			authGroups = append(authGroups, huh.NewGroup(
				huh.NewInput().
					Title("Cloudflare API Token").
					EchoMode(huh.EchoModePassword).
					Value(&cloudflareToken),
			))
		case "custom":
			authGroups = append(authGroups, huh.NewGroup(
				huh.NewInput().
					Title("Registry URL (e.g., https://registry.example.com)").
					Value(&customRegistryURL),
			))
		}

		// If an authentication method was selected, run the secondary form.
		if len(authGroups) > 0 {
			err = huh.NewForm(authGroups...).WithTheme(setup.AeroflareTheme()).Run()
			if err != nil {
				PrintError("Settings cancelled")
				os.Exit(1)
			}
		}

		// Apply the captured settings back into the Viper configuration instance.
		viper.Set("theme", theme)

		switch registryAction {
		case "github":
			viper.Set("git-provider", "github")
			if githubToken != "" {
				viper.Set("git-token", githubToken)
			}
		case "gitlab":
			viper.Set("git-provider", "gitlab")
			if gitlabToken != "" {
				viper.Set("git-token", gitlabToken)
			}
		case "cloudflare":
			if cloudflareToken != "" {
				viper.Set("cloudflare-api-token", cloudflareToken)
			}
		case "custom":
			if customRegistryURL != "" {
				viper.Set("cache-url", customRegistryURL)
			}
		}

		// Finally, persist the updated configuration to the disk (aeroflare.yaml).
		err = viper.WriteConfig()
		if err != nil {
			PrintError(fmt.Sprintf("Failed to save settings: %v", err))
		} else {
			// Provide context-aware success messages.
			// IsNewConfig is determined during root.go's initConfig phase.
			if IsNewConfig {
				PrintSuccess(fmt.Sprintf("Initial config has been saved to %s", viper.ConfigFileUsed()))
			} else {
				PrintSuccess(fmt.Sprintf("Config has been updated in %s", viper.ConfigFileUsed()))
			}
		}
	},
}

func init() {
	// Register the settings command with the root cobra command tree.
	rootCmd.AddCommand(settingsCmd)
}
