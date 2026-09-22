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
