package types

import (
	"time"
	"database/sql/driver"
	"fmt"
)

type UnixTime struct {
	time.Time
}

func (ut *UnixTime) Scan(src interface{}) error {
	if src == nil {
		ut.Time = time.Time{}
		return nil
	}
	
	switch v := src.(type) {
	case int64:
		ut.Time = time.Unix(v, 0).UTC()
	default:
		return fmt.Errorf("unsupported type for UnixTime: %T, expected int64", src)
	}
	return nil
}

func (ut UnixTime) Value() (driver.Value, error) {
	return ut.Unix(), nil
}
