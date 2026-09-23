package domain

import (
	"context"
	"database/sql"
	"fmt"

	"samuellando.com/YNAFB/internal/data"
	"samuellando.com/YNAFB/internal/cache"
)

type Category struct {
	row    data.Category
	budget *Budget
	group  *CategoryGroup
}

// Create a Category object from a data row.
func categoryFromRow(ctx context.Context, row data.Category, budget *Budget, group *CategoryGroup) *Category {
	category, _ := cache.Get(ctx, row.ID, func() (*Category, error) {
		return &Category{
			row:    row,
			budget: budget,
			group:  group,
		}, nil
	})
	return category
}

// Create a new category
func (b *Budget) CreateCategory(ctx context.Context, name string, group *CategoryGroup) (*Category, error) {
	defer cache.InvalidateResults(ctx)
	groupID := sql.NullInt64{}
	if group != nil {
		groupID = sql.NullInt64{Valid: true, Int64: int64(group.ID())}
	}
	row, err := b.service.repo.CreateCategory(ctx, data.CreateCategoryParams{
		Name:            name,
		CategoryGroupID: groupID,
		BudgetID:        int64(b.ID()),
		LoginID:         int64(b.LoginID()),
	})
	if err != nil {
		return nil, err
	}
	category := categoryFromRow(ctx, row, b, group)
	return category, nil
}

// Get an existing category by ID
func (b *Budget) GetCategory(ctx context.Context, categoryID int) (*Category, error) {
	return cache.Get(ctx, int64(categoryID), func() (*Category, error) {
		row, err := b.service.repo.GetCategory(ctx, data.GetCategoryParams{
			LoginID:  int64(b.LoginID()),
			BudgetID: int64(b.ID()),
			ID:       int64(categoryID),
		})
		if err != nil {
			return nil, err
		}
		var group *CategoryGroup
		if row.GroupID.Valid {
			group = categoryGroupFromRow(ctx, data.CategoryGroup{
				ID:       row.GroupID.Int64,
				BudgetID: row.BudgetID,
				Name:     row.GroupName.String,
			}, b)
		}
		return categoryFromRow(ctx, data.Category{
			ID:              row.ID,
			BudgetID:        row.BudgetID,
			Name:            row.Name,
			CategoryGroupID: row.GroupID,
		}, b, group), nil
	})
}

// List all the categories in the budget
func (b *Budget) ListCategories(ctx context.Context) ([]*Category, error) {
	return cache.Result(ctx, fmt.Sprintf("categoryServiceList-%d-%d", b.LoginID(), b.ID()), func() ([]*Category, error) {
		rows, err := b.service.repo.ListCategories(ctx, data.ListCategoriesParams{
			LoginID:  int64(b.LoginID()),
			BudgetID: int64(b.ID()),
		})
		if err != nil {
			return nil, err
		}
		categories := make([]*Category, len(rows))
		for i, row := range rows {
			var group *CategoryGroup
			if row.GroupID.Valid {
				group = categoryGroupFromRow(ctx, data.CategoryGroup{
					ID:       row.GroupID.Int64,
					BudgetID: row.BudgetID,
					Name:     row.GroupName.String,
				}, b)
			}
			categories[i] = categoryFromRow(ctx, data.Category{
				ID:              row.ID,
				BudgetID:        row.BudgetID,
				Name:            row.Name,
				CategoryGroupID: row.GroupID,
			}, b, group)
		}
		return categories, nil
	})
}

// Update the existing category
func (c *Category) Update(ctx context.Context, name string, group *CategoryGroup) error {
	defer cache.InvalidateResults(ctx)
	groupID := sql.NullInt64{}
	if group != nil {
		groupID = sql.NullInt64{Valid: true, Int64: int64(group.ID())}
	}
	row, err := c.budget.service.repo.UpdateCategory(ctx, data.UpdateCategoryParams{
		Name:            name,
		CategoryGroupID: groupID,
		ID:              int64(c.ID()),
		BudgetID:        int64(c.budget.ID()),
		LoginID:         int64(c.budget.LoginID()),
	})
	if err != nil {
		return err
	}
	c.row = row
	c.group = group
	return nil
}

// Delete a category
func (c *Category) Delete(ctx context.Context) error {
	defer cache.InvalidateResults(ctx)
	defer cache.Delete[*Category](ctx, int64(c.ID()))
	return c.budget.service.repo.DeleteCategory(ctx, data.DeleteCategoryParams{
		ID:       int64(c.ID()),
		BudgetID: int64(c.budget.ID()),
		LoginID:  int64(c.budget.LoginID()),
	})
}

// Get the category's ID
func (c *Category) ID() int {
	return int(c.row.ID)
}

// Get the category's name
func (c *Category) Name() string {
	return c.row.Name
}

// Get the category's group, returns an error if none
func (c *Category) Group() (*CategoryGroup, error) {
	if c.group == nil {
		return nil, fmt.Errorf("Category has no group")
	}
	return c.group, nil
}
