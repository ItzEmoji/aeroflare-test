package ui

import (
	"fmt"
	"strings"
)

// BoxField represents a single label-value pair to be displayed in a box.
type BoxField struct {
	Label string
	Value string
}

// PrintSummaryBox prints a formatted box with a title and label-value pairs.
// The box is auto-sized to fit the title and all fields.
func PrintSummaryBox(title string, fields []BoxField) {
	maxLabelLen := 0
	for _, field := range fields {
		if len(field.Label) > maxLabelLen {
			maxLabelLen = len(field.Label)
		}
	}

	maxContentWidth := len(title)
	for _, field := range fields {
		// Width is: label + colon + space + value
		fieldLineWidth := maxLabelLen + 1 + 1 + len(field.Value)
		if fieldLineWidth > maxContentWidth {
			maxContentWidth = fieldLineWidth
		}
	}

	boxWidth := maxContentWidth + 4

	horizontalLine := strings.Repeat("─", boxWidth)

	fmt.Println()
	fmt.Printf("  ╭%s╮\n", horizontalLine)
	fmt.Printf("  │  %-*s  │\n", boxWidth-4, title)
	fmt.Printf("  ├%s┤\n", horizontalLine)

	for _, field := range fields {
		labelStr := field.Label + ":"
		valueColumnWidth := boxWidth - 4 - (maxLabelLen + 2)
		fmt.Printf("  │  %-*s %-*s  │\n", maxLabelLen+1, labelStr, valueColumnWidth, field.Value)
	}

	fmt.Printf("  ╰%s╯\n", horizontalLine)
	fmt.Println()
}
