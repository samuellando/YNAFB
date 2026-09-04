package data_test

import (
	"testing"

	"samuellando.com/YNAFB/data"
)

func TestListAccountTransactionsEmpty(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)

	budget := newBudget(t, queries, ctx, "testBudget")
	account := newAccount(t, queries, ctx, budget.ID, "testaccount")

	rows := listAccountTransactions(t, queries, ctx, account.ID)
	if len(rows) != 0 {
		t.Fatalf("expected no transactions, got %d", len(rows))
	}
}

func TestListAccountTransactionsNonexistentAccount(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)

	rows := listAccountTransactions(t, queries, ctx, 99)
	if len(rows) != 0 {
		t.Fatalf("expected no transactions for nonexistent account, got %d", len(rows))
	}
}

func TestListAccountTransactionsUncategorizedOutflow(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)

	budget := newBudget(t, queries, ctx, "testBudget")
	account := newAccount(t, queries, ctx, budget.ID, "Checking")
	payee := newPayee(t, queries, ctx, budget.ID, "Cafe")
	tx := newTransaction(t, queries, ctx, account.ID, payee.ID, mustTime(t, 2026, 1, 1), 1234, 0, "coffee")

	rows := listAccountTransactions(t, queries, ctx, account.ID)
	if len(rows) != 1 {
		t.Fatalf("expected one row, got %d", len(rows))
	}
	r := rows[0]
	if r.ID != tx.ID {
		t.Errorf("expected id %d, got %d", tx.ID, r.ID)
	}
	if r.Account.ID != account.ID || r.Account.Budget != budget.ID || r.Account.Name != "Checking" {
		t.Errorf("unexpected account embed: %+v", r.Account)
	}
	if !r.Date.Time.Equal(mustTime(t, 2026, 1, 1).Time) {
		t.Errorf("unexpected date: %v", r.Date.Time)
	}
	if r.Payee.ID != payee.ID || r.Payee.Budget != budget.ID || r.Payee.Name != "Cafe" {
		t.Errorf("unexpected payee embed: %+v", r.Payee)
	}
	if r.TotalOutflow != 1234 || r.TotalInflow != 0 {
		t.Errorf("unexpected totals: out %d in %d", r.TotalOutflow, r.TotalInflow)
	}
	if r.Note != "coffee" {
		t.Errorf("unexpected note: %q", r.Note)
	}
	if r.SourceAccountID.Valid || r.TransactionCategoryID.Valid || r.ToAccountID.Valid || r.CategoryID.Valid {
		t.Error("uncategorized row should have no target columns set")
	}
	if r.Income {
		t.Error("uncategorized row should not be income")
	}
	if r.Outflow.Valid || r.Inflow.Valid {
		t.Error("uncategorized row should have null per-line amounts")
	}
	if r.Reconciled {
		t.Error("unreconciled transaction should not be reconciled")
	}
}

func TestListAccountTransactionsUncategorizedInflow(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)

	budget := newBudget(t, queries, ctx, "testBudget")
	account := newAccount(t, queries, ctx, budget.ID, "Checking")
	payee := newPayee(t, queries, ctx, budget.ID, "Employer")
	newTransaction(t, queries, ctx, account.ID, payee.ID, mustTime(t, 2026, 1, 1), 0, 5000, "")

	rows := listAccountTransactions(t, queries, ctx, account.ID)
	if len(rows) != 1 {
		t.Fatalf("expected one row, got %d", len(rows))
	}
	if rows[0].TotalOutflow != 0 || rows[0].TotalInflow != 5000 {
		t.Errorf("unexpected totals: out %d in %d", rows[0].TotalOutflow, rows[0].TotalInflow)
	}
	if rows[0].CategoryID.Valid || rows[0].ToAccountID.Valid || rows[0].Income {
		t.Error("uncategorized inflow should not have a target")
	}
}

