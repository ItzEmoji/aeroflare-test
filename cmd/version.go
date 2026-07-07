package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

// VersionJSON holds the release-please "manifest" JSON (a map of component
// path to version, keyed "." for the root component). main.go populates it
// from a //go:embed'd version.json before calling Execute; it's a package
// var here so main.go can inject it without an import cycle.
var VersionJSON []byte

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of aeroflare",
	Run: func(cmd *cobra.Command, args []string) {
		var v map[string]string
		if len(VersionJSON) > 0 {
			if err := json.Unmarshal(VersionJSON, &v); err == nil {
				if ver, ok := v["."]; ok {
					fmt.Printf("aeroflare version %s\n", ver)
					return
				}
			}
		}
		fmt.Println("aeroflare version unknown")
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
