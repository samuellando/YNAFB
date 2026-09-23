package data_test

import (
	"database/sql"
	"testing"

	"samuellando.com/YNAFB/internal/data"
)

func TestCreateTrx(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	account := newAccount(t, queries, ctx, budget, "testaccount")
	payee := newPayee(t, queries, ctx, budget, "testpayee")
	date := mustTime(t, 2026, 1, 1)
	transaction, err := queries.CreateTrx(ctx, data.CreateTrxParams{
		LoginID:      budget.LoginID,
		BudgetID:     budget.ID,
		AccountID:    account.ID,
		PayeeID:      payee.ID,
		Date:         date,
		TotalOutflow: 1000,
		TotalInflow:  0,
		Note:         "testnote",
	})
	if err != nil {
		t.Fatal(err)
	}
	if transaction.Date.Unix() != date.Unix() {
		t.Error("date does not match")
	}
	if transaction.AccountID != account.ID {
		t.Error("account id does not match")
	}
	if transaction.PayeeID != payee.ID {
		t.Error("payee id does not match")
	}
	if transaction.TotalOutflow != 1000 {
		t.Error("total outflow does not match")
	}
	if transaction.TotalInflow != 0 {
		t.Error("total inflow does not match")
	}
	if transaction.Note != "testnote" {
		t.Error("note does not match")
	}
	if transaction.ID != 1 {
		t.Error("ID of the first transaction should be 1")
	}
}

func TestCreateTrxBothInflowAndOutflowRejected(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	account := newAccount(t, queries, ctx, budget, "testaccount")
	payee := newPayee(t, queries, ctx, budget, "testpayee")
	_, err := queries.CreateTrx(ctx, data.CreateTrxParams{
		LoginID:      budget.LoginID,
		BudgetID:     budget.ID,
		AccountID:    account.ID,
		PayeeID:      payee.ID,
		Date:         mustTime(t, 2026, 1, 1),
		TotalOutflow: 1000,
		TotalInflow:  500,
	})
	if err == nil {
		t.Error("A transaction with both inflow and outflow should be rejected")
	}
}

func TestCreateTrxNegativeOutflowRejected(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	account := newAccount(t, queries, ctx, budget, "testaccount")
	payee := newPayee(t, queries, ctx, budget, "testpayee")
	_, err := queries.CreateTrx(ctx, data.CreateTrxParams{
		LoginID:      budget.LoginID,
		BudgetID:     budget.ID,
		AccountID:    account.ID,
		PayeeID:      payee.ID,
		Date:         mustTime(t, 2026, 1, 1),
		TotalOutflow: -1000,
		TotalInflow:  0,
	})
	if err == nil {
		t.Error("A transaction with a negative outflow should be rejected")
	}
}

func TestCreateTrxNegativeInflowRejected(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	account := newAccount(t, queries, ctx, budget, "testaccount")
	payee := newPayee(t, queries, ctx, budget, "testpayee")
	_, err := queries.CreateTrx(ctx, data.CreateTrxParams{
		LoginID:      budget.LoginID,
		BudgetID:     budget.ID,
		AccountID:    account.ID,
		PayeeID:      payee.ID,
		Date:         mustTime(t, 2026, 1, 1),
		TotalOutflow: 0,
		TotalInflow:  -500,
	})
	if err == nil {
		t.Error("A transaction with a negative inflow should be rejected")
	}
}

func TestCreateTrxZeroAmountRejected(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	account := newAccount(t, queries, ctx, budget, "testaccount")
	payee := newPayee(t, queries, ctx, budget, "testpayee")
	_, err := queries.CreateTrx(ctx, data.CreateTrxParams{
		LoginID:      budget.LoginID,
		BudgetID:     budget.ID,
		AccountID:    account.ID,
		PayeeID:      payee.ID,
		Date:         mustTime(t, 2026, 1, 1),
		TotalOutflow: 0,
		TotalInflow:  0,
	})
	if err == nil {
		t.Error("A transaction with both inflow and outflow zero should be rejected")
	}
}

