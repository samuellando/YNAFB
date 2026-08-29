package importer

import (
	"fmt"
	"sort"
)

// Parser converts a structured statement document into a normalized Statement.
type Parser interface {
	// Name returns a short identifier for the parser.
	Name() string
	// Match reports whether this parser can handle the given document.
	Match(doc Document) bool
	// Parse extracts a normalized Statement from the document.
	Parse(doc Document) (Statement, error)
}

var parsers []Parser

// Register adds a parser to the registry. It is intended to be called from
// package-level init functions so new bank parsers are picked up automatically.
func Register(p Parser) {
	parsers = append(parsers, p)
}

// Parsers returns the registered parsers ordered by name for deterministic
// dispatch and diagnostics.
func Parsers() []Parser {
	ordered := make([]Parser, len(parsers))
	copy(ordered, parsers)
	sort.Slice(ordered, func(i, j int) bool {
		return ordered[i].Name() < ordered[j].Name()
	})
	return ordered
}

// Parse detects the bank and parses the statement document.
func Parse(doc Document) (Statement, error) {
	for _, p := range Parsers() {
		if p.Match(doc) {
			return p.Parse(doc)
		}
	}
	return Statement{}, fmt.Errorf("unsupported statement format")
}

// ImportFile extracts text from a PDF file and parses it into a Statement.
func ImportFile(path string) (Statement, error) {
	doc, err := ExtractFile(path)
	if err != nil {
		return Statement{}, err
	}
	return Parse(doc)
}