func TestListAccountTransactionsZeroAmount(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)

	budget := newBudget(t, queries, ctx, "testBudget")
	account := newAccount(t, queries, ctx, budget.ID, "Checking")
	payee := newPayee(t, queries, ctx, budget.ID, "Cafe")
	category := newCategory(t, queries, ctx, budget.ID, "Groceries")
	tx := newTransaction(t, queries, ctx, account.ID, payee.ID, mustTime(t, 2026, 1, 1), 0, 0, "")
	newCategoryLine(t, queries, ctx, tx.ID, category.ID, 0, 0)

	rows := listAccountTransactions(t, queries, ctx, account.ID)
	if len(rows) != 1 {
		t.Fatalf("expected one row, got %d", len(rows))
	}
	r := rows[0]
	if r.TotalOutflow != 0 || r.TotalInflow != 0 {
		t.Errorf("unexpected totals: out %d in %d", r.TotalOutflow, r.TotalInflow)
	}
	if !r.CategoryID.Valid || r.CategoryID.Int64 != category.ID {
		t.Errorf("expected category %d, got %+v", category.ID, r.CategoryID)
	}
	if !r.Outflow.Valid || r.Outflow.Int64 != 0 {
		t.Errorf("expected zero outflow, got %+v", r.Outflow)
	}
	if !r.Inflow.Valid || r.Inflow.Int64 != 0 {
		t.Errorf("expected zero inflow, got %+v", r.Inflow)
	}
}

func TestListAccountTransactionsEmptyNote(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)

	budget := newBudget(t, queries, ctx, "testBudget")
	account := newAccount(t, queries, ctx, budget.ID, "Checking")
	payee := newPayee(t, queries, ctx, budget.ID, "Cafe")
	newTransaction(t, queries, ctx, account.ID, payee.ID, mustTime(t, 2026, 1, 1), 1000, 0, "")

	rows := listAccountTransactions(t, queries, ctx, account.ID)
	if len(rows) != 1 {
		t.Fatalf("expected one row, got %d", len(rows))
	}
	if rows[0].Note != "" {
		t.Errorf("expected empty note, got %q", rows[0].Note)
	}
}

func TestListAccountTransactionsSingleCategorySpend(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)

	budget := newBudget(t, queries, ctx, "testBudget")
	account := newAccount(t, queries, ctx, budget.ID, "Checking")
	payee := newPayee(t, queries, ctx, budget.ID, "Cafe")
	category := newCategory(t, queries, ctx, budget.ID, "Groceries")
	tx := newTransaction(t, queries, ctx, account.ID, payee.ID, mustTime(t, 2026, 1, 1), 1234, 0, "coffee")
	tc := newCategoryLine(t, queries, ctx, tx.ID, category.ID, 1234, 0)

	rows := listAccountTransactions(t, queries, ctx, account.ID)
	if len(rows) != 1 {
		t.Fatalf("expected one row, got %d", len(rows))
	}
	r := rows[0]
	if !r.TransactionCategoryID.Valid || r.TransactionCategoryID.Int64 != tc.ID {
		t.Errorf("expected transaction category id %d, got %+v", tc.ID, r.TransactionCategoryID)
	}
	if !r.CategoryID.Valid || r.CategoryID.Int64 != category.ID {
		t.Errorf("expected category id %d, got %+v", category.ID, r.CategoryID)
	}
	if !r.CategoryName.Valid || r.CategoryName.String != "Groceries" {
		t.Errorf("expected category name Groceries, got %+v", r.CategoryName)
	}
	if r.ToAccountID.Valid || r.SourceAccountID.Valid {
		t.Error("category spend should not reference accounts")
	}
	if r.Income {
		t.Error("category spend should not be income")
	}
	if !r.Outflow.Valid || r.Outflow.Int64 != 1234 {
		t.Errorf("expected outflow 1234, got %+v", r.Outflow)
	}
	if !r.Inflow.Valid || r.Inflow.Int64 != 0 {
		t.Errorf("expected inflow 0, got %+v", r.Inflow)
	}
	if r.TotalOutflow != 1234 || r.TotalInflow != 0 {
		t.Errorf("unexpected totals: out %d in %d", r.TotalOutflow, r.TotalInflow)
	}
}

