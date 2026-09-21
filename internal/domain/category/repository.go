package category

import (
	"context"
	"samuellando.com/YNAFB/data"
)

type Repository interface {
	ListCategories(context.Context, data.ListCategoriesParams) ([]data.ListCategoriesRow, error)
}
