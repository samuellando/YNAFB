package allocation

import (
	"context"
	"fmt"

	"samuellando.com/YNAFB/data"
	"samuellando.com/YNAFB/internal/cache"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) List(ctx context.Context,  loginID, budgetID int) ([]*Allocation, error) {
	return cache.Result(ctx, fmt.Sprintf("allocationServiceList-%d-%d", loginID, budgetID), func() ([]*Allocation, error) {
		return s.list(ctx, loginID, budgetID)
	})
} 

func (s *Service) list(ctx context.Context, loginID, budgetID int) ([]*Allocation, error) {
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