func TestListAccountTransactionsIncome(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)

	budget := newBudget(t, queries, ctx, "testBudget")
	account := newAccount(t, queries, ctx, budget.ID, "Checking")
	payee := newPayee(t, queries, ctx, budget.ID, "Employer")
	tx := newTransaction(t, queries, ctx, account.ID, payee.ID, mustTime(t, 2026, 1, 1), 0, 5000, "payday")
	tc := newIncomeLine(t, queries, ctx, tx.ID, 5000)

	rows := listAccountTransactions(t, queries, ctx, account.ID)
	if len(rows) != 1 {
		t.Fatalf("expected one row, got %d", len(rows))
	}
	r := rows[0]
	if !r.TransactionCategoryID.Valid || r.TransactionCategoryID.Int64 != tc.ID {
		t.Errorf("expected transaction category id %d, got %+v", tc.ID, r.TransactionCategoryID)
	}
	if !r.Income {
		t.Error("expected income row")
	}
	if r.CategoryID.Valid || r.ToAccountID.Valid || r.SourceAccountID.Valid {
		t.Error("income row should not reference categories or accounts")
	}
	if !r.Inflow.Valid || r.Inflow.Int64 != 5000 {
		t.Errorf("expected inflow 5000, got %+v", r.Inflow)
	}
	if !r.Outflow.Valid || r.Outflow.Int64 != 0 {
		t.Errorf("expected outflow 0, got %+v", r.Outflow)
	}
	if r.TotalOutflow != 0 || r.TotalInflow != 5000 {
		t.Errorf("unexpected totals: out %d in %d", r.TotalOutflow, r.TotalInflow)
	}
}

func TestListAccountTransactionsSplitCategories(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)

	budget := newBudget(t, queries, ctx, "testBudget")
	account := newAccount(t, queries, ctx, budget.ID, "Checking")
	payee := newPayee(t, queries, ctx, budget.ID, "Market")
	cat1 := newCategory(t, queries, ctx, budget.ID, "Groceries")
	cat2 := newCategory(t, queries, ctx, budget.ID, "Household")
	tx := newTransaction(t, queries, ctx, account.ID, payee.ID, mustTime(t, 2026, 1, 1), 3000, 0, "weekly shop")
	newCategoryLine(t, queries, ctx, tx.ID, cat1.ID, 1000, 0)
	newCategoryLine(t, queries, ctx, tx.ID, cat2.ID, 2000, 0)

	rows := listAccountTransactions(t, queries, ctx, account.ID)
	if len(rows) != 2 {
		t.Fatalf("expected two rows, got %d", len(rows))
	}
	if rows[0].ID != tx.ID || rows[1].ID != tx.ID {
		t.Errorf("split rows should share the transaction id, got %d and %d", rows[0].ID, rows[1].ID)
	}
	for _, r := range rows {
		if r.TotalOutflow != 3000 || r.TotalInflow != 0 {
			t.Errorf("split rows should carry the full totals, got out %d in %d", r.TotalOutflow, r.TotalInflow)
		}
	}
	if !rows[0].CategoryID.Valid || rows[0].CategoryID.Int64 != cat1.ID {
		t.Errorf("expected first line to be %d, got %+v", cat1.ID, rows[0].CategoryID)
	}
	if !rows[1].CategoryID.Valid || rows[1].CategoryID.Int64 != cat2.ID {
		t.Errorf("expected second line to be %d, got %+v", cat2.ID, rows[1].CategoryID)
	}
	if rows[0].Outflow.Int64 != 1000 || rows[1].Outflow.Int64 != 2000 {
		t.Errorf("unexpected line amounts: %d %d", rows[0].Outflow.Int64, rows[1].Outflow.Int64)
	}
	if rows[0].TransactionCategoryID.Int64 >= rows[1].TransactionCategoryID.Int64 {
		t.Error("lines should be ordered by transaction category id ascending")
	}
}

func TestListAccountTransactionsMismatchedCategory(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)

	budget := newBudget(t, queries, ctx, "testBudget")
	account := newAccount(t, queries, ctx, budget.ID, "Checking")
	payee := newPayee(t, queries, ctx, budget.ID, "Online Shop")
	category := newCategory(t, queries, ctx, budget.ID, "Groceries")
	tx := newTransaction(t, queries, ctx, account.ID, payee.ID, mustTime(t, 2026, 1, 1), 4200, 0, "import mismatch")
	newCategoryLine(t, queries, ctx, tx.ID, category.ID, 4000, 0)

	rows := listAccountTransactions(t, queries, ctx, account.ID)
	if len(rows) != 1 {
		t.Fatalf("expected one row, got %d", len(rows))
	}
	if rows[0].TotalOutflow != 4200 {
		t.Errorf("expected total outflow 4200, got %d", rows[0].TotalOutflow)
	}
	if !rows[0].Outflow.Valid || rows[0].Outflow.Int64 != 4000 {
		t.Errorf("expected line outflow 4000, got %+v", rows[0].Outflow)
	}
}

