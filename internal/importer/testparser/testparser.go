// Package testparser provides a fake statement parser for tests. It returns a
// deterministic statement so tests don't depend on real (sensitive) bank data.
package testparser

import (
	"bytes"

	"samuellando.com/YNAFB/internal/importer/statement"
)

// marker is matched against statement data to select this parser.
var marker = []byte("ynafb-test-statement")

// Parser is a minimal parser used only in tests.
type Parser struct{}

func (Parser) Name() string { return "test" }

func (Parser) Match(data []byte) bool {
	return bytes.Contains(data, marker)
}

func (Parser) Parse(data []byte) (statement.Statement, error) {
	return statement.Statement{
		Institution: "Test Bank",
		Account:     "1234",
		Entries: []statement.Entry{
			{Payee: "Test Merchant", Outflow: 1234},
			{Payee: "Incoming Transfer", Inflow: 5000},
		},
	}, nil
}
