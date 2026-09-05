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
	budget := newBudget(t, queries, ctx, "testBudget")
	balances, err := queries.ListAccountsBalances(ctx, data.ListAccountsBalancesParams{BudgetID: budget.ID})
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
	budget := newBudget(t, queries, ctx, "testBudget")
	account := newAccount(t, queries, ctx, budget.ID, "testaccount")
	balances, err := queries.ListAccountsBalances(ctx, data.ListAccountsBalancesParams{BudgetID: budget.ID})
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
	budget := newBudget(t, queries, ctx, "testBudget")
	newAccount(t, queries, ctx, budget.ID, "account1")
	newAccount(t, queries, ctx, budget.ID, "account2")
	balances, err := queries.ListAccountsBalances(ctx, data.ListAccountsBalancesParams{BudgetID: budget.ID})
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
	budget1 := newBudget(t, queries, ctx, "testBudget1")
	budget2 := newBudget(t, queries, ctx, "testBudget2")
	account1 := newAccount(t, queries, ctx, budget1.ID, "same")
	account2 := newAccount(t, queries, ctx, budget2.ID, "same")
	balances1, err := queries.ListAccountsBalances(ctx, data.ListAccountsBalancesParams{BudgetID: budget1.ID})
	if err != nil {
		t.Fatal(err)
	}
	if len(balances1) != 1 {
		t.Fatalf("expected one balance for budget1, got %d", len(balances1))
	}
	if balances1[0].ID != account1.ID {
		t.Error("budget1 returned the wrong account")
	}
	balances2, err := queries.ListAccountsBalances(ctx, data.ListAccountsBalancesParams{BudgetID: budget2.ID})
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
	budget := newBudget(t, queries, ctx, "testBudget")
	account := newAccount(t, queries, ctx, budget.ID, "testaccount")
	payee := newPayee(t, queries, ctx, budget.ID, "testpayee")
	category := newCategory(t, queries, ctx, budget.ID, "testcategory")
	transaction := newTrx(t, queries, ctx, account, payee, mustTime(t, 2026, 1, 1), 1000, 0, "")
	newCategoryLine(t, queries, ctx, transaction, category, 1000, 0)
	balances, err := queries.ListAccountsBalances(ctx, data.ListAccountsBalancesParams{BudgetID: budget.ID})
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
	budget := newBudget(t, queries, ctx, "testBudget")
	account := newAccount(t, queries, ctx, budget.ID, "testaccount")
	payee := newPayee(t, queries, ctx, budget.ID, "testpayee")
	transaction := newTrx(t, queries, ctx, account, payee, mustTime(t, 2026, 1, 1), 0, 5000, "")
	newIncomeLine(t, queries, ctx, transaction, 5000)
	balances, err := queries.ListAccountsBalances(ctx, data.ListAccountsBalancesParams{BudgetID: budget.ID})
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
	budget := newBudget(t, queries, ctx, "testBudget")
	source := newAccount(t, queries, ctx, budget.ID, "source")
	target := newAccount(t, queries, ctx, budget.ID, "target")
	payee := newPayee(t, queries, ctx, budget.ID, "testpayee")
	transaction := newTrx(t, queries, ctx, source, payee, mustTime(t, 2026, 1, 1), 3000, 0, "")
	newTransfer(t, queries, ctx, transaction, target, 3000, 0)
	balances, err := queries.ListAccountsBalances(ctx, data.ListAccountsBalancesParams{BudgetID: budget.ID})
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
	budget := newBudget(t, queries, ctx, "testBudget")
	source := newAccount(t, queries, ctx, budget.ID, "source")
	target := newAccount(t, queries, ctx, budget.ID, "target")
	payee := newPayee(t, queries, ctx, budget.ID, "testpayee")
	transaction := newTrx(t, queries, ctx, target, payee, mustTime(t, 2026, 1, 1), 0, 500, "")
	newTransfer(t, queries, ctx, transaction, source, 0, 500)
	balances, err := queries.ListAccountsBalances(ctx, data.ListAccountsBalancesParams{BudgetID: budget.ID})
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
	budget := newBudget(t, queries, ctx, "testBudget")
	source := newAccount(t, queries, ctx, budget.ID, "source")
	target1 := newAccount(t, queries, ctx, budget.ID, "target1")
	target2 := newAccount(t, queries, ctx, budget.ID, "target2")
	payee := newPayee(t, queries, ctx, budget.ID, "testpayee")
	transaction := newTrx(t, queries, ctx, source, payee, mustTime(t, 2026, 1, 1), 3000, 0, "")
	newTransfer(t, queries, ctx, transaction, target1, 1000, 0)
	newTransfer(t, queries, ctx, transaction, target2, 2000, 0)
	balances, err := queries.ListAccountsBalances(ctx, data.ListAccountsBalancesParams{BudgetID: budget.ID})
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