func TestListAccountTransactionsTransferOut(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)

	budget := newBudget(t, queries, ctx, "testBudget")
	source := newAccount(t, queries, ctx, budget.ID, "Checking")
	target := newAccount(t, queries, ctx, budget.ID, "Savings")
	payee := newPayee(t, queries, ctx, budget.ID, "Bank")
	tx := newTransaction(t, queries, ctx, source.ID, payee.ID, mustTime(t, 2026, 1, 1), 3000, 0, "move money")
	tc := newTransfer(t, queries, ctx, tx.ID, target.ID, 3000, 0)

	sourceRows := listAccountTransactions(t, queries, ctx, source.ID)
	if len(sourceRows) != 1 {
		t.Fatalf("expected one source row, got %d", len(sourceRows))
	}
	s := sourceRows[0]
	if !s.TransactionCategoryID.Valid || s.TransactionCategoryID.Int64 != tc.ID {
		t.Errorf("expected source row to carry the category id, got %+v", s.TransactionCategoryID)
	}
	if !s.ToAccountID.Valid || s.ToAccountID.Int64 != target.ID {
		t.Errorf("expected to account %d, got %+v", target.ID, s.ToAccountID)
	}
	if !s.ToAccountName.Valid || s.ToAccountName.String != "Savings" {
		t.Errorf("expected to account name Savings, got %+v", s.ToAccountName)
	}
	if s.SourceAccountID.Valid {
		t.Error("source row should not have a source account")
	}
	if s.CategoryID.Valid || s.Income {
		t.Error("transfer row should not be a category or income")
	}
	if !s.Outflow.Valid || s.Outflow.Int64 != 3000 {
		t.Errorf("expected outflow 3000, got %+v", s.Outflow)
	}
	if s.TotalOutflow != 3000 || s.TotalInflow != 0 {
		t.Errorf("unexpected source totals: out %d in %d", s.TotalOutflow, s.TotalInflow)
	}

	targetRows := listAccountTransactions(t, queries, ctx, target.ID)
	if len(targetRows) != 1 {
		t.Fatalf("expected one target row, got %d", len(targetRows))
	}
	m := targetRows[0]
	if m.ID != tx.ID {
		t.Errorf("expected transaction id %d, got %d", tx.ID, m.ID)
	}
	if !m.SourceAccountID.Valid || m.SourceAccountID.Int64 != source.ID {
		t.Errorf("expected source account %d, got %+v", source.ID, m.SourceAccountID)
	}
	if !m.SourceAccountName.Valid || m.SourceAccountName.String != "Checking" {
		t.Errorf("expected source account name Checking, got %+v", m.SourceAccountName)
	}
	if m.TransactionCategoryID.Valid || m.ToAccountID.Valid || m.CategoryID.Valid {
		t.Error("mirror row should not carry category, to-account, or target columns")
	}
	if m.Income {
		t.Error("mirror row should not be income")
	}
	if m.Outflow.Valid || m.Inflow.Valid {
		t.Error("mirror row should have null per-line amounts")
	}
	if m.TotalOutflow != 0 || m.TotalInflow != 3000 {
		t.Errorf("expected mirror totals out 0 in 3000, got out %d in %d", m.TotalOutflow, m.TotalInflow)
	}
	if m.Account.ID != target.ID || m.Account.Name != "Savings" {
		t.Errorf("unexpected mirror account embed: %+v", m.Account)
	}
	if !m.Date.Time.Equal(mustTime(t, 2026, 1, 1).Time) {
		t.Errorf("unexpected mirror date: %v", m.Date.Time)
	}
	if m.Payee.Name != "Bank" {
		t.Errorf("unexpected mirror payee: %+v", m.Payee)
	}
}

