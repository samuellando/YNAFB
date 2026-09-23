package domain

import (
	"context"
	"fmt"

	"samuellando.com/YNAFB/internal/data"
	"samuellando.com/YNAFB/internal/cache"
)

type Budget struct {
	service *DomainService
	row     data.Budget
}

func budgetFromRow(ctx context.Context, row data.Budget, service *DomainService) *Budget {
	budget, _ := cache.Get(ctx, row.ID, func() (*Budget, error) {
		return &Budget{
			service: service,
			row:     row,
		}, nil
	})
	return budget
}

// List all the budgets for a login
func (s *DomainService) ListBudgets(ctx context.Context, loginID int) ([]*Budget, error) {
	return cache.Result(ctx, fmt.Sprintf("budgetServiceList-%d", loginID), func() ([]*Budget, error) {
		rows, err := s.repo.ListBudgets(ctx, data.ListBudgetsParams{
			LoginID: int64(loginID),
		})
		if err != nil {
			return nil, err
		}
		budgets := make([]*Budget, len(rows))
		for i, row := range rows {
			budgets[i] = budgetFromRow(ctx, row, s)
		}
		return budgets, nil
	})
}

// Create a new budget
func (s *DomainService) CreateBudget(ctx context.Context, loginID int, name string) (*Budget, error) {
	defer cache.InvalidateResults(ctx)
	row, err := s.repo.CreateBudget(ctx, data.CreateBudgetParams{
		LoginID: int64(loginID),
		Name:    name,
	})
	if err != nil {
		return nil, err
	}
	budget := budgetFromRow(ctx, row, s)
	return budget, nil
}

// Get an existing budget by ID
func (s *DomainService) GetBudget(ctx context.Context, loginID, budgetID int) (*Budget, error) {
	return cache.Get(ctx, int64(budgetID), func() (*Budget, error) {
		row, err := s.repo.GetBudget(ctx, data.GetBudgetParams{
			ID:      int64(budgetID),
			LoginID: int64(loginID),
		})
		if err != nil {
			return nil, err
		}
		return budgetFromRow(ctx, row, s), nil
	})
}

func (b *Budget) ID() int {
	return int(b.row.ID)
}

func (b *Budget) LoginID() int {
	return int(b.row.LoginID)
}

func (b *Budget) Name() string {
	return b.row.Name
}

func (b *Budget) ListTransactions(ctx context.Context) ([]*Trx, error) {
	return b.listTransactions(ctx)
}

func (b *Budget) Update(ctx context.Context, name string) error {
	defer cache.InvalidateResults(ctx)
	row, err := b.service.repo.UpdateBudget(ctx, data.UpdateBudgetParams{
		LoginID: int64(b.LoginID()),
		ID:      int64(b.ID()),
		Name:    name,
	})
	if err != nil {
		return err
	}
	b.row = row
	return nil
}

func (b *Budget) Delete(ctx context.Context) error {
	defer cache.InvalidateResults(ctx)
	defer cache.Delete[*Budget](ctx, int64(b.ID()))
	return b.service.repo.DeleteBudget(ctx, data.DeleteBudgetParams{
		LoginID: int64(b.LoginID()),
		ID:      int64(b.ID()),
	})
}
