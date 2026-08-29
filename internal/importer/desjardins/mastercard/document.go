package mastercard

import "strings"

// token is a horizontal run of characters sharing an X position.
type token struct {
	x     float64
	right float64
	text  string
}

// textLine is a single visual line, holding its tokens left-to-right.
type textLine struct {
	y      float64
	tokens []token
}

// document is an ordered (top-to-bottom) list of text lines.
type document []textLine

func (d document) text() string {
	var sb strings.Builder
	for _, line := range d {
		for _, t := range line.tokens {
			sb.WriteString(t.text)
			sb.WriteString(" ")
		}
		sb.WriteString("\n")
	}
	return sb.String()
}
