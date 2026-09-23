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
	row    data.Account
	budget *Budget
}

func accountFromRow(ctx context.Context, row data.Account, budget *Budget) *Account {
	account, _ := cache.Get(ctx, row.ID, func() (*Account, error) {
		return &Account{
			row:    row,
			budget: budget,
		}, nil
	})
	return account
}

// List all the accounts in the budget
func (b *Budget) ListAccounts(ctx context.Context) ([]*Account, error) {
	return cache.Result(ctx, fmt.Sprintf("accountServiceList-%d-%d", b.LoginID(), b.ID()), func() ([]*Account, error) {
		rows, err := b.service.repo.ListAccounts(ctx, data.ListAccountsParams{
			LoginID:  int64(b.LoginID()),
			BudgetID: int64(b.ID()),
		})
		if err != nil {
			return nil, err
		}
		accounts := make([]*Account, len(rows))
		for i, row := range rows {
			accounts[i] = accountFromRow(ctx, data.Account{
				ID:       row.ID,
				BudgetID: row.BudgetID,
				Name:     row.Name,
			}, b)
		}
		return accounts, nil
	})
}

// Create a new account
func (b *Budget) CreateAccount(ctx context.Context, name string) (*Account, error) {
	defer cache.InvalidateResults(ctx)
	row, err := b.service.repo.CreateAccount(ctx, data.CreateAccountParams{
		Name:     name,
		LoginID:  int64(b.LoginID()),
		BudgetID: int64(b.ID()),
	})
	if err != nil {
		return nil, err
	}
	account := accountFromRow(ctx, data.Account{
		ID:       row.ID,
		BudgetID: row.BudgetID,
		Name:     row.Name,
	}, b)
	return account, nil
}

// Get an existing account by ID
func (b *Budget) GetAccount(ctx context.Context, accountID int) (*Account, error) {
	return cache.Get(ctx, int64(accountID), func() (*Account, error) {
		row, err := b.service.repo.GetAccount(ctx, data.GetAccountParams{
			LoginID:  int64(b.LoginID()),
			BudgetID: int64(b.ID()),
			ID:       int64(accountID),
		})
		if err != nil {
			return nil, err
		}
		return accountFromRow(ctx, data.Account{
			ID:       row.ID,
			BudgetID: row.BudgetID,
			Name:     row.Name,
		}, b), nil
	})
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
	_, err = a.budget.service.repo.ReconcileAccountTransactions(ctx, data.ReconcileAccountTransactionsParams{
		BudgetID: int64(a.budget.ID()),
		ID:       int64(a.ID()),
		LoginID:  int64(a.budget.LoginID()),
		Date:     types.UnixTime{Time: date},
	})
	return err
}

func (a *Account) ImportStatement(ctx context.Context, stmt statement.Statement) (int, error) {
	defer cache.InvalidateResults(ctx)
	for _, entry := range stmt.Entries {
		payeeName := strings.TrimSpace(entry.Payee)
		if payeeName == "" {
			payeeName = "unknown"
		}
		payee, err := a.budget.getOrCreatePayee(ctx, payeeName)
		if err != nil {
			return 0, err
		}
		if _, err := a.CreateTransaction(ctx, payee, entry.TransDate, int(entry.Outflow), int(entry.Inflow), entry.Note); err != nil {
			return 0, err
		}
	}
	return len(stmt.Entries), nil
}

func (a *Account) Update(ctx context.Context, name string) error {
	defer cache.InvalidateResults(ctx)
	row, err := a.budget.service.repo.UpdateAccount(ctx, data.UpdateAccountParams{
		Name:     name,
		ID:       int64(a.ID()),
		LoginID:  int64(a.budget.LoginID()),
		BudgetID: int64(a.budget.ID()),
	})
	if err != nil {
		return err
	}
	a.row = row
	return nil
}

func (a *Account) Delete(ctx context.Context) error {
	defer cache.InvalidateResults(ctx)
	defer cache.Delete[*Account](ctx, int64(a.ID()))
	return a.budget.service.repo.DeleteAccount(ctx, data.DeleteAccountParams{
		ID:       int64(a.ID()),
		LoginID:  int64(a.budget.LoginID()),
		BudgetID: int64(a.budget.ID()),
	})
}
