package domain

import (
	"context"
	"samuellando.com/YNAFB/data"
)

type CategoryRepository interface {
	ListCategories(context.Context, data.ListCategoriesParams) ([]data.ListCategoriesRow, error)
}
