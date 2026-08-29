package importer

import (
	"fmt"
	"regexp"
	"strings"
)

type wealthsimpleParser struct{}

func init() { Register(wealthsimpleParser{}) }

func (wealthsimpleParser) Name() string { return "wealthsimple" }

func (wealthsimpleParser) Match(doc Document) bool {
	text := doc.Text()
	return strings.Contains(text, "Wealthsimple") && strings.Contains(text, "Credit card statement")
}

var (
	wsAccountRe = regexp.MustCompile(`\d{4}\s+\d{2}\*{2}\s+\*{4}\s+(\d{4})`)
	fxNoteRe    = regexp.MustCompile(`^\d[\d,]*\.\d{2}\s+[A-Z]{3}\s*•`)
)

func (wealthsimpleParser) Parse(doc Document) (Statement, error) {
	text := doc.Text()

	p, err := parsePeriod(text)
	if err != nil {
		return Statement{}, fmt.Errorf("wealthsimple: %w", err)
	}

	stmt := Statement{
		Institution: "Wealthsimple",
		Start:       p.start,
		End:         p.end,
	}
	if m := wsAccountRe.FindStringSubmatch(text); m != nil {
		stmt.Account = m[1]
	}

	for _, line := range doc {
		if len(line.Columns) >= 5 {
			transMonth, transDay, transErr := parseMonthDay(line.Columns[0])
			postedMonth, postedDay, postedErr := parseMonthDay(line.Columns[1])
			if transErr == nil && postedErr == nil {
				entry := Entry{
					TransDate:  p.resolveDate(transMonth, transDay),
					PostedDate: p.resolveDate(postedMonth, postedDay),
					Type:       strings.ToLower(line.Columns[2]),
					Payee:      line.Columns[3],
				}
				if cents, err := ParseMoney(line.Columns[4]); err == nil {
					applyAmount(&entry, cents)
				}
				stmt.Entries = append(stmt.Entries, entry)
				continue
			}
		}

		if len(stmt.Entries) > 0 && len(line.Columns) == 1 && fxNoteRe.MatchString(line.Columns[0]) {
			last := &stmt.Entries[len(stmt.Entries)-1]
			last.Note = joinNote(last.Note, line.Columns[0])
		}
	}

	return stmt, nil
}

func joinNote(existing, next string) string {
	if existing == "" {
		return next
	}
	return existing + " " + next
}
