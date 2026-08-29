package mastercard

import (
	"fmt"
	"regexp"
	"time"
)

// statementDateRe matches "Day 29 Month 07 Year 2026" appearing on the
// statement header.
var statementDateRe = regexp.MustCompile(`Day\s+(\d{1,2})\s+Month\s+(\d{1,2})\s+Year\s+(\d{4})`)

// statementDate is the statement's closing date, used for year inference.
type statementDate struct {
	year  int
	month time.Month
}

func parseStatementDate(s string) (statementDate, error) {
	m := statementDateRe.FindStringSubmatch(s)
	if m == nil {
		return statementDate{}, fmt.Errorf("could not parse statement date from %q", s)
	}

	day := atoi(m[1])
	month := atoi(m[2])
	year := atoi(m[3])
	if day < 1 || day > 31 || month < 1 || month > 12 || year < 1900 {
		return statementDate{}, fmt.Errorf("invalid statement date %s/%s/%s", m[1], m[2], m[3])
	}

	return statementDate{year: year, month: time.Month(month)}, nil
}

// resolveDate infers the year for a transaction date. Transactions occur
// before the statement closing date, so a month later than the statement month
// must belong to the previous year.
func (d statementDate) resolveDate(month time.Month, day int) time.Time {
	year := d.year
	if month > d.month {
		year--
	}
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}

func atoi(s string) int {
	n := 0
	for _, r := range s {
		if r < '0' || r > '9' {
			return 0
		}
		n = n*10 + int(r-'0')
	}
	return n
}
