package cmd

import (
	"fmt"
	"aeroflare/src/secrets"
	"github.com/spf13/cobra"
)

var (
	githubToken string
	cfToken     string
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Manage Aeroflare authentication secrets",
	Run: func(cmd *cobra.Command, args []string) {
		manager := secrets.NewManager()
		
		if githubToken != "" {
			manager.Set("github-token", githubToken)
			fmt.Println("Saved github-token")
		}
		
		if cfToken != "" {
			manager.Set("cf-token", cfToken)
			fmt.Println("Saved cf-token")
		}
		
		if githubToken == "" && cfToken == "" {
			fmt.Println("Interactive mode not fully implemented in CLI yet, please use flags.")
		}
	},
}

var authSetCmd = &cobra.Command{
	Use:   "set [key] [value]",
	Short: "Set an arbitrary secret",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		manager := secrets.NewManager()
		err := manager.Set(args[0], args[1])
		if err != nil {
			PrintError(err.Error())
			return
		}
		fmt.Printf("Saved %s\n", args[0])
	},
}

func init() {
	authCmd.Flags().StringVar(&githubToken, "github-token", "", "GitHub Token")
	authCmd.Flags().StringVar(&cfToken, "cf-token", "", "Cloudflare Token")
	
	authCmd.AddCommand(authSetCmd)
	rootCmd.AddCommand(authCmd)
}
