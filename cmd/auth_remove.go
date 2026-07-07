package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var authRemoveCmd = &cobra.Command{
	Use:   "remove <service>",
	Short: "Remove a stored credential (github, gitlab, cloudflare, oci <host>)",
	Long: `Delete every field of a service's credential from the secrets store.

Examples:
  aeroflare auth remove github
  aeroflare auth remove oci registry.example.com`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		manager := getSecretsManager()

		svc, _, err := serviceFromArgs(args)
		if err != nil {
			PrintError(err.Error())
			return err
		}

		removed := 0
		for _, f := range svc.Fields {
			// Only delete fields that are actually stored, so the summary count
			// is accurate.
			if _, gerr := manager.Get(f.SecretKey); gerr != nil {
				continue
			}
			if err := manager.Delete(f.SecretKey); err != nil {
				PrintError(err.Error())
				return err
			}
			removed++
		}

		if removed == 0 {
			fmt.Printf("No stored credentials for %s.\n", svc.DisplayName)
		} else {
			fmt.Printf("Removed %s credentials.\n", svc.DisplayName)
		}
		return nil
	},
}

func init() {
	authCmd.AddCommand(authRemoveCmd)
}
