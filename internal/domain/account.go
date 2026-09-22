package domain 

import (
	"context"
	"fmt"
	"strings"
	"time"

	"samuellando.com/YNAFB/data"
	"samuellando.com/YNAFB/internal/cache"
	"samuellando.com/YNAFB/internal/db/types"
	"samuellando.com/YNAFB/internal/importer/statement"
)

type Account struct {
	service *AccountService
	row data.Account
	budget *Budget
}

func (s *AccountService) fromRow(ctx context.Context, row data.Account, budget *Budget) *Account {
	if cached, ok := cache.Get[*Account](ctx, row.ID); ok {
		return cached
	}
	account := &Account{
		service: s,
		row: row,
		budget: budget,
	}
	cache.Store(ctx, row.ID, account)
	return account
}

func (a *Account) ID() int {
	return int(a.row.ID)
}

func (a *Account) Budget() *Budget {
	return a.budget
}

func (a *Account) Name() string {
	return a.row.Name
}

func (a *Account) Balance(ctx context.Context) (int, error) {
	trxs, err := a.ListTransactions(ctx)
	if err != nil {
		return 0, err
	}
	balance := 0
	for _, trx := range trxs {
		balance += trx.TotalInflow() - trx.TotalOutflow()
	}
	return balance, nil
}

func (a *Account) ReconciledBalance(ctx context.Context) (int, error) {
	trxs, err := a.ListTransactions(ctx)
	if err != nil {
		return 0, err
	}
	balance := 0
	for _, trx := range trxs {
		if trx.Reconciled() {
			balance += trx.TotalInflow() - trx.TotalOutflow()
		}
	}
	return balance, nil
}

func (a *Account) BalanceAsOf(ctx context.Context, asOf time.Time) (int, error) {
	trxs, err := a.ListTransactions(ctx)
	if err != nil {
		return 0, err
	}
	balance := 0
	for _, trx := range trxs {
		if trx.Date().After(asOf) {
			continue
		}
		balance += trx.TotalInflow() - trx.TotalOutflow()
	}
	return balance, nil
}

func (a *Account) Reconcile(ctx context.Context, date time.Time, expectedBalance int) error {
	defer cache.InvalidateResults(ctx)
	balance, err := a.BalanceAsOf(ctx, date)
	if err != nil {
		return err
	}
	if balance != expectedBalance {
		return fmt.Errorf("balance mismatch: statement %d != calculated %d", expectedBalance, balance)
	}
	_, err = a.service.repo.ReconcileAccountTransactions(ctx, data.ReconcileAccountTransactionsParams{
		BudgetID: a.row.BudgetID,
		ID:       a.row.ID,
		LoginID:  int64(a.budget.LoginID()),
		Date:     types.UnixTime{Time: date},
	})
	return err
}

func (a *Account) ImportStatement(ctx context.Context, stmt statement.Statement) (int, error) {
	defer cache.InvalidateResults(ctx)
	loginID := a.budget.LoginID()
	budgetID := int(a.row.BudgetID)
	for _, entry := range stmt.Entries {
		payeeName := strings.TrimSpace(entry.Payee)
		if payeeName == "" {
			payeeName = "unknown"
		}
		payee, err := a.service.payeeService.GetOrCreate(ctx, loginID, budgetID, payeeName)
		if err != nil {
			return 0, err
		}
		if _, err := a.service.trxService.create(ctx, a, payee, entry.TransDate, entry.Outflow, entry.Inflow, entry.Note); err != nil {
			return 0, err
		}
	}
	return len(stmt.Entries), nil
}

func (a *Account) ListTransactions(ctx context.Context) ([]*Trx, error) {
	budgetTrxs, err := a.budget.ListTransactions(ctx)
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
}

func (a *Account) Update(ctx context.Context, name string) error {
	row, err := a.service.repo.UpdateAccount(ctx, data.UpdateAccountParams{
		Name:     name,
		ID:       a.row.ID,
		LoginID:  int64(a.budget.LoginID()),
		BudgetID: a.row.BudgetID,
	})
	if err != nil {
		return err
	}
	a.row = row
	return nil
}

func (a *Account) Delete(ctx context.Context) error {
	defer cache.InvalidateResults(ctx)
	err := a.service.repo.DeleteAccount(ctx, data.DeleteAccountParams{
		ID:       a.row.ID,
		LoginID:  int64(a.budget.LoginID()),
		BudgetID: a.row.BudgetID,
	})
	return err
}
