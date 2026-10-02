package domain

import (
	"context"
	"fmt"
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
	return nil, fmt.Errorf("expense share not implemented")
}

func (t *ExpenseShareTransaction) AddCategorization(ctx context.Context, category *Category, outflow, inflow int) (*SplitCategorization, error) {
	return nil, fmt.Errorf("expense share not implemented")
}

func (t *ExpenseShareTransaction) GetCategorization(ctx context.Context, lineID int) (*SplitCategorization, error) {
	return nil, fmt.Errorf("expense share not implemented")
}
