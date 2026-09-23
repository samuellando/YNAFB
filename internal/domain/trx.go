package domain

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"samuellando.com/YNAFB/internal/data"
	"samuellando.com/YNAFB/internal/cache"
	"samuellando.com/YNAFB/internal/db/types"
)

type Trx struct {
	row        data.Trx
	account    *Account
	payee      *Payee
	lines      []*TrxLine
	reconciled bool
}

// Load a trx and its lines from a set of rows. Returns the trx and the number of rows consumed
func trxFromRows(ctx context.Context, rows []data.ListTrxsAndLinesRow, budget *Budget) (int, *Trx) {
	if len(rows) == 0 {
		return 0, nil
	}
	n := 0
	trx, _ := cache.Get(ctx, rows[0].Trx.ID, func() (*Trx, error) {
		account := accountFromRow(ctx, data.Account{
			ID:   rows[0].Trx.AccountID,
			BudgetID: rows[0].Trx.BudgetID,
			Name: rows[0].AccountName,
		}, budget)
		trx := &Trx{
			row:     rows[0].Trx,
			account: account,
			payee: payeeFromRow(ctx,
				data.Payee{
					ID:       rows[0].Trx.PayeeID,
					BudgetID: rows[0].Trx.BudgetID,
					Name:     rows[0].PayeeName,
				}, budget),
			reconciled: rows[0].Reconciled,
		}
		for _, row := range rows {
			if row.Trx.ID != trx.row.ID {
				return trx, nil
			}
			n += 1
			line := trxLineFromRow(ctx, row, trx)
			if line != nil {
				trx.lines = append(trx.lines, line)
			}
		}
		return trx, nil
	})
	// Cache hit, count the rows belonging to the cached trx
	if n == 0 {
		for _, row := range rows {
			if row.Trx.ID != int64(trx.ID()) {
				break
			}
			n += 1
		}
	}
	return n, trx
}

// List all the transactions in the budget
func (b *Budget) listTransactions(ctx context.Context) ([]*Trx, error) {
	return cache.Result(ctx, fmt.Sprintf("budgetTrxList-%d-%d", b.LoginID(), b.ID()), func() ([]*Trx, error) {
		rows, err := b.service.repo.ListTrxsAndLines(ctx, data.ListTrxsAndLinesParams{
			LoginID:  int64(b.LoginID()),
			BudgetID: int64(b.ID()),
		})
		if err != nil {
			return nil, err
		}
		if len(rows) == 0 {
			return nil, nil
		}
		trxs := make([]*Trx, 0)
		n := 0
		for n < len(rows) {
			consumed, trx := trxFromRows(ctx, rows[n:], b)
			trxs = append(trxs, trx)
			n += consumed
		}
		return trxs, nil
	})
}

// List all the transactions in an account
func (a *Account) ListTransactions(ctx context.Context) ([]*Trx, error) {
	return cache.Result(ctx, fmt.Sprintf("accountTrxList-%d-%d-%d", a.budget.LoginID(), a.budget.ID(), a.ID()), func() ([]*Trx, error) {
		budgetTrxs, err := a.budget.listTransactions(ctx)
		if err != nil {
			return nil, err
		}
		trxs := make([]*Trx, 0)
		for _, trx := range budgetTrxs {
			if trx.account == a {
				trxs = append(trxs, trx)
			}
		}
		return trxs, nil
	})
}

// Create a new transaction in an account
func (a *Account) CreateTransaction(ctx context.Context, payee *Payee, date time.Time, outflow, inflow int, note string) (*Trx, error) {
	defer cache.InvalidateResults(ctx)
	row, err := a.budget.service.repo.CreateTrx(ctx, data.CreateTrxParams{
		LoginID:      int64(a.budget.LoginID()),
		BudgetID:     int64(a.budget.ID()),
		Date:         types.UnixTime{Time: date},
		AccountID:    int64(a.ID()),
		PayeeID:      int64(payee.ID()),
		TotalOutflow: int64(outflow),
		TotalInflow:  int64(inflow),
		Note:         note,
	})
	if err != nil {
		return nil, err
	}
	_, trx := trxFromRows(ctx, []data.ListTrxsAndLinesRow{{
		Trx:         row,
		AccountName: a.Name(),
		PayeeName:   payee.Name(),
		Reconciled:  false,
	}}, a.budget)
	return trx, nil
}

