package api

import (
	"context"
	"fmt"
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

