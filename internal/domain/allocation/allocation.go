package allocation

import (
	"time"
	"context"

	"samuellando.com/YNAFB/data"
	"samuellando.com/YNAFB/internal/cache"
)

type Allocation struct {
	service *Service
	row     data.Allocation
}

func (s *Service) FromRow(ctx context.Context, row data.Allocation) *Allocation {
	if cached, ok := cache.Get[*Allocation](ctx, row.ID); ok {
		return cached
	}
	group := &Allocation{
		service: s,
		row: row,
	}
	cache.Store(ctx, row.ID, group)
	return group
}

func (a *Allocation) Category() int {
	return int(a.row.CategoryID)
}

func (a *Allocation) Month() time.Time {
	return a.row.Month.Time
}

func (a *Allocation) Amount() int {
	return int(a.row.Amount)
}
