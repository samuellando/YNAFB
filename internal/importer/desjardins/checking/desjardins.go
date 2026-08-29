// Package desjardins parses Desjardins (AccesD) bank account statements.
package checking

import (
	"fmt"
	"regexp"
	"strings"

	"samuellando.com/YNAFB/internal/importer/statement"
)

// Parser implements importer.Parser for Desjardins statements.
type Parser struct{}

func (Parser) Name() string { return "desjardins" }

func (Parser) Match(data []byte) bool {
	doc, err := extract(data)
	if err != nil {
		return false
	}
	// French Desjardins account statements ("RELEVÉ DE COMPTE"), distinct from
	// the English Mastercard statements handled by the mastercard parser.
	return strings.Contains(strings.ToUpper(doc.text()), "RELEV")
}

var (
	accountRe = regexp.MustCompile(`(\d{3}-\d{5}-\d)`)
	dateRe    = regexp.MustCompile(`^\d{1,2}\s+[A-Z]{3}$`)
	codeRe    = regexp.MustCompile(`^[A-Z]{2,3}$`)
	amountRe  = regexp.MustCompile(`^\d[\d\s,\x{00A0}\x{202F}\x{2009}\x{2007}]*\.\d{2}$`)
)

// columnSet holds the X positions of the money columns, read from the header.
type columnSet struct {
	fraisX   float64
	retraitX float64
	depotX   float64
	soldeX   float64
}

func (Parser) Parse(data []byte) (statement.Statement, error) {
	doc, err := extract(data)
	if err != nil {
		return statement.Statement{}, err
	}

	text := doc.text()

	p, err := parsePeriod(text)
	if err != nil {
		return statement.Statement{}, fmt.Errorf("desjardins: %w", err)
	}

	stmt := statement.Statement{
		Institution: "Desjardins",
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

	var pending *pendingEntry

	flush := func() {
		if pending == nil {
			return
		}
		pending.entry.Payee = counterpartyFromDesc(pending.desc)
		stmt.Entries = append(stmt.Entries, pending.entry)
		pending = nil
	}

	for _, line := range doc {
		dateIdx := -1
		for i, tok := range line.tokens {
			if dateRe.MatchString(strings.TrimSpace(tok.text)) {
				dateIdx = i
				break
			}
		}

		if dateIdx < 0 {
			if pending == nil {
				continue
			}
			for _, tok := range line.tokens {
				s := strings.TrimSpace(tok.text)
				switch {
				case amountRe.MatchString(s) && tok.x >= cols.fraisX && tok.right < cols.soldeX:
					if cents, err := parseAmount(s); err == nil && !pending.hasAmount {
						applyAmount(&pending.entry, tok.right, cols, cents)
						pending.hasAmount = true
					}
				case amountRe.MatchString(s) && tok.right >= cols.soldeX:
					// running balance, ignore
				case tok.x >= cols.fraisX:
					// non-numeric cell in a money column, ignore
				default:
					pending.desc += " " + s
				}
			}
			if pending.hasAmount {
				flush()
			}
			continue
		}

		flush()

		entry, hasAmount := p.parseTransactionLine(line, dateIdx, cols)
		pending = &pendingEntry{entry: entry, desc: entry.Payee, hasAmount: hasAmount}
		if hasAmount {
			flush()
		}
	}

	flush()

	return stmt, nil
}

type pendingEntry struct {
	entry     statement.Entry
	desc      string
	hasAmount bool
}

func (p period) parseTransactionLine(line textLine, dateIdx int, cols columnSet) (statement.Entry, bool) {
	dateTok := strings.TrimSpace(line.tokens[dateIdx].text)
	month, day, err := parseDayMonth(dateTok)
	if err != nil {
		return statement.Entry{}, false
	}

	entry := statement.Entry{
		TransDate:  p.resolveDate(month, day),
		PostedDate: p.resolveDate(month, day),
	}

	var descParts []string
	hasAmount := false

	for i := dateIdx + 1; i < len(line.tokens); i++ {
		tok := line.tokens[i]
		s := strings.TrimSpace(tok.text)
		switch {
		case i == dateIdx+1 && codeRe.MatchString(s):
			entry.Type = strings.ToLower(s)
		case amountRe.MatchString(s) && tok.x >= cols.fraisX && tok.right < cols.soldeX:
			if cents, err := parseAmount(s); err == nil {
				applyAmount(&entry, tok.right, cols, cents)
				hasAmount = true
			}
		case amountRe.MatchString(s) && tok.right >= cols.soldeX:
			// running balance, ignore
		case tok.x >= cols.fraisX:
			// non-numeric cell in a money column, ignore
		default:
			descParts = append(descParts, s)
		}
	}

	entry.Payee = strings.Join(descParts, " ")
	return entry, hasAmount
}

func applyAmount(e *statement.Entry, right float64, cols columnSet, cents int64) {
	if right >= cols.depotX {
		e.Inflow = cents
	} else {
		e.Outflow = cents
	}
}

func findColumns(doc document) (columnSet, error) {
	for _, line := range doc {
		cs := columnSet{}
		found := map[string]bool{}
		for _, tok := range line.tokens {
			switch normalizeHeader(tok.text) {
			case "frais":
				cs.fraisX, found["frais"] = tok.x, true
			case "retrait":
				cs.retraitX, found["retrait"] = tok.x, true
			case "depot":
				cs.depotX, found["depot"] = tok.x, true
			case "solde":
				cs.soldeX, found["solde"] = tok.x, true
			}
		}
		if found["frais"] && found["retrait"] && found["depot"] && found["solde"] {
			return cs, nil
		}
	}
	return columnSet{}, fmt.Errorf("could not find transaction table header")
}

func normalizeHeader(s string) string {
	return accentReplacer.Replace(strings.ToLower(strings.TrimSpace(s)))
}

// counterpartyFromDesc returns the counterparty from a description like
// "Virement envoyé à / Maurice Lando /papa".
func counterpartyFromDesc(desc string) string {
	collapsed := strings.Join(strings.Fields(desc), " ")
	parts := strings.Split(collapsed, "/")
	payee := strings.TrimSpace(parts[0])
	if len(parts) > 1 {
		payee = strings.TrimSpace(parts[1])
	}
	if payee == "" {
		payee = collapsed
	}
	return payee
}
