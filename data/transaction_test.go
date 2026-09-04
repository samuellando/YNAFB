package data_test

import (
	"testing"

	"samuellando.com/YNAFB/data"
)

func TestCreateTransaction(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	account := newAccount(t, queries, ctx, budget.ID, "testaccount")
	payee := newPayee(t, queries, ctx, budget.ID, "testpayee")
	date := mustTime(t, 2026, 1, 1)
	transaction := newTransaction(t, queries, ctx, account.ID, payee.ID, date, 1000, 0, "testnote")
	if transaction.Date.Unix() != date.Unix() {
		t.Error("date does not match")
	}
	if transaction.Account != account.ID {
		t.Error("account id does not match")
	}
	if transaction.Payee != payee.ID {
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

func TestCreateTransactionBothInflowAndOutflowRejected(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	account := newAccount(t, queries, ctx, budget.ID, "testaccount")
	payee := newPayee(t, queries, ctx, budget.ID, "testpayee")
	_, err := queries.CreateTransaction(ctx, data.CreateTransactionParams{
		Date:         mustTime(t, 2026, 1, 1),
		Account:      account.ID,
		Payee:        payee.ID,
		TotalOutflow: 1000,
		TotalInflow:  500,
	})
	if err == nil {
		t.Error("A transaction with both inflow and outflow should be rejected")
	}
}

func TestCreateTransactionNegativeOutflowRejected(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	account := newAccount(t, queries, ctx, budget.ID, "testaccount")
	payee := newPayee(t, queries, ctx, budget.ID, "testpayee")
	_, err := queries.CreateTransaction(ctx, data.CreateTransactionParams{
		Date:         mustTime(t, 2026, 1, 1),
		Account:      account.ID,
		Payee:        payee.ID,
		TotalOutflow: -1000,
		TotalInflow:  0,
	})
	if err == nil {
		t.Error("A transaction with a negative outflow should be rejected")
	}
}

func TestCreateTransactionNegativeInflowRejected(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	account := newAccount(t, queries, ctx, budget.ID, "testaccount")
	payee := newPayee(t, queries, ctx, budget.ID, "testpayee")
	_, err := queries.CreateTransaction(ctx, data.CreateTransactionParams{
		Date:         mustTime(t, 2026, 1, 1),
		Account:      account.ID,
		Payee:        payee.ID,
		TotalOutflow: 0,
		TotalInflow:  -500,
	})
	if err == nil {
		t.Error("A transaction with a negative inflow should be rejected")
	}
}

func TestUpdateTransactionBothInflowAndOutflowRejected(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	account := newAccount(t, queries, ctx, budget.ID, "testaccount")
	payee := newPayee(t, queries, ctx, budget.ID, "testpayee")
	transaction := newTransaction(t, queries, ctx, account.ID, payee.ID, mustTime(t, 2026, 1, 1), 1000, 0, "")
	_, err := queries.UpdateTransaction(ctx, data.UpdateTransactionParams{
		Date:         mustTime(t, 2026, 1, 1),
		Account:      account.ID,
		Payee:        payee.ID,
		TotalOutflow: 1000,
		TotalInflow:  500,
		ID:           transaction.ID,
	})
	if err == nil {
		t.Error("Updating a transaction to have both inflow and outflow should be rejected")
	}
}

func TestCreateTransactionNonExistingAccount(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	payee := newPayee(t, queries, ctx, budget.ID, "testpayee")
	_, err := queries.CreateTransaction(ctx, data.CreateTransactionParams{
		Date:         mustTime(t, 2026, 1, 1),
		Account:      99,
		Payee:        payee.ID,
		TotalOutflow: 1000,
		TotalInflow:  0,
	})
	if err == nil {
		t.Error("Non existing account should raise an error")
	}
}

func TestCreateTransactionNonExistingPayee(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	account := newAccount(t, queries, ctx, budget.ID, "testaccount")
	_, err := queries.CreateTransaction(ctx, data.CreateTransactionParams{
		Date:         mustTime(t, 2026, 1, 1),
		Account:      account.ID,
		Payee:        99,
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
	account := newAccount(t, queries, ctx, budget.ID, "testaccount")
	payee := newPayee(t, queries, ctx, budget.ID, "testpayee")
	transaction := newTransaction(t, queries, ctx, account.ID, payee.ID, mustTime(t, 2026, 1, 1), 1000, 0, "")
	transactions, err := queries.ListTransactions(ctx, budget.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(transactions) != 1 {
		t.Fatal("There should be one transaction before")
	}
	err = queries.DeleteAccount(ctx, account.ID)
	if err != nil {
		t.Fatal(err)
	}
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM "transaction" WHERE id = ?`, transaction.ID).Scan(&count); err != nil {
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
	account := newAccount(t, queries, ctx, budget.ID, "testaccount")
	payee := newPayee(t, queries, ctx, budget.ID, "testpayee")
	transaction := newTransaction(t, queries, ctx, account.ID, payee.ID, mustTime(t, 2026, 1, 1), 1000, 0, "")
	transactions, err := queries.ListTransactions(ctx, budget.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(transactions) != 1 {
		t.Fatal("There should be one transaction before")
	}
	err = queries.DeletePayee(ctx, payee.ID)
	if err != nil {
		t.Fatal(err)
	}
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM "transaction" WHERE id = ?`, transaction.ID).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatal("Deleting a payee should cascade delete its transactions")
	}
}

func TestUpdateTransaction(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	account := newAccount(t, queries, ctx, budget.ID, "testaccount")
	payee := newPayee(t, queries, ctx, budget.ID, "testpayee")
	transaction := newTransaction(t, queries, ctx, account.ID, payee.ID, mustTime(t, 2026, 1, 1), 1000, 0, "testnote")
	date := mustTime(t, 2026, 2, 1)
	n, err := queries.UpdateTransaction(ctx, data.UpdateTransactionParams{
		Date:         date,
		Account:      account.ID,
		Payee:        payee.ID,
		TotalOutflow: 2000,
		TotalInflow:  0,
		Note:         "newnote",
		ID:           transaction.ID,
	})
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Error("The number of affected rows should be 1")
	}
	transactions, err := queries.ListTransactions(ctx, budget.ID)
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

func TestDeleteTransaction(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	account := newAccount(t, queries, ctx, budget.ID, "testaccount")
	payee := newPayee(t, queries, ctx, budget.ID, "testpayee")
	transaction := newTransaction(t, queries, ctx, account.ID, payee.ID, mustTime(t, 2026, 1, 1), 1000, 0, "")
	transactions, err := queries.ListTransactions(ctx, budget.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(transactions) != 1 {
		t.Fatal("There should be one transaction before")
	}
	err = queries.DeleteTransaction(ctx, transaction.ID)
	if err != nil {
		t.Fatal(err)
	}
	transactions, err = queries.ListTransactions(ctx, budget.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(transactions) != 0 {
		t.Fatal("There should be no transaction after")
	}
}

func TestListTransactions(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	account := newAccount(t, queries, ctx, budget.ID, "testaccount")
	payee := newPayee(t, queries, ctx, budget.ID, "testpayee")
	tx1 := newTransaction(t, queries, ctx, account.ID, payee.ID, mustTime(t, 2026, 1, 1), 1000, 0, "first")
	newTransaction(t, queries, ctx, account.ID, payee.ID, mustTime(t, 2026, 2, 1), 2000, 0, "second")
	transactions, err := queries.ListTransactions(ctx, budget.ID)
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