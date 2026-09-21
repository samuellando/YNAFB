package goal

import (
	"context"

	"samuellando.com/YNAFB/data"
	"samuellando.com/YNAFB/internal/domain/allocation"
)

type Service struct {
	repo Repository
	allocationService *allocation.Service
}

func NewService(repo Repository, allocationService *allocation.Service) *Service {
	return &Service{
		repo: repo,
		allocationService: allocationService,
	}
}

func (s *Service) List(ctx context.Context, loginID, budgetID int) ([]*Goal, error) {
	rows, err := s.repo.ListGoals(ctx, data.ListGoalsParams{
		BudgetID: int64(budgetID),
		LoginID:  int64(loginID),
	})
	if err != nil {
		return nil, err
	}
	goals := make([]*Goal, len(rows))
	for i, row := range rows {
		goals[i] = s.fromRow(ctx, row)
	}
	return goals, nil
}
