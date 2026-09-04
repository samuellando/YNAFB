package data_test

import (
	"database/sql"
	"testing"

	"samuellando.com/YNAFB/data"
)

func TestReconcileAccountTransactions(t *testing.T) {
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
	tx1, err := queries.CreateTransaction(ctx, data.CreateTransactionParams{
		Date:         mustTime(t, 2026, 1, 1),
		Account:      account.ID,
		Payee:        payee.ID,
		TotalOutflow: 1000,
		TotalInflow:  0,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := queries.CreateTransaction(ctx, data.CreateTransactionParams{
		Date:         mustTime(t, 2026, 3, 1),
		Account:      account.ID,
		Payee:        payee.ID,
		TotalOutflow: 2000,
		TotalInflow:  0,
	}); err != nil {
		t.Fatal(err)
	}
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
	tx, err := queries.CreateTransaction(ctx, data.CreateTransactionParams{
		Date:         mustTime(t, 2026, 1, 15),
		Account:      otherAccount.ID,
		Payee:        payee.ID,
		TotalOutflow: 3000,
		TotalInflow:  0,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := queries.CreateTransactionCategory(ctx, data.CreateTransactionCategoryParams{
		Transaction:  tx.ID,
		OtherAccount: sql.NullInt64{Int64: account.ID, Valid: true},
		Category:     sql.NullInt64{},
		Income:       false,
		Outflow:      3000,
		Inflow:       0,
	}); err != nil {
		t.Fatal(err)
	}
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
	if _, err := queries.ReconcileAccountTransactions(ctx, data.ReconcileAccountTransactionsParams{
		AccountID: account.ID,
		Date:      mustTime(t, 2026, 2, 1),
	}); err != nil {
		t.Fatal(err)
	}
	err = queries.DeleteAccount(ctx, account.ID)
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
		TotalOutflow: 1000,
		TotalInflow:  0,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := queries.ReconcileAccountTransactions(ctx, data.ReconcileAccountTransactionsParams{
		AccountID: account.ID,
		Date:      mustTime(t, 2026, 2, 1),
	}); err != nil {
		t.Fatal(err)
	}
	err = queries.DeleteTransaction(ctx, transaction.ID)
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
