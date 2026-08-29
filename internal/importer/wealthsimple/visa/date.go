package visa

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var monthNames = map[string]time.Month{
	"jan":       time.January,
	"january":   time.January,
	"feb":       time.February,
	"february":  time.February,
	"mar":       time.March,
	"march":     time.March,
	"apr":       time.April,
	"april":     time.April,
	"may":       time.May,
	"jun":       time.June,
	"june":      time.June,
	"jul":       time.July,
	"july":      time.July,
	"aug":       time.August,
	"august":    time.August,
	"sep":       time.September,
	"sept":      time.September,
	"september": time.September,
	"oct":       time.October,
	"october":   time.October,
	"nov":       time.November,
	"november":  time.November,
	"dec":       time.December,
	"december":  time.December,
}

func parseMonth(s string) (time.Month, error) {
	m, ok := monthNames[strings.ToLower(strings.TrimSpace(s))]
	if !ok {
		return 0, fmt.Errorf("unknown month %q", s)
	}
	return m, nil
}

// parseMonthDay parses strings like "Jul 16" into a month and day.
func parseMonthDay(s string) (time.Month, int, error) {
	parts := strings.Fields(strings.TrimSpace(s))
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid date %q", s)
	}
	month, err := parseMonth(parts[0])
	if err != nil {
		return 0, 0, err
	}
	day, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, fmt.Errorf("invalid day in %q: %w", s, err)
	}
	return month, day, nil
}

const monthPattern = `(January|February|March|April|May|June|July|August|September|October|November|December|Jan|Feb|Mar|Apr|Jun|Jul|Aug|Sep|Oct|Nov|Dec)`

// periodRe matches statement period strings like "Jul 15 — Aug 14, 2026".
// The year is only present on the end date.
var periodRe = regexp.MustCompile(monthPattern + `\s+(\d{1,2})\s*[—–\-−]\s*` + monthPattern + `\s+(\d{1,2}),?\s+(\d{4})`)

type period struct {
	start time.Time
	end   time.Time
}

// parsePeriod parses a statement period string into start/end times with year
// inference so that "Dec 15 — Jan 14, 2026" maps December to 2025.
func parsePeriod(s string) (period, error) {
	m := periodRe.FindStringSubmatch(s)
	if m == nil {
		return period{}, fmt.Errorf("could not parse statement period from %q", s)
	}

	startMonth, err := parseMonth(m[1])
	if err != nil {
		return period{}, err
	}
	startDay, err := strconv.Atoi(m[2])
	if err != nil {
		return period{}, err
	}
	endMonth, err := parseMonth(m[3])
	if err != nil {
		return period{}, err
	}
	endDay, err := strconv.Atoi(m[4])
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
// It prefers a year that keeps the date within the period.
func (p period) resolveDate(month time.Month, day int) time.Time {
	for _, year := range []int{p.end.Year(), p.start.Year()} {
		d := time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
		if !d.Before(p.start) && !d.After(p.end) {
			return d
		}
	}
	return time.Date(p.end.Year(), month, day, 0, 0, 0, 0, time.UTC)
}
