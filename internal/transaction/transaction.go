package transaction

import (
	"time"
	"samuellando.com/YNAFB/internal/category"
	"samuellando.com/YNAFB/internal/payee"
)

type Transaction struct {
	Date time.Time
	Payee *payee.Payee
	Splits []Split
	Reconciled bool
	Note string
}

type Split struct {
	Transfer bool
	Category *category.Category
	Outflow int
	Inflow int
}
