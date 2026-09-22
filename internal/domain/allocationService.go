package domain

import (
	"context"
	"fmt"

	"samuellando.com/YNAFB/data"
	"samuellando.com/YNAFB/internal/cache"
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
		allocations[i] = s.FromRow(ctx, row)
	}
	return allocations, nil
}
