package main

import (
	"aeroflare/cmd"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra/doc"
)

func main() {
	outDir := "./docs-site/docs/reference/cli"
	
	// Create directory if it doesn't exist
	if err := os.MkdirAll(outDir, 0755); err != nil {
		fmt.Printf("Error creating directory: %v\n", err)
		os.Exit(1)
	}

	// Generate markdown documentation tree
	rootCmd := cmd.GetRootCmd()
	err := doc.GenMarkdownTree(rootCmd, outDir)
	if err != nil {
		fmt.Printf("Error generating markdown docs: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Successfully generated Cobra CLI markdown documentation in docs-site/docs/reference/cli!")

	// Create _category_.json for the CLI reference directory
	categoryJSON := `{
  "label": "CLI Reference (Auto-generated)",
  "position": 1,
  "link": {
    "type": "generated-index",
    "description": "Auto-generated reference documentation for Aeroflare CLI commands and flags."
  }
}
`
	err = os.WriteFile(filepath.Join(outDir, "_category_.json"), []byte(categoryJSON), 0644)
	if err != nil {
		fmt.Printf("Error writing _category_.json: %v\n", err)
		os.Exit(1)
	}
}
