// Package wealthsimple parses Wealthsimple credit card statements.
package wealthsimple

import (
	"fmt"
	"regexp"
	"strings"

	"samuellando.com/YNAFB/internal/importer/statement"
)

type Parser struct{}

func (Parser) Name() string { return "wealthsimple" }

func (Parser) Match(data []byte) bool {
	doc, err := extract(data)
	if err != nil {
		return false
	}
	text := doc.text()
	return strings.Contains(text, "Wealthsimple") && strings.Contains(text, "Credit card statement")
}

var (
	accountRe = regexp.MustCompile(`\d{4}\s+\d{2}\*{2}\s+\*{4}\s+(\d{4})`)
	fxNoteRe  = regexp.MustCompile(`^\d[\d,]*\.\d{2}\s+[A-Z]{3}\s*•`)
)

func (Parser) Parse(data []byte) (statement.Statement, error) {
	doc, err := extract(data)
	if err != nil {
		return statement.Statement{}, err
	}

	text := doc.text()

	p, err := parsePeriod(text)
	if err != nil {
		return statement.Statement{}, fmt.Errorf("wealthsimple: %w", err)
	}

	stmt := statement.Statement{
		Institution: "Wealthsimple",
		Start:       p.start,
		End:         p.end,
	}
	if m := accountRe.FindStringSubmatch(text); m != nil {
		stmt.Account = m[1]
	}

	for _, line := range doc {
		if len(line.columns) >= 5 {
			transMonth, transDay, transErr := parseMonthDay(line.columns[0])
			postedMonth, postedDay, postedErr := parseMonthDay(line.columns[1])
			if transErr == nil && postedErr == nil {
				entry := statement.Entry{
					TransDate:  p.resolveDate(transMonth, transDay),
					PostedDate: p.resolveDate(postedMonth, postedDay),
					Type:       strings.ToLower(line.columns[2]),
					Payee:      line.columns[3],
				}
				if cents, err := parseMoney(line.columns[4]); err == nil {
					if cents < 0 {
						entry.Inflow = -cents
					} else {
						entry.Outflow = cents
					}
				}
				stmt.Entries = append(stmt.Entries, entry)
				continue
			}
		}

		if len(stmt.Entries) > 0 && len(line.columns) == 1 && fxNoteRe.MatchString(line.columns[0]) {
			last := &stmt.Entries[len(stmt.Entries)-1]
			if last.Note == "" {
				last.Note = line.columns[0]
			} else {
				last.Note += " " + line.columns[0]
			}
		}
	}

	return stmt, nil
}
