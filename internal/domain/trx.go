package domain

import (
	"context"
	"time"

	"samuellando.com/YNAFB/data"
	"samuellando.com/YNAFB/internal/cache"
	"samuellando.com/YNAFB/internal/db/types"
)

type Trx struct {
	service *TrxService
	row data.Trx
	lines []*TrxLine
	payee *Payee
	budget *Budget
	account *Account
	reconciled bool
}

func (t *Trx) ID() int {
	return int(t.row.ID)
}

func (t *Trx) Date() time.Time {
	return t.row.Date.Time
}

func (t *Trx) Reconciled() bool {
	return t.reconciled
}

func (t *Trx) TotalOutflow() int {
	return int(t.row.TotalOutflow)
}

func (t *Trx) TotalInflow() int {
	return int(t.row.TotalInflow)
}

func (t *Trx) Lines() []*TrxLine {
	return t.lines
}

func (t *Trx) Payee() *Payee {
	return t.payee
}

func (t *Trx) Account() *Account {
	return t.account
}

func (t *Trx) Note() string {
	return t.row.Note
}

func (t *Trx) Update(ctx context.Context, loginID, accountID, payeeID int, date time.Time, outflow, inflow int, note string) error {
	defer cache.InvalidateResults(ctx)
	row, err := t.service.repo.UpdateTrx(ctx, data.UpdateTrxParams{
		Date:         types.UnixTime{Time: date},
		AccountID:    int64(accountID),
		PayeeID:      int64(payeeID),
		TotalOutflow: int64(outflow),
		TotalInflow:  int64(inflow),
		Note:         note,
		ID:           t.row.ID,
		BudgetID:     t.row.BudgetID,
		LoginID:      int64(loginID),
	})
	if err != nil {
		return err
	}
	t.row = row
	// AccountID/PayeeID read through the refs, so refresh them when they moved.
	if t.account == nil || t.account.ID() != accountID {
		account, err := t.service.accountService.Get(ctx, loginID, int(t.row.BudgetID), accountID)
		if err != nil {
			return err
		}
		t.account = account
		t.budget = account.Budget()
	}
	if t.payee == nil || t.payee.ID() != payeeID {
		payee, err := t.service.payeeService.Get(ctx, loginID, int(t.row.BudgetID), payeeID)
		if err != nil {
			return err
		}
		t.payee = payee
	}
	return nil
}

func (t *Trx) Delete(ctx context.Context, loginID int) error {
	defer cache.InvalidateResults(ctx)
	return t.service.repo.DeleteTrx(ctx, data.DeleteTrxParams{
		ID:       t.row.ID,
		BudgetID: t.row.BudgetID,
		LoginID:  int64(loginID),
	})
}