func TestUpdateTrxBothInflowAndOutflowRejected(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	account := newAccount(t, queries, ctx, budget, "testaccount")
	payee := newPayee(t, queries, ctx, budget, "testpayee")
	transaction := newTrx(t, queries, ctx, budget, account, payee, mustTime(t, 2026, 1, 1), 1000, 0, "")
	_, err := queries.UpdateTrx(ctx, data.UpdateTrxParams{
		LoginID:      budget.LoginID,
		Date:         mustTime(t, 2026, 1, 1),
		PayeeID:      payee.ID,
		TotalOutflow: 1000,
		TotalInflow:  500,
		ID:           transaction.ID,
		BudgetID:     budget.ID,
	})
	if err == nil {
		t.Error("Updating a transaction to have both inflow and outflow should be rejected")
	}
}

func TestCreateTrxNonExistingAccount(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	payee := newPayee(t, queries, ctx, budget, "testpayee")
	_, err := queries.CreateTrx(ctx, data.CreateTrxParams{
		LoginID:      budget.LoginID,
		BudgetID:     budget.ID,
		AccountID:    99,
		PayeeID:      payee.ID,
		Date:         mustTime(t, 2026, 1, 1),
		TotalOutflow: 1000,
		TotalInflow:  0,
	})
	if err == nil {
		t.Error("Non existing account should raise an error")
	}
}

func TestCreateTrxNonExistingPayee(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	account := newAccount(t, queries, ctx, budget, "testaccount")
	_, err := queries.CreateTrx(ctx, data.CreateTrxParams{
		LoginID:      budget.LoginID,
		BudgetID:     budget.ID,
		AccountID:    account.ID,
		PayeeID:      99,
		Date:         mustTime(t, 2026, 1, 1),
		TotalOutflow: 1000,
		TotalInflow:  0,
	})
	if err == nil {
		t.Error("Non existing payee should raise an error")
	}
}

func TestDeleteAccountCascadesTransactions(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	account := newAccount(t, queries, ctx, budget, "testaccount")
	payee := newPayee(t, queries, ctx, budget, "testpayee")
	transaction := newTrx(t, queries, ctx, budget, account, payee, mustTime(t, 2026, 1, 1), 1000, 0, "")
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM trx WHERE id = ?`, transaction.ID).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatal("There should be one transaction before")
	}
	err := queries.DeleteAccount(ctx, data.DeleteAccountParams{ID: account.ID, BudgetID: budget.ID, LoginID: budget.LoginID})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`SELECT COUNT(*) FROM trx WHERE id = ?`, transaction.ID).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatal("Deleting an account should cascade delete its transactions")
	}
}

func TestDeletePayeeCascadesTransactions(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	account := newAccount(t, queries, ctx, budget, "testaccount")
	payee := newPayee(t, queries, ctx, budget, "testpayee")
	transaction := newTrx(t, queries, ctx, budget, account, payee, mustTime(t, 2026, 1, 1), 1000, 0, "")
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM trx WHERE id = ?`, transaction.ID).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatal("There should be one transaction before")
	}
	err := queries.DeletePayee(ctx, data.DeletePayeeParams{ID: payee.ID, BudgetID: budget.ID, LoginID: budget.LoginID})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`SELECT COUNT(*) FROM trx WHERE id = ?`, transaction.ID).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatal("Deleting a payee should cascade delete its transactions")
	}
}

func TestUpdateTrx(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	account := newAccount(t, queries, ctx, budget, "testaccount")
	payee := newPayee(t, queries, ctx, budget, "testpayee")
	transaction := newTrx(t, queries, ctx, budget, account, payee, mustTime(t, 2026, 1, 1), 1000, 0, "testnote")
	date := mustTime(t, 2026, 2, 1)
	updated, err := queries.UpdateTrx(ctx, data.UpdateTrxParams{
		LoginID:      budget.LoginID,
		Date:         date,
		PayeeID:      payee.ID,
		TotalOutflow: 2000,
		TotalInflow:  0,
		Note:         "newnote",
		ID:           transaction.ID,
		BudgetID:     budget.ID,
	})
	if err != nil {
		t.Fatal(err)
	}
	if updated.ID != transaction.ID {
		t.Error("ID changed on update")
	}
	if updated.Date.Unix() != date.Unix() {
		t.Error("transaction date was not updated")
	}
	if updated.TotalOutflow != 2000 {
		t.Error("transaction total outflow was not updated")
	}
	if updated.Note != "newnote" {
		t.Error("transaction note was not updated")
	}
}

