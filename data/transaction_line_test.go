package data_test

import (
	"context"
	"database/sql"
	"testing"

	"samuellando.com/YNAFB/data"
)

type transactionContext struct {
	transaction  data.Trx
	category     data.Category
	account      data.Account
	otherAccount data.Account
}

func createTransactionContext(t *testing.T, queries *data.Queries, ctx context.Context) transactionContext {
	t.Helper()
	budget := newBudget(t, queries, ctx, "testBudget")
	account := newAccount(t, queries, ctx, budget.ID, "testaccount")
	otherAccount := newAccount(t, queries, ctx, budget.ID, "otheraccount")
	payee := newPayee(t, queries, ctx, budget.ID, "testpayee")
	category := newCategory(t, queries, ctx, budget.ID, "testcategory")
	transaction := newTrx(t, queries, ctx, account, payee, mustTime(t, 2026, 1, 1), 1000, 0, "")
	return transactionContext{
		transaction:  transaction,
		category:     category,
		account:      account,
		otherAccount: otherAccount,
	}
}

func TestCreateTrxLineCategory(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	tc := createTransactionContext(t, queries, ctx)
	txCategory, err := queries.CreateTrxLine(ctx, data.CreateTrxLineParams{
		BudgetID:   tc.transaction.BudgetID,
		TrxID:      tc.transaction.ID,
		DestAccountID: sql.NullInt64{},
		CategoryID: sql.NullInt64{Int64: tc.category.ID, Valid: true},
		Income:     false,
		Outflow:    1000,
		Inflow:     0,
	})
	if err != nil {
		t.Fatal(err)
	}
	if txCategory.TrxID != tc.transaction.ID {
		t.Error("transaction id does not match")
	}
	if txCategory.DestAccountID.Valid {
		t.Error("dest account should be null for a category spend")
	}
	if !txCategory.CategoryID.Valid || txCategory.CategoryID.Int64 != tc.category.ID {
		t.Error("category id does not match")
	}
	if txCategory.Income {
		t.Error("income should be false")
	}
	if txCategory.Outflow != 1000 {
		t.Error("outflow does not match")
	}
	if txCategory.Inflow != 0 {
		t.Error("inflow does not match")
	}
	if txCategory.ID != 1 {
		t.Error("ID of the first transaction line should be 1")
	}
}

