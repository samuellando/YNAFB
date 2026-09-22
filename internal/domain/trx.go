package domain

import (
	"time"

	"samuellando.com/YNAFB/data"
)

type Trx struct {
	service *TrxService
	row data.Trx
	lines []*TrxLine
	payee *Payee
	budget *Budget
	account *Account
	reconciled bool
}

func (t *Trx) Date() time.Time {
	return t.row.Date.Time
}

func (t *Trx) Reconciled() bool {
	return t.reconciled
}

func (t *Trx) TotalOutflow() int {
	return int(t.row.TotalOutflow)
}

func (t *Trx) TotalInflow() int {
	return int(t.row.TotalInflow)
}

func (t *Trx) Lines() []*TrxLine {
	return t.lines
}

