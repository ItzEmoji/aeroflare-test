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
	authCmd.AddCommand(authSetCmd)
	rootCmd.AddCommand(authCmd)
}
