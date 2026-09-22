package domain

import (
	"context"

	"samuellando.com/YNAFB/data"
)

type GoalService struct {
	repo GoalRepository
	allocationService *AllocationService
}

func NewGoalService(repo GoalRepository, allocationService *AllocationService) *GoalService {
	return &GoalService{
		repo: repo,
		allocationService: allocationService,
	}
}

func (s *GoalService) List(ctx context.Context, loginID, budgetID int) ([]*Goal, error) {
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