func TestListAccountTransactionsTransferIn(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)

	budget := newBudget(t, queries, ctx, "testBudget")
	source := newAccount(t, queries, ctx, budget.ID, "Checking")
	target := newAccount(t, queries, ctx, budget.ID, "Savings")
	payee := newPayee(t, queries, ctx, budget.ID, "Bank")
	tx := newTransaction(t, queries, ctx, target.ID, payee.ID, mustTime(t, 2026, 1, 1), 0, 500, "moved in")
	newTransfer(t, queries, ctx, tx.ID, source.ID, 0, 500)

	targetRows := listAccountTransactions(t, queries, ctx, target.ID)
	if len(targetRows) != 1 {
		t.Fatalf("expected one target row, got %d", len(targetRows))
	}
	if !targetRows[0].ToAccountID.Valid || targetRows[0].ToAccountID.Int64 != source.ID {
		t.Errorf("expected to account %d, got %+v", source.ID, targetRows[0].ToAccountID)
	}
	if !targetRows[0].Inflow.Valid || targetRows[0].Inflow.Int64 != 500 {
		t.Errorf("expected inflow 500, got %+v", targetRows[0].Inflow)
	}
	if targetRows[0].TotalOutflow != 0 || targetRows[0].TotalInflow != 500 {
		t.Errorf("unexpected target totals: out %d in %d", targetRows[0].TotalOutflow, targetRows[0].TotalInflow)
	}

	sourceRows := listAccountTransactions(t, queries, ctx, source.ID)
	if len(sourceRows) != 1 {
		t.Fatalf("expected one source row, got %d", len(sourceRows))
	}
	if !sourceRows[0].SourceAccountID.Valid || sourceRows[0].SourceAccountID.Int64 != target.ID {
		t.Errorf("expected source account %d, got %+v", target.ID, sourceRows[0].SourceAccountID)
	}
	if sourceRows[0].TotalOutflow != 500 || sourceRows[0].TotalInflow != 0 {
		t.Errorf("expected mirror totals out 500 in 0, got out %d in %d", sourceRows[0].TotalOutflow, sourceRows[0].TotalInflow)
	}
}

func TestListAccountTransactionsSplitTransfer(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)

	budget := newBudget(t, queries, ctx, "testBudget")
	source := newAccount(t, queries, ctx, budget.ID, "Checking")
	target := newAccount(t, queries, ctx, budget.ID, "Savings")
	payee := newPayee(t, queries, ctx, budget.ID, "superC")
	category := newCategory(t, queries, ctx, budget.ID, "Groceries")
	tx := newTransaction(t, queries, ctx, source.ID, payee.ID, mustTime(t, 2026, 1, 1), 3000, 0, "")
	newCategoryLine(t, queries, ctx, tx.ID, category.ID, 1000, 0)
	newTransfer(t, queries, ctx, tx.ID, target.ID, 2000, 0)

	sourceRows := listAccountTransactions(t, queries, ctx, source.ID)
	if len(sourceRows) != 2 {
		t.Fatalf("expected two source rows, got %d", len(sourceRows))
	}
	if sourceRows[0].CategoryID.Int64 != category.ID || sourceRows[0].Outflow.Int64 != 1000 {
		t.Errorf("expected first line to be the category, got %+v %+v", sourceRows[0].CategoryID, sourceRows[0].Outflow)
	}
	if sourceRows[1].ToAccountID.Int64 != target.ID || sourceRows[1].Outflow.Int64 != 2000 {
		t.Errorf("expected second line to be the transfer, got %+v %+v", sourceRows[1].ToAccountID, sourceRows[1].Outflow)
	}

	targetRows := listAccountTransactions(t, queries, ctx, target.ID)
	if len(targetRows) != 1 {
		t.Fatalf("expected one target row, got %d", len(targetRows))
	}
	if targetRows[0].TotalInflow != 2000 {
		t.Errorf("expected mirror inflow 2000, got %d", targetRows[0].TotalInflow)
	}
}

