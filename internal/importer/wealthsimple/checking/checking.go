// Package checking parses Wealthsimple chequing account statements.
package checking

import (
	"fmt"
	"math"
	"regexp"
	"strings"

	"samuellando.com/YNAFB/internal/importer/statement"
)

// Parser implements importer.Parser for Wealthsimple chequing statements.
type Parser struct{}

func (Parser) Name() string { return "wealthsimple-checking" }

func (Parser) Match(data []byte) bool {
	doc, err := extract(data)
	if err != nil {
		return false
	}
	text := strings.ToUpper(doc.text())
	return strings.Contains(text, "WEALTHSIMPLE") && strings.Contains(text, "CHEQUING")
}

var accountRe = regexp.MustCompile(`Account number:\s*(\d+)`)

// columnSet holds the X positions of the table columns, read from the header.
// Amount and balance columns are right-aligned, so their left edge varies and
// the header X is used as the column's left boundary.
type columnSet struct {
	dateX     float64
	postedX   float64
	descX     float64
	amountLo  float64
	balanceLo float64
}

func (Parser) Parse(data []byte) (statement.Statement, error) {
	doc, err := extract(data)
	if err != nil {
		return statement.Statement{}, err
	}

	text := doc.text()

	p, err := parsePeriod(text)
	if err != nil {
		return statement.Statement{}, fmt.Errorf("wealthsimple-checking: %w", err)
	}

	stmt := statement.Statement{
		Institution: "Wealthsimple",
		Start:       p.start,
		End:         p.end,
	}
	if m := accountRe.FindStringSubmatch(text); m != nil {
		stmt.Account = m[1]
	}

	cols, err := findColumns(doc)
	if err != nil {
		return statement.Statement{}, err
	}

	for _, line := range doc {
		entry, ok := parseTransactionLine(line, cols)
		if ok {
			stmt.Entries = append(stmt.Entries, entry)
		}
	}

	return stmt, nil
}

func findColumns(doc document) (columnSet, error) {
	for _, line := range doc {
		cs := columnSet{}
		for _, c := range line.columns {
			switch c.text {
			case "DATE":
				cs.dateX = c.x
			case "POSTED DATE":
				cs.postedX = c.x
			case "DESCRIPTION":
				cs.descX = c.x
			case "AMOUNT (CAD)":
				cs.amountLo = c.x
			case "BALANCE (CAD)":
				cs.balanceLo = c.x
			}
		}
		if cs.dateX != 0 && cs.postedX != 0 && cs.descX != 0 && cs.amountLo != 0 && cs.balanceLo != 0 {
			return cs, nil
		}
	}
	return columnSet{}, fmt.Errorf("could not find transaction table header")
}

func parseTransactionLine(line textLine, cols columnSet) (statement.Entry, bool) {
	dateCol := columnNear(line, cols.dateX, 5)
	if dateCol == nil {
		return statement.Entry{}, false
	}

	date, err := parseDate(dateCol.text)
	if err != nil {
		return statement.Entry{}, false
	}

	entry := statement.Entry{TransDate: date}

	if posted := columnNear(line, cols.postedX, 5); posted != nil {
		if pd, err := parseDate(posted.text); err == nil {
			entry.PostedDate = pd
		}
	}
	if desc := columnNear(line, cols.descX, 5); desc != nil {
		entry.Payee = desc.text
	}

	if amt := columnInRange(line, cols.amountLo, cols.balanceLo); amt != nil {
		cents, err := parseMoney(amt.text)
		if err != nil {
			return statement.Entry{}, false
		}
		if cents < 0 {
			entry.Outflow = -cents
		} else {
			entry.Inflow = cents
		}
	}

	return entry, true
}

func columnNear(line textLine, x, tol float64) *column {
	for i := range line.columns {
		if math.Abs(line.columns[i].x-x) <= tol {
			return &line.columns[i]
		}
	}
	return nil
}

func columnInRange(line textLine, lo, hi float64) *column {
	for i := range line.columns {
		if line.columns[i].x >= lo && line.columns[i].x < hi {
			return &line.columns[i]
		}
	}
	return nil
}
