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
			group = s.groupFromRow(ctx, data.CategoryGroup{
				ID:       row.GroupID.Int64,
				BudgetID: row.BudgetID,
				Name:     row.GroupName.String,
			})
		}
	categories[i] = s.fromRow(ctx, data.Category{
		ID:              row.ID,
		BudgetID:        row.BudgetID,
		Name:            row.Name,
		CategoryGroupID: row.GroupID,
	}, group)
	categories[i].service = s
	}
	return categories, nil
}

func (s *CategoryService) ListGroups(ctx context.Context, loginID, budgetID int) ([]*Group, error) {
	return cache.Result(ctx, fmt.Sprintf("categoryServiceListGroups-%d-%d", loginID, budgetID), func() ([]*Group, error) {
		return s.listGroups(ctx, loginID, budgetID)
	})
}

func (s *CategoryService) listGroups(ctx context.Context, loginID, budgetID int) ([]*Group, error) {
	rows, err := s.repo.ListCategoryGroups(ctx, data.ListCategoryGroupsParams{
		LoginID:  int64(loginID),
		BudgetID: int64(budgetID),
	})
	if err != nil {
		return nil, err
	}
	groups := make([]*Group, len(rows))
	for i, row := range rows {
		groups[i] = s.groupFromRow(ctx, row)
	}
	return groups, nil
}

func (s *CategoryService) CreateGroup(ctx context.Context, loginID, budgetID int, name string) (*Group, error) {
	defer cache.InvalidateResults(ctx)
	id, err := s.repo.CreateCategoryGroup(ctx, data.CreateCategoryGroupParams{
		Name:     name,
		BudgetID: int64(budgetID),
		LoginID:  int64(loginID),
	})
	if err != nil {
		return nil, err
	}
	group := s.groupFromRow(ctx, data.CategoryGroup{
		ID:       id,
		BudgetID: int64(budgetID),
		Name:     name,
	})
	cache.Store(ctx, group.row.ID, group)
	return group, nil
}

func (s *CategoryService) GetGroup(ctx context.Context, loginID, budgetID, groupID int) (*Group, error) {
	if v, ok := cache.Get[*Group](ctx, int64(groupID)); ok {
		return v, nil
	}
	row, err := s.repo.GetCategoryGroup(ctx, data.GetCategoryGroupParams{
		LoginID:  int64(loginID),
		BudgetID: int64(budgetID),
		ID:       int64(groupID),
	})
	if err != nil {
		return nil, err
	}
	group := s.groupFromRow(ctx, row)
	cache.Store(ctx, group.row.ID, group)
	return group, nil
}