func TestListAccountsBalancesMixed(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	account := newAccount(t, queries, ctx, budget.ID, "testaccount")
	otherAccount := newAccount(t, queries, ctx, budget.ID, "otheraccount")
	payee := newPayee(t, queries, ctx, budget.ID, "testpayee")
	category := newCategory(t, queries, ctx, budget.ID, "testcategory")
	spend := newTrx(t, queries, ctx, account, payee, mustTime(t, 2026, 1, 1), 1000, 0, "")
	newCategoryLine(t, queries, ctx, spend, category, 1000, 0)
	income := newTrx(t, queries, ctx, account, payee, mustTime(t, 2026, 1, 2), 0, 5000, "")
	newIncomeLine(t, queries, ctx, income, 5000)
	transfer := newTrx(t, queries, ctx, account, payee, mustTime(t, 2026, 1, 3), 2000, 0, "")
	newTransfer(t, queries, ctx, transfer, otherAccount, 2000, 0)
	balances, err := queries.ListAccountsBalances(ctx, data.ListAccountsBalancesParams{BudgetID: budget.ID})
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
	budget := newBudget(t, queries, ctx, "testBudget")
	account := newAccount(t, queries, ctx, budget.ID, "testaccount")
	payee := newPayee(t, queries, ctx, budget.ID, "testpayee")
	category := newCategory(t, queries, ctx, budget.ID, "testcategory")
	reconciled := newTrx(t, queries, ctx, account, payee, mustTime(t, 2026, 1, 1), 1000, 0, "")
	newCategoryLine(t, queries, ctx, reconciled, category, 1000, 0)
	unreconciled := newTrx(t, queries, ctx, account, payee, mustTime(t, 2026, 2, 1), 2000, 0, "")
	newCategoryLine(t, queries, ctx, unreconciled, category, 2000, 0)
	if _, err := queries.ReconcileAccountTransactions(ctx, data.ReconcileAccountTransactionsParams{
		BudgetID:  budget.ID,
		AccountID: account.ID,
		Date:      mustTime(t, 2026, 1, 31),
	}); err != nil {
		t.Fatal(err)
	}
	balances, err := queries.ListAccountsBalances(ctx, data.ListAccountsBalancesParams{BudgetID: budget.ID})
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
	budget := newBudget(t, queries, ctx, "testBudget")
	source := newAccount(t, queries, ctx, budget.ID, "source")
	target := newAccount(t, queries, ctx, budget.ID, "target")
	payee := newPayee(t, queries, ctx, budget.ID, "testpayee")
	transaction := newTrx(t, queries, ctx, source, payee, mustTime(t, 2026, 1, 1), 3000, 0, "")
	newTransfer(t, queries, ctx, transaction, target, 3000, 0)
	if _, err := queries.ReconcileAccountTransactions(ctx, data.ReconcileAccountTransactionsParams{
		BudgetID:  budget.ID,
		AccountID: target.ID,
		Date:      mustTime(t, 2026, 2, 1),
	}); err != nil {
		t.Fatal(err)
	}
	balances, err := queries.ListAccountsBalances(ctx, data.ListAccountsBalancesParams{BudgetID: budget.ID})
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
	budget := newBudget(t, queries, ctx, "testBudget")
	account := newAccount(t, queries, ctx, budget.ID, "testaccount")
	balances, err := queries.GetAccountBalances(ctx, data.GetAccountBalancesParams{BudgetID: budget.ID, AccountID: account.ID})
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
	budget := newBudget(t, queries, ctx, "testBudget")
	account := newAccount(t, queries, ctx, budget.ID, "testaccount")
	payee := newPayee(t, queries, ctx, budget.ID, "testpayee")
	category := newCategory(t, queries, ctx, budget.ID, "testcategory")
	transaction := newTrx(t, queries, ctx, account, payee, mustTime(t, 2026, 1, 1), 1000, 0, "")
	newCategoryLine(t, queries, ctx, transaction, category, 1000, 0)
	balances, err := queries.GetAccountBalances(ctx, data.GetAccountBalancesParams{BudgetID: budget.ID, AccountID: account.ID})
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
	budget := newBudget(t, queries, ctx, "testBudget")
	account := newAccount(t, queries, ctx, budget.ID, "testaccount")
	payee := newPayee(t, queries, ctx, budget.ID, "testpayee")
	transaction := newTrx(t, queries, ctx, account, payee, mustTime(t, 2026, 1, 1), 0, 5000, "")
	newIncomeLine(t, queries, ctx, transaction, 5000)
	balances, err := queries.GetAccountBalances(ctx, data.GetAccountBalancesParams{BudgetID: budget.ID, AccountID: account.ID})
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
	budget := newBudget(t, queries, ctx, "testBudget")
	source := newAccount(t, queries, ctx, budget.ID, "source")
	target := newAccount(t, queries, ctx, budget.ID, "target")
	payee := newPayee(t, queries, ctx, budget.ID, "testpayee")
	transaction := newTrx(t, queries, ctx, source, payee, mustTime(t, 2026, 1, 1), 3000, 0, "")
	newTransfer(t, queries, ctx, transaction, target, 3000, 0)
	balances, err := queries.GetAccountBalances(ctx, data.GetAccountBalancesParams{BudgetID: budget.ID, AccountID: target.ID})
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
	budget := newBudget(t, queries, ctx, "testBudget")
	source := newAccount(t, queries, ctx, budget.ID, "source")
	target := newAccount(t, queries, ctx, budget.ID, "target")
	payee := newPayee(t, queries, ctx, budget.ID, "testpayee")
	transaction := newTrx(t, queries, ctx, source, payee, mustTime(t, 2026, 1, 1), 3000, 0, "")
	newTransfer(t, queries, ctx, transaction, target, 3000, 0)
	balances, err := queries.GetAccountBalances(ctx, data.GetAccountBalancesParams{BudgetID: budget.ID, AccountID: source.ID})
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
	budget := newBudget(t, queries, ctx, "testBudget")
	source := newAccount(t, queries, ctx, budget.ID, "source")
	target := newAccount(t, queries, ctx, budget.ID, "target")
	payee := newPayee(t, queries, ctx, budget.ID, "testpayee")
	transaction := newTrx(t, queries, ctx, target, payee, mustTime(t, 2026, 1, 1), 0, 500, "")
	newTransfer(t, queries, ctx, transaction, source, 0, 500)
	balances, err := queries.GetAccountBalances(ctx, data.GetAccountBalancesParams{BudgetID: budget.ID, AccountID: target.ID})
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
	budget := newBudget(t, queries, ctx, "testBudget")
	account := newAccount(t, queries, ctx, budget.ID, "testaccount")
	payee := newPayee(t, queries, ctx, budget.ID, "testpayee")
	category := newCategory(t, queries, ctx, budget.ID, "testcategory")
	reconciled := newTrx(t, queries, ctx, account, payee, mustTime(t, 2026, 1, 1), 1000, 0, "")
	newCategoryLine(t, queries, ctx, reconciled, category, 1000, 0)
	unreconciled := newTrx(t, queries, ctx, account, payee, mustTime(t, 2026, 2, 1), 2000, 0, "")
	newCategoryLine(t, queries, ctx, unreconciled, category, 2000, 0)
	if _, err := queries.ReconcileAccountTransactions(ctx, data.ReconcileAccountTransactionsParams{
		BudgetID:  budget.ID,
		AccountID: account.ID,
		Date:      mustTime(t, 2026, 1, 31),
	}); err != nil {
		t.Fatal(err)
	}
	balances, err := queries.GetAccountBalances(ctx, data.GetAccountBalancesParams{BudgetID: budget.ID, AccountID: account.ID})
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

func TestGetAccountBalancesNonExistent(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	_, err := queries.GetAccountBalances(ctx, data.GetAccountBalancesParams{BudgetID: budget.ID, AccountID: 99})
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("expected sql.ErrNoRows, got %v", err)
	}
}

