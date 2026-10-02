package domain

import (
	"context"
	"database/sql"
	"fmt"

	"samuellando.com/YNAFB/internal/cache"
	"samuellando.com/YNAFB/internal/data"
)

// ExpenseShareTransaction is a read-only live view of a source transaction
// carrying split or settlement lines in a share. Like a mirror transaction it
// is never stored and never goes through cache.Get; totals, payee, date and
// note resolve live from the source.
type ExpenseShareTransaction struct {
	membership   *ExpenseShareMembership
	sourceBudget *Budget
	sourceTrx    *Trx
	// How the source member is displayed to other members, hydrated by
	// the share-scoped query that builds this view.
	sourceDisplayName string
	categorizations   []*SplitCategorization
}

func expenseShareTransactionFromSource(membership *ExpenseShareMembership, sourceBudget *Budget, sourceTrx *Trx, sourceDisplayName string) *ExpenseShareTransaction {
	return &ExpenseShareTransaction{
		membership:        membership,
		sourceBudget:      sourceBudget,
		sourceTrx:         sourceTrx,
		sourceDisplayName: sourceDisplayName,
	}
}

func (t *ExpenseShareTransaction) SourceBudget() *Budget {
	return t.sourceBudget
}

func (t *ExpenseShareTransaction) SourceBudgetDisplayName() string {
	return t.sourceDisplayName
}

func (t *ExpenseShareTransaction) PayeeName() string {
	return t.sourceTrx.Payee().Name()
}

func (t *ExpenseShareTransaction) Date() string {
	return t.sourceTrx.Date().Format("2006-01-02T15:04:05Z07:00")
}

func (t *ExpenseShareTransaction) TotalOutflow() int {
	return t.sourceTrx.TotalOutflow()
}

func (t *ExpenseShareTransaction) TotalInflow() int {
	return t.sourceTrx.TotalInflow()
}

func (t *ExpenseShareTransaction) Note() string {
	return t.sourceTrx.Note()
}

func (t *ExpenseShareTransaction) Membership() *ExpenseShareMembership {
	return t.membership
}

func (t *ExpenseShareTransaction) Share() *ExpenseShare {
	return t.membership.Share()
}

func (t *ExpenseShareTransaction) SourceTrx() *Trx {
	return t.sourceTrx
}

func (t *ExpenseShareTransaction) MyCategorizations(ctx context.Context) ([]*SplitCategorization, error) {
	return cache.Result(ctx, fmt.Sprintf("shareCategorizations-%d-%d-%d-%d", t.membership.budget.LoginID(), t.membership.budget.ID(), t.membership.share.ID(), t.sourceTrx.ID()), func() ([]*SplitCategorization, error) {
		rows, err := t.membership.budget.service.repo.ListCategorizations(ctx, data.ListCategorizationsParams{
			LoginID:        int64(t.membership.budget.LoginID()),
			BudgetID:       int64(t.membership.budget.ID()),
			ExpenseShareID: int64(t.membership.share.ID()),
			SourceTrxID:    int64(t.sourceTrx.ID()),
		})
		if err != nil {
			return nil, err
		}
		if len(rows) == 0 {
			return nil, nil
		}
		cats := make([]*SplitCategorization, len(rows))
		for i, row := range rows {
			cats[i] = splitCategorizationFromListRow(ctx, row.ExpenseShareTrxSplitLine, t, row.CategoryName, row.CategoryGroupID, row.CategoryGroupName)
		}
		return cats, nil
	})
}

func (t *ExpenseShareTransaction) AddCategorization(ctx context.Context, category *Category, outflow, inflow int) (*SplitCategorization, error) {
	defer cache.InvalidateResults(ctx)
	categoryID := sql.NullInt64{}
	if category != nil {
		categoryID = sql.NullInt64{Valid: true, Int64: int64(category.ID())}
	}
	row, err := t.membership.budget.service.repo.CreateSplitLine(ctx, data.CreateSplitLineParams{
		ExpenseShareID: int64(t.membership.share.ID()),
		SourceBudgetID: int64(t.sourceBudget.ID()),
		SourceTrxID:    int64(t.sourceTrx.ID()),
		CategoryID:     categoryID,
		Outflow:        int64(outflow),
		Inflow:         int64(inflow),
		BudgetID:       int64(t.membership.budget.ID()),
		LoginID:        int64(t.membership.budget.LoginID()),
	})
	if err != nil {
		return nil, err
	}
	return splitCategorizationFromRow(ctx, row, t, category), nil
}

func (t *ExpenseShareTransaction) GetCategorization(ctx context.Context, lineID int) (*SplitCategorization, error) {
	return cache.Get(ctx, int64(lineID), func() (*SplitCategorization, error) {
		row, err := t.membership.budget.service.repo.GetSplitLine(ctx, data.GetSplitLineParams{
			LoginID:        int64(t.membership.budget.LoginID()),
			BudgetID:       int64(t.membership.budget.ID()),
			ExpenseShareID: int64(t.membership.share.ID()),
			ID:             int64(lineID),
		})
		if err != nil {
			return nil, err
		}
		return splitCategorizationFromGetRow(ctx, row, t), nil
	})
}
