package domain

import (
	"context"
	"fmt"

	"samuellando.com/YNAFB/data"
	"samuellando.com/YNAFB/internal/cache"
)

type CategoryService struct {
	repo CategoryRepository
}

func NewCategoryService(repo CategoryRepository) *CategoryService {
	return &CategoryService{repo: repo}
}

func (s *CategoryService) List(ctx context.Context, loginID, budgetID int) ([]*Category, error) {
	return cache.Result(ctx, fmt.Sprintf("categoryServiceList-%d-%d", loginID, budgetID), func() ([]*Category, error) {
		return s.list(ctx, loginID, budgetID)
	})
}

func (s *CategoryService) list(ctx context.Context, loginID, budgetID int) ([]*Category, error) {
	rows, err := s.repo.ListCategories(ctx, data.ListCategoriesParams{
		BudgetID: int64(budgetID),
		LoginID:  int64(loginID),
	})
	if err != nil {
		return nil, err
	}
	categories := make([]*Category, len(rows))
	for i, row := range rows {
		var group *Group
		if row.GroupID.Valid {
			group = s.GroupFromRow(ctx, data.CategoryGroup{
				ID:       row.GroupID.Int64,
				BudgetID: row.BudgetID,
				Name:     row.GroupName.String,
			})
		}
		categories[i] = s.FromRow(ctx, data.Category{
			ID:              row.ID,
			BudgetID:        row.BudgetID,
			Name:            row.Name,
			CategoryGroupID: row.GroupID,
		}, group)
		categories[i].service = s
	}
	return categories, nil
}
