package allocation

import (
	"context"
	"samuellando.com/YNAFB/data"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) List(ctx context.Context, loginID, budgetID int) ([]*Allocation, error) {
	rows, err := s.repo.ListAllocations(ctx, data.ListAllocationsParams{
		BudgetID: int64(budgetID),
		LoginID:  int64(loginID),
	})
	if err != nil {
		return nil, err
	}
	allocations := make([]*Allocation, len(rows))
	for i, row := range rows {
		allocations[i] = &Allocation{
			service: s,
			row:     row,
		}
	}
	return allocations, nil
}
