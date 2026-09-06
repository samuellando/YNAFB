package data_test

import (
	"testing"

	"samuellando.com/YNAFB/data"
)

func TestReconcileAccountTransactions(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	account := newAccount(t, queries, ctx, budget, "testaccount")
	payee := newPayee(t, queries, ctx, budget, "testpayee")
	tx1 := newTrx(t, queries, ctx, budget, account, payee, mustTime(t, 2026, 1, 1), 1000, 0, "")
	newTrx(t, queries, ctx, budget, account, payee, mustTime(t, 2026, 3, 1), 2000, 0, "")
	n, err := queries.ReconcileAccountTransactions(ctx, data.ReconcileAccountTransactionsParams{
		BudgetID: budget.ID,
		ID:       account.ID,
		LoginID:  budget.LoginID,
		Date:     mustTime(t, 2026, 2, 1),
	})
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("Expected 1 reconciled transaction, got %d", n)
	}
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM reconciliation WHERE account_id = ?`, account.ID).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("Expected 1 reconciliation row, got %d", count)
	}
	if err := db.QueryRow(`SELECT COUNT(*) FROM reconciliation WHERE account_id = ? AND trx_id = ?`, account.ID, tx1.ID).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("Expected the transaction before the cutoff date to be reconciled by trx_id, got %d", count)
	}
}

func TestReconcileAccountTransactionsReconcilesTransferLine(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	account := newAccount(t, queries, ctx, budget, "testaccount")
	otherAccount := newAccount(t, queries, ctx, budget, "otheraccount")
	payee := newPayee(t, queries, ctx, budget, "testpayee")
	tx := newTrx(t, queries, ctx, budget, otherAccount, payee, mustTime(t, 2026, 1, 15), 3000, 0, "")
	line := newTransfer(t, queries, ctx, budget, tx, account, 3000, 0)
	n, err := queries.ReconcileAccountTransactions(ctx, data.ReconcileAccountTransactionsParams{
		BudgetID: budget.ID,
		ID:       account.ID,
		LoginID:  budget.LoginID,
		Date:     mustTime(t, 2026, 2, 1),
	})
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("Expected 1 reconciled transaction, got %d", n)
	}
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM reconciliation WHERE account_id = ? AND trx_line_id = ?`, account.ID, line.ID).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("Expected the transfer line to be reconciled by trx_line_id, got %d", count)
	}
	if err := db.QueryRow(`SELECT COUNT(*) FROM reconciliation WHERE account_id = ? AND trx_id = ?`, account.ID, tx.ID).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("A transfer into the account should reconcile the line, not the transaction, got %d trx_id rows", count)
	}
}

func TestReconcileAccountTransactionsOwnedAndTransferred(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	account := newAccount(t, queries, ctx, budget, "testaccount")
	target := newAccount(t, queries, ctx, budget, "target")
	payee := newPayee(t, queries, ctx, budget, "testpayee")
	newTrx(t, queries, ctx, budget, account, payee, mustTime(t, 2026, 1, 1), 1000, 0, "")
	transferOut := newTrx(t, queries, ctx, budget, account, payee, mustTime(t, 2026, 1, 2), 2000, 0, "")
	newTransfer(t, queries, ctx, budget, transferOut, target, 2000, 0)
	n, err := queries.ReconcileAccountTransactions(ctx, data.ReconcileAccountTransactionsParams{
		BudgetID: budget.ID,
		ID:       account.ID,
		LoginID:  budget.LoginID,
		Date:     mustTime(t, 2026, 2, 1),
	})
	if err != nil {
		t.Fatal(err)
	}
	if n != 2 {
		t.Fatalf("Expected 2 owned transactions reconciled, got %d", n)
	}
	n, err = queries.ReconcileAccountTransactions(ctx, data.ReconcileAccountTransactionsParams{
		BudgetID: budget.ID,
		ID:       target.ID,
		LoginID:  budget.LoginID,
		Date:     mustTime(t, 2026, 2, 1),
	})
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("Expected the transfer line reconciled in the target, got %d", n)
	}
}

func TestReconcileAccountTransactionsIdempotent(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	account := newAccount(t, queries, ctx, budget, "testaccount")
	payee := newPayee(t, queries, ctx, budget, "testpayee")
	newTrx(t, queries, ctx, budget, account, payee, mustTime(t, 2026, 1, 1), 1000, 0, "")
	params := data.ReconcileAccountTransactionsParams{
		BudgetID: budget.ID,
		ID:       account.ID,
		LoginID:  budget.LoginID,
		Date:     mustTime(t, 2026, 2, 1),
	}
	n, err := queries.ReconcileAccountTransactions(ctx, params)
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("Expected 1 reconciled transaction on first reconcile, got %d", n)
	}
	n, err = queries.ReconcileAccountTransactions(ctx, params)
	if err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Fatalf("Expected 0 affected rows on second reconcile, got %d", n)
	}
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM reconciliation WHERE account_id = ?`, account.ID).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("Expected a single reconciliation row after repeated reconciles, got %d", count)
	}
}

func TestReconcileAccountTransactionsEmpty(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	account := newAccount(t, queries, ctx, budget, "testaccount")
	n, err := queries.ReconcileAccountTransactions(ctx, data.ReconcileAccountTransactionsParams{
		BudgetID: budget.ID,
		ID:       account.ID,
		LoginID:  budget.LoginID,
		Date:     mustTime(t, 2026, 2, 1),
	})
	if err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Fatalf("Expected 0 reconciled transactions, got %d", n)
	}
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM reconciliation WHERE account_id = ?`, account.ID).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("Expected no reconciliation rows, got %d", count)
	}
}