// Get an accounts transaction
func (a *Account) GetTransaction(ctx context.Context, id int) (*Trx, error) {
	return cache.Get(ctx, int64(id), func() (*Trx, error) {
		rows, err := a.budget.service.repo.GetTrxAndLines(ctx, data.GetTrxAndLinesParams{
			LoginID:   int64(a.budget.LoginID()),
			BudgetID:  int64(a.budget.ID()),
			AccountID: int64(a.ID()),
			ID:        int64(id),
		})
		if err != nil {
			return nil, err
		}
		_, trx := trxFromRows(ctx, asListRows(rows), a.budget)
		if trx == nil {
			return nil, sql.ErrNoRows
		}
		return trx, nil
	})
}

// Helper method for converting get account rows to trxFromRows rows
func asListRows(rows []data.GetTrxAndLinesRow) []data.ListTrxsAndLinesRow {
	out := make([]data.ListTrxsAndLinesRow, len(rows))
	for i, r := range rows {
		out[i] = data.ListTrxsAndLinesRow{
			Trx:               r.Trx,
			AccountName:       r.AccountName,
			PayeeName:         r.PayeeName,
			LineID:            r.LineID,
			LineIncome:        r.LineIncome,
			LineInflow:        r.LineInflow,
			LineOutflow:       r.LineOutflow,
			CategoryID:        r.CategoryID,
			CategoryName:      r.CategoryName,
			CategoryGroupID:   r.CategoryGroupID,
			CategoryGroupName: r.CategoryGroupName,
			DestAccountID:     r.DestAccountID,
			DestAccountName:   r.DestAccountName,
			Reconciled:        r.Reconciled,
		}
	}
	return out
}

// Get the id of the transaction
func (t *Trx) ID() int {
	return int(t.row.ID)
}

// Get the date of the transaction
func (t *Trx) Date() time.Time {
	return t.row.Date.Time
}

// Returns true if the transaction has been reconciled
func (t *Trx) Reconciled() bool {
	return t.reconciled
}

// Get the total outflow of the transaction
func (t *Trx) TotalOutflow() int {
	return int(t.row.TotalOutflow)
}

// Get the total inflow of the transaction
func (t *Trx) TotalInflow() int {
	return int(t.row.TotalInflow)
}

// Get the payee of the transaction
func (t *Trx) Payee() *Payee {
	return t.payee
}

// Get the account of the transaction
func (t *Trx) Account() *Account {
	return t.account
}

// Get the note on the transaction
func (t *Trx) Note() string {
	return t.row.Note
}

// Get a list of the transaction's lines if any
func (t *Trx) Lines() []*TrxLine {
	return t.lines
}

// Update the transaction
func (t *Trx) Update(ctx context.Context, payee *Payee, date time.Time, outflow, inflow int, note string) error {
	defer cache.InvalidateResults(ctx)
	row, err := t.account.budget.service.repo.UpdateTrx(ctx, data.UpdateTrxParams{
		Date:         types.UnixTime{Time: date},
		PayeeID:      int64(payee.ID()),
		TotalOutflow: int64(outflow),
		TotalInflow:  int64(inflow),
		Note:         note,
		ID:           int64(t.ID()),
		BudgetID:     int64(t.account.budget.ID()),
		LoginID:      int64(t.account.budget.LoginID()),
	})
	if err != nil {
		return err
	}
	t.row = row
	t.payee = payee
	return nil
}

// Delete the transaction
func (t *Trx) Delete(ctx context.Context) error {
	defer cache.InvalidateResults(ctx)
	defer cache.Delete[*Trx](ctx, int64(t.ID()))
	return t.account.budget.service.repo.DeleteTrx(ctx, data.DeleteTrxParams{
		ID:       int64(t.ID()),
		BudgetID: int64(t.account.budget.ID()),
		LoginID:  int64(t.account.budget.LoginID()),
	})
}
