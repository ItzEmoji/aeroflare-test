package ui

import (
	"fmt"
	"io"
	"os"
	"strings"
)

// ANSI color codes for terminal output.
const (
	colorGray  = "\x1b[90m" // Gray for borders and separators
	colorCyan  = "\x1b[36m" // Cyan for column headers
	colorReset = "\x1b[0m"  // Reset to default color
)

// PrintTable prints a formatted table with borders, headers, and rows to
// stdout. Columns are auto-sized to fit the widest content in each column.
func PrintTable(headers []string, rows [][]string) {
	PrintTableTo(os.Stdout, headers, rows)
}

// PrintTableTo prints a formatted table with borders, headers, and rows to w.
// Columns are auto-sized to fit the widest content in each column.
func PrintTableTo(w io.Writer, headers []string, rows [][]string) {
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
	fmt.Fprintln(w, "  "+drawLine("╭", "┬", "╮", "─"))

	// Print column headers with cyan highlighting
	fmt.Fprint(w, "  "+colorGray+"│"+colorReset)
	for i, header := range headers {
		fmt.Fprintf(w, " "+colorCyan+"%-*s"+colorReset+" "+colorGray+"│"+colorReset, columnWidths[i], header)
	}
	fmt.Fprintln(w)

	// Header separator
	fmt.Fprintln(w, "  "+drawLine("├", "┼", "┤", "─"))

	// Print table rows, padding cells to match column widths
	for _, row := range rows {
		fmt.Fprint(w, "  "+colorGray+"│"+colorReset)
		for i, cell := range row {
			columnWidth := columnWidths[i]
			if i < len(columnWidths) {
				fmt.Fprintf(w, " %-*s "+colorGray+"│"+colorReset, columnWidth, cell)
			}
		}
		// Pad with empty cells if row has fewer columns than headers
		for i := len(row); i < len(headers); i++ {
			fmt.Fprintf(w, " %-*s "+colorGray+"│"+colorReset, columnWidths[i], "")
		}
		fmt.Fprintln(w)
	}

	// Bottom border
	fmt.Fprintln(w, "  "+drawLine("╰", "┴", "╯", "─"))
}
