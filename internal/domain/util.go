package domain

import (
	"database/sql"
	"time"
)

func getStartAndEndOfMonth(month time.Time) (time.Time, time.Time) {
	startOfMonth := getStartOfMonth(month)
	endOfMonth := startOfMonth.AddDate(0, 1, 0).Add(-1 * time.Nanosecond)
	return startOfMonth, endOfMonth
}

func getStartOfMonth(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
}

func timeLE(t1, t2 time.Time) bool {
	return t1.Before(t2) || t1.Equal(t2)
}

func timeGE(t1, t2 time.Time) bool {
	return t1.After(t2) || t1.Equal(t2)
}

func abs(n int) int {
	if n < 0 {
		return -1 * n
	}
	return n
}

func timeInsideMonth(t, start, end time.Time) bool {
	return t.Equal(start) || t.Equal(end) || (t.After(start) && t.Before(end))
}

func nullInt64FromInt(id *int) sql.NullInt64 {
	if id == nil {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: int64(*id), Valid: true}
}

func maxTime(a, b time.Time) time.Time {
	if a.After(b) {
		return a
	} else {
		return b
	}
}

func minTime(a, b time.Time) time.Time {
	if a.Before(b) {
		return a
	} else {
		return b
	}
}

func monthDiff(t1, t2 time.Time) int {
	years := t2.Year() - t1.Year()
	months := int(t2.Month()) - int(t1.Month())
	return years*12 + months
}