func TestReconcileAccountTransactionsScopedToBudget(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget1 := newBudget(t, queries, ctx, "budget1")
	budget2 := newBudget(t, queries, ctx, "budget2")
	account1 := newAccount(t, queries, ctx, budget1, "account1")
	payee1 := newPayee(t, queries, ctx, budget1, "payee1")
	newTrx(t, queries, ctx, budget1, account1, payee1, mustTime(t, 2026, 1, 1), 1000, 0, "")
	n, err := queries.ReconcileAccountTransactions(ctx, data.ReconcileAccountTransactionsParams{
		BudgetID: budget2.ID,
		ID:       account1.ID,
		LoginID:  budget1.LoginID,
		Date:     mustTime(t, 2026, 2, 1),
	})
	if err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Fatalf("Reconciling with a mismatched budget should affect 0 rows, got %d", n)
	}
}

func TestDeleteAccountCascadesReconciliations(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	account := newAccount(t, queries, ctx, budget, "testaccount")
	payee := newPayee(t, queries, ctx, budget, "testpayee")
	newTrx(t, queries, ctx, budget, account, payee, mustTime(t, 2026, 1, 1), 1000, 0, "")
	if _, err := queries.ReconcileAccountTransactions(ctx, data.ReconcileAccountTransactionsParams{
		BudgetID: budget.ID,
		ID:       account.ID,
		LoginID:  budget.LoginID,
		Date:     mustTime(t, 2026, 2, 1),
	}); err != nil {
		t.Fatal(err)
	}
	err := queries.DeleteAccount(ctx, data.DeleteAccountParams{ID: account.ID, BudgetID: budget.ID, LoginID: budget.LoginID})
	if err != nil {
		t.Fatal(err)
	}
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM reconciliation WHERE account_id = ?`, account.ID).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatal("Deleting an account should cascade delete its reconciliations")
	}
}

func TestDeleteTrxCascadesReconciliations(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	account := newAccount(t, queries, ctx, budget, "testaccount")
	payee := newPayee(t, queries, ctx, budget, "testpayee")
	transaction := newTrx(t, queries, ctx, budget, account, payee, mustTime(t, 2026, 1, 1), 1000, 0, "")
	if _, err := queries.ReconcileAccountTransactions(ctx, data.ReconcileAccountTransactionsParams{
		BudgetID: budget.ID,
		ID:       account.ID,
		LoginID:  budget.LoginID,
		Date:     mustTime(t, 2026, 2, 1),
	}); err != nil {
		t.Fatal(err)
	}
	err := queries.DeleteTrx(ctx, data.DeleteTrxParams{ID: transaction.ID, BudgetID: budget.ID, LoginID: budget.LoginID})
	if err != nil {
		t.Fatal(err)
	}
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM reconciliation WHERE trx_id = ?`, transaction.ID).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatal("Deleting a transaction should cascade delete its reconciliations")
	}
}

func TestDeleteTrxLineCascadesReconciliations(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	account := newAccount(t, queries, ctx, budget, "testaccount")
	target := newAccount(t, queries, ctx, budget, "target")
	payee := newPayee(t, queries, ctx, budget, "testpayee")
	tx := newTrx(t, queries, ctx, budget, account, payee, mustTime(t, 2026, 1, 1), 2000, 0, "")
	line := newTransfer(t, queries, ctx, budget, tx, target, 2000, 0)
	if _, err := queries.ReconcileAccountTransactions(ctx, data.ReconcileAccountTransactionsParams{
		BudgetID: budget.ID,
		ID:       target.ID,
		LoginID:  budget.LoginID,
		Date:     mustTime(t, 2026, 2, 1),
	}); err != nil {
		t.Fatal(err)
	}
	err := queries.DeleteTrxLine(ctx, data.DeleteTrxLineParams{ID: line.ID, BudgetID: budget.ID, LoginID: budget.LoginID})
	if err != nil {
		t.Fatal(err)
	}
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM reconciliation WHERE trx_line_id = ?`, line.ID).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatal("Deleting a transaction line should cascade delete its reconciliations")
	}
}

func TestScopingReconcileAccountTransactionsScopedByLogin(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budgetA := newBudget(t, queries, ctx, "budgetA")
	budgetB := newBudget(t, queries, ctx, "budgetB")
	accountB := newAccount(t, queries, ctx, budgetB, "accountB")
	payeeB := newPayee(t, queries, ctx, budgetB, "payeeB")
	newTrx(t, queries, ctx, budgetB, accountB, payeeB, mustTime(t, 2026, 1, 1), 1000, 0, "")

	n, err := queries.ReconcileAccountTransactions(ctx, data.ReconcileAccountTransactionsParams{
		BudgetID: budgetB.ID,
		ID:       accountB.ID,
		LoginID:  budgetA.LoginID,
		Date:     mustTime(t, 2026, 2, 1),
	})
	if err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Fatalf("expected 0 reconciled transactions across logins, got %d", n)
	}
}
