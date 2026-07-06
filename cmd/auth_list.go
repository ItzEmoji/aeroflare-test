package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"aeroflare/internal/ui"

	"github.com/spf13/cobra"
)

var authListJSON bool

var authListCmd = &cobra.Command{
	Use:   "list",
	Short: "List saved authentication credentials",
	RunE: func(cmd *cobra.Command, args []string) error {
		manager := getSecretsManager()
		keys, err := manager.List()
		if err != nil {
			if authListJSON {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			} else {
				PrintError(err.Error())
			}
			return err
		}

		if len(keys) == 0 {
			if authListJSON {
				fmt.Println("[]")
			} else {
				fmt.Println("No credentials saved.")
			}
			return nil
		}

		type Entry struct {
			Service string `json:"service"`
			Info    string `json:"info"`
			Key     string `json:"key"`
		}
		var entries []Entry
		// OCI credentials are stored as two separate keys per registry
		// ("oci-<registry>-username" and "oci-<registry>-token"); collect them
		// here keyed by registry so they can be merged into a single row below.
		ociRegistries := make(map[string]map[string]string)

		for _, key := range keys {
			val, _ := manager.Get(key)

			if key == "github-token" {
				info := "Token"
				if user, scopes := getGithubUser(val); user != "" {
					info = "Username: " + user
					hasWritePackages := false
					hasWorkflow := false
					for _, s := range scopes {
						if s == "write:packages" {
							hasWritePackages = true
						}
						if s == "workflow" {
							hasWorkflow = true
						}
					}
					var warnings []string
					if !hasWritePackages {
						warnings = append(warnings, "write:packages")
					}
					if !hasWorkflow {
						warnings = append(warnings, "workflow")
					}
					if len(warnings) > 0 {
						info += fmt.Sprintf(" (⚠️ Missing scopes: %s)", strings.Join(warnings, ", "))
					}
				}
				entries = append(entries, Entry{Service: "GitHub", Info: info, Key: key})
			} else if key == "gitlab-token" {
				info := "Token"
				if user := getGitlabUser(val); user != "" {
					info = "Username: " + user
				}
				entries = append(entries, Entry{Service: "GitLab", Info: info, Key: key})
			} else if key == "cf-token" {
				entries = append(entries, Entry{Service: "Cloudflare", Info: "API Token", Key: key})
			} else if key == "cf-user-id" {
				entries = append(entries, Entry{Service: "Cloudflare", Info: "Account ID", Key: key})
			} else if strings.HasPrefix(key, "oci-") {
				// "oci-<registry>-<suffix>": the registry itself may contain
				// hyphens (e.g. "oci-registry.gitlab.com-token"), so split off
				// only the last "-"-delimited segment as the suffix.
				parts := strings.Split(key, "-")
				if len(parts) >= 3 {
					registry := strings.Join(parts[1:len(parts)-1], "-")
					suffix := parts[len(parts)-1]
					if _, ok := ociRegistries[registry]; !ok {
						ociRegistries[registry] = make(map[string]string)
					}
					ociRegistries[registry][suffix] = key
				} else {
					entries = append(entries, Entry{Service: "Custom", Info: "Secret", Key: key})
				}
			} else {
				entries = append(entries, Entry{Service: "Custom", Info: "Secret", Key: key})
			}
		}

		for reg, data := range ociRegistries {
			info := "Unknown"
			keyStr := ""

			if userKey, ok := data["username"]; ok {
				user, _ := manager.Get(userKey)
				info = "Username: " + user + " (" + reg + ")"
				keyStr = userKey
			}
			if tokenKey, ok := data["token"]; ok {
				if keyStr != "" {
					keyStr += ", " + tokenKey
				} else {
					info = "Token (" + reg + ")"
					keyStr = tokenKey
				}
			}

			entries = append(entries, Entry{Service: "OCI Registry", Info: info, Key: keyStr})
		}

		if authListJSON {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			if err := enc.Encode(entries); err != nil {
				return err
			}
			fmt.Fprintln(os.Stderr, "✅ The list has been successfully exported to json.")
			return nil
		}

		var rows [][]string
		for _, e := range entries {
			rows = append(rows, []string{e.Service, e.Info, e.Key})
		}

		fmt.Println("Saved credentials:")
		ui.PrintTable([]string{"Service", "Info", "Key(s)"}, rows)
		return nil
	},
}

func init() {
	authListCmd.Flags().BoolVar(&authListJSON, "json", false, "Export credentials list as JSON")
}

func getGithubUser(token string) (string, []string) {
	client := &http.Client{Timeout: 2 * time.Second}
	req, err := http.NewRequest("GET", "https://api.github.com/user", nil)
	if err != nil {
		return "", nil
	}
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != 200 {
		return "", nil
	}
	defer func() { _ = resp.Body.Close() }()

	scopesStr := resp.Header.Get("X-OAuth-Scopes")
	var scopes []string
	for _, s := range strings.Split(scopesStr, ",") {
		s = strings.TrimSpace(s)
		if s != "" {
			scopes = append(scopes, s)
		}
	}

	var data struct {
		Login string `json:"login"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", nil
	}
	return data.Login, scopes
}

func getGitlabUser(token string) string {
	client := &http.Client{Timeout: 2 * time.Second}
	req, err := http.NewRequest("GET", "https://gitlab.com/api/v4/user", nil)
	if err != nil {
		return ""
	}
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != 200 {
		return ""
	}
	defer func() { _ = resp.Body.Close() }()

	var data struct {
		Username string `json:"username"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return ""
	}
	return data.Username
}