func TestDeleteTrx(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	account := newAccount(t, queries, ctx, budget, "testaccount")
	payee := newPayee(t, queries, ctx, budget, "testpayee")
	transaction := newTrx(t, queries, ctx, budget, account, payee, mustTime(t, 2026, 1, 1), 1000, 0, "")
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM trx WHERE id = ?`, transaction.ID).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatal("There should be one transaction before")
	}
	err := queries.DeleteTrx(ctx, data.DeleteTrxParams{ID: transaction.ID, BudgetID: budget.ID, LoginID: budget.LoginID})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`SELECT COUNT(*) FROM trx WHERE id = ?`, transaction.ID).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatal("There should be no transaction after")
	}
}

func TestScopingCreateTrxScopedByLogin(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budgetA := newBudget(t, queries, ctx, "budgetA")
	budgetB := newBudget(t, queries, ctx, "budgetB")
	accountB := newAccount(t, queries, ctx, budgetB, "accountB")
	payeeB := newPayee(t, queries, ctx, budgetB, "payeeB")

	_, err := queries.CreateTrx(ctx, data.CreateTrxParams{
		AccountID:    accountB.ID,
		PayeeID:      payeeB.ID,
		Date:         mustTime(t, 2026, 1, 1),
		TotalOutflow: 1000,
		TotalInflow:  0,
		Note:         "",
		BudgetID:     budgetB.ID,
		LoginID:      budgetA.LoginID,
	})
	if err == nil {
		t.Fatal("expected creating a transaction for another login's budget to fail")
	}
}

func TestScopingDeleteTrxScopedByLogin(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budgetA := newBudget(t, queries, ctx, "budgetA")
	budgetB := newBudget(t, queries, ctx, "budgetB")
	accountB := newAccount(t, queries, ctx, budgetB, "accountB")
	payeeB := newPayee(t, queries, ctx, budgetB, "payeeB")
	trxB := newTrx(t, queries, ctx, budgetB, accountB, payeeB, mustTime(t, 2026, 1, 1), 1000, 0, "")

	err := queries.DeleteTrx(ctx, data.DeleteTrxParams{
		ID:       trxB.ID,
		BudgetID: budgetB.ID,
		LoginID:  budgetA.LoginID,
	})
	if err != nil {
		t.Fatal(err)
	}
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM trx WHERE id = ?`, trxB.ID).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("expected trxB to survive a cross-login delete, got %d transactions", count)
	}
}