func TestGetAccountBalanceAsOfNoTransactions(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	account := newAccount(t, queries, ctx, budget.ID, "testaccount")
	balance, err := queries.GetAccountBalanceAsOf(ctx, data.GetAccountBalanceAsOfParams{
		BudgetID:  budget.ID,
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
	budget := newBudget(t, queries, ctx, "testBudget")
	account := newAccount(t, queries, ctx, budget.ID, "testaccount")
	payee := newPayee(t, queries, ctx, budget.ID, "testpayee")
	newTrx(t, queries, ctx, account, payee, mustTime(t, 2026, 1, 15), 1000, 0, "")
	balance, err := queries.GetAccountBalanceAsOf(ctx, data.GetAccountBalanceAsOfParams{
		BudgetID:  budget.ID,
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
	budget := newBudget(t, queries, ctx, "testBudget")
	account := newAccount(t, queries, ctx, budget.ID, "testaccount")
	payee := newPayee(t, queries, ctx, budget.ID, "testpayee")
	newTrx(t, queries, ctx, account, payee, mustTime(t, 2026, 1, 15), 1000, 0, "")
	balance, err := queries.GetAccountBalanceAsOf(ctx, data.GetAccountBalanceAsOfParams{
		BudgetID:  budget.ID,
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
	budget := newBudget(t, queries, ctx, "testBudget")
	account := newAccount(t, queries, ctx, budget.ID, "testaccount")
	payee := newPayee(t, queries, ctx, budget.ID, "testpayee")
	category := newCategory(t, queries, ctx, budget.ID, "testcategory")
	transaction := newTrx(t, queries, ctx, account, payee, mustTime(t, 2026, 1, 1), 1000, 0, "")
	newCategoryLine(t, queries, ctx, transaction, category, 1000, 0)
	balance, err := queries.GetAccountBalanceAsOf(ctx, data.GetAccountBalanceAsOfParams{
		BudgetID:  budget.ID,
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
	budget := newBudget(t, queries, ctx, "testBudget")
	account := newAccount(t, queries, ctx, budget.ID, "testaccount")
	payee := newPayee(t, queries, ctx, budget.ID, "testpayee")
	transaction := newTrx(t, queries, ctx, account, payee, mustTime(t, 2026, 1, 1), 0, 5000, "")
	newIncomeLine(t, queries, ctx, transaction, 5000)
	balance, err := queries.GetAccountBalanceAsOf(ctx, data.GetAccountBalanceAsOfParams{
		BudgetID:  budget.ID,
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
	budget := newBudget(t, queries, ctx, "testBudget")
	source := newAccount(t, queries, ctx, budget.ID, "source")
	target := newAccount(t, queries, ctx, budget.ID, "target")
	payee := newPayee(t, queries, ctx, budget.ID, "testpayee")
	transaction := newTrx(t, queries, ctx, source, payee, mustTime(t, 2026, 1, 1), 3000, 0, "")
	newTransfer(t, queries, ctx, transaction, target, 3000, 0)
	sourceBalance, err := queries.GetAccountBalanceAsOf(ctx, data.GetAccountBalanceAsOfParams{
		BudgetID:  budget.ID,
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
		BudgetID:  budget.ID,
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
	budget := newBudget(t, queries, ctx, "testBudget")
	source := newAccount(t, queries, ctx, budget.ID, "source")
	target := newAccount(t, queries, ctx, budget.ID, "target")
	payee := newPayee(t, queries, ctx, budget.ID, "testpayee")
	transaction := newTrx(t, queries, ctx, target, payee, mustTime(t, 2026, 1, 1), 0, 500, "")
	newTransfer(t, queries, ctx, transaction, source, 0, 500)
	targetBalance, err := queries.GetAccountBalanceAsOf(ctx, data.GetAccountBalanceAsOfParams{
		BudgetID:  budget.ID,
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
		BudgetID:  budget.ID,
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
	budget := newBudget(t, queries, ctx, "testBudget")
	account := newAccount(t, queries, ctx, budget.ID, "testaccount")
	payee := newPayee(t, queries, ctx, budget.ID, "testpayee")
	newTrx(t, queries, ctx, account, payee, mustTime(t, 2026, 1, 1), 1000, 0, "")
	newTrx(t, queries, ctx, account, payee, mustTime(t, 2026, 2, 1), 2000, 0, "")
	balance, err := queries.GetAccountBalanceAsOf(ctx, data.GetAccountBalanceAsOfParams{
		BudgetID:  budget.ID,
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
		BudgetID:  budget.ID,
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
	budget := newBudget(t, queries, ctx, "testBudget")
	source := newAccount(t, queries, ctx, budget.ID, "source")
	target := newAccount(t, queries, ctx, budget.ID, "target")
	payee := newPayee(t, queries, ctx, budget.ID, "testpayee")
	transaction := newTrx(t, queries, ctx, source, payee, mustTime(t, 2026, 2, 1), 3000, 0, "")
	newTransfer(t, queries, ctx, transaction, target, 3000, 0)
	balance, err := queries.GetAccountBalanceAsOf(ctx, data.GetAccountBalanceAsOfParams{
		BudgetID:  budget.ID,
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
	budget := newBudget(t, queries, ctx, "testBudget")
	account := newAccount(t, queries, ctx, budget.ID, "testaccount")
	otherAccount := newAccount(t, queries, ctx, budget.ID, "otheraccount")
	payee := newPayee(t, queries, ctx, budget.ID, "testpayee")
	category := newCategory(t, queries, ctx, budget.ID, "testcategory")
	spend := newTrx(t, queries, ctx, account, payee, mustTime(t, 2026, 1, 1), 1000, 0, "")
	newCategoryLine(t, queries, ctx, spend, category, 1000, 0)
	income := newTrx(t, queries, ctx, account, payee, mustTime(t, 2026, 1, 2), 0, 5000, "")
	newIncomeLine(t, queries, ctx, income, 5000)
	incoming := newTrx(t, queries, ctx, otherAccount, payee, mustTime(t, 2026, 1, 3), 2000, 0, "")
	newTransfer(t, queries, ctx, incoming, account, 2000, 0)
	before, err := queries.GetAccountBalanceAsOf(ctx, data.GetAccountBalanceAsOfParams{
		BudgetID:  budget.ID,
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
		BudgetID:  budget.ID,
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