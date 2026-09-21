package trx

import (
	"time"

	"samuellando.com/YNAFB/data"
	"samuellando.com/YNAFB/internal/domain/payee"
)

type Trx struct {
	service *Service
	row data.Trx
	lines []*Line
	payee *payee.Payee
}

func (t *Trx) Date() time.Time {
	return t.row.Date.Time
}

func (t *Trx) TotalOutflow() int {
	return int(t.row.TotalOutflow)
}

func (t *Trx) TotalInflow() int {
	return int(t.row.TotalInflow)
}

func (t *Trx) Lines() []*Line {
	return t.lines
}

