package config

import (
	"encoding/json"
	"time"
)

type Duration struct {
	time.Duration
}

func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.String())
}

func (d *Duration) UnmarshalJSON(b []byte) error {
	var v string
	var err error
	if err = json.Unmarshal(b, &v); err != nil {
		return err
	}
	d.Duration, err = time.ParseDuration(v)
	if err != nil {
		return err
	}
	return nil
}
