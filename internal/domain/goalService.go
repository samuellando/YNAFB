package domain

import (
	"context"
	"fmt"

	"samuellando.com/YNAFB/data"
	"samuellando.com/YNAFB/internal/cache"
	"samuellando.com/YNAFB/internal/db/types"
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
	return cache.Result(ctx, fmt.Sprintf("goalServiceList-%d-%d", loginID, budgetID), func() ([]*Goal, error) {
		return s.list(ctx, loginID, budgetID)
	})
}

func (s *GoalService) list(ctx context.Context, loginID, budgetID int) ([]*Goal, error) {
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

func (s *GoalService) GetByCategory(ctx context.Context, loginID, budgetID, categoryID int) (*Goal, error) {
	row, err := s.repo.GetGoalByCategory(ctx, data.GetGoalByCategoryParams{
		LoginID:    int64(loginID),
		BudgetID:   int64(budgetID),
		CategoryID: int64(categoryID),
	})
	if err != nil {
		return nil, err
	}
	goal := s.fromRow(ctx, data.Goal{
		ID:         row.ID,
		BudgetID:   row.BudgetID,
		CategoryID: row.CategoryID,
		Type:       row.Type,
		StartDate:  row.StartDate,
		EndDate:    row.EndDate,
		Amount:     row.Amount,
	})
	cache.Store(ctx, goal.row.ID, goal)
	return goal, nil
}

func (s *GoalService) Create(ctx context.Context, loginID, budgetID, categoryID int, goalType string, start types.UnixTime, end types.NullUnixTime, amount int) (*Goal, error) {
	defer cache.InvalidateResults(ctx)
	row, err := s.repo.CreateGoal(ctx, data.CreateGoalParams{
		Type:       goalType,
		StartDate:  start,
		EndDate:    end,
		CategoryID: int64(categoryID),
		Amount:     int64(amount),
		BudgetID:   int64(budgetID),
		LoginID:    int64(loginID),
	})
	if err != nil {
		return nil, err
	}
	goal := s.fromRow(ctx, row)
	cache.Store(ctx, goal.row.ID, goal)
	return goal, nil
}
