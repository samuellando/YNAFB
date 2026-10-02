package domain

import (
	"context"
	"fmt"

	"samuellando.com/YNAFB/internal/cache"
	"samuellando.com/YNAFB/internal/data"
)

type SplitCategorization struct {
	row      data.ExpenseShareTrxSplitLine
	trx      *ExpenseShareTransaction
	category *Category
}

func splitCategorizationFromRow(ctx context.Context, row data.ExpenseShareTrxSplitLine, trx *ExpenseShareTransaction, category *Category) *SplitCategorization {
	cat, _ := cache.Get(ctx, row.ID, func() (*SplitCategorization, error) {
		return &SplitCategorization{
			row:      row,
			trx:      trx,
			category: category,
		}, nil
	})
	return cat
}

func (s *SplitCategorization) ID() int {
	return int(s.row.ID)
}

func (s *SplitCategorization) Category() (*Category, error) {
	if s.row.CategoryID.Valid {
		return s.category, nil
	}
	return nil, fmt.Errorf("Split categorization has no category")
}

func (s *SplitCategorization) Outflow() int {
	return int(s.row.Outflow)
}

func (s *SplitCategorization) Inflow() int {
	return int(s.row.Inflow)
}

func (s *SplitCategorization) Transaction() *ExpenseShareTransaction {
	return s.trx
}

func (s *SplitCategorization) Update(ctx context.Context, category *Category, outflow, inflow int) error {
	return fmt.Errorf("expense share not implemented")
}

func (s *SplitCategorization) Delete(ctx context.Context) error {
	return fmt.Errorf("expense share not implemented")
}
