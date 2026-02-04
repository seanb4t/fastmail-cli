package output

import (
	"strings"
)

// FormatTable renders data as an ASCII table.
// Returns an empty string if headers or rows are empty.
func FormatTable(headers []string, rows [][]string) string {
	if len(headers) == 0 {
		return ""
	}

	// Calculate column widths
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = len(h)
	}
	for _, row := range rows {
		for i := 0; i < len(widths) && i < len(row); i++ {
			if len(row[i]) > widths[i] {
				widths[i] = len(row[i])
			}
		}
	}

	var sb strings.Builder

	// Header row
	sb.WriteString(formatRow(headers, widths))
	sb.WriteString("\n")

	// Separator
	sb.WriteString(formatSeparator(widths))
	sb.WriteString("\n")

	// Data rows
	for _, row := range rows {
		sb.WriteString(formatRow(row, widths))
		sb.WriteString("\n")
	}

	return sb.String()
}

// formatRow formats a single row with proper column alignment.
func formatRow(cells []string, widths []int) string {
	var sb strings.Builder
	for i, w := range widths {
		if i > 0 {
			sb.WriteString("  ")
		}
		cell := ""
		if i < len(cells) {
			cell = cells[i]
		}
		sb.WriteString(padRight(cell, w))
	}
	return sb.String()
}

// formatSeparator creates a separator line with dashes.
func formatSeparator(widths []int) string {
	var sb strings.Builder
	for i, w := range widths {
		if i > 0 {
			sb.WriteString("  ")
		}
		sb.WriteString(strings.Repeat("-", w))
	}
	return sb.String()
}

// padRight pads s with spaces to reach width.
func padRight(s string, width int) string {
	if len(s) >= width {
		return s
	}
	return s + strings.Repeat(" ", width-len(s))
}
