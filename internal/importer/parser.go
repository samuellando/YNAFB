package importer

import (
	"fmt"
	"sort"

	desjardinsChecking "samuellando.com/YNAFB/internal/importer/desjardins/checking"
	desjardinsMastercard "samuellando.com/YNAFB/internal/importer/desjardins/mastercard"
	"samuellando.com/YNAFB/internal/importer/statement"
	wealthsimpleChecking "samuellando.com/YNAFB/internal/importer/wealthsimple/checking"
	wealthsimpleVisa "samuellando.com/YNAFB/internal/importer/wealthsimple/visa"
)

func init() {
	Register(desjardinsChecking.Parser{})
	Register(desjardinsMastercard.Parser{})
	Register(wealthsimpleChecking.Parser{})
	Register(wealthsimpleVisa.Parser{})
}

// Parser converts a raw statement (PDF bytes) into a normalized Statement.
type Parser interface {
	// Name returns a short identifier for the parser.
	Name() string
	// Match reports whether this parser can handle the given statement data.
	Match(data []byte) bool
	// Parse extracts a normalized Statement from the data.
	Parse(data []byte) (statement.Statement, error)
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

// Parse detects the bank and parses the statement data.
func Parse(data []byte) (statement.Statement, error) {
	for _, p := range Parsers() {
		if p.Match(data) {
			return p.Parse(data)
		}
	}
	return statement.Statement{}, fmt.Errorf("unsupported statement format")
}
