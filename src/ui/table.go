package ui

import (
	"fmt"
	"strings"
)

// ANSI color codes
const (
	colorGray  = "\x1b[90m"
	colorCyan  = "\x1b[36m"
	colorReset = "\x1b[0m"
)

// PrintTable prints a beautiful Nushell-style table.
func PrintTable(headers []string, rows [][]string) {
	if len(headers) == 0 {
		return
	}

	// Calculate column widths
	colWidths := make([]int, len(headers))
	for i, h := range headers {
		colWidths[i] = len(h)
	}

	for _, row := range rows {
		for i, cell := range row {
			if i < len(colWidths) && len(cell) > colWidths[i] {
				colWidths[i] = len(cell)
			}
		}
	}

	// Helper to draw horizontal lines
	drawLine := func(left, mid, right, fill string) string {
		parts := make([]string, len(colWidths))
		for i, w := range colWidths {
			parts[i] = strings.Repeat(fill, w+2)
		}
		return colorGray + left + strings.Join(parts, mid) + right + colorReset
	}

	// Top border
	fmt.Println("  " + drawLine("╭", "┬", "╮", "─"))

	// Headers
	fmt.Print("  " + colorGray + "│" + colorReset)
	for i, h := range headers {
		fmt.Printf(" " + colorCyan + "%-*s" + colorReset + " " + colorGray + "│" + colorReset, colWidths[i], h)
	}
	fmt.Println()

	// Header separator
	fmt.Println("  " + drawLine("├", "┼", "┤", "─"))

	// Rows
	for _, row := range rows {
		fmt.Print("  " + colorGray + "│" + colorReset)
		for i, cell := range row {
			w := colWidths[i]
			if i < len(colWidths) {
				fmt.Printf(" %-*s " + colorGray + "│" + colorReset, w, cell)
			}
		}
		// In case row has fewer elements than headers
		for i := len(row); i < len(headers); i++ {
			fmt.Printf(" %-*s " + colorGray + "│" + colorReset, colWidths[i], "")
		}
		fmt.Println()
	}

	// Bottom border
	fmt.Println("  " + drawLine("╰", "┴", "╯", "─"))
}
