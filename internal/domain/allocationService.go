package domain

import (
	"context"
	"fmt"
	"time"

	"samuellando.com/YNAFB/data"
	"samuellando.com/YNAFB/internal/cache"
	"samuellando.com/YNAFB/internal/db/types"
)

type AllocationService struct {
	repo AloocationRepository
}

func NewAllocationService(repo AloocationRepository) *AllocationService {
	return &AllocationService{repo: repo}
}

func (s *AllocationService) List(ctx context.Context,  loginID, budgetID int) ([]*Allocation, error) {
	return cache.Result(ctx, fmt.Sprintf("allocationServiceList-%d-%d", loginID, budgetID), func() ([]*Allocation, error) {
		return s.list(ctx, loginID, budgetID)
	})
} 

func (s *AllocationService) list(ctx context.Context, loginID, budgetID int) ([]*Allocation, error) {
	rows, err := s.repo.ListAllocations(ctx, data.ListAllocationsParams{
		BudgetID: int64(budgetID),
		LoginID:  int64(loginID),
	})
	if err != nil {
		return nil, err
	}
	allocations := make([]*Allocation, len(rows))
	for i, row := range rows {
		allocations[i] = s.fromRow(ctx, row)
	}
	return allocations, nil
}

func (s *AllocationService) Set(ctx context.Context, loginID, budgetID, categoryID int, month time.Time, amount int) (*Allocation, error) {
	defer cache.InvalidateResults(ctx)
	row, err := s.repo.SetAllocation(ctx, data.SetAllocationParams{
		CategoryID: int64(categoryID),
		Month:      types.UnixTime{Time: month},
		Amount:     int64(amount),
		BudgetID:   int64(budgetID),
		LoginID:    int64(loginID),
	})
	if err != nil {
		return nil, err
	}
	allocation := s.fromRow(ctx, row)
	cache.Store(ctx, allocation.row.ID, allocation)
	return allocation, nil
}
