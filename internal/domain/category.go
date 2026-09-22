package domain

import (
	"context"

	"samuellando.com/YNAFB/data"
	"samuellando.com/YNAFB/internal/cache"
)

type Category struct {
	service *CategoryService
	row     data.Category
	group   *Group
}

func (s *CategoryService) fromRow(ctx context.Context, row data.Category, group *Group) *Category {
	if cached, ok := cache.Get[*Category](ctx, row.ID); ok {
		return cached
	}
	category := &Category{
		service: s,
		row:     row,
		group:   group,
	}
	cache.Store(ctx, row.ID, category)
	return category
}

func (c *Category) ID() int {
	return int(c.row.ID)
}

func (c *Category) Name() string {
	return c.row.Name
}

func (c *Category) Group() *Group {
	return c.group
}

func (c *Category) GroupID() *int {
	if !c.row.CategoryGroupID.Valid {
		return nil
	}
	id := int(c.row.CategoryGroupID.Int64)
	return &id
}

// Update renames the category and/or moves it to another group (nil clears
// the group). loginID is passed explicitly because, unlike Account, Category
// holds no budget reference to derive it from.
func (c *Category) Update(ctx context.Context, loginID int, name string, groupID *int) error {
	defer cache.InvalidateResults(ctx)
	row, err := c.service.repo.UpdateCategory(ctx, data.UpdateCategoryParams{
		Name:            name,
		CategoryGroupID: nullInt64FromInt(groupID),
		ID:              c.row.ID,
		BudgetID:        c.row.BudgetID,
		LoginID:         int64(loginID),
	})
	if err != nil {
		return err
	}
	group, err := c.service.resolveGroup(ctx, loginID, int(c.row.BudgetID), row.CategoryGroupID)
	if err != nil {
		return err
	}
	c.row = row
	c.group = group
	return nil
}

// Delete removes the category. loginID is passed explicitly, see Update.
func (c *Category) Delete(ctx context.Context, loginID int) error {
	defer cache.InvalidateResults(ctx)
	return c.service.repo.DeleteCategory(ctx, data.DeleteCategoryParams{
		ID:       c.row.ID,
		BudgetID: c.row.BudgetID,
		LoginID:  int64(loginID),
	})
}

type Group struct {
	service *CategoryService
	row     data.CategoryGroup
}

func (s *CategoryService) groupFromRow(ctx context.Context, row data.CategoryGroup) *Group {
	if cached, ok := cache.Get[*Group](ctx, row.ID); ok {
		return cached
	}
	group := &Group{
		service: s,
		row:     row,
	}
	cache.Store(ctx, row.ID, group)
	return group
}

func (g *Group) ID() int {
	return int(g.row.ID)
}

func (g *Group) Name() string {
	return g.row.Name
}

// Update renames the group. loginID is passed explicitly because,
// unlike Account, Group holds no budget reference to derive it from.
func (g *Group) Update(ctx context.Context, loginID int, name string) error {
	defer cache.InvalidateResults(ctx)
	row, err := g.service.repo.UpdateCategoryGroup(ctx, data.UpdateCategoryGroupParams{
		Name:     name,
		ID:       g.row.ID,
		BudgetID: g.row.BudgetID,
		LoginID:  int64(loginID),
	})
	if err != nil {
		return err
	}
	g.row = row
	return nil
}

// Delete removes the group. loginID is passed explicitly, see Update.
func (g *Group) Delete(ctx context.Context, loginID int) error {
	defer cache.InvalidateResults(ctx)
	return g.service.repo.DeleteCategoryGroup(ctx, data.DeleteCategoryGroupParams{
		ID:       g.row.ID,
		BudgetID: g.row.BudgetID,
		LoginID:  int64(loginID),
	})
}
