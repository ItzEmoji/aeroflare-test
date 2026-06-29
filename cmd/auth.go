package cmd

import (
	"fmt"
	"strings"
	
	"aeroflare/src/secrets"
	"aeroflare/src/ui"
	"github.com/spf13/cobra"
)

// SecretsManager allows mocking in tests
var SecretsManager secrets.Manager

func getSecretsManager() secrets.Manager {
	if SecretsManager != nil {
		return SecretsManager
	}
	return secrets.NewManager()
}

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Manage Aeroflare authentication secrets",
}

var authLoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate interactively",
	RunE: func(cmd *cobra.Command, args []string) error {
		manager := getSecretsManager()
		
		tokens := []struct {
			key string
			val string
		}{
			{"github-token", globalGithubToken},
			{"gitlab-token", globalGitlabToken},
			{"cf-token", globalCfToken},
			{"cf-user-id", globalCfUserID},
		}

		savedAny := false
		for _, t := range tokens {
			if t.val != "" {
				err := manager.Set(t.key, t.val)
				if err != nil {
					PrintError(err.Error())
					return err
				}
				fmt.Printf("Saved %s\n", t.key)
				savedAny = true
			}
		}
		
		if !savedAny {
			runInteractiveAuth()
		}
		return nil
	},
}

var authListCmd = &cobra.Command{
	Use:   "list",
	Short: "List saved authentication credentials",
	RunE: func(cmd *cobra.Command, args []string) error {
		manager := getSecretsManager()
		keys, err := manager.List()
		if err != nil {
			PrintError(err.Error())
			return err
		}
		
		if len(keys) == 0 {
			fmt.Println("No credentials saved.")
			return nil
		}
		
		var rows [][]string
		for _, key := range keys {
			service := "Custom"
			info := "Secret"

			if key == "github-token" {
				service = "GitHub"
				info = "Token"
			} else if key == "gitlab-token" {
				service = "GitLab"
				info = "Token"
			} else if key == "cf-token" {
				service = "Cloudflare"
				info = "API Token"
			} else if key == "cf-user-id" {
				service = "Cloudflare"
				info = "Account ID"
			} else if strings.HasPrefix(key, "oci-") {
				service = "OCI Registry"
				parts := strings.Split(key, "-")
				if len(parts) >= 3 {
					// oci-<registry>-username or oci-<registry>-token
					registry := strings.Join(parts[1:len(parts)-1], "-")
					suffix := parts[len(parts)-1]
					if suffix == "username" {
						info = "Username (" + registry + ")"
					} else if suffix == "token" {
						info = "Token (" + registry + ")"
					}
				}
			}

			rows = append(rows, []string{service, info, key})
		}

		fmt.Println("Saved credentials:")
		ui.PrintTable([]string{"Service", "Info", "Key"}, rows)
		return nil
	},
}

var authRemoveCmd = &cobra.Command{
	Use:   "remove [key]",
	Short: "Remove a saved credential",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		manager := getSecretsManager()
		key := args[0]
		err := manager.Delete(key)
		if err != nil {
			PrintError(err.Error())
			return err
		}
		fmt.Printf("Removed %s\n", key)
		return nil
	},
}

var authSetCmd = &cobra.Command{
	Use:   "set [key] [value]",
	Short: "Set an arbitrary secret",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		manager := getSecretsManager()
		err := manager.Set(args[0], args[1])
		if err != nil {
			PrintError(err.Error())
			return err
		}
		fmt.Printf("Saved %s\n", args[0])
		return nil
	},
}

func init() {
	authCmd.AddCommand(authLoginCmd)
	authCmd.AddCommand(authListCmd)
	authCmd.AddCommand(authRemoveCmd)
	authCmd.AddCommand(authSetCmd)
	rootCmd.AddCommand(authCmd)
}
