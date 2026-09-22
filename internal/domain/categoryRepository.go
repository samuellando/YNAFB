package domain

import (
	"context"
	"samuellando.com/YNAFB/data"
)

type CategoryRepository interface {
	ListCategories(context.Context, data.ListCategoriesParams) ([]data.ListCategoriesRow, error)
	GetCategory(context.Context, data.GetCategoryParams) (data.Category, error)
	CreateCategory(context.Context, data.CreateCategoryParams) (data.Category, error)
	UpdateCategory(context.Context, data.UpdateCategoryParams) (data.Category, error)
	DeleteCategory(context.Context, data.DeleteCategoryParams) error
	ListCategoryGroups(context.Context, data.ListCategoryGroupsParams) ([]data.CategoryGroup, error)
	GetCategoryGroup(context.Context, data.GetCategoryGroupParams) (data.CategoryGroup, error)
	CreateCategoryGroup(context.Context, data.CreateCategoryGroupParams) (int64, error)
	UpdateCategoryGroup(context.Context, data.UpdateCategoryGroupParams) (data.CategoryGroup, error)
	DeleteCategoryGroup(context.Context, data.DeleteCategoryGroupParams) error
}
