package api

import (
	"context"
	"fmt"
	"database/sql"
	"time"
	"samuellando.com/YNAFB/internal/db/types"
)

func getLoginID(ctx context.Context) (int64, error) {
	idA := ctx.Value("loginID")
	if idA == nil {
		return 0, fmt.Errorf("No login ID")
	}
	if id, ok := idA.(int64); ok {
		return id, nil
	} else {
		return 0, fmt.Errorf("Id not int64")
	}
}

func parseMonth(s string) (types.UnixTime, error) {
	t, err := time.Parse("2006-01", s)
	if err != nil {
		return types.UnixTime{}, err
	}
	return types.UnixTime{Time: t}, nil
}

func parseNullableMonth(s *string) (types.NullUnixTime, error) {
	if s == nil {
		return types.NullUnixTime{}, nil
	}
	t, err := time.Parse("2006-01", *s)
	if err != nil {
		return types.NullUnixTime{}, err
	}
	return types.NullUnixTime{Time: t, Valid: true}, nil
}

func nullUnixTimeToMonthString(t types.NullUnixTime) *string {
	if !t.Valid {
		return nil
	}
	s := t.Time.Format(time.RFC3339)
	return &s
}

func nullStringToPtr(s sql.NullString) *string {
	if !s.Valid {
		return nil
	}
	name := s.String
	return &name
}

func BoolPtrToBool(income *bool) bool {
	if income == nil {
		return false
	}
	return *income
}

func intToNullInt64(id *int) sql.NullInt64 {
	if id == nil {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: int64(*id), Valid: true}
}

func nullInt64ToInt(id sql.NullInt64) *int {
	if !id.Valid {
		return nil
	}
	v := int(id.Int64)
	return &v
}

func StrPtrToStr(note *string) string {
	if note == nil {
		return ""
	}
	return *note
}

