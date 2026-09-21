package budget

import (
	"context"

	"samuellando.com/YNAFB/data"
	"samuellando.com/YNAFB/internal/cache"
	"samuellando.com/YNAFB/internal/domain/allocation"
	"samuellando.com/YNAFB/internal/domain/category"
	"samuellando.com/YNAFB/internal/domain/goal"
	"samuellando.com/YNAFB/internal/domain/trx"
)

type Budget struct {
	repo Repository
	trxService *trx.Service
	allocationService *allocation.Service
	goalService *goal.Service
	categoryService *category.Service
	row  data.Budget
}

func (s *Service) fromRow(ctx context.Context, row data.Budget) *Budget {
	if v, ok := cache.Get[*Budget](ctx, row.ID); ok {
		return v
	}
	budget := &Budget{
		repo:              s.repo,
		trxService:        s.trxService,
		allocationService: s.allocationService,
		goalService:       s.goalService,
		categoryService:   s.categoryService,
		row:               row,
	}
	cache.Store(ctx, row.ID, budget)
	return budget
}

func (b *Budget) ID() int {
	return int(b.row.ID)
}

func (b *Budget) Name() string {
	return b.row.Name
}

func (b *Budget) Update(ctx context.Context, name string) error {
	row, err := b.repo.UpdateBudget(ctx, data.UpdateBudgetParams{
		LoginID: b.row.LoginID,
		ID:      b.row.ID,
		Name:    name,
	})
	if err != nil {
		return err
	}
	b.row = row
	return nil
}

func (b *Budget) Delete(ctx context.Context) error {
	err := b.repo.DeleteBudget(ctx, data.DeleteBudgetParams{
		LoginID: b.row.LoginID,
		ID:      b.row.ID,
	})
	cache.InvalidateResults(ctx)
	return err
}
