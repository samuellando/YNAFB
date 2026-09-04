package data_test

import (
	"database/sql"
	"errors"
	"testing"

	"samuellando.com/YNAFB/data"
)

func TestListAccountsBalancesEmpty(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget, err := queries.CreateBudget(ctx, "testBudget")
	if err != nil {
		t.Fatal(err)
	}
	balances, err := queries.ListAccountsBalances(ctx, budget.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(balances) != 0 {
		t.Fatalf("expected no balances, got %d", len(balances))
	}
}

func TestListAccountsBalancesNoTransactions(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget, err := queries.CreateBudget(ctx, "testBudget")
	if err != nil {
		t.Fatal(err)
	}
	account, err := queries.CreateAccount(ctx, data.CreateAccountParams{
		Budget: budget.ID,
		Name:   "testaccount",
	})
	if err != nil {
		t.Fatal(err)
	}
	balances, err := queries.ListAccountsBalances(ctx, budget.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(balances) != 1 {
		t.Fatalf("expected one balance, got %d", len(balances))
	}
	if balances[0].ID != account.ID {
		t.Error("account id does not match")
	}
	if balances[0].Name != "testaccount" {
		t.Error("account name does not match")
	}
	if balances[0].Balance != 0 {
		t.Errorf("expected zero balance, got %d", balances[0].Balance)
	}
	if balances[0].ReconciledBalance != 0 {
		t.Errorf("expected zero reconciled balance, got %d", balances[0].ReconciledBalance)
	}
}

func TestListAccountsBalancesMultipleAccounts(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget, err := queries.CreateBudget(ctx, "testBudget")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := queries.CreateAccount(ctx, data.CreateAccountParams{
		Budget: budget.ID,
		Name:   "account1",
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := queries.CreateAccount(ctx, data.CreateAccountParams{
		Budget: budget.ID,
		Name:   "account2",
	}); err != nil {
		t.Fatal(err)
	}
	balances, err := queries.ListAccountsBalances(ctx, budget.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(balances) != 2 {
		t.Fatalf("expected two balances, got %d", len(balances))
	}
	byName := map[string]data.ListAccountsBalancesRow{}
	for _, b := range balances {
		byName[b.Name] = b
	}
	for _, name := range []string{"account1", "account2"} {
		b, ok := byName[name]
		if !ok {
			t.Fatalf("expected balance for account %q", name)
		}
		if b.Balance != 0 || b.ReconciledBalance != 0 {
			t.Errorf("account %q should have zero balances, got %d/%d", name, b.Balance, b.ReconciledBalance)
		}
	}
}

func TestListAccountsBalancesScopedToBudget(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget1, err := queries.CreateBudget(ctx, "testBudget1")
	if err != nil {
		t.Fatal(err)
	}
	budget2, err := queries.CreateBudget(ctx, "testBudget2")
	if err != nil {
		t.Fatal(err)
	}
	account1, err := queries.CreateAccount(ctx, data.CreateAccountParams{
		Budget: budget1.ID,
		Name:   "same",
	})
	if err != nil {
		t.Fatal(err)
	}
	account2, err := queries.CreateAccount(ctx, data.CreateAccountParams{
		Budget: budget2.ID,
		Name:   "same",
	})
	if err != nil {
		t.Fatal(err)
	}
	balances1, err := queries.ListAccountsBalances(ctx, budget1.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(balances1) != 1 {
		t.Fatalf("expected one balance for budget1, got %d", len(balances1))
	}
	if balances1[0].ID != account1.ID {
		t.Error("budget1 returned the wrong account")
	}
	balances2, err := queries.ListAccountsBalances(ctx, budget2.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(balances2) != 1 {
		t.Fatalf("expected one balance for budget2, got %d", len(balances2))
	}
	if balances2[0].ID != account2.ID {
		t.Error("budget2 returned the wrong account")
	}
}

func TestListAccountsBalancesCategorySpend(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget, err := queries.CreateBudget(ctx, "testBudget")
	if err != nil {
		t.Fatal(err)
	}
	account, err := queries.CreateAccount(ctx, data.CreateAccountParams{
		Budget: budget.ID,
		Name:   "testaccount",
	})
	if err != nil {
		t.Fatal(err)
	}
	payee, err := queries.CreatePayee(ctx, data.CreatePayeeParams{
		Budget: budget.ID,
		Name:   "testpayee",
	})
	if err != nil {
		t.Fatal(err)
	}
	category, err := queries.CreateCategory(ctx, data.CreateCategoryParams{
		Budget: budget.ID,
		Name:   "testcategory",
	})
	if err != nil {
		t.Fatal(err)
	}
	transaction, err := queries.CreateTransaction(ctx, data.CreateTransactionParams{
		Date:         mustTime(t, 2026, 1, 1),
		Account:      account.ID,
		Payee:        payee.ID,
		TotalOutflow: 1000,
		TotalInflow:  0,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := queries.CreateTransactionCategory(ctx, data.CreateTransactionCategoryParams{
		Transaction:  transaction.ID,
		OtherAccount: sql.NullInt64{},
		Category:     sql.NullInt64{Int64: category.ID, Valid: true},
		Income:       false,
		Outflow:      1000,
		Inflow:       0,
	}); err != nil {
		t.Fatal(err)
	}
	balances, err := queries.ListAccountsBalances(ctx, budget.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(balances) != 1 {
		t.Fatalf("expected one balance, got %d", len(balances))
	}
	if balances[0].Balance != -1000 {
		t.Errorf("expected balance -1000, got %d", balances[0].Balance)
	}
	if balances[0].ReconciledBalance != 0 {
		t.Errorf("expected reconciled balance 0, got %d", balances[0].ReconciledBalance)
	}
}

func TestListAccountsBalancesIncome(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget, err := queries.CreateBudget(ctx, "testBudget")
	if err != nil {
		t.Fatal(err)
	}
	account, err := queries.CreateAccount(ctx, data.CreateAccountParams{
		Budget: budget.ID,
		Name:   "testaccount",
	})
	if err != nil {
		t.Fatal(err)
	}
	payee, err := queries.CreatePayee(ctx, data.CreatePayeeParams{
		Budget: budget.ID,
		Name:   "testpayee",
	})
	if err != nil {
		t.Fatal(err)
	}
	transaction, err := queries.CreateTransaction(ctx, data.CreateTransactionParams{
		Date:         mustTime(t, 2026, 1, 1),
		Account:      account.ID,
		Payee:        payee.ID,
		TotalOutflow: 0,
		TotalInflow:  5000,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := queries.CreateTransactionCategory(ctx, data.CreateTransactionCategoryParams{
		Transaction:  transaction.ID,
		OtherAccount: sql.NullInt64{},
		Category:     sql.NullInt64{},
		Income:       true,
		Outflow:      0,
		Inflow:       5000,
	}); err != nil {
		t.Fatal(err)
	}
	balances, err := queries.ListAccountsBalances(ctx, budget.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(balances) != 1 {
		t.Fatalf("expected one balance, got %d", len(balances))
	}
	if balances[0].Balance != 5000 {
		t.Errorf("expected balance 5000, got %d", balances[0].Balance)
	}
	if balances[0].ReconciledBalance != 0 {
		t.Errorf("expected reconciled balance 0, got %d", balances[0].ReconciledBalance)
	}
}

func TestListAccountsBalancesTransferOutSourcePrimary(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget, err := queries.CreateBudget(ctx, "testBudget")
	if err != nil {
		t.Fatal(err)
	}
	source, err := queries.CreateAccount(ctx, data.CreateAccountParams{
		Budget: budget.ID,
		Name:   "source",
	})
	if err != nil {
		t.Fatal(err)
	}
	target, err := queries.CreateAccount(ctx, data.CreateAccountParams{
		Budget: budget.ID,
		Name:   "target",
	})
	if err != nil {
		t.Fatal(err)
	}
	payee, err := queries.CreatePayee(ctx, data.CreatePayeeParams{
		Budget: budget.ID,
		Name:   "testpayee",
	})
	if err != nil {
		t.Fatal(err)
	}
	transaction, err := queries.CreateTransaction(ctx, data.CreateTransactionParams{
		Date:         mustTime(t, 2026, 1, 1),
		Account:      source.ID,
		Payee:        payee.ID,
		TotalOutflow: 3000,
		TotalInflow:  0,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := queries.CreateTransactionCategory(ctx, data.CreateTransactionCategoryParams{
		Transaction:  transaction.ID,
		OtherAccount: sql.NullInt64{Int64: target.ID, Valid: true},
		Category:     sql.NullInt64{},
		Income:       false,
		Outflow:      3000,
		Inflow:       0,
	}); err != nil {
		t.Fatal(err)
	}
	balances, err := queries.ListAccountsBalances(ctx, budget.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(balances) != 2 {
		t.Fatalf("expected two balances, got %d", len(balances))
	}
	byName := map[string]data.ListAccountsBalancesRow{}
	for _, b := range balances {
		byName[b.Name] = b
	}
	if b := byName["source"]; b.Balance != -3000 {
		t.Errorf("expected source balance -3000, got %d", b.Balance)
	}
	if b := byName["target"]; b.Balance != 3000 {
		t.Errorf("expected target balance 3000, got %d", b.Balance)
	}
}

func TestListAccountsBalancesTransferInMirrored(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget, err := queries.CreateBudget(ctx, "testBudget")
	if err != nil {
		t.Fatal(err)
	}
	source, err := queries.CreateAccount(ctx, data.CreateAccountParams{
		Budget: budget.ID,
		Name:   "source",
	})
	if err != nil {
		t.Fatal(err)
	}
	target, err := queries.CreateAccount(ctx, data.CreateAccountParams{
		Budget: budget.ID,
		Name:   "target",
	})
	if err != nil {
		t.Fatal(err)
	}
	payee, err := queries.CreatePayee(ctx, data.CreatePayeeParams{
		Budget: budget.ID,
		Name:   "testpayee",
	})
	if err != nil {
		t.Fatal(err)
	}
	transaction, err := queries.CreateTransaction(ctx, data.CreateTransactionParams{
		Date:         mustTime(t, 2026, 1, 1),
		Account:      target.ID,
		Payee:        payee.ID,
		TotalOutflow: 0,
		TotalInflow:  500,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := queries.CreateTransactionCategory(ctx, data.CreateTransactionCategoryParams{
		Transaction:  transaction.ID,
		OtherAccount: sql.NullInt64{Int64: source.ID, Valid: true},
		Category:     sql.NullInt64{},
		Income:       false,
		Outflow:      0,
		Inflow:       500,
	}); err != nil {
		t.Fatal(err)
	}
	balances, err := queries.ListAccountsBalances(ctx, budget.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(balances) != 2 {
		t.Fatalf("expected two balances, got %d", len(balances))
	}
	byName := map[string]data.ListAccountsBalancesRow{}
	for _, b := range balances {
		byName[b.Name] = b
	}
	if b := byName["source"]; b.Balance != -500 {
		t.Errorf("expected source balance -500, got %d", b.Balance)
	}
	if b := byName["target"]; b.Balance != 500 {
		t.Errorf("expected target balance 500, got %d", b.Balance)
	}
}

func TestListAccountsBalancesSplitTransfer(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget, err := queries.CreateBudget(ctx, "testBudget")
	if err != nil {
		t.Fatal(err)
	}
	source, err := queries.CreateAccount(ctx, data.CreateAccountParams{
		Budget: budget.ID,
		Name:   "source",
	})
	if err != nil {
		t.Fatal(err)
	}
	target1, err := queries.CreateAccount(ctx, data.CreateAccountParams{
		Budget: budget.ID,
		Name:   "target1",
	})
	if err != nil {
		t.Fatal(err)
	}
	target2, err := queries.CreateAccount(ctx, data.CreateAccountParams{
		Budget: budget.ID,
		Name:   "target2",
	})
	if err != nil {
		t.Fatal(err)
	}
	payee, err := queries.CreatePayee(ctx, data.CreatePayeeParams{
		Budget: budget.ID,
		Name:   "testpayee",
	})
	if err != nil {
		t.Fatal(err)
	}
	transaction, err := queries.CreateTransaction(ctx, data.CreateTransactionParams{
		Date:         mustTime(t, 2026, 1, 1),
		Account:      source.ID,
		Payee:        payee.ID,
		TotalOutflow: 3000,
		TotalInflow:  0,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := queries.CreateTransactionCategory(ctx, data.CreateTransactionCategoryParams{
		Transaction:  transaction.ID,
		OtherAccount: sql.NullInt64{Int64: target1.ID, Valid: true},
		Category:     sql.NullInt64{},
		Income:       false,
		Outflow:      1000,
		Inflow:       0,
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := queries.CreateTransactionCategory(ctx, data.CreateTransactionCategoryParams{
		Transaction:  transaction.ID,
		OtherAccount: sql.NullInt64{Int64: target2.ID, Valid: true},
		Category:     sql.NullInt64{},
		Income:       false,
		Outflow:      2000,
		Inflow:       0,
	}); err != nil {
		t.Fatal(err)
	}
	balances, err := queries.ListAccountsBalances(ctx, budget.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(balances) != 3 {
		t.Fatalf("expected three balances, got %d", len(balances))
	}
	byName := map[string]data.ListAccountsBalancesRow{}
	for _, b := range balances {
		byName[b.Name] = b
	}
	if b := byName["source"]; b.Balance != -3000 {
		t.Errorf("expected source balance -3000, got %d", b.Balance)
	}
	if b := byName["target1"]; b.Balance != 1000 {
		t.Errorf("expected target1 balance 1000, got %d", b.Balance)
	}
	if b := byName["target2"]; b.Balance != 2000 {
		t.Errorf("expected target2 balance 2000, got %d", b.Balance)
	}
}

func TestListAccountsBalancesZeroAmountTransaction(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget, err := queries.CreateBudget(ctx, "testBudget")
	if err != nil {
		t.Fatal(err)
	}
	account, err := queries.CreateAccount(ctx, data.CreateAccountParams{
		Budget: budget.ID,
		Name:   "testaccount",
	})
	if err != nil {
		t.Fatal(err)
	}
	payee, err := queries.CreatePayee(ctx, data.CreatePayeeParams{
		Budget: budget.ID,
		Name:   "testpayee",
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := queries.CreateTransaction(ctx, data.CreateTransactionParams{
		Date:         mustTime(t, 2026, 1, 1),
		Account:      account.ID,
		Payee:        payee.ID,
		TotalOutflow: 0,
		TotalInflow:  0,
	}); err != nil {
		t.Fatal(err)
	}
	balances, err := queries.ListAccountsBalances(ctx, budget.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(balances) != 1 {
		t.Fatalf("expected one balance, got %d", len(balances))
	}
	if balances[0].Balance != 0 {
		t.Errorf("expected balance 0, got %d", balances[0].Balance)
	}
	if balances[0].ReconciledBalance != 0 {
		t.Errorf("expected reconciled balance 0, got %d", balances[0].ReconciledBalance)
	}
}

func TestListAccountsBalancesMixed(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget, err := queries.CreateBudget(ctx, "testBudget")
	if err != nil {
		t.Fatal(err)
	}
	account, err := queries.CreateAccount(ctx, data.CreateAccountParams{
		Budget: budget.ID,
		Name:   "testaccount",
	})
	if err != nil {
		t.Fatal(err)
	}
	otherAccount, err := queries.CreateAccount(ctx, data.CreateAccountParams{
		Budget: budget.ID,
		Name:   "otheraccount",
	})
	if err != nil {
		t.Fatal(err)
	}
	payee, err := queries.CreatePayee(ctx, data.CreatePayeeParams{
		Budget: budget.ID,
		Name:   "testpayee",
	})
	if err != nil {
		t.Fatal(err)
	}
	category, err := queries.CreateCategory(ctx, data.CreateCategoryParams{
		Budget: budget.ID,
		Name:   "testcategory",
	})
	if err != nil {
		t.Fatal(err)
	}
	spend, err := queries.CreateTransaction(ctx, data.CreateTransactionParams{
		Date:         mustTime(t, 2026, 1, 1),
		Account:      account.ID,
		Payee:        payee.ID,
		TotalOutflow: 1000,
		TotalInflow:  0,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := queries.CreateTransactionCategory(ctx, data.CreateTransactionCategoryParams{
		Transaction:  spend.ID,
		OtherAccount: sql.NullInt64{},
		Category:     sql.NullInt64{Int64: category.ID, Valid: true},
		Income:       false,
		Outflow:      1000,
		Inflow:       0,
	}); err != nil {
		t.Fatal(err)
	}
	income, err := queries.CreateTransaction(ctx, data.CreateTransactionParams{
		Date:         mustTime(t, 2026, 1, 2),
		Account:      account.ID,
		Payee:        payee.ID,
		TotalOutflow: 0,
		TotalInflow:  5000,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := queries.CreateTransactionCategory(ctx, data.CreateTransactionCategoryParams{
		Transaction:  income.ID,
		OtherAccount: sql.NullInt64{},
		Category:     sql.NullInt64{},
		Income:       true,
		Outflow:      0,
		Inflow:       5000,
	}); err != nil {
		t.Fatal(err)
	}
	transfer, err := queries.CreateTransaction(ctx, data.CreateTransactionParams{
		Date:         mustTime(t, 2026, 1, 3),
		Account:      account.ID,
		Payee:        payee.ID,
		TotalOutflow: 2000,
		TotalInflow:  0,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := queries.CreateTransactionCategory(ctx, data.CreateTransactionCategoryParams{
		Transaction:  transfer.ID,
		OtherAccount: sql.NullInt64{Int64: otherAccount.ID, Valid: true},
		Category:     sql.NullInt64{},
		Income:       false,
		Outflow:      2000,
		Inflow:       0,
	}); err != nil {
		t.Fatal(err)
	}
	balances, err := queries.ListAccountsBalances(ctx, budget.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(balances) != 2 {
		t.Fatalf("expected two balances, got %d", len(balances))
	}
	byName := map[string]data.ListAccountsBalancesRow{}
	for _, b := range balances {
		byName[b.Name] = b
	}
	if b := byName["testaccount"]; b.Balance != 2000 {
		t.Errorf("expected testaccount balance 2000, got %d", b.Balance)
	}
	if b := byName["otheraccount"]; b.Balance != 2000 {
		t.Errorf("expected otheraccount balance 2000, got %d", b.Balance)
	}
}

func TestListAccountsBalancesReconciledAndUnreconciled(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget, err := queries.CreateBudget(ctx, "testBudget")
	if err != nil {
		t.Fatal(err)
	}
	account, err := queries.CreateAccount(ctx, data.CreateAccountParams{
		Budget: budget.ID,
		Name:   "testaccount",
	})
	if err != nil {
		t.Fatal(err)
	}
	payee, err := queries.CreatePayee(ctx, data.CreatePayeeParams{
		Budget: budget.ID,
		Name:   "testpayee",
	})
	if err != nil {
		t.Fatal(err)
	}
	category, err := queries.CreateCategory(ctx, data.CreateCategoryParams{
		Budget: budget.ID,
		Name:   "testcategory",
	})
	if err != nil {
		t.Fatal(err)
	}
	reconciled, err := queries.CreateTransaction(ctx, data.CreateTransactionParams{
		Date:         mustTime(t, 2026, 1, 1),
		Account:      account.ID,
		Payee:        payee.ID,
		TotalOutflow: 1000,
		TotalInflow:  0,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := queries.CreateTransactionCategory(ctx, data.CreateTransactionCategoryParams{
		Transaction:  reconciled.ID,
		OtherAccount: sql.NullInt64{},
		Category:     sql.NullInt64{Int64: category.ID, Valid: true},
		Income:       false,
		Outflow:      1000,
		Inflow:       0,
	}); err != nil {
		t.Fatal(err)
	}
	unreconciled, err := queries.CreateTransaction(ctx, data.CreateTransactionParams{
		Date:         mustTime(t, 2026, 2, 1),
		Account:      account.ID,
		Payee:        payee.ID,
		TotalOutflow: 2000,
		TotalInflow:  0,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := queries.CreateTransactionCategory(ctx, data.CreateTransactionCategoryParams{
		Transaction:  unreconciled.ID,
		OtherAccount: sql.NullInt64{},
		Category:     sql.NullInt64{Int64: category.ID, Valid: true},
		Income:       false,
		Outflow:      2000,
		Inflow:       0,
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := queries.ReconcileAccountTransactions(ctx, data.ReconcileAccountTransactionsParams{
		AccountID: account.ID,
		Date:      mustTime(t, 2026, 1, 31),
	}); err != nil {
		t.Fatal(err)
	}
	balances, err := queries.ListAccountsBalances(ctx, budget.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(balances) != 1 {
		t.Fatalf("expected one balance, got %d", len(balances))
	}
	if balances[0].Balance != -3000 {
		t.Errorf("expected balance -3000, got %d", balances[0].Balance)
	}
	if balances[0].ReconciledBalance != -1000 {
		t.Errorf("expected reconciled balance -1000, got %d", balances[0].ReconciledBalance)
	}
}

func TestListAccountsBalancesTransferReconciledPerAccount(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget, err := queries.CreateBudget(ctx, "testBudget")
	if err != nil {
		t.Fatal(err)
	}
	source, err := queries.CreateAccount(ctx, data.CreateAccountParams{
		Budget: budget.ID,
		Name:   "source",
	})
	if err != nil {
		t.Fatal(err)
	}
	target, err := queries.CreateAccount(ctx, data.CreateAccountParams{
		Budget: budget.ID,
		Name:   "target",
	})
	if err != nil {
		t.Fatal(err)
	}
	payee, err := queries.CreatePayee(ctx, data.CreatePayeeParams{
		Budget: budget.ID,
		Name:   "testpayee",
	})
	if err != nil {
		t.Fatal(err)
	}
	transaction, err := queries.CreateTransaction(ctx, data.CreateTransactionParams{
		Date:         mustTime(t, 2026, 1, 1),
		Account:      source.ID,
		Payee:        payee.ID,
		TotalOutflow: 3000,
		TotalInflow:  0,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := queries.CreateTransactionCategory(ctx, data.CreateTransactionCategoryParams{
		Transaction:  transaction.ID,
		OtherAccount: sql.NullInt64{Int64: target.ID, Valid: true},
		Category:     sql.NullInt64{},
		Income:       false,
		Outflow:      3000,
		Inflow:       0,
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := queries.ReconcileAccountTransactions(ctx, data.ReconcileAccountTransactionsParams{
		AccountID: target.ID,
		Date:      mustTime(t, 2026, 2, 1),
	}); err != nil {
		t.Fatal(err)
	}
	balances, err := queries.ListAccountsBalances(ctx, budget.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(balances) != 2 {
		t.Fatalf("expected two balances, got %d", len(balances))
	}
	byName := map[string]data.ListAccountsBalancesRow{}
	for _, b := range balances {
		byName[b.Name] = b
	}
	if b := byName["source"]; b.Balance != -3000 || b.ReconciledBalance != 0 {
		t.Errorf("expected source balance -3000 reconciled 0, got %d/%d", b.Balance, b.ReconciledBalance)
	}
	if b := byName["target"]; b.Balance != 3000 || b.ReconciledBalance != 3000 {
		t.Errorf("expected target balance 3000 reconciled 3000, got %d/%d", b.Balance, b.ReconciledBalance)
	}
}

func TestGetAccountBalancesNoTransactions(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget, err := queries.CreateBudget(ctx, "testBudget")
	if err != nil {
		t.Fatal(err)
	}
	account, err := queries.CreateAccount(ctx, data.CreateAccountParams{
		Budget: budget.ID,
		Name:   "testaccount",
	})
	if err != nil {
		t.Fatal(err)
	}
	balances, err := queries.GetAccountBalances(ctx, account.ID)
	if err != nil {
		t.Fatal(err)
	}
	if balances.ID != account.ID {
		t.Error("account id does not match")
	}
	if balances.Name != "testaccount" {
		t.Error("account name does not match")
	}
	if balances.Balance != 0 {
		t.Errorf("expected balance 0, got %d", balances.Balance)
	}
	if balances.ReconciledBalance != 0 {
		t.Errorf("expected reconciled balance 0, got %d", balances.ReconciledBalance)
	}
}

func TestGetAccountBalancesCategorySpend(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget, err := queries.CreateBudget(ctx, "testBudget")
	if err != nil {
		t.Fatal(err)
	}
	account, err := queries.CreateAccount(ctx, data.CreateAccountParams{
		Budget: budget.ID,
		Name:   "testaccount",
	})
	if err != nil {
		t.Fatal(err)
	}
	payee, err := queries.CreatePayee(ctx, data.CreatePayeeParams{
		Budget: budget.ID,
		Name:   "testpayee",
	})
	if err != nil {
		t.Fatal(err)
	}
	category, err := queries.CreateCategory(ctx, data.CreateCategoryParams{
		Budget: budget.ID,
		Name:   "testcategory",
	})
	if err != nil {
		t.Fatal(err)
	}
	transaction, err := queries.CreateTransaction(ctx, data.CreateTransactionParams{
		Date:         mustTime(t, 2026, 1, 1),
		Account:      account.ID,
		Payee:        payee.ID,
		TotalOutflow: 1000,
		TotalInflow:  0,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := queries.CreateTransactionCategory(ctx, data.CreateTransactionCategoryParams{
		Transaction:  transaction.ID,
		OtherAccount: sql.NullInt64{},
		Category:     sql.NullInt64{Int64: category.ID, Valid: true},
		Income:       false,
		Outflow:      1000,
		Inflow:       0,
	}); err != nil {
		t.Fatal(err)
	}
	balances, err := queries.GetAccountBalances(ctx, account.ID)
	if err != nil {
		t.Fatal(err)
	}
	if balances.Balance != -1000 {
		t.Errorf("expected balance -1000, got %d", balances.Balance)
	}
	if balances.ReconciledBalance != 0 {
		t.Errorf("expected reconciled balance 0, got %d", balances.ReconciledBalance)
	}
}

func TestGetAccountBalancesIncome(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget, err := queries.CreateBudget(ctx, "testBudget")
	if err != nil {
		t.Fatal(err)
	}
	account, err := queries.CreateAccount(ctx, data.CreateAccountParams{
		Budget: budget.ID,
		Name:   "testaccount",
	})
	if err != nil {
		t.Fatal(err)
	}
	payee, err := queries.CreatePayee(ctx, data.CreatePayeeParams{
		Budget: budget.ID,
		Name:   "testpayee",
	})
	if err != nil {
		t.Fatal(err)
	}
	transaction, err := queries.CreateTransaction(ctx, data.CreateTransactionParams{
		Date:         mustTime(t, 2026, 1, 1),
		Account:      account.ID,
		Payee:        payee.ID,
		TotalOutflow: 0,
		TotalInflow:  5000,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := queries.CreateTransactionCategory(ctx, data.CreateTransactionCategoryParams{
		Transaction:  transaction.ID,
		OtherAccount: sql.NullInt64{},
		Category:     sql.NullInt64{},
		Income:       true,
		Outflow:      0,
		Inflow:       5000,
	}); err != nil {
		t.Fatal(err)
	}
	balances, err := queries.GetAccountBalances(ctx, account.ID)
	if err != nil {
		t.Fatal(err)
	}
	if balances.Balance != 5000 {
		t.Errorf("expected balance 5000, got %d", balances.Balance)
	}
	if balances.ReconciledBalance != 0 {
		t.Errorf("expected reconciled balance 0, got %d", balances.ReconciledBalance)
	}
}

func TestGetAccountBalancesTransferTarget(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget, err := queries.CreateBudget(ctx, "testBudget")
	if err != nil {
		t.Fatal(err)
	}
	source, err := queries.CreateAccount(ctx, data.CreateAccountParams{
		Budget: budget.ID,
		Name:   "source",
	})
	if err != nil {
		t.Fatal(err)
	}
	target, err := queries.CreateAccount(ctx, data.CreateAccountParams{
		Budget: budget.ID,
		Name:   "target",
	})
	if err != nil {
		t.Fatal(err)
	}
	payee, err := queries.CreatePayee(ctx, data.CreatePayeeParams{
		Budget: budget.ID,
		Name:   "testpayee",
	})
	if err != nil {
		t.Fatal(err)
	}
	transaction, err := queries.CreateTransaction(ctx, data.CreateTransactionParams{
		Date:         mustTime(t, 2026, 1, 1),
		Account:      source.ID,
		Payee:        payee.ID,
		TotalOutflow: 3000,
		TotalInflow:  0,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := queries.CreateTransactionCategory(ctx, data.CreateTransactionCategoryParams{
		Transaction:  transaction.ID,
		OtherAccount: sql.NullInt64{Int64: target.ID, Valid: true},
		Category:     sql.NullInt64{},
		Income:       false,
		Outflow:      3000,
		Inflow:       0,
	}); err != nil {
		t.Fatal(err)
	}
	balances, err := queries.GetAccountBalances(ctx, target.ID)
	if err != nil {
		t.Fatal(err)
	}
	if balances.Balance != 3000 {
		t.Errorf("expected target balance 3000, got %d", balances.Balance)
	}
}

func TestGetAccountBalancesTransferSource(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget, err := queries.CreateBudget(ctx, "testBudget")
	if err != nil {
		t.Fatal(err)
	}
	source, err := queries.CreateAccount(ctx, data.CreateAccountParams{
		Budget: budget.ID,
		Name:   "source",
	})
	if err != nil {
		t.Fatal(err)
	}
	target, err := queries.CreateAccount(ctx, data.CreateAccountParams{
		Budget: budget.ID,
		Name:   "target",
	})
	if err != nil {
		t.Fatal(err)
	}
	payee, err := queries.CreatePayee(ctx, data.CreatePayeeParams{
		Budget: budget.ID,
		Name:   "testpayee",
	})
	if err != nil {
		t.Fatal(err)
	}
	transaction, err := queries.CreateTransaction(ctx, data.CreateTransactionParams{
		Date:         mustTime(t, 2026, 1, 1),
		Account:      source.ID,
		Payee:        payee.ID,
		TotalOutflow: 3000,
		TotalInflow:  0,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := queries.CreateTransactionCategory(ctx, data.CreateTransactionCategoryParams{
		Transaction:  transaction.ID,
		OtherAccount: sql.NullInt64{Int64: target.ID, Valid: true},
		Category:     sql.NullInt64{},
		Income:       false,
		Outflow:      3000,
		Inflow:       0,
	}); err != nil {
		t.Fatal(err)
	}
	balances, err := queries.GetAccountBalances(ctx, source.ID)
	if err != nil {
		t.Fatal(err)
	}
	if balances.Balance != -3000 {
		t.Errorf("expected source balance -3000, got %d", balances.Balance)
	}
}

func TestGetAccountBalancesMirroredTransfer(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget, err := queries.CreateBudget(ctx, "testBudget")
	if err != nil {
		t.Fatal(err)
	}
	source, err := queries.CreateAccount(ctx, data.CreateAccountParams{
		Budget: budget.ID,
		Name:   "source",
	})
	if err != nil {
		t.Fatal(err)
	}
	target, err := queries.CreateAccount(ctx, data.CreateAccountParams{
		Budget: budget.ID,
		Name:   "target",
	})
	if err != nil {
		t.Fatal(err)
	}
	payee, err := queries.CreatePayee(ctx, data.CreatePayeeParams{
		Budget: budget.ID,
		Name:   "testpayee",
	})
	if err != nil {
		t.Fatal(err)
	}
	transaction, err := queries.CreateTransaction(ctx, data.CreateTransactionParams{
		Date:         mustTime(t, 2026, 1, 1),
		Account:      target.ID,
		Payee:        payee.ID,
		TotalOutflow: 0,
		TotalInflow:  500,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := queries.CreateTransactionCategory(ctx, data.CreateTransactionCategoryParams{
		Transaction:  transaction.ID,
		OtherAccount: sql.NullInt64{Int64: source.ID, Valid: true},
		Category:     sql.NullInt64{},
		Income:       false,
		Outflow:      0,
		Inflow:       500,
	}); err != nil {
		t.Fatal(err)
	}
	balances, err := queries.GetAccountBalances(ctx, target.ID)
	if err != nil {
		t.Fatal(err)
	}
	if balances.Balance != 500 {
		t.Errorf("expected target balance 500, got %d", balances.Balance)
	}
}

func TestGetAccountBalancesReconciled(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget, err := queries.CreateBudget(ctx, "testBudget")
	if err != nil {
		t.Fatal(err)
	}
	account, err := queries.CreateAccount(ctx, data.CreateAccountParams{
		Budget: budget.ID,
		Name:   "testaccount",
	})
	if err != nil {
		t.Fatal(err)
	}
	payee, err := queries.CreatePayee(ctx, data.CreatePayeeParams{
		Budget: budget.ID,
		Name:   "testpayee",
	})
	if err != nil {
		t.Fatal(err)
	}
	category, err := queries.CreateCategory(ctx, data.CreateCategoryParams{
		Budget: budget.ID,
		Name:   "testcategory",
	})
	if err != nil {
		t.Fatal(err)
	}
	reconciled, err := queries.CreateTransaction(ctx, data.CreateTransactionParams{
		Date:         mustTime(t, 2026, 1, 1),
		Account:      account.ID,
		Payee:        payee.ID,
		TotalOutflow: 1000,
		TotalInflow:  0,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := queries.CreateTransactionCategory(ctx, data.CreateTransactionCategoryParams{
		Transaction:  reconciled.ID,
		OtherAccount: sql.NullInt64{},
		Category:     sql.NullInt64{Int64: category.ID, Valid: true},
		Income:       false,
		Outflow:      1000,
		Inflow:       0,
	}); err != nil {
		t.Fatal(err)
	}
	unreconciled, err := queries.CreateTransaction(ctx, data.CreateTransactionParams{
		Date:         mustTime(t, 2026, 2, 1),
		Account:      account.ID,
		Payee:        payee.ID,
		TotalOutflow: 2000,
		TotalInflow:  0,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := queries.CreateTransactionCategory(ctx, data.CreateTransactionCategoryParams{
		Transaction:  unreconciled.ID,
		OtherAccount: sql.NullInt64{},
		Category:     sql.NullInt64{Int64: category.ID, Valid: true},
		Income:       false,
		Outflow:      2000,
		Inflow:       0,
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := queries.ReconcileAccountTransactions(ctx, data.ReconcileAccountTransactionsParams{
		AccountID: account.ID,
		Date:      mustTime(t, 2026, 1, 31),
	}); err != nil {
		t.Fatal(err)
	}
	balances, err := queries.GetAccountBalances(ctx, account.ID)
	if err != nil {
		t.Fatal(err)
	}
	if balances.Balance != -3000 {
		t.Errorf("expected balance -3000, got %d", balances.Balance)
	}
	if balances.ReconciledBalance != -1000 {
		t.Errorf("expected reconciled balance -1000, got %d", balances.ReconciledBalance)
	}
}

func TestGetAccountBalancesZeroAmountTransaction(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget, err := queries.CreateBudget(ctx, "testBudget")
	if err != nil {
		t.Fatal(err)
	}
	account, err := queries.CreateAccount(ctx, data.CreateAccountParams{
		Budget: budget.ID,
		Name:   "testaccount",
	})
	if err != nil {
		t.Fatal(err)
	}
	payee, err := queries.CreatePayee(ctx, data.CreatePayeeParams{
		Budget: budget.ID,
		Name:   "testpayee",
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := queries.CreateTransaction(ctx, data.CreateTransactionParams{
		Date:         mustTime(t, 2026, 1, 1),
		Account:      account.ID,
		Payee:        payee.ID,
		TotalOutflow: 0,
		TotalInflow:  0,
	}); err != nil {
		t.Fatal(err)
	}
	balances, err := queries.GetAccountBalances(ctx, account.ID)
	if err != nil {
		t.Fatal(err)
	}
	if balances.Balance != 0 {
		t.Errorf("expected balance 0, got %d", balances.Balance)
	}
	if balances.ReconciledBalance != 0 {
		t.Errorf("expected reconciled balance 0, got %d", balances.ReconciledBalance)
	}
}

func TestGetAccountBalancesNonExistent(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	_, err := queries.GetAccountBalances(ctx, 99)
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("expected sql.ErrNoRows, got %v", err)
	}
}

func TestGetAccountBalanceAsOfNoTransactions(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget, err := queries.CreateBudget(ctx, "testBudget")
	if err != nil {
		t.Fatal(err)
	}
	account, err := queries.CreateAccount(ctx, data.CreateAccountParams{
		Budget: budget.ID,
		Name:   "testaccount",
	})
	if err != nil {
		t.Fatal(err)
	}
	balance, err := queries.GetAccountBalanceAsOf(ctx, data.GetAccountBalanceAsOfParams{
		AccountID: account.ID,
		Date:      mustTime(t, 2026, 1, 15),
	})
	if err != nil {
		t.Fatal(err)
	}
	if balance != 0 {
		t.Errorf("expected balance 0, got %d", balance)
	}
}

func TestGetAccountBalanceAsOfBeforeFirstTransaction(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget, err := queries.CreateBudget(ctx, "testBudget")
	if err != nil {
		t.Fatal(err)
	}
	account, err := queries.CreateAccount(ctx, data.CreateAccountParams{
		Budget: budget.ID,
		Name:   "testaccount",
	})
	if err != nil {
		t.Fatal(err)
	}
	payee, err := queries.CreatePayee(ctx, data.CreatePayeeParams{
		Budget: budget.ID,
		Name:   "testpayee",
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := queries.CreateTransaction(ctx, data.CreateTransactionParams{
		Date:         mustTime(t, 2026, 1, 15),
		Account:      account.ID,
		Payee:        payee.ID,
		TotalOutflow: 1000,
		TotalInflow:  0,
	}); err != nil {
		t.Fatal(err)
	}
	balance, err := queries.GetAccountBalanceAsOf(ctx, data.GetAccountBalanceAsOfParams{
		AccountID: account.ID,
		Date:      mustTime(t, 2026, 1, 1),
	})
	if err != nil {
		t.Fatal(err)
	}
	if balance != 0 {
		t.Errorf("expected balance 0, got %d", balance)
	}
}

func TestGetAccountBalanceAsOfInclusiveOfDate(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget, err := queries.CreateBudget(ctx, "testBudget")
	if err != nil {
		t.Fatal(err)
	}
	account, err := queries.CreateAccount(ctx, data.CreateAccountParams{
		Budget: budget.ID,
		Name:   "testaccount",
	})
	if err != nil {
		t.Fatal(err)
	}
	payee, err := queries.CreatePayee(ctx, data.CreatePayeeParams{
		Budget: budget.ID,
		Name:   "testpayee",
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := queries.CreateTransaction(ctx, data.CreateTransactionParams{
		Date:         mustTime(t, 2026, 1, 15),
		Account:      account.ID,
		Payee:        payee.ID,
		TotalOutflow: 1000,
		TotalInflow:  0,
	}); err != nil {
		t.Fatal(err)
	}
	balance, err := queries.GetAccountBalanceAsOf(ctx, data.GetAccountBalanceAsOfParams{
		AccountID: account.ID,
		Date:      mustTime(t, 2026, 1, 15),
	})
	if err != nil {
		t.Fatal(err)
	}
	if balance != -1000 {
		t.Errorf("expected balance -1000, got %d", balance)
	}
}

func TestGetAccountBalanceAsOfCategorySpend(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget, err := queries.CreateBudget(ctx, "testBudget")
	if err != nil {
		t.Fatal(err)
	}
	account, err := queries.CreateAccount(ctx, data.CreateAccountParams{
		Budget: budget.ID,
		Name:   "testaccount",
	})
	if err != nil {
		t.Fatal(err)
	}
	payee, err := queries.CreatePayee(ctx, data.CreatePayeeParams{
		Budget: budget.ID,
		Name:   "testpayee",
	})
	if err != nil {
		t.Fatal(err)
	}
	category, err := queries.CreateCategory(ctx, data.CreateCategoryParams{
		Budget: budget.ID,
		Name:   "testcategory",
	})
	if err != nil {
		t.Fatal(err)
	}
	transaction, err := queries.CreateTransaction(ctx, data.CreateTransactionParams{
		Date:         mustTime(t, 2026, 1, 1),
		Account:      account.ID,
		Payee:        payee.ID,
		TotalOutflow: 1000,
		TotalInflow:  0,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := queries.CreateTransactionCategory(ctx, data.CreateTransactionCategoryParams{
		Transaction:  transaction.ID,
		OtherAccount: sql.NullInt64{},
		Category:     sql.NullInt64{Int64: category.ID, Valid: true},
		Income:       false,
		Outflow:      1000,
		Inflow:       0,
	}); err != nil {
		t.Fatal(err)
	}
	balance, err := queries.GetAccountBalanceAsOf(ctx, data.GetAccountBalanceAsOfParams{
		AccountID: account.ID,
		Date:      mustTime(t, 2026, 1, 31),
	})
	if err != nil {
		t.Fatal(err)
	}
	if balance != -1000 {
		t.Errorf("expected balance -1000, got %d", balance)
	}
}

func TestGetAccountBalanceAsOfIncome(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget, err := queries.CreateBudget(ctx, "testBudget")
	if err != nil {
		t.Fatal(err)
	}
	account, err := queries.CreateAccount(ctx, data.CreateAccountParams{
		Budget: budget.ID,
		Name:   "testaccount",
	})
	if err != nil {
		t.Fatal(err)
	}
	payee, err := queries.CreatePayee(ctx, data.CreatePayeeParams{
		Budget: budget.ID,
		Name:   "testpayee",
	})
	if err != nil {
		t.Fatal(err)
	}
	transaction, err := queries.CreateTransaction(ctx, data.CreateTransactionParams{
		Date:         mustTime(t, 2026, 1, 1),
		Account:      account.ID,
		Payee:        payee.ID,
		TotalOutflow: 0,
		TotalInflow:  5000,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := queries.CreateTransactionCategory(ctx, data.CreateTransactionCategoryParams{
		Transaction:  transaction.ID,
		OtherAccount: sql.NullInt64{},
		Category:     sql.NullInt64{},
		Income:       true,
		Outflow:      0,
		Inflow:       5000,
	}); err != nil {
		t.Fatal(err)
	}
	balance, err := queries.GetAccountBalanceAsOf(ctx, data.GetAccountBalanceAsOfParams{
		AccountID: account.ID,
		Date:      mustTime(t, 2026, 1, 31),
	})
	if err != nil {
		t.Fatal(err)
	}
	if balance != 5000 {
		t.Errorf("expected balance 5000, got %d", balance)
	}
}

func TestGetAccountBalanceAsOfTransferSourcePrimary(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget, err := queries.CreateBudget(ctx, "testBudget")
	if err != nil {
		t.Fatal(err)
	}
	source, err := queries.CreateAccount(ctx, data.CreateAccountParams{
		Budget: budget.ID,
		Name:   "source",
	})
	if err != nil {
		t.Fatal(err)
	}
	target, err := queries.CreateAccount(ctx, data.CreateAccountParams{
		Budget: budget.ID,
		Name:   "target",
	})
	if err != nil {
		t.Fatal(err)
	}
	payee, err := queries.CreatePayee(ctx, data.CreatePayeeParams{
		Budget: budget.ID,
		Name:   "testpayee",
	})
	if err != nil {
		t.Fatal(err)
	}
	transaction, err := queries.CreateTransaction(ctx, data.CreateTransactionParams{
		Date:         mustTime(t, 2026, 1, 1),
		Account:      source.ID,
		Payee:        payee.ID,
		TotalOutflow: 3000,
		TotalInflow:  0,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := queries.CreateTransactionCategory(ctx, data.CreateTransactionCategoryParams{
		Transaction:  transaction.ID,
		OtherAccount: sql.NullInt64{Int64: target.ID, Valid: true},
		Category:     sql.NullInt64{},
		Income:       false,
		Outflow:      3000,
		Inflow:       0,
	}); err != nil {
		t.Fatal(err)
	}
	sourceBalance, err := queries.GetAccountBalanceAsOf(ctx, data.GetAccountBalanceAsOfParams{
		AccountID: source.ID,
		Date:      mustTime(t, 2026, 1, 31),
	})
	if err != nil {
		t.Fatal(err)
	}
	if sourceBalance != -3000 {
		t.Errorf("expected source balance -3000, got %d", sourceBalance)
	}
	targetBalance, err := queries.GetAccountBalanceAsOf(ctx, data.GetAccountBalanceAsOfParams{
		AccountID: target.ID,
		Date:      mustTime(t, 2026, 1, 31),
	})
	if err != nil {
		t.Fatal(err)
	}
	if targetBalance != 3000 {
		t.Errorf("expected target balance 3000, got %d", targetBalance)
	}
}

func TestGetAccountBalanceAsOfTransferMirrored(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget, err := queries.CreateBudget(ctx, "testBudget")
	if err != nil {
		t.Fatal(err)
	}
	source, err := queries.CreateAccount(ctx, data.CreateAccountParams{
		Budget: budget.ID,
		Name:   "source",
	})
	if err != nil {
		t.Fatal(err)
	}
	target, err := queries.CreateAccount(ctx, data.CreateAccountParams{
		Budget: budget.ID,
		Name:   "target",
	})
	if err != nil {
		t.Fatal(err)
	}
	payee, err := queries.CreatePayee(ctx, data.CreatePayeeParams{
		Budget: budget.ID,
		Name:   "testpayee",
	})
	if err != nil {
		t.Fatal(err)
	}
	transaction, err := queries.CreateTransaction(ctx, data.CreateTransactionParams{
		Date:         mustTime(t, 2026, 1, 1),
		Account:      target.ID,
		Payee:        payee.ID,
		TotalOutflow: 0,
		TotalInflow:  500,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := queries.CreateTransactionCategory(ctx, data.CreateTransactionCategoryParams{
		Transaction:  transaction.ID,
		OtherAccount: sql.NullInt64{Int64: source.ID, Valid: true},
		Category:     sql.NullInt64{},
		Income:       false,
		Outflow:      0,
		Inflow:       500,
	}); err != nil {
		t.Fatal(err)
	}
	targetBalance, err := queries.GetAccountBalanceAsOf(ctx, data.GetAccountBalanceAsOfParams{
		AccountID: target.ID,
		Date:      mustTime(t, 2026, 1, 31),
	})
	if err != nil {
		t.Fatal(err)
	}
	if targetBalance != 500 {
		t.Errorf("expected target balance 500, got %d", targetBalance)
	}
	sourceBalance, err := queries.GetAccountBalanceAsOf(ctx, data.GetAccountBalanceAsOfParams{
		AccountID: source.ID,
		Date:      mustTime(t, 2026, 1, 31),
	})
	if err != nil {
		t.Fatal(err)
	}
	if sourceBalance != -500 {
		t.Errorf("expected source balance -500, got %d", sourceBalance)
	}
}

func TestGetAccountBalanceAsOfExcludesLaterTransactions(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget, err := queries.CreateBudget(ctx, "testBudget")
	if err != nil {
		t.Fatal(err)
	}
	account, err := queries.CreateAccount(ctx, data.CreateAccountParams{
		Budget: budget.ID,
		Name:   "testaccount",
	})
	if err != nil {
		t.Fatal(err)
	}
	payee, err := queries.CreatePayee(ctx, data.CreatePayeeParams{
		Budget: budget.ID,
		Name:   "testpayee",
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := queries.CreateTransaction(ctx, data.CreateTransactionParams{
		Date:         mustTime(t, 2026, 1, 1),
		Account:      account.ID,
		Payee:        payee.ID,
		TotalOutflow: 1000,
		TotalInflow:  0,
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := queries.CreateTransaction(ctx, data.CreateTransactionParams{
		Date:         mustTime(t, 2026, 2, 1),
		Account:      account.ID,
		Payee:        payee.ID,
		TotalOutflow: 2000,
		TotalInflow:  0,
	}); err != nil {
		t.Fatal(err)
	}
	balance, err := queries.GetAccountBalanceAsOf(ctx, data.GetAccountBalanceAsOfParams{
		AccountID: account.ID,
		Date:      mustTime(t, 2026, 1, 31),
	})
	if err != nil {
		t.Fatal(err)
	}
	if balance != -1000 {
		t.Errorf("expected balance -1000, got %d", balance)
	}
	fullBalance, err := queries.GetAccountBalanceAsOf(ctx, data.GetAccountBalanceAsOfParams{
		AccountID: account.ID,
		Date:      mustTime(t, 2026, 2, 28),
	})
	if err != nil {
		t.Fatal(err)
	}
	if fullBalance != -3000 {
		t.Errorf("expected full balance -3000, got %d", fullBalance)
	}
}

func TestGetAccountBalanceAsOfExcludesTransfersAfterDate(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget, err := queries.CreateBudget(ctx, "testBudget")
	if err != nil {
		t.Fatal(err)
	}
	source, err := queries.CreateAccount(ctx, data.CreateAccountParams{
		Budget: budget.ID,
		Name:   "source",
	})
	if err != nil {
		t.Fatal(err)
	}
	target, err := queries.CreateAccount(ctx, data.CreateAccountParams{
		Budget: budget.ID,
		Name:   "target",
	})
	if err != nil {
		t.Fatal(err)
	}
	payee, err := queries.CreatePayee(ctx, data.CreatePayeeParams{
		Budget: budget.ID,
		Name:   "testpayee",
	})
	if err != nil {
		t.Fatal(err)
	}
	transaction, err := queries.CreateTransaction(ctx, data.CreateTransactionParams{
		Date:         mustTime(t, 2026, 2, 1),
		Account:      source.ID,
		Payee:        payee.ID,
		TotalOutflow: 3000,
		TotalInflow:  0,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := queries.CreateTransactionCategory(ctx, data.CreateTransactionCategoryParams{
		Transaction:  transaction.ID,
		OtherAccount: sql.NullInt64{Int64: target.ID, Valid: true},
		Category:     sql.NullInt64{},
		Income:       false,
		Outflow:      3000,
		Inflow:       0,
	}); err != nil {
		t.Fatal(err)
	}
	balance, err := queries.GetAccountBalanceAsOf(ctx, data.GetAccountBalanceAsOfParams{
		AccountID: target.ID,
		Date:      mustTime(t, 2026, 1, 31),
	})
	if err != nil {
		t.Fatal(err)
	}
	if balance != 0 {
		t.Errorf("expected balance 0, got %d", balance)
	}
}

func TestGetAccountBalanceAsOfCombined(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget, err := queries.CreateBudget(ctx, "testBudget")
	if err != nil {
		t.Fatal(err)
	}
	account, err := queries.CreateAccount(ctx, data.CreateAccountParams{
		Budget: budget.ID,
		Name:   "testaccount",
	})
	if err != nil {
		t.Fatal(err)
	}
	otherAccount, err := queries.CreateAccount(ctx, data.CreateAccountParams{
		Budget: budget.ID,
		Name:   "otheraccount",
	})
	if err != nil {
		t.Fatal(err)
	}
	payee, err := queries.CreatePayee(ctx, data.CreatePayeeParams{
		Budget: budget.ID,
		Name:   "testpayee",
	})
	if err != nil {
		t.Fatal(err)
	}
	category, err := queries.CreateCategory(ctx, data.CreateCategoryParams{
		Budget: budget.ID,
		Name:   "testcategory",
	})
	if err != nil {
		t.Fatal(err)
	}
	spend, err := queries.CreateTransaction(ctx, data.CreateTransactionParams{
		Date:         mustTime(t, 2026, 1, 1),
		Account:      account.ID,
		Payee:        payee.ID,
		TotalOutflow: 1000,
		TotalInflow:  0,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := queries.CreateTransactionCategory(ctx, data.CreateTransactionCategoryParams{
		Transaction:  spend.ID,
		OtherAccount: sql.NullInt64{},
		Category:     sql.NullInt64{Int64: category.ID, Valid: true},
		Income:       false,
		Outflow:      1000,
		Inflow:       0,
	}); err != nil {
		t.Fatal(err)
	}
	income, err := queries.CreateTransaction(ctx, data.CreateTransactionParams{
		Date:         mustTime(t, 2026, 1, 2),
		Account:      account.ID,
		Payee:        payee.ID,
		TotalOutflow: 0,
		TotalInflow:  5000,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := queries.CreateTransactionCategory(ctx, data.CreateTransactionCategoryParams{
		Transaction:  income.ID,
		OtherAccount: sql.NullInt64{},
		Category:     sql.NullInt64{},
		Income:       true,
		Outflow:      0,
		Inflow:       5000,
	}); err != nil {
		t.Fatal(err)
	}
	incoming, err := queries.CreateTransaction(ctx, data.CreateTransactionParams{
		Date:         mustTime(t, 2026, 1, 3),
		Account:      otherAccount.ID,
		Payee:        payee.ID,
		TotalOutflow: 2000,
		TotalInflow:  0,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := queries.CreateTransactionCategory(ctx, data.CreateTransactionCategoryParams{
		Transaction:  incoming.ID,
		OtherAccount: sql.NullInt64{Int64: account.ID, Valid: true},
		Category:     sql.NullInt64{},
		Income:       false,
		Outflow:      2000,
		Inflow:       0,
	}); err != nil {
		t.Fatal(err)
	}
	before, err := queries.GetAccountBalanceAsOf(ctx, data.GetAccountBalanceAsOfParams{
		AccountID: account.ID,
		Date:      mustTime(t, 2026, 1, 1),
	})
	if err != nil {
		t.Fatal(err)
	}
	if before != -1000 {
		t.Errorf("expected balance -1000 as of Jan 1, got %d", before)
	}
	after, err := queries.GetAccountBalanceAsOf(ctx, data.GetAccountBalanceAsOfParams{
		AccountID: account.ID,
		Date:      mustTime(t, 2026, 1, 31),
	})
	if err != nil {
		t.Fatal(err)
	}
	if after != 6000 {
		t.Errorf("expected balance 6000 as of Jan 31, got %d", after)
	}
}