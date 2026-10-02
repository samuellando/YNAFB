package domain

import (
	"context"
	"fmt"

	"samuellando.com/YNAFB/internal/cache"
	"samuellando.com/YNAFB/internal/data"
)

type ExpenseShare struct {
	service *DomainService
	row data.ExpenseShare
}

func expenseShareFromRow(ctx context.Context, row data.ExpenseShare, service *DomainService) *ExpenseShare {
	share, _ := cache.Get(ctx, row.ID, func() (*ExpenseShare, error) {
		return &ExpenseShare{
			service: service,
			row: row,
		}, nil
	})
	return share
}

func (s *DomainService) createExpenseShare(ctx context.Context, defaultName string) (*ExpenseShare, error) {
	return nil, fmt.Errorf("expense share not implemented")
}

func (e *ExpenseShare) delete(ctx context.Context, service *DomainService) error {
	return fmt.Errorf("expense share not implemented")
}

func (e *ExpenseShare) ID() int {
	return int(e.row.ID)
}

func (e *ExpenseShare) DefaultName() string {
	return e.row.DefaultName
}
