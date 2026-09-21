package category

import (
	"context"

	"samuellando.com/YNAFB/data"
	"samuellando.com/YNAFB/internal/cache"
)

type Category struct {
	service *Service
	row     data.Category
	group   *Group
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

type Group struct {
	service *Service
	row     data.CategoryGroup
}

func (g *Group) ID() int {
	return int(g.row.ID)
}

func (g *Group) Name() string {
	return g.row.Name
}

func (s *Service) GroupFromRow(ctx context.Context, row data.CategoryGroup) *Group {
	if cached, ok := cache.Get[*Group](ctx, "categoryGroup", row.ID); ok {
		return cached
	}
	group := &Group{
		service: s,
		row: row,
	}
	cache.Store(ctx, "categoryGroup", row.ID, group)
	return group
}

func (s *Service) FromRow(ctx context.Context, row data.Category, group *Group) *Category {
	if cached, ok := cache.Get[*Category](ctx, "category", row.ID); ok {
		return cached
	}
	category := &Category{
		service: s,
		row:   row,
		group: group,
	}
	cache.Store(ctx, "category", row.ID, category)
	return category
}
