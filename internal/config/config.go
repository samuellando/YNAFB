package config

import (
	"encoding/json"
	"log"
	"os"
)

type ConfigValues struct {
	ExpenseShareCodeTtl Duration
}

var Values ConfigValues

func LoadConfigFromFile(p string) error {
	data, err := os.ReadFile(p)
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, &Values)
	if err != nil {
		return err
	}
	log.Printf("Loaded config %+v", Values)
	return nil
}
