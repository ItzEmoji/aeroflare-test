package ui

import (
	"fmt"
	"strings"
)

// ANSI color codes for terminal output.
const (
	colorGray  = "\x1b[90m" // Gray for borders and separators
	colorCyan  = "\x1b[36m" // Cyan for column headers
	colorReset = "\x1b[0m"  // Reset to default color
)

// PrintTable prints a formatted table with borders, headers, and rows.
// Columns are auto-sized to fit the widest content in each column.
func PrintTable(headers []string, rows [][]string) {
	if len(headers) == 0 {
		return
	}

	// Calculate the width needed for each column based on headers and data.
	columnWidths := make([]int, len(headers))
	for i, header := range headers {
		columnWidths[i] = len(header)
	}

	for _, row := range rows {
		for i, cell := range row {
			if i < len(columnWidths) && len(cell) > columnWidths[i] {
				columnWidths[i] = len(cell)
			}
		}
	}

	// drawLine constructs a horizontal line with the given junction characters
	// and fill character, properly sized for each column.
	drawLine := func(left, mid, right, fill string) string {
		columnSeparators := make([]string, len(columnWidths))
		for i, columnWidth := range columnWidths {
			columnSeparators[i] = strings.Repeat(fill, columnWidth+2)
		}
		return colorGray + left + strings.Join(columnSeparators, mid) + right + colorReset
	}

	// Top border
	fmt.Println("  " + drawLine("╭", "┬", "╮", "─"))

	// Print column headers with cyan highlighting
	fmt.Print("  " + colorGray + "│" + colorReset)
	for i, header := range headers {
		fmt.Printf(" " + colorCyan + "%-*s" + colorReset + " " + colorGray + "│" + colorReset, columnWidths[i], header)
	}
	fmt.Println()

	// Header separator
	fmt.Println("  " + drawLine("├", "┼", "┤", "─"))

	// Print table rows, padding cells to match column widths
	for _, row := range rows {
		fmt.Print("  " + colorGray + "│" + colorReset)
		for i, cell := range row {
			columnWidth := columnWidths[i]
			if i < len(columnWidths) {
				fmt.Printf(" %-*s " + colorGray + "│" + colorReset, columnWidth, cell)
			}
		}
		// Pad with empty cells if row has fewer columns than headers
		for i := len(row); i < len(headers); i++ {
			fmt.Printf(" %-*s " + colorGray + "│" + colorReset, columnWidths[i], "")
		}
		fmt.Println()
	}

	// Bottom border
	fmt.Println("  " + drawLine("╰", "┴", "╯", "─"))
}
