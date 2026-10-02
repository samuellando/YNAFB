package domain

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"samuellando.com/YNAFB/internal/cache"
	"samuellando.com/YNAFB/internal/data"
	"samuellando.com/YNAFB/internal/db/types"
)

type Trx struct {
	row        data.Trx
	account    *Account
	payee      *Payee
	lines      []*TrxLine
	reconciled bool
	// mirrorSource is nil for real transactions.
	mirrorSource *Account
}

// Load a trx and its lines from a set of rows. Returns the trx and the number of rows consumed
func trxFromRows(ctx context.Context, rows []data.ListTrxsAndLinesRow, budget *Budget) (int, *Trx) {
	if len(rows) == 0 {
		return 0, nil
	}
	n := 0
	trx, _ := cache.Get(ctx, rows[0].Trx.ID, func() (*Trx, error) {
		account := accountFromRow(ctx, data.Account{
			ID:       rows[0].Trx.AccountID,
			BudgetID: rows[0].Trx.BudgetID,
			Name:     rows[0].AccountName,
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
			if mirror, err := a.mirrorTrx(ctx, trx); mirror != nil {
				trxs = append(trxs, mirror)
			} else if err != nil {
				return nil, err
			}
		}
		return trxs, nil
	})
}

// reconciledLineIDs returns the set of trx line ids reconciled in the account.
func (a *Account) reconciledLineIDs(ctx context.Context) (map[int64]bool, error) {
	return cache.Result(ctx, fmt.Sprintf("Reconcilked lines-%d-%d-%d", a.budget.LoginID(), a.budget.ID(), a.ID()), func() (map[int64]bool, error) {
		rows, err := a.budget.service.repo.ListReconciledTrxLines(ctx, data.ListReconciledTrxLinesParams{
			LoginID:  int64(a.budget.LoginID()),
			BudgetID: int64(a.budget.ID()),
			ID:       int64(a.ID()),
		})
		if err != nil {
			return nil, err
		}
		reconciled := make(map[int64]bool, len(rows))
		for _, row := range rows {
			if row.Valid {
				reconciled[row.Int64] = true
			}
		}
		return reconciled, nil
	})
}

// Mirror the trx for this account. Returns nil if the transaction is not a transfer into this account and error if
// any error occurrs.
func (a *Account) mirrorTrx(ctx context.Context, source *Trx) (*Trx, error) {
	if source.account.ID() == a.ID() {
		return nil, nil
	}
	reconciled, err := a.reconciledLineIDs(ctx)
	if err != nil {
		return nil, err
	}
	outflow, inflow := 0, 0
	matched := false
	allReconciled := true
	for _, line := range source.lines {
		destAccount, err := line.DestinationAccount()
		if err != nil || destAccount.ID() != a.ID() {
			continue
		}
		matched = true
		outflow += line.Inflow()
		inflow += line.Outflow()
		if !reconciled[int64(line.ID())] {
			allReconciled = false
		}
	}
	if !matched {
		return nil, nil
	}
	row := source.row
	row.AccountID = int64(a.ID())
	row.TotalOutflow = int64(outflow)
	row.TotalInflow = int64(inflow)
	return &Trx{
		row:          row,
		account:      a,
		payee:        source.payee,
		reconciled:   allReconciled,
		mirrorSource: source.account,
	}, nil
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
	trx, err := cache.Get(ctx, int64(id), func() (*Trx, error) {
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
	if err == nil {
		return trx, nil
	} else if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	return a.getMirrorTransaction(ctx, id)
}

// getMirrorTransaction returns the read-only mirror of the transfer with the
// given source transaction id into this account, or sql.ErrNoRows.
func (a *Account) getMirrorTransaction(ctx context.Context, id int) (*Trx, error) {
	budgetTrxs, err := a.budget.listTransactions(ctx)
	if err != nil {
		return nil, err
	}
	for _, trx := range budgetTrxs {
		if trx.ID() != id {
			continue
		}
		if mirror, err := a.mirrorTrx(ctx, trx); mirror != nil {
			return mirror, nil
		} else if err != nil {
			return nil, err
		}
	}
	return nil, sql.ErrNoRows
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

// Returns true if the transaction is a read-only mirror of a transfer made in
// another account. Mirrors cannot be updated, deleted, or extended.
func (t *Trx) IsMirror() bool {
	return t.mirrorSource != nil
}

// The account the funds were transferred from. Returns an error if the
// transaction is not a mirror.
func (t *Trx) SourceAccount() (*Account, error) {
	if t.mirrorSource != nil {
		return t.mirrorSource, nil
	}
	return nil, fmt.Errorf("Transaction is not a mirror transfer")
}

// Update the transaction
func (t *Trx) Update(ctx context.Context, payee *Payee, date time.Time, outflow, inflow int, note string) error {
	if t.IsMirror() {
		return fmt.Errorf("Mirror transactions cannot be edited")
	}
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
	if t.IsMirror() {
		return fmt.Errorf("Mirror transactions cannot be edited")
	}
	defer cache.InvalidateResults(ctx)
	defer cache.Delete[*Trx](ctx, int64(t.ID()))
	return t.account.budget.service.repo.DeleteTrx(ctx, data.DeleteTrxParams{
		ID:       int64(t.ID()),
		BudgetID: int64(t.account.budget.ID()),
		LoginID:  int64(t.account.budget.LoginID()),
	})
}
