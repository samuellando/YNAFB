package allocation

import (
	"samuellando.com/YNAFB/data"
	"time"
)

type Allocation struct {
	service *Service
	row     data.Allocation
}

func (a *Allocation) Month() time.Time {
	return a.row.Month.Time
}

func (a *Allocation) Amount() int {
	return int(a.row.Amount)
}