func TestListAccountTransactionsReconciled(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)

	budget := newBudget(t, queries, ctx, "testBudget")
	account := newAccount(t, queries, ctx, budget.ID, "Checking")
	payee := newPayee(t, queries, ctx, budget.ID, "Cafe")
	newTransaction(t, queries, ctx, account.ID, payee.ID, mustTime(t, 2026, 1, 1), 1000, 0, "")
	newTransaction(t, queries, ctx, account.ID, payee.ID, mustTime(t, 2026, 2, 1), 2000, 0, "")

	if _, err := queries.ReconcileAccountTransactions(ctx, data.ReconcileAccountTransactionsParams{
		AccountID: account.ID,
		Date:      mustTime(t, 2026, 1, 31),
	}); err != nil {
		t.Fatal(err)
	}

	rows := listAccountTransactions(t, queries, ctx, account.ID)
	if len(rows) != 2 {
		t.Fatalf("expected two rows, got %d", len(rows))
	}
	byID := map[int64]data.ListAccountTransactionsRow{}
	for _, r := range rows {
		byID[r.ID] = r
	}
	if !byID[1].Reconciled {
		t.Error("expected transaction 1 to be reconciled")
	}
	if byID[2].Reconciled {
		t.Error("expected transaction 2 to not be reconciled")
	}
}

func TestListAccountTransactionsReconciledPerAccount(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)

	budget := newBudget(t, queries, ctx, "testBudget")
	source := newAccount(t, queries, ctx, budget.ID, "Checking")
	target := newAccount(t, queries, ctx, budget.ID, "Savings")
	payee := newPayee(t, queries, ctx, budget.ID, "Bank")
	tx := newTransaction(t, queries, ctx, source.ID, payee.ID, mustTime(t, 2026, 1, 1), 3000, 0, "")
	newTransfer(t, queries, ctx, tx.ID, target.ID, 3000, 0)

	if _, err := queries.ReconcileAccountTransactions(ctx, data.ReconcileAccountTransactionsParams{
		AccountID: target.ID,
		Date:      mustTime(t, 2026, 2, 1),
	}); err != nil {
		t.Fatal(err)
	}

	sourceRows := listAccountTransactions(t, queries, ctx, source.ID)
	if sourceRows[0].Reconciled {
		t.Error("reconciling the target should not reconcile the source's view")
	}
	targetRows := listAccountTransactions(t, queries, ctx, target.ID)
	if !targetRows[0].Reconciled {
		t.Error("expected the mirror row to be reconciled in the target")
	}

	if _, err := queries.ReconcileAccountTransactions(ctx, data.ReconcileAccountTransactionsParams{
		AccountID: source.ID,
		Date:      mustTime(t, 2026, 2, 1),
	}); err != nil {
		t.Fatal(err)
	}

	sourceRows = listAccountTransactions(t, queries, ctx, source.ID)
	if !sourceRows[0].Reconciled {
		t.Error("expected the source's native row to be reconciled after reconciling the source")
	}
}

func TestListAccountTransactionsScopedToAccount(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)

	budget := newBudget(t, queries, ctx, "testBudget")
	account1 := newAccount(t, queries, ctx, budget.ID, "Checking")
	account2 := newAccount(t, queries, ctx, budget.ID, "Savings")
	payee := newPayee(t, queries, ctx, budget.ID, "Cafe")
	tx1 := newTransaction(t, queries, ctx, account1.ID, payee.ID, mustTime(t, 2026, 1, 1), 1000, 0, "")
	tx2 := newTransaction(t, queries, ctx, account2.ID, payee.ID, mustTime(t, 2026, 1, 2), 2000, 0, "")

	rows1 := listAccountTransactions(t, queries, ctx, account1.ID)
	if len(rows1) != 1 || rows1[0].ID != tx1.ID {
		t.Errorf("account1 should only contain its own transaction, got %+v", rows1)
	}
	rows2 := listAccountTransactions(t, queries, ctx, account2.ID)
	if len(rows2) != 1 || rows2[0].ID != tx2.ID {
		t.Errorf("account2 should only contain its own transaction, got %+v", rows2)
	}
}

