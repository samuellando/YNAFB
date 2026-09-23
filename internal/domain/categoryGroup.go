package domain

import (
	"context"
	"fmt"

	"samuellando.com/YNAFB/internal/data"
	"samuellando.com/YNAFB/internal/cache"
)

type CategoryGroup struct {
	row    data.CategoryGroup
	budget *Budget
}

// Create a CategoryGroup object from a data row.
func categoryGroupFromRow(ctx context.Context, row data.CategoryGroup, budget *Budget) *CategoryGroup {
	group, _ := cache.Get(ctx, row.ID, func() (*CategoryGroup, error) {
		return &CategoryGroup{
			row:    row,
			budget: budget,
		}, nil
	})
	return group
}

// Create a new category group
func (b *Budget) CreateCategoryGroup(ctx context.Context, name string) (*CategoryGroup, error) {
	defer cache.InvalidateResults(ctx)
	id, err := b.service.repo.CreateCategoryGroup(ctx, data.CreateCategoryGroupParams{
		Name:     name,
		BudgetID: int64(b.ID()),
		LoginID:  int64(b.LoginID()),
	})
	if err != nil {
		return nil, err
	}
	group := categoryGroupFromRow(ctx, data.CategoryGroup{
		ID:       id,
		BudgetID: int64(b.ID()),
		Name:     name,
	}, b)
	return group, nil
}

// Get an existing category group by ID
func (b *Budget) GetCategoryGroup(ctx context.Context, groupID int) (*CategoryGroup, error) {
	return cache.Get(ctx, int64(groupID), func() (*CategoryGroup, error) {
		row, err := b.service.repo.GetCategoryGroup(ctx, data.GetCategoryGroupParams{
			LoginID:  int64(b.LoginID()),
			BudgetID: int64(b.ID()),
			ID:       int64(groupID),
		})
		if err != nil {
			return nil, err
		}
		return categoryGroupFromRow(ctx, row, b), nil
	})
}

// List all the category groups in the budget
func (b *Budget) ListCategoryGroups(ctx context.Context) ([]*CategoryGroup, error) {
	return cache.Result(ctx, fmt.Sprintf("categoryServiceListGroups-%d-%d", b.LoginID(), b.ID()), func() ([]*CategoryGroup, error) {
		rows, err := b.service.repo.ListCategoryGroups(ctx, data.ListCategoryGroupsParams{
			LoginID:  int64(b.LoginID()),
			BudgetID: int64(b.ID()),
		})
		if err != nil {
			return nil, err
		}
		groups := make([]*CategoryGroup, len(rows))
		for i, row := range rows {
			groups[i] = categoryGroupFromRow(ctx, row, b)
		}
		return groups, nil
	})
}

// Update the existing category group
func (g *CategoryGroup) Update(ctx context.Context, name string) error {
	defer cache.InvalidateResults(ctx)
	row, err := g.budget.service.repo.UpdateCategoryGroup(ctx, data.UpdateCategoryGroupParams{
		Name:     name,
		ID:       int64(g.ID()),
		BudgetID: int64(g.budget.ID()),
		LoginID:  int64(g.budget.LoginID()),
	})
	if err != nil {
		return err
	}
	g.row = row
	return nil
}

// Delete a category group
func (g *CategoryGroup) Delete(ctx context.Context) error {
	defer cache.InvalidateResults(ctx)
	defer cache.Delete[*CategoryGroup](ctx, int64(g.ID()))
	return g.budget.service.repo.DeleteCategoryGroup(ctx, data.DeleteCategoryGroupParams{
		ID:       int64(g.ID()),
		BudgetID: int64(g.budget.ID()),
		LoginID:  int64(g.budget.LoginID()),
	})
}

// Get the category group's ID
func (g *CategoryGroup) ID() int {
	return int(g.row.ID)
}

// Get the category group's name
func (g *CategoryGroup) Name() string {
	return g.row.Name
}
