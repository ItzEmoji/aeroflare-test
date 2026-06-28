package cmd

import (
	"fmt"
	"aeroflare/src/secrets"
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
	Run: func(cmd *cobra.Command, args []string) {
		manager := getSecretsManager()
		
		savedAny := false
		if globalGithubToken != "" {
			err := manager.Set("github-token", globalGithubToken)
			if err != nil {
				PrintError(err.Error())
				return
			}
			fmt.Println("Saved github-token")
			savedAny = true
		}
		if globalGitlabToken != "" {
			err := manager.Set("gitlab-token", globalGitlabToken)
			if err != nil {
				PrintError(err.Error())
				return
			}
			fmt.Println("Saved gitlab-token")
			savedAny = true
		}
		if globalCfToken != "" {
			err := manager.Set("cf-token", globalCfToken)
			if err != nil {
				PrintError(err.Error())
				return
			}
			fmt.Println("Saved cf-token")
			savedAny = true
		}
		if globalCfUserID != "" {
			err := manager.Set("cf-user-id", globalCfUserID)
			if err != nil {
				PrintError(err.Error())
				return
			}
			fmt.Println("Saved cf-user-id")
			savedAny = true
		}
		
		if !savedAny {
			runInteractiveAuth()
		}
	},
}

var authSetCmd = &cobra.Command{
	Use:   "set [key] [value]",
	Short: "Set an arbitrary secret",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		manager := getSecretsManager()
		err := manager.Set(args[0], args[1])
		if err != nil {
			PrintError(err.Error())
			return
		}
		fmt.Printf("Saved %s\n", args[0])
	},
}

func init() {
	authCmd.AddCommand(authSetCmd)
	rootCmd.AddCommand(authCmd)
}
