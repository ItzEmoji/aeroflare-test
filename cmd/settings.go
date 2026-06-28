package cmd

import (
	"fmt"
	"os"

	"aeroflare/src/init"
	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var settingsCmd = &cobra.Command{
	Use:   "settings",
	Short: "Configure Aeroflare interactively",
	Run: func(cmd *cobra.Command, args []string) {
		var theme string
		var registryAction string
		var cloudflareToken string
		var gitProvider string
		var githubToken string
		var gitlabToken string
		var customRegistryURL string

		// Preload existing config
		theme = viper.GetString("theme")
		if theme == "" {
			theme = "default"
		}
		cloudflareToken = viper.GetString("cloudflare-api-token")
		gitProvider = viper.GetString("git-provider")
		if gitProvider == "" {
			gitProvider = "none"
		}

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
						huh.NewOption("Cloudflare R2", "cloudflare"),
						huh.NewOption("Custom OCI Registry", "custom"),
						huh.NewOption("None", "none"),
					).
					Value(&registryAction),
			),
		).WithTheme(setup.AeroflareTheme()).Run()

		if err != nil {
			PrintError("Settings cancelled")
			os.Exit(1)
		}

		// Show secondary forms based on registry action
		var authGroups []*huh.Group
		if registryAction == "github" {
			authGroups = append(authGroups, huh.NewGroup(
				huh.NewInput().
					Title("GitHub Personal Access Token").
					EchoMode(huh.EchoModePassword).
					Value(&githubToken),
			))
		} else if registryAction == "gitlab" {
			authGroups = append(authGroups, huh.NewGroup(
				huh.NewInput().
					Title("GitLab Personal Access Token").
					EchoMode(huh.EchoModePassword).
					Value(&gitlabToken),
			))
		} else if registryAction == "cloudflare" {
			authGroups = append(authGroups, huh.NewGroup(
				huh.NewInput().
					Title("Cloudflare API Token").
					EchoMode(huh.EchoModePassword).
					Value(&cloudflareToken),
			))
		} else if registryAction == "custom" {
			authGroups = append(authGroups, huh.NewGroup(
				huh.NewInput().
					Title("Registry URL (e.g., https://registry.example.com)").
					Value(&customRegistryURL),
			))
		}

		if len(authGroups) > 0 {
			err = huh.NewForm(authGroups...).WithTheme(setup.AeroflareTheme()).Run()
			if err != nil {
				PrintError("Settings cancelled")
				os.Exit(1)
			}
		}

		// Apply changes
		viper.Set("theme", theme)

		if registryAction == "github" {
			viper.Set("git-provider", "github")
			if githubToken != "" {
				viper.Set("git-token", githubToken)
			}
		} else if registryAction == "gitlab" {
			viper.Set("git-provider", "gitlab")
			if gitlabToken != "" {
				viper.Set("git-token", gitlabToken)
			}
		} else if registryAction == "cloudflare" {
			if cloudflareToken != "" {
				viper.Set("cloudflare-api-token", cloudflareToken)
			}
		} else if registryAction == "custom" {
			if customRegistryURL != "" {
				viper.Set("cache-url", customRegistryURL)
			}
		}

		err = viper.WriteConfig()
		if err != nil {
			PrintError(fmt.Sprintf("Failed to save settings: %v", err))
		} else {
			if IsNewConfig {
				PrintSuccess(fmt.Sprintf("Initial config has been saved to %s", viper.ConfigFileUsed()))
			} else {
				PrintSuccess(fmt.Sprintf("Config has been updated in %s", viper.ConfigFileUsed()))
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(settingsCmd)
}
