package cmd

import (
	"encoding/json"
	"fmt"
	"sort"

	"aeroflare/internal/auth"
	"aeroflare/internal/secrets"
	"aeroflare/internal/ui"

	"github.com/spf13/cobra"
)

var (
	authStatusJSON     bool
	authStatusNoVerify bool
)

// statusFieldJSON / statusEntryJSON are the JSON shapes emitted by
// `auth status --json`.
type statusFieldJSON struct {
	Name   string `json:"name"`
	Set    bool   `json:"set"`
	Secret bool   `json:"secret"`
	Value  string `json:"value,omitempty"`
}

type statusEntryJSON struct {
	Service    string            `json:"service"`
	ID         string            `json:"id"`
	Fields     []statusFieldJSON `json:"fields"`
	Verifiable bool              `json:"verifiable"`
	User       string            `json:"user,omitempty"`
	Warnings   []string          `json:"warnings,omitempty"`
	Error      string            `json:"error,omitempty"`
}

// storedServices returns the services that have at least one field stored, in
// stable order (fixed services first, then OCI registries sorted by host),
// along with any leftover keys that belong to no known service.
func storedServices(m secrets.Manager, keys []string) (services []auth.Service, custom []string) {
	stored := make(map[string]bool, len(keys))
	for _, k := range keys {
		stored[k] = true
	}

	seen := make(map[string]bool)
	add := func(svc auth.Service) {
		if seen[svc.ID] {
			return
		}
		for _, f := range svc.Fields {
			if stored[f.SecretKey] {
				seen[svc.ID] = true
				services = append(services, svc)
				return
			}
		}
	}

	for _, svc := range auth.Services() {
		add(svc)
	}

	// Discover OCI registry hosts from the stored keys, then add each once.
	var ociHosts []string
	ociSeen := make(map[string]bool)
	for _, k := range keys {
		svc, ok := auth.ServiceForSecretKey(k)
		if !ok {
			custom = append(custom, k)
			continue
		}
		if _, isFixed := auth.ServiceByID(svc.ID); isFixed {
			continue
		}
		if !ociSeen[svc.ID] {
			ociSeen[svc.ID] = true
			ociHosts = append(ociHosts, k)
		}
	}
	sort.Strings(custom)

	var ociServices []auth.Service
	for _, k := range ociHosts {
		if svc, ok := auth.ServiceForSecretKey(k); ok {
			ociServices = append(ociServices, svc)
		}
	}
	sort.Slice(ociServices, func(i, j int) bool { return ociServices[i].ID < ociServices[j].ID })
	for _, svc := range ociServices {
		add(svc)
	}

	return services, custom
}

var authStatusCmd = &cobra.Command{
	Use:     "status",
	Aliases: []string{"list"},
	Short:   "Show stored credentials and their validity",
	RunE: func(cmd *cobra.Command, args []string) error {
		manager := getSecretsManager()
		keys, err := manager.List()
		if err != nil {
			PrintError(err.Error())
			return err
		}

		services, custom := storedServices(manager, keys)

		var entries []statusEntryJSON
		for _, svc := range services {
			entry := statusEntryJSON{Service: svc.DisplayName, ID: svc.ID, Verifiable: svc.Validate != nil}
			for _, f := range svc.Fields {
				val, gerr := manager.Get(f.SecretKey)
				set := gerr == nil && val != ""
				fj := statusFieldJSON{Name: f.Name, Set: set, Secret: f.Secret}
				if set {
					if f.Secret {
						fj.Value = redact(val)
					} else {
						fj.Value = val
					}
				}
				entry.Fields = append(entry.Fields, fj)
			}

			if !authStatusNoVerify {
				if id, verr := validateService(svc, manager); verr != nil {
					entry.Error = verr.Error()
				} else if id != nil {
					entry.User = id.User
					entry.Warnings = id.Warnings
				}
			}
			entries = append(entries, entry)
		}

		if authStatusJSON {
			enc := json.NewEncoder(cmd.OutOrStdout())
			enc.SetIndent("", "  ")
			return enc.Encode(entries)
		}

		out := cmd.OutOrStdout()
		if len(entries) == 0 && len(custom) == 0 {
			fmt.Fprintln(out, "No credentials saved.")
			return nil
		}

		// Build one row per field. The Service, ID, and Status columns are
		// filled on the first row of each service and left blank on its
		// remaining field rows, so each credential reads as a visual block.
		var rows [][]string
		var footnotes []string
		for _, e := range entries {
			for i, f := range e.Fields {
				service, id, status := "", "", ""
				if i == 0 {
					service = e.Service
					id = e.ID
					status = statusCell(e)
				}
				value := "not set"
				if f.Set {
					value = f.Value
				}
				rows = append(rows, []string{service, id, f.Name, value, status})
			}
			for _, w := range e.Warnings {
				footnotes = append(footnotes, fmt.Sprintf("%s: %s", e.Service, w))
			}
		}

		for _, k := range custom {
			rows = append(rows, []string{"Custom", "-", k, "(hidden)", "-"})
		}

		ui.PrintTableTo(out, []string{"Service", "ID", "Field", "Value", "Status"}, rows)

		if len(footnotes) > 0 {
			fmt.Fprintln(out)
			fmt.Fprintln(out, "Warnings:")
			for _, note := range footnotes {
				fmt.Fprintf(out, "  • %s\n", note)
			}
		}
		return nil
	},
}

// statusCell renders the Status column for a service entry: whether it was
// verified, and if so, who it authenticated as or why it failed.
func statusCell(e statusEntryJSON) string {
	if authStatusNoVerify {
		return "-"
	}
	if !e.Verifiable {
		return "n/a"
	}
	if e.Error != "" {
		return "invalid"
	}
	status := "valid"
	if e.User != "" {
		status = "valid (" + e.User + ")"
	}
	if len(e.Warnings) > 0 {
		status += " ⚠"
	}
	return status
}

// printIdentity renders a validated identity for a service to stdout, used
// after an interactive `auth set` to confirm the credential works.
func printIdentity(svc auth.Service, id *auth.Identity) {
	if id.User != "" {
		fmt.Printf("✓ %s authenticated as %s\n", svc.DisplayName, id.User)
	}
	for _, w := range id.Warnings {
		fmt.Printf("⚠️  %s\n", w)
	}
}

func init() {
	authStatusCmd.Flags().BoolVar(&authStatusJSON, "json", false, "Output as JSON")
	authStatusCmd.Flags().BoolVar(&authStatusNoVerify, "no-verify", false, "Skip live validation of credentials")
	authCmd.AddCommand(authStatusCmd)
}
