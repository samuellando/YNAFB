package visa

import "strings"

// textLine is a single visual line from a PDF, with its text grouped into
// left-to-right columns by X position.
type textLine struct {
	y       float64
	columns []string
}

// document is an ordered (top-to-bottom) list of text lines.
type document []textLine

func (d document) text() string {
	var sb strings.Builder
	for _, line := range d {
		sb.WriteString(strings.Join(line.columns, " "))
		sb.WriteString("\n")
	}
	return sb.String()
}