func TestCreateTrxLineDestAccount(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	tc := createTransactionContext(t, queries, ctx)
	txCategory, err := queries.CreateTrxLine(ctx, data.CreateTrxLineParams{
		BudgetID:      tc.transaction.BudgetID,
		TrxID:         tc.transaction.ID,
		DestAccountID: sql.NullInt64{Int64: tc.otherAccount.ID, Valid: true},
		CategoryID:    sql.NullInt64{},
		Income:        false,
		Outflow:       0,
		Inflow:        1000,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !txCategory.DestAccountID.Valid || txCategory.DestAccountID.Int64 != tc.otherAccount.ID {
		t.Error("dest account id does not match")
	}
	if txCategory.CategoryID.Valid {
		t.Error("category should be null for a transfer")
	}
	if txCategory.Income {
		t.Error("income should be false")
	}
}

func TestCreateTrxLineIncome(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	tc := createTransactionContext(t, queries, ctx)
	txCategory, err := queries.CreateTrxLine(ctx, data.CreateTrxLineParams{
		BudgetID:   tc.transaction.BudgetID,
		TrxID:      tc.transaction.ID,
		DestAccountID: sql.NullInt64{},
		CategoryID: sql.NullInt64{},
		Income:     true,
		Outflow:    0,
		Inflow:     5000,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !txCategory.Income {
		t.Error("income should be true")
	}
	if txCategory.CategoryID.Valid {
		t.Error("category should be null for income")
	}
	if txCategory.DestAccountID.Valid {
		t.Error("dest account should be null for income")
	}
}

func TestCreateTrxLineIncomeInvalidOutflow(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	tc := createTransactionContext(t, queries, ctx)
	_, err := queries.CreateTrxLine(ctx, data.CreateTrxLineParams{
		BudgetID:   tc.transaction.BudgetID,
		TrxID:      tc.transaction.ID,
		DestAccountID: sql.NullInt64{},
		CategoryID: sql.NullInt64{},
		Income:     true,
		Outflow:    1000,
		Inflow:     0,
	})
	if err == nil {
		t.Error("Income with a non zero outflow should be rejected")
	}
}

func TestCreateTrxLineIncomeZeroInflow(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	tc := createTransactionContext(t, queries, ctx)
	_, err := queries.CreateTrxLine(ctx, data.CreateTrxLineParams{
		BudgetID:   tc.transaction.BudgetID,
		TrxID:      tc.transaction.ID,
		DestAccountID: sql.NullInt64{},
		CategoryID: sql.NullInt64{},
		Income:     true,
		Outflow:    0,
		Inflow:     0,
	})
	if err == nil {
		t.Error("Income with a zero inflow should be rejected")
	}
}

func TestUpdateTrxLineIncomeViolation(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	tc := createTransactionContext(t, queries, ctx)
	txCategory := newCategoryLine(t, queries, ctx, tc.transaction, tc.category, 1000, 0)
	_, err := queries.UpdateTrxLine(ctx, data.UpdateTrxLineParams{
		TrxID:         tc.transaction.ID,
		DestAccountID: sql.NullInt64{},
		CategoryID:    sql.NullInt64{},
		Income:        true,
		Outflow:       1000,
		Inflow:        0,
		ID:            txCategory.ID,
		BudgetID:      tc.transaction.BudgetID,
	})
	if err == nil {
		t.Error("Updating a transaction line to income with a non zero outflow should be rejected")
	}
}

func TestCreateTrxLineXorViolation(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	tc := createTransactionContext(t, queries, ctx)
	_, err := queries.CreateTrxLine(ctx, data.CreateTrxLineParams{
		BudgetID:      tc.transaction.BudgetID,
		TrxID:         tc.transaction.ID,
		DestAccountID: sql.NullInt64{Int64: tc.otherAccount.ID, Valid: true},
		CategoryID:    sql.NullInt64{Int64: tc.category.ID, Valid: true},
		Income:        false,
		Outflow:       1000,
		Inflow:        0,
	})
	if err == nil {
		t.Error("Setting both dest_account and category should raise an error")
	}
	_, err = queries.CreateTrxLine(ctx, data.CreateTrxLineParams{
		BudgetID:      tc.transaction.BudgetID,
		TrxID:         tc.transaction.ID,
		DestAccountID: sql.NullInt64{},
		CategoryID:    sql.NullInt64{},
		Income:        false,
		Outflow:       1000,
		Inflow:        0,
	})
	if err == nil {
		t.Error("Setting neither dest_account, category nor income should raise an error")
	}
}

func TestCreateTrxLineBothInflowAndOutflowRejected(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	tc := createTransactionContext(t, queries, ctx)
	_, err := queries.CreateTrxLine(ctx, data.CreateTrxLineParams{
		BudgetID:      tc.transaction.BudgetID,
		TrxID:         tc.transaction.ID,
		DestAccountID: sql.NullInt64{},
		CategoryID:    sql.NullInt64{Int64: tc.category.ID, Valid: true},
		Income:        false,
		Outflow:       1000,
		Inflow:        500,
	})
	if err == nil {
		t.Error("A transaction line with both inflow and outflow should be rejected")
	}
}

func TestCreateTrxLineNegativeOutflowRejected(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	tc := createTransactionContext(t, queries, ctx)
	_, err := queries.CreateTrxLine(ctx, data.CreateTrxLineParams{
		BudgetID:      tc.transaction.BudgetID,
		TrxID:         tc.transaction.ID,
		DestAccountID: sql.NullInt64{},
		CategoryID:    sql.NullInt64{Int64: tc.category.ID, Valid: true},
		Income:        false,
		Outflow:       -1000,
		Inflow:        0,
	})
	if err == nil {
		t.Error("A transaction line with a negative outflow should be rejected")
	}
}

func TestCreateTrxLineNegativeInflowRejected(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	tc := createTransactionContext(t, queries, ctx)
	_, err := queries.CreateTrxLine(ctx, data.CreateTrxLineParams{
		BudgetID:      tc.transaction.BudgetID,
		TrxID:         tc.transaction.ID,
		DestAccountID: sql.NullInt64{},
		CategoryID:    sql.NullInt64{Int64: tc.category.ID, Valid: true},
		Income:        false,
		Outflow:       0,
		Inflow:        -500,
	})
	if err == nil {
		t.Error("A transaction line with a negative inflow should be rejected")
	}
}

func TestCreateTrxLineNonExistingTrx(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	tc := createTransactionContext(t, queries, ctx)
	_, err := queries.CreateTrxLine(ctx, data.CreateTrxLineParams{
		BudgetID:      tc.transaction.BudgetID,
		TrxID:         99,
		DestAccountID: sql.NullInt64{},
		CategoryID:    sql.NullInt64{Int64: tc.category.ID, Valid: true},
		Income:        false,
		Outflow:       1000,
		Inflow:        0,
	})
	if err == nil {
		t.Error("Non existing transaction should raise an error")
	}
}

func TestCreateTrxLineNonExistingDestAccount(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	tc := createTransactionContext(t, queries, ctx)
	_, err := queries.CreateTrxLine(ctx, data.CreateTrxLineParams{
		BudgetID:      tc.transaction.BudgetID,
		TrxID:         tc.transaction.ID,
		DestAccountID: sql.NullInt64{Int64: 99, Valid: true},
		CategoryID:    sql.NullInt64{},
		Income:        false,
		Outflow:       0,
		Inflow:        1000,
	})
	if err == nil {
		t.Error("Non existing dest account should raise an error")
	}
}

func TestCreateTrxLineNonExistingCategory(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	tc := createTransactionContext(t, queries, ctx)
	_, err := queries.CreateTrxLine(ctx, data.CreateTrxLineParams{
		BudgetID:      tc.transaction.BudgetID,
		TrxID:         tc.transaction.ID,
		DestAccountID: sql.NullInt64{},
		CategoryID:    sql.NullInt64{Int64: 99, Valid: true},
		Income:        false,
		Outflow:       1000,
		Inflow:        0,
	})
	if err == nil {
		t.Error("Non existing category should raise an error")
	}
}

func TestDeleteTrxCascadesTrxLines(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	tc := createTransactionContext(t, queries, ctx)
	txCategory := newCategoryLine(t, queries, ctx, tc.transaction, tc.category, 1000, 0)
	err := queries.DeleteTrx(ctx, data.DeleteTrxParams{ID: tc.transaction.ID, BudgetID: tc.transaction.BudgetID})
	if err != nil {
		t.Fatal(err)
	}
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM trx_line WHERE id = ?`, txCategory.ID).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatal("Deleting a transaction should cascade delete its transaction lines")
	}
}

func TestDeleteAccountCascadesTrxLines(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	tc := createTransactionContext(t, queries, ctx)
	newTransfer(t, queries, ctx, tc.transaction, tc.otherAccount, 0, 1000)
	err := queries.DeleteAccount(ctx, data.DeleteAccountParams{ID: tc.otherAccount.ID, BudgetID: tc.otherAccount.BudgetID})
	if err != nil {
		t.Fatal(err)
	}
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM trx_line WHERE dest_account_id = ?`, tc.otherAccount.ID).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatal("Deleting an account should cascade delete transaction lines referencing it")
	}
}

func TestDeleteCategoryCascadesTrxLines(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	tc := createTransactionContext(t, queries, ctx)
	newCategoryLine(t, queries, ctx, tc.transaction, tc.category, 1000, 0)
	err := queries.DeleteCategory(ctx, data.DeleteCategoryParams{ID: tc.category.ID, BudgetID: tc.category.BudgetID})
	if err != nil {
		t.Fatal(err)
	}
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM trx_line WHERE category_id = ?`, tc.category.ID).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatal("Deleting a category should cascade delete transaction lines referencing it")
	}
}

func TestUpdateTrxLineBothInflowAndOutflowRejected(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	tc := createTransactionContext(t, queries, ctx)
	txCategory := newCategoryLine(t, queries, ctx, tc.transaction, tc.category, 1000, 0)
	_, err := queries.UpdateTrxLine(ctx, data.UpdateTrxLineParams{
		TrxID:         tc.transaction.ID,
		DestAccountID: sql.NullInt64{},
		CategoryID:    sql.NullInt64{Int64: tc.category.ID, Valid: true},
		Income:        false,
		Outflow:       1000,
		Inflow:        500,
		ID:            txCategory.ID,
		BudgetID:      tc.transaction.BudgetID,
	})
	if err == nil {
		t.Error("Updating a transaction line to have both inflow and outflow should be rejected")
	}
}

func TestUpdateTrxLine(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	tc := createTransactionContext(t, queries, ctx)
	txCategory := newCategoryLine(t, queries, ctx, tc.transaction, tc.category, 1000, 0)
	n, err := queries.UpdateTrxLine(ctx, data.UpdateTrxLineParams{
		TrxID:         tc.transaction.ID,
		DestAccountID: sql.NullInt64{Int64: tc.otherAccount.ID, Valid: true},
		CategoryID:    sql.NullInt64{},
		Income:        false,
		Outflow:       0,
		Inflow:        1000,
		ID:            txCategory.ID,
		BudgetID:      tc.transaction.BudgetID,
	})
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Error("The number of affected rows should be 1")
	}
	var destAccount sql.NullInt64
	if err := db.QueryRow(`SELECT dest_account_id FROM trx_line WHERE id = ?`, txCategory.ID).Scan(&destAccount); err != nil {
		t.Fatal(err)
	}
	if !destAccount.Valid || destAccount.Int64 != tc.otherAccount.ID {
		t.Error("transaction line was not updated to a transfer")
	}
}

func TestDeleteTrxLine(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	tc := createTransactionContext(t, queries, ctx)
	txCategory := newCategoryLine(t, queries, ctx, tc.transaction, tc.category, 1000, 0)
	err := queries.DeleteTrxLine(ctx, data.DeleteTrxLineParams{ID: txCategory.ID, BudgetID: tc.transaction.BudgetID})
	if err != nil {
		t.Fatal(err)
	}
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM trx_line WHERE id = ?`, txCategory.ID).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatal("There should be no transaction line after")
	}
}

func TestDeleteTrxLinesByTrx(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	tc := createTransactionContext(t, queries, ctx)
	for i := 0; i < 2; i++ {
		newCategoryLine(t, queries, ctx, tc.transaction, tc.category, 1000, 0)
	}
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM trx_line WHERE trx_id = ?`, tc.transaction.ID).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 2 {
		t.Fatal("There should be two transaction lines before")
	}
	err := queries.DeleteTrxLinesByTrx(ctx, data.DeleteTrxLinesByTrxParams{TrxID: tc.transaction.ID, BudgetID: tc.transaction.BudgetID})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`SELECT COUNT(*) FROM trx_line WHERE trx_id = ?`, tc.transaction.ID).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatal("There should be no transaction line after")
	}
}