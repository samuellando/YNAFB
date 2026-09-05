package data_test

import (
	"context"
	"database/sql"
	"testing"

	"samuellando.com/YNAFB/data"
)

type transactionContext struct {
	transaction  data.Transaction
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
	transaction := newTransaction(t, queries, ctx, account.ID, payee.ID, mustTime(t, 2026, 1, 1), 1000, 0, "")
	return transactionContext{
		transaction:  transaction,
		category:     category,
		account:      account,
		otherAccount: otherAccount,
	}
}

func TestCreateTransactionCategoryCategory(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	tc := createTransactionContext(t, queries, ctx)
	txCategory, err := queries.CreateTransactionCategory(ctx, data.CreateTransactionCategoryParams{
		Transaction:  tc.transaction.ID,
		OtherAccount: sql.NullInt64{},
		Category:     sql.NullInt64{Int64: tc.category.ID, Valid: true},
		Income:       false,
		Outflow:      1000,
		Inflow:       0,
	})
	if err != nil {
		t.Fatal(err)
	}
	if txCategory.Transaction != tc.transaction.ID {
		t.Error("transaction id does not match")
	}
	if txCategory.OtherAccount.Valid {
		t.Error("other account should be null for a category spend")
	}
	if !txCategory.Category.Valid || txCategory.Category.Int64 != tc.category.ID {
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
		t.Error("ID of the first transaction category should be 1")
	}
}

