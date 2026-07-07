package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/itzemoji/aeroflare/internal/auth"
	"github.com/itzemoji/aeroflare/internal/secrets"

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

// serviceFromArgs maps positional CLI args to a catalog service. The first arg
// is the service name; the special name "oci" consumes a second arg as the
// registry host. It returns the resolved service, the remaining positional
// args (field values for `set`, an optional field name for `get`), and an
// error naming the valid services if the name is not recognized.
func serviceFromArgs(args []string) (auth.Service, []string, error) {
	name := args[0]
	if name == "oci" {
		if len(args) < 2 || args[1] == "" {
			return auth.Service{}, nil, fmt.Errorf("usage: oci <host> [username] [token]")
		}
		return auth.ServiceForRegistry(args[1]), args[2:], nil
	}
	svc, ok := auth.ServiceByID(name)
	if !ok {
		return auth.Service{}, nil, fmt.Errorf("unknown service %q (valid: github, gitlab, cloudflare, or oci <host>)", name)
	}
	return svc, args[1:], nil
}

// redact masks a secret value for display, revealing only its last few
// characters so a credential can be recognized without being exposed.
func redact(val string) string {
	if len(val) <= 4 {
		return "****"
	}
	return "****" + val[len(val)-4:]
}

// validateService resolves a service's fields and runs its live validation
// check, returning the authenticated identity. It returns (nil, nil) when the
// service declares no validation.
func validateService(svc auth.Service, m secrets.Manager) (*auth.Identity, error) {
	if svc.Validate == nil {
		return nil, nil
	}
	vals, err := svc.Resolve(m)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return svc.Validate(ctx, vals)
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

func init() {
	authCmd.AddCommand(authLoginCmd)
	rootCmd.AddCommand(authCmd)
}
