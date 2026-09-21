package budget

import (
	"context"
	"time"

	"samuellando.com/YNAFB/internal/domain/allocation"
	"samuellando.com/YNAFB/internal/domain/category"
	"samuellando.com/YNAFB/internal/domain/goal"
	"samuellando.com/YNAFB/internal/domain/trx"
)

type fetchedData struct {
	trxs        []*trx.Trx
	allocations []*allocation.Allocation
	goals       []*goal.Goal
	categories  []*category.Category
}

func (b *Budget) getCalculationData(ctx context.Context) (*fetchedData, error) {
	trxs, err := b.trxService.List(ctx, int(b.row.LoginID), int(b.row.ID))
	if err != nil {
		return nil, err
	}
	allocations, err := b.allocationService.List(ctx, int(b.row.LoginID), int(b.row.ID))
	if err != nil {
		return nil, err
	}
	goals, err := b.goalService.List(ctx, int(b.row.LoginID), int(b.row.ID))
	if err != nil {
		return nil, err
	}
	categories, err := b.categoryService.List(ctx, int(b.row.LoginID), int(b.row.ID))
	if err != nil {
		return nil, err
	}
	return &fetchedData{
		trxs:        trxs,
		allocations: allocations,
		goals:       goals,
		categories: categories,
	}, nil
}

func getStartAndEndOfMonth(month time.Time) (time.Time, time.Time) {
	startOfMonth := time.Date(month.Year(), month.Month(), 1, 0, 0, 0, 0, month.Location())
	endOfMonth := startOfMonth.AddDate(0, 1, 0).Add(-1 * time.Nanosecond)
	return startOfMonth, endOfMonth
}

func timeInsideMonth(t, start, end time.Time) bool {
	return t.Equal(start) || t.Equal(end) || (t.After(start) && t.Before(end))
}
