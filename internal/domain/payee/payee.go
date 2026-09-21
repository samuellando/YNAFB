package payee

import (
	"context"

	"samuellando.com/YNAFB/data"
	"samuellando.com/YNAFB/internal/cache"
)

type Payee struct {
	row data.Payee
}

func (s *Service) FromRow(ctx context.Context, row data.Payee) *Payee {
	if cached, ok := cache.Get[*Payee](ctx, "payee", row.ID); ok {
		return cached
	}
	payee := &Payee{
		row: row,
	}
	cache.Store(ctx, "payee", row.ID, payee)
	return payee
}
