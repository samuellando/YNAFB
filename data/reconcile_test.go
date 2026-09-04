package data_test

import (
	"testing"

	"samuellando.com/YNAFB/data"
)

func TestReconcileAccountTransactions(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	account := newAccount(t, queries, ctx, budget.ID, "testaccount")
	payee := newPayee(t, queries, ctx, budget.ID, "testpayee")
	tx1 := newTransaction(t, queries, ctx, account.ID, payee.ID, mustTime(t, 2026, 1, 1), 1000, 0, "")
	newTransaction(t, queries, ctx, account.ID, payee.ID, mustTime(t, 2026, 3, 1), 2000, 0, "")
	n, err := queries.ReconcileAccountTransactions(ctx, data.ReconcileAccountTransactionsParams{
		AccountID: account.ID,
		Date:      mustTime(t, 2026, 2, 1),
	})
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("Expected 1 reconciled transaction, got %d", n)
	}
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM reconciliation WHERE account = ?`, account.ID).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("Expected 1 reconciliation row, got %d", count)
	}
	if err := db.QueryRow(`SELECT COUNT(*) FROM reconciliation WHERE account = ? AND "transaction" = ?`, account.ID, tx1.ID).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("Expected the transaction before the cutoff date to be reconciled, got %d", count)
	}
}

func TestReconcileAccountTransactionsIncludesTransfers(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	account := newAccount(t, queries, ctx, budget.ID, "testaccount")
	otherAccount := newAccount(t, queries, ctx, budget.ID, "otheraccount")
	payee := newPayee(t, queries, ctx, budget.ID, "testpayee")
	tx := newTransaction(t, queries, ctx, otherAccount.ID, payee.ID, mustTime(t, 2026, 1, 15), 3000, 0, "")
	newTransfer(t, queries, ctx, tx.ID, account.ID, 3000, 0)
	n, err := queries.ReconcileAccountTransactions(ctx, data.ReconcileAccountTransactionsParams{
		AccountID: account.ID,
		Date:      mustTime(t, 2026, 2, 1),
	})
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("Expected 1 reconciled transaction, got %d", n)
	}
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM reconciliation WHERE account = ? AND "transaction" = ?`, account.ID, tx.ID).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("Expected the transfer into the account to be reconciled, got %d", count)
	}
}

func TestReconcileAccountTransactionsIdempotent(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	account := newAccount(t, queries, ctx, budget.ID, "testaccount")
	payee := newPayee(t, queries, ctx, budget.ID, "testpayee")
	newTransaction(t, queries, ctx, account.ID, payee.ID, mustTime(t, 2026, 1, 1), 1000, 0, "")
	params := data.ReconcileAccountTransactionsParams{
		AccountID: account.ID,
		Date:      mustTime(t, 2026, 2, 1),
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
	if err := db.QueryRow(`SELECT COUNT(*) FROM reconciliation WHERE account = ?`, account.ID).Scan(&count); err != nil {
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
	account := newAccount(t, queries, ctx, budget.ID, "testaccount")
	n, err := queries.ReconcileAccountTransactions(ctx, data.ReconcileAccountTransactionsParams{
		AccountID: account.ID,
		Date:      mustTime(t, 2026, 2, 1),
	})
	if err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Fatalf("Expected 0 reconciled transactions, got %d", n)
	}
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM reconciliation WHERE account = ?`, account.ID).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("Expected no reconciliation rows, got %d", count)
	}
}

func TestDeleteAccountCascadesReconciliations(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	account := newAccount(t, queries, ctx, budget.ID, "testaccount")
	payee := newPayee(t, queries, ctx, budget.ID, "testpayee")
	newTransaction(t, queries, ctx, account.ID, payee.ID, mustTime(t, 2026, 1, 1), 1000, 0, "")
	if _, err := queries.ReconcileAccountTransactions(ctx, data.ReconcileAccountTransactionsParams{
		AccountID: account.ID,
		Date:      mustTime(t, 2026, 2, 1),
	}); err != nil {
		t.Fatal(err)
	}
	err := queries.DeleteAccount(ctx, account.ID)
	if err != nil {
		t.Fatal(err)
	}
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM reconciliation WHERE account = ?`, account.ID).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatal("Deleting an account should cascade delete its reconciliations")
	}
}

func TestDeleteTransactionCascadesReconciliations(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	account := newAccount(t, queries, ctx, budget.ID, "testaccount")
	payee := newPayee(t, queries, ctx, budget.ID, "testpayee")
	transaction := newTransaction(t, queries, ctx, account.ID, payee.ID, mustTime(t, 2026, 1, 1), 1000, 0, "")
	if _, err := queries.ReconcileAccountTransactions(ctx, data.ReconcileAccountTransactionsParams{
		AccountID: account.ID,
		Date:      mustTime(t, 2026, 2, 1),
	}); err != nil {
		t.Fatal(err)
	}
	err := queries.DeleteTransaction(ctx, transaction.ID)
	if err != nil {
		t.Fatal(err)
	}
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM reconciliation WHERE "transaction" = ?`, transaction.ID).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatal("Deleting a transaction should cascade delete its reconciliations")
	}
}