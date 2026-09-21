package payee

import (
	"samuellando.com/YNAFB/data"
)

type Payee struct {
	row data.Payee
}

func FromRow(row data.Payee) *Payee {
	return &Payee{
		row: row,
	}
}
