package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var authSetCmd = &cobra.Command{
	Use:   "set <service> [value...]",
	Short: "Save a credential for a known service (github, gitlab, cloudflare, oci <host>)",
	Long: `Save a credential for a service aeroflare understands.

Examples:
  aeroflare auth set github <token>
  aeroflare auth set cloudflare <api-token> <account-id>
  aeroflare auth set oci registry.example.com <username> <token>

With no values, you are prompted for each field (requires a terminal).`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		manager := getSecretsManager()

		svc, values, err := serviceFromArgs(args)
		if err != nil {
			PrintError(err.Error())
			return err
		}

		if len(values) > len(svc.Fields) {
			err := fmt.Errorf("%s takes at most %d value(s), got %d", svc.DisplayName, len(svc.Fields), len(values))
			PrintError(err.Error())
			return err
		}

		// With no positional values, fall back to interactive prompting.
		if len(values) == 0 {
			if !isTerminal() {
				err := fmt.Errorf("no values provided and not running interactively; pass values as arguments")
				PrintError(err.Error())
				return err
			}
			prompted := promptServiceFields(svc)
			for _, f := range svc.Fields {
				if v, ok := prompted[f.Name]; ok && v != "" {
					if err := manager.Set(f.SecretKey, v); err != nil {
						PrintError(err.Error())
						return err
					}
					fmt.Printf("Saved %s %s\n", svc.DisplayName, f.Name)
				}
			}
			return nil
		}

		// Positional values map onto fields in declared order.
		for i, v := range values {
			f := svc.Fields[i]
			if err := manager.Set(f.SecretKey, v); err != nil {
				PrintError(err.Error())
				return err
			}
			fmt.Printf("Saved %s %s\n", svc.DisplayName, f.Name)
		}

		// Live-validate on save only when interactive, so scripted/test runs
		// stay offline and fast.
		if isTerminal() {
			if id, err := validateService(svc, manager); err == nil && id != nil {
				printIdentity(svc, id)
			}
		}
		return nil
	},
}

func init() {
	authCmd.AddCommand(authSetCmd)
}