func TestListAccountTransactionsCrossBudgetTransfer(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)

	budget1 := newBudget(t, queries, ctx, "budget1")
	budget2 := newBudget(t, queries, ctx, "budget2")
	account1 := newAccount(t, queries, ctx, budget1.ID, "Checking")
	account2 := newAccount(t, queries, ctx, budget2.ID, "Shared")
	payee1 := newPayee(t, queries, ctx, budget1.ID, "Bank")
	tx := newTransaction(t, queries, ctx, account1.ID, payee1.ID, mustTime(t, 2026, 1, 1), 2000, 0, "")
	newTransfer(t, queries, ctx, tx.ID, account2.ID, 2000, 0)

	rows := listAccountTransactions(t, queries, ctx, account2.ID)
	if len(rows) != 1 {
		t.Fatalf("expected one mirror row, got %d", len(rows))
	}
	r := rows[0]
	if r.Account.ID != account2.ID || r.Account.Budget != budget2.ID {
		t.Errorf("mirror row should embed the target account, got %+v", r.Account)
	}
	if !r.SourceAccountID.Valid || r.SourceAccountID.Int64 != account1.ID {
		t.Errorf("expected source account %d, got %+v", account1.ID, r.SourceAccountID)
	}
	if r.TotalInflow != 2000 {
		t.Errorf("expected inflow 2000, got %d", r.TotalInflow)
	}
	if r.Payee.Budget != budget1.ID {
		t.Errorf("mirror row carries the source budget's payee, got %+v", r.Payee)
	}
}

func TestListAccountTransactionsOrdering(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)

	budget := newBudget(t, queries, ctx, "testBudget")
	account := newAccount(t, queries, ctx, budget.ID, "Checking")
	payee := newPayee(t, queries, ctx, budget.ID, "Cafe")
	t1 := newTransaction(t, queries, ctx, account.ID, payee.ID, mustTime(t, 2026, 1, 3), 1000, 0, "")
	t2 := newTransaction(t, queries, ctx, account.ID, payee.ID, mustTime(t, 2026, 1, 1), 2000, 0, "")
	t3 := newTransaction(t, queries, ctx, account.ID, payee.ID, mustTime(t, 2026, 1, 2), 3000, 0, "")

	rows := listAccountTransactions(t, queries, ctx, account.ID)
	if len(rows) != 3 {
		t.Fatalf("expected three rows, got %d", len(rows))
	}
	want := []int64{t1.ID, t3.ID, t2.ID}
	for i, id := range want {
		if rows[i].ID != id {
			t.Errorf("expected row %d to be transaction %d, got %d", i, id, rows[i].ID)
		}
	}
}

func TestListAccountTransactionsSameDateOrdersByPayeeName(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)

	budget := newBudget(t, queries, ctx, "testBudget")
	account := newAccount(t, queries, ctx, budget.ID, "Checking")
	zeta := newPayee(t, queries, ctx, budget.ID, "Zeta")
	alpha := newPayee(t, queries, ctx, budget.ID, "Alpha")
	t1 := newTransaction(t, queries, ctx, account.ID, zeta.ID, mustTime(t, 2026, 1, 1), 1000, 0, "")
	t2 := newTransaction(t, queries, ctx, account.ID, alpha.ID, mustTime(t, 2026, 1, 1), 2000, 0, "")

	rows := listAccountTransactions(t, queries, ctx, account.ID)
	if len(rows) != 2 {
		t.Fatalf("expected two rows, got %d", len(rows))
	}
	if rows[0].ID != t2.ID || rows[0].Payee.Name != "Alpha" {
		t.Errorf("expected Alpha (payee name, id %d) first, got %+v", alpha.ID, rows[0])
	}
	if rows[1].ID != t1.ID || rows[1].Payee.Name != "Zeta" {
		t.Errorf("expected Zeta second, got %+v", rows[1])
	}
}

func TestListAccountTransactionsSameDateSamePayeeOrderByTransactionID(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)

	budget := newBudget(t, queries, ctx, "testBudget")
	account := newAccount(t, queries, ctx, budget.ID, "Checking")
	payee := newPayee(t, queries, ctx, budget.ID, "Cafe")
	t1 := newTransaction(t, queries, ctx, account.ID, payee.ID, mustTime(t, 2026, 1, 1), 1000, 0, "")
	t2 := newTransaction(t, queries, ctx, account.ID, payee.ID, mustTime(t, 2026, 1, 1), 2000, 0, "")

	rows := listAccountTransactions(t, queries, ctx, account.ID)
	if len(rows) != 2 {
		t.Fatalf("expected two rows, got %d", len(rows))
	}
	if rows[0].ID != t1.ID || rows[1].ID != t2.ID {
		t.Errorf("expected insertion order t1 then t2, got %d then %d", rows[0].ID, rows[1].ID)
	}
}