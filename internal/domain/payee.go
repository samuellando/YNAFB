package domain

import (
	"context"

	"samuellando.com/YNAFB/data"
	"samuellando.com/YNAFB/internal/cache"
)

type Payee struct {
	row data.Payee
}

func (s *PayeeService) fromRow(ctx context.Context, row data.Payee) *Payee {
	if cached, ok := cache.Get[*Payee](ctx, row.ID); ok {
		return cached
	}
	payee := &Payee{
		row: row,
	}
	cache.Store(ctx, row.ID, payee)
	return payee
}
