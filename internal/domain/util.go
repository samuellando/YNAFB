package domain

import (
	"time"
)

func getStartAndEndOfMonth(month time.Time) (time.Time, time.Time) {
	startOfMonth := time.Date(month.Year(), month.Month(), 1, 0, 0, 0, 0, month.Location())
	endOfMonth := startOfMonth.AddDate(0, 1, 0).Add(-1 * time.Nanosecond)
	return startOfMonth, endOfMonth
}

func timeInsideMonth(t, start, end time.Time) bool {
	return t.Equal(start) || t.Equal(end) || (t.After(start) && t.Before(end))
}
