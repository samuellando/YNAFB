package checking

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var accentReplacer = strings.NewReplacer(
	"é", "e", "è", "e", "ê", "e", "ë", "e",
	"ô", "o", "ö", "o",
	"û", "u", "ù", "u", "ü", "u",
	"à", "a", "â", "a",
	"î", "i", "ï", "i",
	"ç", "c",
)

func normalizeMonth(s string) string {
	return accentReplacer.Replace(strings.ToLower(strings.TrimSpace(s)))
}

var frenchMonths = map[string]time.Month{
	"jan":       time.January,
	"janv":      time.January,
	"janvier":   time.January,
	"fev":       time.February,
	"fevr":      time.February,
	"fevrier":   time.February,
	"mar":       time.March,
	"mars":      time.March,
	"avr":       time.April,
	"avril":     time.April,
	"mai":       time.May,
	"jun":       time.June,
	"juin":      time.June,
	"jul":       time.July,
	"juil":      time.July,
	"juillet":   time.July,
	"aou":       time.August,
	"aout":      time.August,
	"sep":       time.September,
	"sept":      time.September,
	"septembre": time.September,
	"oct":       time.October,
	"octobre":   time.October,
	"nov":       time.November,
	"novembre":  time.November,
	"dec":       time.December,
	"decembre":  time.December,
}

func parseMonth(s string) (time.Month, error) {
	m, ok := frenchMonths[normalizeMonth(s)]
	if !ok {
		return 0, fmt.Errorf("unknown month %q", s)
	}
	return m, nil
}

// parseDayMonth parses transaction dates like "1 JUL" (day then month).
func parseDayMonth(s string) (time.Month, int, error) {
	parts := strings.Fields(strings.TrimSpace(s))
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid date %q", s)
	}
	day, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, fmt.Errorf("invalid day in %q: %w", s, err)
	}
	month, err := parseMonth(parts[1])
	if err != nil {
		return 0, 0, err
	}
	return month, day, nil
}

// periodRe matches French statement periods like "du 1er juillet au 31 juillet 2026".
var periodRe = regexp.MustCompile(`du\s+(\d{1,2})(?:er|e)?\s+([a-zA-Zûôéèàç]+)\s+au\s+(\d{1,2})(?:er|e)?\s+([a-zA-Zûôéèàç]+)\s+(\d{4})`)

type period struct {
	start time.Time
	end   time.Time
}

func parsePeriod(s string) (period, error) {
	m := periodRe.FindStringSubmatch(s)
	if m == nil {
		return period{}, fmt.Errorf("could not parse statement period from %q", s)
	}

	startDay, err := strconv.Atoi(m[1])
	if err != nil {
		return period{}, err
	}
	startMonth, err := parseMonth(m[2])
	if err != nil {
		return period{}, err
	}
	endDay, err := strconv.Atoi(m[3])
	if err != nil {
		return period{}, err
	}
	endMonth, err := parseMonth(m[4])
	if err != nil {
		return period{}, err
	}
	endYear, err := strconv.Atoi(m[5])
	if err != nil {
		return period{}, err
	}

	startYear := endYear
	if startMonth > endMonth {
		startYear = endYear - 1
	}

	return period{
		start: time.Date(startYear, startMonth, startDay, 0, 0, 0, 0, time.UTC),
		end:   time.Date(endYear, endMonth, endDay, 0, 0, 0, 0, time.UTC),
	}, nil
}

// resolveDate infers the year for a month/day pair using the statement period.
func (p period) resolveDate(month time.Month, day int) time.Time {
	for _, year := range []int{p.end.Year(), p.start.Year()} {
		d := time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
		if !d.Before(p.start) && !d.After(p.end) {
			return d
		}
	}
	return time.Date(p.end.Year(), month, day, 0, 0, 0, 0, time.UTC)
}
