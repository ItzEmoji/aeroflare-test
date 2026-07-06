package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var authGetCmd = &cobra.Command{
	Use:   "get <service> [field]",
	Short: "Print a stored credential value (for scripting)",
	Long: `Print the raw stored value of a credential to stdout, for use in scripts.

Examples:
  aeroflare auth get github
  aeroflare auth get cloudflare account_id
  aeroflare auth get oci registry.example.com token

For a multi-field service with no field given, each field is printed as
"name=value".`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		manager := getSecretsManager()

		svc, rest, err := serviceFromArgs(args)
		if err != nil {
			PrintError(err.Error())
			return err
		}

		// A single trailing positional selects one field by name.
		if len(rest) == 1 {
			field, ok := svc.Field(rest[0])
			if !ok {
				err := fmt.Errorf("service %s has no field %q", svc.DisplayName, rest[0])
				PrintError(err.Error())
				return err
			}
			val, err := field.Resolve(manager)
			if err != nil {
				PrintError(fmt.Sprintf("no value found for %s %s", svc.DisplayName, field.Name))
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), val)
			return nil
		}
		if len(rest) > 1 {
			err := fmt.Errorf("get takes at most one field name")
			PrintError(err.Error())
			return err
		}

		vals, err := svc.Resolve(manager)
		if err != nil {
			PrintError(err.Error())
			return err
		}
		if len(vals) == 0 {
			err := fmt.Errorf("no credentials stored for %s", svc.DisplayName)
			PrintError(err.Error())
			return err
		}

		// Single-field service: print the bare value so it can be piped.
		if len(svc.Fields) == 1 {
			fmt.Fprintln(cmd.OutOrStdout(), vals[svc.Fields[0].Name])
			return nil
		}

		// Multi-field service: print each field as name=value, in declared order.
		for _, f := range svc.Fields {
			if v, ok := vals[f.Name]; ok {
				fmt.Fprintf(cmd.OutOrStdout(), "%s=%s\n", f.Name, v)
			}
		}
		return nil
	},
}

func init() {
	authCmd.AddCommand(authGetCmd)
}
