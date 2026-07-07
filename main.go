package main

import (
	"github.com/itzemoji/aeroflare/cmd"
	_ "embed"
)

//go:embed version.json
var versionJSON []byte

func main() {
	cmd.VersionJSON = versionJSON
	cmd.Execute()
}
