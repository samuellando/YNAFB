package domain

import (
	"context"
	"fmt"

	"samuellando.com/YNAFB/data"
	"samuellando.com/YNAFB/internal/cache"
)

type BudgetService struct {
	repo              BudgetRepository
	trxService        *TrxService
	allocationService *AllocationService
	goalService       *GoalService
	categoryService   *CategoryService
}

func NewBudgetService(
	repo BudgetRepository,
	trxService *TrxService,
	allocationService *AllocationService,
	goalService *GoalService,
	categoryService *CategoryService,
) *BudgetService {
	return &BudgetService{
		repo:              repo,
		trxService:        trxService,
		allocationService: allocationService,
		goalService:       goalService,
		categoryService:   categoryService,
	}
}

func (s *BudgetService) List(ctx context.Context, loginID int) ([]*Budget, error) {
	return cache.Result(ctx, fmt.Sprintf("busgetServiceList-%d", loginID), func() ([]*Budget, error) {
		return s.list(ctx, loginID)
	})
}

func (s *BudgetService) list(ctx context.Context, loginID int) ([]*Budget, error) {
	rows, err := s.repo.ListBudgets(ctx, data.ListBudgetsParams{
		LoginID: int64(loginID),
	})
	if err != nil {
		return nil, err
	}
	budgets := make([]*Budget, len(rows))
	for i, row := range rows {
		budgets[i] = s.fromRow(ctx, row)
	}
	return budgets, nil
}

func (s *BudgetService) Create(ctx context.Context, loginID int, name string) (*Budget, error) {
	defer cache.InvalidateResults(ctx)
	row, err := s.repo.CreateBudget(ctx, data.CreateBudgetParams{
		LoginID: int64(loginID),
		Name:    name,
	})
	if err != nil {
		return nil, err
	}
	budget := s.fromRow(ctx, row)
	return budget, nil
}

func (s *BudgetService) Get(ctx context.Context, loginID, budgetID int) (*Budget, error) {
	if v, ok := cache.Get[*Budget](ctx, int64(budgetID)); ok {
		return v, nil
	}
	row, err := s.repo.GetBudget(ctx, data.GetBudgetParams{
		ID:      int64(budgetID),
		LoginID: int64(loginID),
	})
	if err != nil {
		return nil, err
	}
	return s.fromRow(ctx, row), nil
}