func TestScopingListTrxsAndLinesScopedByLogin(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budgetA := newBudget(t, queries, ctx, "budgetA")
	budgetB := newBudget(t, queries, ctx, "budgetB")
	accountB := newAccount(t, queries, ctx, budgetB, "accountB")
	payeeB := newPayee(t, queries, ctx, budgetB, "payeeB")
	newTrx(t, queries, ctx, budgetB, accountB, payeeB, mustTime(t, 2026, 1, 1), 1000, 0, "")

	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM trx WHERE budget_id = ?`, budgetB.ID).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("expected 1 transaction for budgetB, got %d", count)
	}

	rows, err := queries.ListTrxsAndLines(ctx, data.ListTrxsAndLinesParams{
		LoginID:  budgetA.LoginID,
		BudgetID: budgetB.ID,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 0 {
		t.Fatalf("expected 0 transactions for another login's budget, got %d", len(rows))
	}
}

func TestScopingUpdateTrxScopedByLogin(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budgetA := newBudget(t, queries, ctx, "budgetA")
	budgetB := newBudget(t, queries, ctx, "budgetB")
	accountB := newAccount(t, queries, ctx, budgetB, "accountB")
	payeeB := newPayee(t, queries, ctx, budgetB, "payeeB")
	trxB := newTrx(t, queries, ctx, budgetB, accountB, payeeB, mustTime(t, 2026, 1, 1), 1000, 0, "")

	_, err := queries.UpdateTrx(ctx, data.UpdateTrxParams{
		Date:         mustTime(t, 2026, 2, 1),
		PayeeID:      payeeB.ID,
		TotalOutflow: 5000,
		TotalInflow:  0,
		Note:         "",
		ID:           trxB.ID,
		BudgetID:     budgetB.ID,
		LoginID:      budgetA.LoginID,
	})
	if err != sql.ErrNoRows {
		t.Fatalf("expected sql.ErrNoRows when updating across logins, got %v", err)
	}
}

func TestGetTrxAndLines(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	account := newAccount(t, queries, ctx, budget, "testaccount")
	otherAccount := newAccount(t, queries, ctx, budget, "otheraccount")
	payee := newPayee(t, queries, ctx, budget, "testpayee")
	category := newCategory(t, queries, ctx, budget, "testcategory")
	transaction := newTrx(t, queries, ctx, budget, account, payee, mustTime(t, 2026, 1, 1), 1000, 0, "testnote")
	newCategoryLine(t, queries, ctx, budget, transaction, category, 600, 0)
	newTransfer(t, queries, ctx, budget, transaction, otherAccount, 400, 0)

	rows, err := queries.GetTrxAndLines(ctx, data.GetTrxAndLinesParams{
		LoginID:   budget.LoginID,
		BudgetID:  budget.ID,
		ID:        transaction.ID,
		AccountID: account.ID,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 2 {
		t.Fatalf("There should be two rows (one per line), got %d", len(rows))
	}
	if rows[0].Trx.ID != transaction.ID {
		t.Error("trx id does not match")
	}
	if rows[0].AccountName != "testaccount" {
		t.Error("trx account name does not match")
	}
	if rows[0].PayeeName != "testpayee" {
		t.Error("trx payee name does not match")
	}
	if rows[0].Trx.Note != "testnote" {
		t.Error("trx note does not match")
	}
	if rows[0].Reconciled {
		t.Error("trx should not be reconciled")
	}
	byCategory := map[bool]data.GetTrxAndLinesRow{}
	for _, row := range rows {
		byCategory[row.CategoryID.Valid] = row
	}
	categoryRow := byCategory[true]
	if categoryRow.CategoryName.String != "testcategory" {
		t.Error("line category name does not match")
	}
	if categoryRow.LineOutflow.Int64 != 600 {
		t.Error("line outflow does not match")
	}
	transferRow := byCategory[false]
	if !transferRow.DestAccountID.Valid || transferRow.DestAccountID.Int64 != otherAccount.ID {
		t.Error("line dest account id does not match")
	}
	if transferRow.DestAccountName.String != "otheraccount" {
		t.Error("line dest account name does not match")
	}
}

func TestGetTrxAndLinesNoLines(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	account := newAccount(t, queries, ctx, budget, "testaccount")
	payee := newPayee(t, queries, ctx, budget, "testpayee")
	transaction := newTrx(t, queries, ctx, budget, account, payee, mustTime(t, 2026, 1, 1), 1000, 0, "")

	rows, err := queries.GetTrxAndLines(ctx, data.GetTrxAndLinesParams{
		LoginID:   budget.LoginID,
		BudgetID:  budget.ID,
		ID:        transaction.ID,
		AccountID: account.ID,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 {
		t.Fatalf("There should be one row for a transaction without lines, got %d", len(rows))
	}
	if rows[0].LineID.Valid {
		t.Error("line id should be null when the transaction has no lines")
	}
}

func TestGetTrxAndLinesWrongAccount(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	account := newAccount(t, queries, ctx, budget, "testaccount")
	otherAccount := newAccount(t, queries, ctx, budget, "otheraccount")
	payee := newPayee(t, queries, ctx, budget, "testpayee")
	transaction := newTrx(t, queries, ctx, budget, account, payee, mustTime(t, 2026, 1, 1), 1000, 0, "")

	rows, err := queries.GetTrxAndLines(ctx, data.GetTrxAndLinesParams{
		LoginID:   budget.LoginID,
		BudgetID:  budget.ID,
		ID:        transaction.ID,
		AccountID: otherAccount.ID,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 0 {
		t.Fatalf("expected 0 rows when the account does not own the transaction, got %d", len(rows))
	}
}

func TestGetTrxAndLinesReconciled(t *testing.T) {
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

	rows, err := queries.GetTrxAndLines(ctx, data.GetTrxAndLinesParams{
		LoginID:   budget.LoginID,
		BudgetID:  budget.ID,
		ID:        transaction.ID,
		AccountID: account.ID,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 {
		t.Fatalf("There should be one row, got %d", len(rows))
	}
	if !rows[0].Reconciled {
		t.Error("trx should be reconciled after reconciling the account")
	}
}

func TestListTrxsAndLines(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	budget2 := newBudget(t, queries, ctx, "testBudget2")
	account := newAccount(t, queries, ctx, budget, "testaccount")
	payee := newPayee(t, queries, ctx, budget, "testpayee")
	category := newCategory(t, queries, ctx, budget, "testcategory")
	otherAccount := newAccount(t, queries, ctx, budget2, "otheraccount")
	otherPayee := newPayee(t, queries, ctx, budget2, "otherpayee")
	newTrx(t, queries, ctx, budget2, otherAccount, otherPayee, mustTime(t, 2026, 1, 1), 1000, 0, "")
	tx1 := newTrx(t, queries, ctx, budget, account, payee, mustTime(t, 2026, 1, 1), 1000, 0, "first")
	tx2 := newTrx(t, queries, ctx, budget, account, payee, mustTime(t, 2026, 2, 1), 2000, 0, "second")
	newCategoryLine(t, queries, ctx, budget, tx2, category, 2000, 0)

	rows, err := queries.ListTrxsAndLines(ctx, data.ListTrxsAndLinesParams{LoginID: budget.LoginID, BudgetID: budget.ID})
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 2 {
		t.Fatalf("There should be two rows (tx1 without lines, tx2 with one line), got %d", len(rows))
	}
	if rows[0].Trx.ID != tx2.ID {
		t.Error("First row should belong to the most recent transaction (ordered by date desc)")
	}
	if rows[0].AccountName != "testaccount" {
		t.Error("row account name does not match")
	}
	if rows[0].PayeeName != "testpayee" {
		t.Error("row payee name does not match")
	}
	if rows[0].CategoryName.String != "testcategory" {
		t.Error("row category name does not match")
	}
	if rows[1].Trx.ID != tx1.ID {
		t.Error("Second row should belong to the older transaction")
	}
	if rows[1].LineID.Valid {
		t.Error("older transaction has no lines, line id should be null")
	}
}

func TestScopingGetTrxAndLinesScopedByLogin(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budgetA := newBudget(t, queries, ctx, "budgetA")
	budgetB := newBudget(t, queries, ctx, "budgetB")
	accountB := newAccount(t, queries, ctx, budgetB, "accountB")
	payeeB := newPayee(t, queries, ctx, budgetB, "payeeB")
	trxB := newTrx(t, queries, ctx, budgetB, accountB, payeeB, mustTime(t, 2026, 1, 1), 1000, 0, "")

	rows, err := queries.GetTrxAndLines(ctx, data.GetTrxAndLinesParams{
		LoginID:   budgetA.LoginID,
		BudgetID:  budgetB.ID,
		ID:        trxB.ID,
		AccountID: accountB.ID,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 0 {
		t.Fatalf("expected 0 rows for another login's transaction, got %d", len(rows))
	}
}