func TestCreateTransactionCategoryOtherAccount(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	tc := createTransactionContext(t, queries, ctx)
	txCategory, err := queries.CreateTransactionCategory(ctx, data.CreateTransactionCategoryParams{
		Transaction:  tc.transaction.ID,
		OtherAccount: sql.NullInt64{Int64: tc.otherAccount.ID, Valid: true},
		Category:     sql.NullInt64{},
		Income:       false,
		Outflow:      0,
		Inflow:       1000,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !txCategory.OtherAccount.Valid || txCategory.OtherAccount.Int64 != tc.otherAccount.ID {
		t.Error("other account id does not match")
	}
	if txCategory.Category.Valid {
		t.Error("category should be null for a transfer")
	}
	if txCategory.Income {
		t.Error("income should be false")
	}
}

func TestCreateTransactionCategoryIncome(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	tc := createTransactionContext(t, queries, ctx)
	txCategory, err := queries.CreateTransactionCategory(ctx, data.CreateTransactionCategoryParams{
		Transaction:  tc.transaction.ID,
		OtherAccount: sql.NullInt64{},
		Category:     sql.NullInt64{},
		Income:       true,
		Outflow:      0,
		Inflow:       5000,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !txCategory.Income {
		t.Error("income should be true")
	}
	if txCategory.Category.Valid {
		t.Error("category should be null for income")
	}
	if txCategory.OtherAccount.Valid {
		t.Error("other account should be null for income")
	}
}

func TestCreateTransactionCategoryIncomeInvalidOutflow(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	tc := createTransactionContext(t, queries, ctx)
	_, err := queries.CreateTransactionCategory(ctx, data.CreateTransactionCategoryParams{
		Transaction:  tc.transaction.ID,
		OtherAccount: sql.NullInt64{},
		Category:     sql.NullInt64{},
		Income:       true,
		Outflow:      1000,
		Inflow:       0,
	})
	if err == nil {
		t.Error("Income with a non zero outflow should be rejected")
	}
}

func TestCreateTransactionCategoryIncomeZeroInflow(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	tc := createTransactionContext(t, queries, ctx)
	_, err := queries.CreateTransactionCategory(ctx, data.CreateTransactionCategoryParams{
		Transaction:  tc.transaction.ID,
		OtherAccount: sql.NullInt64{},
		Category:     sql.NullInt64{},
		Income:       true,
		Outflow:      0,
		Inflow:       0,
	})
	if err == nil {
		t.Error("Income with a zero inflow should be rejected")
	}
}

func TestUpdateTransactionCategoryIncomeViolation(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	tc := createTransactionContext(t, queries, ctx)
	txCategory := newCategoryLine(t, queries, ctx, tc.transaction.ID, tc.category.ID, 1000, 0)
	_, err := queries.UpdateTransactionCategory(ctx, data.UpdateTransactionCategoryParams{
		Transaction:  tc.transaction.ID,
		OtherAccount: sql.NullInt64{},
		Category:     sql.NullInt64{},
		Income:       true,
		Outflow:      1000,
		Inflow:       0,
		ID:           txCategory.ID,
	})
	if err == nil {
		t.Error("Updating a transaction category to income with a non zero outflow should be rejected")
	}
}

func TestCreateTransactionCategoryXorViolation(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	tc := createTransactionContext(t, queries, ctx)
	_, err := queries.CreateTransactionCategory(ctx, data.CreateTransactionCategoryParams{
		Transaction:  tc.transaction.ID,
		OtherAccount: sql.NullInt64{Int64: tc.otherAccount.ID, Valid: true},
		Category:     sql.NullInt64{Int64: tc.category.ID, Valid: true},
		Income:       false,
		Outflow:      1000,
		Inflow:       0,
	})
	if err == nil {
		t.Error("Setting both other_account and category should raise an error")
	}
	_, err = queries.CreateTransactionCategory(ctx, data.CreateTransactionCategoryParams{
		Transaction:  tc.transaction.ID,
		OtherAccount: sql.NullInt64{},
		Category:     sql.NullInt64{},
		Income:       false,
		Outflow:      1000,
		Inflow:       0,
	})
	if err == nil {
		t.Error("Setting neither other_account, category nor income should raise an error")
	}
}

func TestCreateTransactionCategoryBothInflowAndOutflowRejected(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	tc := createTransactionContext(t, queries, ctx)
	_, err := queries.CreateTransactionCategory(ctx, data.CreateTransactionCategoryParams{
		Transaction:  tc.transaction.ID,
		OtherAccount: sql.NullInt64{},
		Category:     sql.NullInt64{Int64: tc.category.ID, Valid: true},
		Income:       false,
		Outflow:      1000,
		Inflow:       500,
	})
	if err == nil {
		t.Error("A transaction category with both inflow and outflow should be rejected")
	}
}

func TestCreateTransactionCategoryNegativeOutflowRejected(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	tc := createTransactionContext(t, queries, ctx)
	_, err := queries.CreateTransactionCategory(ctx, data.CreateTransactionCategoryParams{
		Transaction:  tc.transaction.ID,
		OtherAccount: sql.NullInt64{},
		Category:     sql.NullInt64{Int64: tc.category.ID, Valid: true},
		Income:       false,
		Outflow:      -1000,
		Inflow:       0,
	})
	if err == nil {
		t.Error("A transaction category with a negative outflow should be rejected")
	}
}

func TestCreateTransactionCategoryNegativeInflowRejected(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	tc := createTransactionContext(t, queries, ctx)
	_, err := queries.CreateTransactionCategory(ctx, data.CreateTransactionCategoryParams{
		Transaction:  tc.transaction.ID,
		OtherAccount: sql.NullInt64{},
		Category:     sql.NullInt64{Int64: tc.category.ID, Valid: true},
		Income:       false,
		Outflow:      0,
		Inflow:       -500,
	})
	if err == nil {
		t.Error("A transaction category with a negative inflow should be rejected")
	}
}

func TestCreateTransactionCategoryNonExistingTransaction(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	tc := createTransactionContext(t, queries, ctx)
	_, err := queries.CreateTransactionCategory(ctx, data.CreateTransactionCategoryParams{
		Transaction:  99,
		OtherAccount: sql.NullInt64{},
		Category:     sql.NullInt64{Int64: tc.category.ID, Valid: true},
		Income:       false,
		Outflow:      1000,
		Inflow:       0,
	})
	if err == nil {
		t.Error("Non existing transaction should raise an error")
	}
}

func TestCreateTransactionCategoryNonExistingOtherAccount(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	tc := createTransactionContext(t, queries, ctx)
	_, err := queries.CreateTransactionCategory(ctx, data.CreateTransactionCategoryParams{
		Transaction:  tc.transaction.ID,
		OtherAccount: sql.NullInt64{Int64: 99, Valid: true},
		Category:     sql.NullInt64{},
		Income:       false,
		Outflow:      0,
		Inflow:       1000,
	})
	if err == nil {
		t.Error("Non existing other account should raise an error")
	}
}

func TestCreateTransactionCategoryNonExistingCategory(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	tc := createTransactionContext(t, queries, ctx)
	_, err := queries.CreateTransactionCategory(ctx, data.CreateTransactionCategoryParams{
		Transaction:  tc.transaction.ID,
		OtherAccount: sql.NullInt64{},
		Category:     sql.NullInt64{Int64: 99, Valid: true},
		Income:       false,
		Outflow:      1000,
		Inflow:       0,
	})
	if err == nil {
		t.Error("Non existing category should raise an error")
	}
}

func TestDeleteTransactionCascadesTransactionCategories(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	tc := createTransactionContext(t, queries, ctx)
	txCategory := newCategoryLine(t, queries, ctx, tc.transaction.ID, tc.category.ID, 1000, 0)
	err := queries.DeleteTransaction(ctx, tc.transaction.ID)
	if err != nil {
		t.Fatal(err)
	}
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM "transaction_category" WHERE id = ?`, txCategory.ID).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatal("Deleting a transaction should cascade delete its transaction categories")
	}
}

func TestDeleteAccountCascadesTransactionCategories(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	tc := createTransactionContext(t, queries, ctx)
	newTransfer(t, queries, ctx, tc.transaction.ID, tc.otherAccount.ID, 0, 1000)
	err := queries.DeleteAccount(ctx, tc.otherAccount.ID)
	if err != nil {
		t.Fatal(err)
	}
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM "transaction_category" WHERE other_account = ?`, tc.otherAccount.ID).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatal("Deleting an account should cascade delete transaction categories referencing it")
	}
}

func TestDeleteCategoryCascadesTransactionCategories(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	tc := createTransactionContext(t, queries, ctx)
	newCategoryLine(t, queries, ctx, tc.transaction.ID, tc.category.ID, 1000, 0)
	err := queries.DeleteCategory(ctx, tc.category.ID)
	if err != nil {
		t.Fatal(err)
	}
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM "transaction_category" WHERE category = ?`, tc.category.ID).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatal("Deleting a category should cascade delete transaction categories referencing it")
	}
}

func TestUpdateTransactionCategoryBothInflowAndOutflowRejected(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	tc := createTransactionContext(t, queries, ctx)
	txCategory := newCategoryLine(t, queries, ctx, tc.transaction.ID, tc.category.ID, 1000, 0)
	_, err := queries.UpdateTransactionCategory(ctx, data.UpdateTransactionCategoryParams{
		Transaction:  tc.transaction.ID,
		OtherAccount: sql.NullInt64{},
		Category:     sql.NullInt64{Int64: tc.category.ID, Valid: true},
		Income:       false,
		Outflow:      1000,
		Inflow:       500,
		ID:           txCategory.ID,
	})
	if err == nil {
		t.Error("Updating a transaction category to have both inflow and outflow should be rejected")
	}
}

func TestUpdateTransactionCategory(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	tc := createTransactionContext(t, queries, ctx)
	txCategory := newCategoryLine(t, queries, ctx, tc.transaction.ID, tc.category.ID, 1000, 0)
	n, err := queries.UpdateTransactionCategory(ctx, data.UpdateTransactionCategoryParams{
		Transaction:  tc.transaction.ID,
		OtherAccount: sql.NullInt64{Int64: tc.otherAccount.ID, Valid: true},
		Category:     sql.NullInt64{},
		Income:       false,
		Outflow:      0,
		Inflow:       1000,
		ID:           txCategory.ID,
	})
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Error("The number of affected rows should be 1")
	}
	var otherAccount sql.NullInt64
	if err := db.QueryRow(`SELECT other_account FROM "transaction_category" WHERE id = ?`, txCategory.ID).Scan(&otherAccount); err != nil {
		t.Fatal(err)
	}
	if !otherAccount.Valid || otherAccount.Int64 != tc.otherAccount.ID {
		t.Error("transaction category was not updated to a transfer")
	}
}

func TestDeleteTransactionCategory(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	tc := createTransactionContext(t, queries, ctx)
	txCategory := newCategoryLine(t, queries, ctx, tc.transaction.ID, tc.category.ID, 1000, 0)
	err := queries.DeleteTransactionCategory(ctx, txCategory.ID)
	if err != nil {
		t.Fatal(err)
	}
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM "transaction_category" WHERE id = ?`, txCategory.ID).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatal("There should be no transaction category after")
	}
}

func TestDeleteTransactionCategoriesByTransaction(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	tc := createTransactionContext(t, queries, ctx)
	for i := 0; i < 2; i++ {
		newCategoryLine(t, queries, ctx, tc.transaction.ID, tc.category.ID, 1000, 0)
	}
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM "transaction_category" WHERE "transaction" = ?`, tc.transaction.ID).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 2 {
		t.Fatal("There should be two transaction categories before")
	}
	err := queries.DeleteTransactionCategoriesByTransaction(ctx, tc.transaction.ID)
	if err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`SELECT COUNT(*) FROM "transaction_category" WHERE "transaction" = ?`, tc.transaction.ID).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatal("There should be no transaction category after")
	}
}