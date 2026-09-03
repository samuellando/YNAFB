package types

import (
	"time"
	"database/sql/driver"
	"fmt"
)


type NullUnixTime struct {
	Time  time.Time
	Valid bool
}

func (nut *NullUnixTime) Scan(src interface{}) error {
	if src == nil {
		nut.Time = time.Time{}
		nut.Valid = false
		return nil
	}

	switch v := src.(type) {
	case int64:
		nut.Time = time.Unix(v, 0).UTC()
		nut.Valid = true
	default:
		return fmt.Errorf("unsupported type for NullUnixTime: %T, expected int64", src)
	}
	return nil
}

func (nut NullUnixTime) Value() (driver.Value, error) {
	if !nut.Valid {
		return nil, nil
	}
	return nut.Time.Unix(), nil
}

func (nut NullUnixTime) IsZero() bool {
	return !nut.Valid || nut.Time.IsZero()
}
