package checking

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// parseDate parses ISO dates like "2026-07-01".
func parseDate(s string) (time.Time, error) {
	return time.Parse("2006-01-02", strings.TrimSpace(s))
}

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

const monthPattern = `(January|February|March|April|May|June|July|August|September|October|November|December|Jan|Feb|Mar|Apr|Jun|Jul|Aug|Sep|Oct|Nov|Dec)`

// periodRe matches statement periods like "Jul 1 - Jul 31, 2026".
var periodRe = regexp.MustCompile(monthPattern + `\s+(\d{1,2})\s*[—–\-−]\s*` + monthPattern + `\s+(\d{1,2}),?\s+(\d{4})`)

type period struct {
	start time.Time
	end   time.Time
}

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
