// Package mastercard parses Desjardins credit card (Mastercard) statements.
package mastercard

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"

	"samuellando.com/YNAFB/internal/importer/statement"
)

// Parser implements importer.Parser for Desjardins Mastercard statements.
type Parser struct{}

func (Parser) Name() string { return "desjardins-mastercard" }

func (Parser) Match(data []byte) bool {
	doc, err := extract(data)
	if err != nil {
		return false
	}
	text := strings.ToUpper(doc.text())
	return strings.Contains(text, "MASTERCARD") && strings.Contains(text, "DESJARDINS")
}

// Column X anchors for the transaction table (left edges). Amounts are
// rendered in a fixed-width, right-aligned field so their left edge is stable.
const (
	transDayX   = 91.0
	transMonthX = 115.0
	postDayX    = 154.0
	postMonthX  = 187.0
	descLo      = 210.0
	descHi      = 420.0
	amountX     = 475.0
)

var (
	accountRe  = regexp.MustCompile(`\d{4}\s+\d{2}\*{2}\s+\*{4}\s+(\d{4})`)
	multiSpace = regexp.MustCompile(`\s{2,}`)
)

func (Parser) Parse(data []byte) (statement.Statement, error) {
	doc, err := extract(data)
	if err != nil {
		return statement.Statement{}, err
	}

	text := doc.text()

	sd, err := parseStatementDate(text)
	if err != nil {
		return statement.Statement{}, fmt.Errorf("mastercard: %w", err)
	}

	stmt := statement.Statement{
		Institution: "Desjardins",
	}
	if m := accountRe.FindStringSubmatch(text); m != nil {
		stmt.Account = m[1]
	}

	for _, line := range doc {
		entry, ok := parseTransactionLine(line, sd)
		if ok {
			stmt.Entries = append(stmt.Entries, entry)
		}
	}

	return stmt, nil
}

func parseTransactionLine(line textLine, sd statementDate) (statement.Entry, bool) {
	transDayTok := tokenNear(line, transDayX, 8)
	transMonthTok := tokenNear(line, transMonthX, 8)
	if transDayTok == nil || transMonthTok == nil {
		return statement.Entry{}, false
	}

	day, err1 := strconv.Atoi(strings.TrimSpace(transDayTok.text))
	month, err2 := strconv.Atoi(strings.TrimSpace(transMonthTok.text))
	if err1 != nil || err2 != nil || day < 1 || day > 31 || month < 1 || month > 12 {
		return statement.Entry{}, false
	}

	entry := statement.Entry{
		TransDate: sd.resolveDate(time.Month(month), day),
	}

	if postDayTok := tokenNear(line, postDayX, 8); postDayTok != nil {
		if postMonthTok := tokenNear(line, postMonthX, 14); postMonthTok != nil {
			if pd, e1 := strconv.Atoi(strings.TrimSpace(postDayTok.text)); e1 == nil {
				if pm, e2 := strconv.Atoi(strings.TrimSpace(postMonthTok.text)); e2 == nil {
					entry.PostedDate = sd.resolveDate(time.Month(pm), pd)
				}
			}
		}
	}

	if descTok := tokenInRange(line, descLo, descHi); descTok != nil {
		entry.Payee = merchantFromDesc(descTok.text)
	}

	amountTok := tokenNear(line, amountX, 10)
	if amountTok == nil {
		return statement.Entry{}, false
	}

	cents, isCredit, err := parseCreditAmount(amountTok.text)
	if err != nil {
		return statement.Entry{}, false
	}
	if isCredit {
		entry.Inflow = cents
		entry.Type = "credit"
	} else {
		entry.Outflow = cents
		entry.Type = "debit"
	}

	return entry, true
}

func tokenNear(line textLine, x, tol float64) *token {
	for i := range line.tokens {
		if math.Abs(line.tokens[i].x-x) <= tol {
			return &line.tokens[i]
		}
	}
	return nil
}

func tokenInRange(line textLine, lo, hi float64) *token {
	for i := range line.tokens {
		if line.tokens[i].x >= lo && line.tokens[i].x < hi {
			return &line.tokens[i]
		}
	}
	return nil
}

// merchantFromDesc returns the merchant from a description like
// "COMPTOIR DU CHEF         MONTREAL   QC".
func merchantFromDesc(s string) string {
	s = strings.TrimSpace(s)
	return multiSpace.Split(s, -1)[0]
}
