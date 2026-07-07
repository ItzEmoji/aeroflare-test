package ui

import (
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
func PrintTable(headers []string, rows [][]string) error {
	return PrintTableTo(os.Stdout, headers, rows)
}

// PrintTableTo renders a formatted table with borders, headers, and rows and
// writes it to w. Columns are auto-sized to fit the widest content in each
// column. It returns any error encountered writing to w.
func PrintTableTo(w io.Writer, headers []string, rows [][]string) error {
	if len(headers) == 0 {
		return nil
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

	// padRight left-justifies s in a field of the given width, matching the
	// "%-*s" verb but without a fmt call so the render path can build the
	// output purely through strings.Builder (whose writes never fail).
	padRight := func(s string, width int) string {
		if pad := width - len(s); pad > 0 {
			return s + strings.Repeat(" ", pad)
		}
		return s
	}

	// Build the whole table in memory first so the writer is touched exactly
	// once; strings.Builder never fails, keeping the render loop error-free.
	var b strings.Builder

	// Top border
	b.WriteString("  " + drawLine("╭", "┬", "╮", "─") + "\n")

	// Print column headers with cyan highlighting
	b.WriteString("  " + colorGray + "│" + colorReset)
	for i, header := range headers {
		b.WriteString(" " + colorCyan + padRight(header, columnWidths[i]) + colorReset + " " + colorGray + "│" + colorReset)
	}
	b.WriteString("\n")

	// Header separator
	b.WriteString("  " + drawLine("├", "┼", "┤", "─") + "\n")

	// Print table rows, padding cells to match column widths
	for _, row := range rows {
		b.WriteString("  " + colorGray + "│" + colorReset)
		for i, cell := range row {
			if i < len(columnWidths) {
				b.WriteString(" " + padRight(cell, columnWidths[i]) + " " + colorGray + "│" + colorReset)
			}
		}
		// Pad with empty cells if row has fewer columns than headers
		for i := len(row); i < len(headers); i++ {
			b.WriteString(" " + padRight("", columnWidths[i]) + " " + colorGray + "│" + colorReset)
		}
		b.WriteString("\n")
	}

	// Bottom border
	b.WriteString("  " + drawLine("╰", "┴", "╯", "─") + "\n")

	_, err := io.WriteString(w, b.String())
	return err
}
