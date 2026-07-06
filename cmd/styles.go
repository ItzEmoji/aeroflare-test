package cmd

import (
	"fmt"
	"os"
)

// PrintError prints msg to stderr, prefixed with "Error: ".
func PrintError(msg string) {
	fmt.Fprintf(os.Stderr, "Error: %s\n", msg)
}

// PrintSuccess prints msg to stdout as-is. It's a distinct function from
// PrintInfo mainly so call sites read as intent ("this succeeded" vs.
// "just letting you know"), even though the formatting is currently identical.
func PrintSuccess(msg string) {
	fmt.Println(msg)
}

// PrintInfo prints msg to stdout as-is.
func PrintInfo(msg string) {
	fmt.Println(msg)
}

// PrintWarning prints msg to stdout, prefixed with "Warning: ".
func PrintWarning(msg string) {
	fmt.Println("Warning: " + msg)
}
