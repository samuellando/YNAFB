package checking

import "strings"

// column is a horizontal run of characters at a shared X position.
type column struct {
	x    float64
	text string
}

// textLine is a single visual line, holding its columns left-to-right.
type textLine struct {
	y       float64
	columns []column
}

// document is an ordered (top-to-bottom) list of text lines.
type document []textLine

func (d document) text() string {
	var sb strings.Builder
	for _, line := range d {
		for _, c := range line.columns {
			sb.WriteString(c.text)
			sb.WriteString(" ")
		}
		sb.WriteString("\n")
	}
	return sb.String()
}
