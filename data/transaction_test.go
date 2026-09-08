package data_test

import (
	"database/sql"
	"testing"

	"samuellando.com/YNAFB/data"
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
		AccountID:    account.ID,
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
	transactions, err := queries.ListTrxs(ctx, data.ListTrxsParams{LoginID: budget.LoginID, BudgetID: budget.ID})
	if err != nil {
		t.Fatal(err)
	}
	if len(transactions) != 1 {
		t.Fatal("There should be one transaction before")
	}
	err = queries.DeleteAccount(ctx, data.DeleteAccountParams{ID: account.ID, BudgetID: budget.ID, LoginID: budget.LoginID})
	if err != nil {
		t.Fatal(err)
	}
	var count int
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
	transactions, err := queries.ListTrxs(ctx, data.ListTrxsParams{LoginID: budget.LoginID, BudgetID: budget.ID})
	if err != nil {
		t.Fatal(err)
	}
	if len(transactions) != 1 {
		t.Fatal("There should be one transaction before")
	}
	err = queries.DeletePayee(ctx, data.DeletePayeeParams{ID: payee.ID, BudgetID: budget.ID, LoginID: budget.LoginID})
	if err != nil {
		t.Fatal(err)
	}
	var count int
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
		AccountID:    account.ID,
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
	transactions, err := queries.ListTrxs(ctx, data.ListTrxsParams{LoginID: budget.LoginID, BudgetID: budget.ID})
	if err != nil {
		t.Fatal(err)
	}
	if len(transactions) != 1 {
		t.Fatal("There should be one transaction")
	}
	if transactions[0].ID != transaction.ID {
		t.Error("ID changed on update")
	}
	if transactions[0].Date.Unix() != date.Unix() {
		t.Error("transaction date was not updated")
	}
	if transactions[0].TotalOutflow != 2000 {
		t.Error("transaction total outflow was not updated")
	}
	if transactions[0].Note != "newnote" {
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
	transactions, err := queries.ListTrxs(ctx, data.ListTrxsParams{LoginID: budget.LoginID, BudgetID: budget.ID})
	if err != nil {
		t.Fatal(err)
	}
	if len(transactions) != 1 {
		t.Fatal("There should be one transaction before")
	}
	err = queries.DeleteTrx(ctx, data.DeleteTrxParams{ID: transaction.ID, BudgetID: budget.ID, LoginID: budget.LoginID})
	if err != nil {
		t.Fatal(err)
	}
	transactions, err = queries.ListTrxs(ctx, data.ListTrxsParams{LoginID: budget.LoginID, BudgetID: budget.ID})
	if err != nil {
		t.Fatal(err)
	}
	if len(transactions) != 0 {
		t.Fatal("There should be no transaction after")
	}
}

func TestListTrxs(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	account := newAccount(t, queries, ctx, budget, "testaccount")
	payee := newPayee(t, queries, ctx, budget, "testpayee")
	tx1 := newTrx(t, queries, ctx, budget, account, payee, mustTime(t, 2026, 1, 1), 1000, 0, "first")
	newTrx(t, queries, ctx, budget, account, payee, mustTime(t, 2026, 2, 1), 2000, 0, "second")
	transactions, err := queries.ListTrxs(ctx, data.ListTrxsParams{LoginID: budget.LoginID, BudgetID: budget.ID})
	if err != nil {
		t.Fatal(err)
	}
	if len(transactions) != 2 {
		t.Fatal("There should be two transactions")
	}
	if transactions[0].ID == tx1.ID {
		t.Error("First transaction should be the most recent one (ordered by date desc)")
	}
	if transactions[0].AccountName != "testaccount" {
		t.Error("transaction account name does not match")
	}
	if transactions[0].PayeeName != "testpayee" {
		t.Error("transaction payee name does not match")
	}
	if transactions[0].TotalOutflow != 2000 {
		t.Error("first transaction total outflow does not match")
	}
	if transactions[1].TotalOutflow != 1000 {
		t.Error("second transaction total outflow does not match")
	}
}

func TestListTrxsScopedToBudget(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget1 := newBudget(t, queries, ctx, "budget1")
	budget2 := newBudget(t, queries, ctx, "budget2")
	account1 := newAccount(t, queries, ctx, budget1, "account1")
	payee1 := newPayee(t, queries, ctx, budget1, "payee1")
	account2 := newAccount(t, queries, ctx, budget2, "account2")
	payee2 := newPayee(t, queries, ctx, budget2, "payee2")
	newTrx(t, queries, ctx, budget1, account1, payee1, mustTime(t, 2026, 1, 1), 1000, 0, "")
	newTrx(t, queries, ctx, budget2, account2, payee2, mustTime(t, 2026, 1, 2), 2000, 0, "")
	transactions, err := queries.ListTrxs(ctx, data.ListTrxsParams{LoginID: budget1.LoginID, BudgetID: budget1.ID})
	if err != nil {
		t.Fatal(err)
	}
	if len(transactions) != 1 {
		t.Fatalf("There should be one transaction for budget1, got %d", len(transactions))
	}
	if transactions[0].AccountName != "account1" {
		t.Error("budget1 returned the wrong transaction")
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
	trxs, err := queries.ListTrxs(ctx, data.ListTrxsParams{LoginID: budgetB.LoginID, BudgetID: budgetB.ID})
	if err != nil {
		t.Fatal(err)
	}
	if len(trxs) != 1 {
		t.Fatalf("expected trxB to survive a cross-login delete, got %d transactions", len(trxs))
	}
}

func TestScopingListTrxsScopedByLogin(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budgetA := newBudget(t, queries, ctx, "budgetA")
	budgetB := newBudget(t, queries, ctx, "budgetB")
	accountB := newAccount(t, queries, ctx, budgetB, "accountB")
	payeeB := newPayee(t, queries, ctx, budgetB, "payeeB")
	newTrx(t, queries, ctx, budgetB, accountB, payeeB, mustTime(t, 2026, 1, 1), 1000, 0, "")

	rows, err := queries.ListTrxs(ctx, data.ListTrxsParams{
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
		AccountID:    accountB.ID,
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
