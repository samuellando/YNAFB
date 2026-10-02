package domain

import (
	"context"
	"database/sql"
	"fmt"

	"samuellando.com/YNAFB/internal/cache"
	"samuellando.com/YNAFB/internal/data"
)

type SplitCategorization struct {
	row      data.ExpenseShareTrxSplitLine
	trx      *ExpenseShareTransaction
	category *Category
}

func splitCategorizationFromRow(ctx context.Context, row data.ExpenseShareTrxSplitLine, trx *ExpenseShareTransaction, category *Category) *SplitCategorization {
	cat, _ := cache.Get(ctx, row.ID, func() (*SplitCategorization, error) {
		return &SplitCategorization{
			row:      row,
			trx:      trx,
			category: category,
		}, nil
	})
	return cat
}

// Hydrate a categorization from a wide row carrying the category join.
func splitCategorizationFromListRow(ctx context.Context, row data.ExpenseShareTrxSplitLine, trx *ExpenseShareTransaction, categoryName sql.NullString, groupID sql.NullInt64, groupName sql.NullString) *SplitCategorization {
	return splitCategorizationFromRow(ctx, row, trx, categorizationCategoryFromRow(ctx, row, trx, categoryName, groupID, groupName))
}

func splitCategorizationFromGetRow(ctx context.Context, row data.GetSplitLineRow, trx *ExpenseShareTransaction) *SplitCategorization {
	return splitCategorizationFromListRow(ctx, row.ExpenseShareTrxSplitLine, trx, row.CategoryName, row.CategoryGroupID, row.CategoryGroupName)
}

func categorizationCategoryFromRow(ctx context.Context, row data.ExpenseShareTrxSplitLine, trx *ExpenseShareTransaction, categoryName sql.NullString, groupID sql.NullInt64, groupName sql.NullString) *Category {
	if !row.CategoryID.Valid {
		return nil
	}
	owner := trx.membership.budget
	var group *CategoryGroup
	if groupID.Valid {
		group = categoryGroupFromRow(ctx, data.CategoryGroup{
			ID:       groupID.Int64,
			BudgetID: row.BudgetID,
			Name:     groupName.String,
		}, owner)
	}
	return categoryFromRow(ctx, data.Category{
		ID:              row.CategoryID.Int64,
		BudgetID:        row.BudgetID,
		Name:            categoryName.String,
		CategoryGroupID: groupID,
	}, owner, group)
}

func (s *SplitCategorization) ID() int {
	return int(s.row.ID)
}

func (s *SplitCategorization) Category() (*Category, error) {
	if s.row.CategoryID.Valid {
		return s.category, nil
	}
	return nil, fmt.Errorf("Split categorization has no category")
}

func (s *SplitCategorization) Outflow() int {
	return int(s.row.Outflow)
}

func (s *SplitCategorization) Inflow() int {
	return int(s.row.Inflow)
}

func (s *SplitCategorization) Transaction() *ExpenseShareTransaction {
	return s.trx
}

func (s *SplitCategorization) Update(ctx context.Context, category *Category, outflow, inflow int) error {
	defer cache.InvalidateResults(ctx)
	categoryID := sql.NullInt64{}
	if category != nil {
		categoryID = sql.NullInt64{Valid: true, Int64: int64(category.ID())}
	}
	row, err := s.trx.membership.budget.service.repo.UpdateSplitLine(ctx, data.UpdateSplitLineParams{
		CategoryID:     categoryID,
		Outflow:        int64(outflow),
		Inflow:         int64(inflow),
		ID:             int64(s.ID()),
		BudgetID:       s.row.BudgetID,
		ExpenseShareID: s.row.ExpenseShareID,
		LoginID:        int64(s.trx.membership.budget.LoginID()),
	})
	if err != nil {
		return err
	}
	s.row = row
	s.category = category
	return nil
}

func (s *SplitCategorization) Delete(ctx context.Context) error {
	defer cache.InvalidateResults(ctx)
	defer cache.Delete[*SplitCategorization](ctx, int64(s.ID()))
	return s.trx.membership.budget.service.repo.DeleteSplitLine(ctx, data.DeleteSplitLineParams{
		ID:             int64(s.ID()),
		BudgetID:       s.row.BudgetID,
		ExpenseShareID: s.row.ExpenseShareID,
		LoginID:        int64(s.trx.membership.budget.LoginID()),
	})
}
