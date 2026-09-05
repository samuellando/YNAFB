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

	rows := listAccountTransactions(t, queries, ctx, budget.ID, account.ID)
	if len(rows) != 0 {
		t.Fatalf("expected no transactions, got %d", len(rows))
	}
}

func TestListAccountTransactionsNonexistentAccount(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)

	budget := newBudget(t, queries, ctx, "testBudget")
	rows := listAccountTransactions(t, queries, ctx, budget.ID, 99)
	if len(rows) != 0 {
		t.Fatalf("expected no transactions for nonexistent account, got %d", len(rows))
	}
}

func TestListAccountTransactionsScopedToBudget(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)

	budget1 := newBudget(t, queries, ctx, "budget1")
	budget2 := newBudget(t, queries, ctx, "budget2")
	account := newAccount(t, queries, ctx, budget1.ID, "Checking")
	payee := newPayee(t, queries, ctx, budget1.ID, "Cafe")
	newTrx(t, queries, ctx, account, payee, mustTime(t, 2026, 1, 1), 1234, 0, "")

	rows := listAccountTransactions(t, queries, ctx, budget2.ID, account.ID)
	if len(rows) != 0 {
		t.Fatalf("expected no rows for an account queried under the wrong budget, got %d", len(rows))
	}
}

func TestListAccountTransactionsUncategorizedOutflow(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)

	budget := newBudget(t, queries, ctx, "testBudget")
	account := newAccount(t, queries, ctx, budget.ID, "Checking")
	payee := newPayee(t, queries, ctx, budget.ID, "Cafe")
	tx := newTrx(t, queries, ctx, account, payee, mustTime(t, 2026, 1, 1), 1234, 0, "coffee")

	rows := listAccountTransactions(t, queries, ctx, budget.ID, account.ID)
	if len(rows) != 1 {
		t.Fatalf("expected one row, got %d", len(rows))
	}
	r := rows[0]
	if r.TrxID != tx.ID {
		t.Errorf("expected id %d, got %d", tx.ID, r.TrxID)
	}
	if r.AccountID != account.ID || r.BudgetID != budget.ID || !r.AccountName.Valid || r.AccountName.String != "Checking" {
		t.Errorf("unexpected account fields: id %d budget %d name %+v", r.AccountID, r.BudgetID, r.AccountName)
	}
	if !r.Date.Time.Equal(mustTime(t, 2026, 1, 1).Time) {
		t.Errorf("unexpected date: %v", r.Date.Time)
	}
	if r.PayeeID != payee.ID || !r.PayeeName.Valid || r.PayeeName.String != "Cafe" {
		t.Errorf("unexpected payee fields: id %d name %+v", r.PayeeID, r.PayeeName)
	}
	if r.Outflow != 1234 || r.Inflow != 0 {
		t.Errorf("unexpected totals: out %d in %d", r.Outflow, r.Inflow)
	}
	if r.Note != "coffee" {
		t.Errorf("unexpected note: %q", r.Note)
	}
	if r.SourceAccountID.Valid || r.TrxLineID.Valid || r.DestAccountID.Valid || r.CategoryID.Valid {
		t.Error("uncategorized row should have no target columns set")
	}
	if r.Income {
		t.Error("uncategorized row should not be income")
	}
	if r.LineOutflow.Valid || r.LineInflow.Valid {
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
	newTrx(t, queries, ctx, account, payee, mustTime(t, 2026, 1, 1), 0, 5000, "")

	rows := listAccountTransactions(t, queries, ctx, budget.ID, account.ID)
	if len(rows) != 1 {
		t.Fatalf("expected one row, got %d", len(rows))
	}
	if rows[0].Outflow != 0 || rows[0].Inflow != 5000 {
		t.Errorf("unexpected totals: out %d in %d", rows[0].Outflow, rows[0].Inflow)
	}
	if rows[0].CategoryID.Valid || rows[0].DestAccountID.Valid || rows[0].Income {
		t.Error("uncategorized inflow should not have a target")
	}
}

func TestListAccountTransactionsEmptyNote(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)

	budget := newBudget(t, queries, ctx, "testBudget")
	account := newAccount(t, queries, ctx, budget.ID, "Checking")
	payee := newPayee(t, queries, ctx, budget.ID, "Cafe")
	newTrx(t, queries, ctx, account, payee, mustTime(t, 2026, 1, 1), 1000, 0, "")

	rows := listAccountTransactions(t, queries, ctx, budget.ID, account.ID)
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
	tx := newTrx(t, queries, ctx, account, payee, mustTime(t, 2026, 1, 1), 1234, 0, "coffee")
	tc := newCategoryLine(t, queries, ctx, tx, category, 1234, 0)

	rows := listAccountTransactions(t, queries, ctx, budget.ID, account.ID)
	if len(rows) != 1 {
		t.Fatalf("expected one row, got %d", len(rows))
	}
	r := rows[0]
	if !r.TrxLineID.Valid || r.TrxLineID.Int64 != tc.ID {
		t.Errorf("expected trx line id %d, got %+v", tc.ID, r.TrxLineID)
	}
	if !r.CategoryID.Valid || r.CategoryID.Int64 != category.ID {
		t.Errorf("expected category id %d, got %+v", category.ID, r.CategoryID)
	}
	if !r.CategoryName.Valid || r.CategoryName.String != "Groceries" {
		t.Errorf("expected category name Groceries, got %+v", r.CategoryName)
	}
	if r.DestAccountID.Valid || r.SourceAccountID.Valid {
		t.Error("category spend should not reference accounts")
	}
	if r.Income {
		t.Error("category spend should not be income")
	}
	if !r.LineOutflow.Valid || r.LineOutflow.Int64 != 1234 {
		t.Errorf("expected line outflow 1234, got %+v", r.LineOutflow)
	}
	if !r.LineInflow.Valid || r.LineInflow.Int64 != 0 {
		t.Errorf("expected line inflow 0, got %+v", r.LineInflow)
	}
	if r.Outflow != 1234 || r.Inflow != 0 {
		t.Errorf("unexpected totals: out %d in %d", r.Outflow, r.Inflow)
	}
}

func TestListAccountTransactionsIncome(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)

	budget := newBudget(t, queries, ctx, "testBudget")
	account := newAccount(t, queries, ctx, budget.ID, "Checking")
	payee := newPayee(t, queries, ctx, budget.ID, "Employer")
	tx := newTrx(t, queries, ctx, account, payee, mustTime(t, 2026, 1, 1), 0, 5000, "payday")
	tc := newIncomeLine(t, queries, ctx, tx, 5000)

	rows := listAccountTransactions(t, queries, ctx, budget.ID, account.ID)
	if len(rows) != 1 {
		t.Fatalf("expected one row, got %d", len(rows))
	}
	r := rows[0]
	if !r.TrxLineID.Valid || r.TrxLineID.Int64 != tc.ID {
		t.Errorf("expected trx line id %d, got %+v", tc.ID, r.TrxLineID)
	}
	if !r.Income {
		t.Error("expected income row")
	}
	if r.CategoryID.Valid || r.DestAccountID.Valid || r.SourceAccountID.Valid {
		t.Error("income row should not reference categories or accounts")
	}
	if !r.LineInflow.Valid || r.LineInflow.Int64 != 5000 {
		t.Errorf("expected line inflow 5000, got %+v", r.LineInflow)
	}
	if !r.LineOutflow.Valid || r.LineOutflow.Int64 != 0 {
		t.Errorf("expected line outflow 0, got %+v", r.LineOutflow)
	}
	if r.Outflow != 0 || r.Inflow != 5000 {
		t.Errorf("unexpected totals: out %d in %d", r.Outflow, r.Inflow)
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
	tx := newTrx(t, queries, ctx, account, payee, mustTime(t, 2026, 1, 1), 3000, 0, "weekly shop")
	newCategoryLine(t, queries, ctx, tx, cat1, 1000, 0)
	newCategoryLine(t, queries, ctx, tx, cat2, 2000, 0)

	rows := listAccountTransactions(t, queries, ctx, budget.ID, account.ID)
	if len(rows) != 2 {
		t.Fatalf("expected two rows, got %d", len(rows))
	}
	if rows[0].TrxID != tx.ID || rows[1].TrxID != tx.ID {
		t.Errorf("split rows should share the transaction id, got %d and %d", rows[0].TrxID, rows[1].TrxID)
	}
	for _, r := range rows {
		if r.Outflow != 3000 || r.Inflow != 0 {
			t.Errorf("split rows should carry the full totals, got out %d in %d", r.Outflow, r.Inflow)
		}
	}
	if !rows[0].CategoryID.Valid || rows[0].CategoryID.Int64 != cat1.ID {
		t.Errorf("expected first line to be %d, got %+v", cat1.ID, rows[0].CategoryID)
	}
	if !rows[1].CategoryID.Valid || rows[1].CategoryID.Int64 != cat2.ID {
		t.Errorf("expected second line to be %d, got %+v", cat2.ID, rows[1].CategoryID)
	}
	if rows[0].LineOutflow.Int64 != 1000 || rows[1].LineOutflow.Int64 != 2000 {
		t.Errorf("unexpected line amounts: %d %d", rows[0].LineOutflow.Int64, rows[1].LineOutflow.Int64)
	}
	if rows[0].TrxLineID.Int64 >= rows[1].TrxLineID.Int64 {
		t.Error("lines should be ordered by trx line id ascending")
	}
}

func TestListAccountTransactionsMismatchedCategory(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)

	budget := newBudget(t, queries, ctx, "testBudget")
	account := newAccount(t, queries, ctx, budget.ID, "Checking")
	payee := newPayee(t, queries, ctx, budget.ID, "Online Shop")
	category := newCategory(t, queries, ctx, budget.ID, "Groceries")
	tx := newTrx(t, queries, ctx, account, payee, mustTime(t, 2026, 1, 1), 4200, 0, "import mismatch")
	newCategoryLine(t, queries, ctx, tx, category, 4000, 0)

	rows := listAccountTransactions(t, queries, ctx, budget.ID, account.ID)
	if len(rows) != 1 {
		t.Fatalf("expected one row, got %d", len(rows))
	}
	if rows[0].Outflow != 4200 {
		t.Errorf("expected total outflow 4200, got %d", rows[0].Outflow)
	}
	if !rows[0].LineOutflow.Valid || rows[0].LineOutflow.Int64 != 4000 {
		t.Errorf("expected line outflow 4000, got %+v", rows[0].LineOutflow)
	}
}

func TestListAccountTransactionsTransferOut(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)

	budget := newBudget(t, queries, ctx, "testBudget")
	source := newAccount(t, queries, ctx, budget.ID, "Checking")
	target := newAccount(t, queries, ctx, budget.ID, "Savings")
	payee := newPayee(t, queries, ctx, budget.ID, "Bank")
	tx := newTrx(t, queries, ctx, source, payee, mustTime(t, 2026, 1, 1), 3000, 0, "move money")
	tc := newTransfer(t, queries, ctx, tx, target, 3000, 0)

	sourceRows := listAccountTransactions(t, queries, ctx, budget.ID, source.ID)
	if len(sourceRows) != 1 {
		t.Fatalf("expected one source row, got %d", len(sourceRows))
	}
	s := sourceRows[0]
	if !s.TrxLineID.Valid || s.TrxLineID.Int64 != tc.ID {
		t.Errorf("expected source row to carry the line id, got %+v", s.TrxLineID)
	}
	if !s.DestAccountID.Valid || s.DestAccountID.Int64 != target.ID {
		t.Errorf("expected dest account %d, got %+v", target.ID, s.DestAccountID)
	}
	if !s.DestAccountName.Valid || s.DestAccountName.String != "Savings" {
		t.Errorf("expected dest account name Savings, got %+v", s.DestAccountName)
	}
	if s.SourceAccountID.Valid {
		t.Error("source row should not have a source account")
	}
	if s.CategoryID.Valid || s.Income {
		t.Error("transfer row should not be a category or income")
	}
	if !s.LineOutflow.Valid || s.LineOutflow.Int64 != 3000 {
		t.Errorf("expected line outflow 3000, got %+v", s.LineOutflow)
	}
	if s.Outflow != 3000 || s.Inflow != 0 {
		t.Errorf("unexpected source totals: out %d in %d", s.Outflow, s.Inflow)
	}

	targetRows := listAccountTransactions(t, queries, ctx, budget.ID, target.ID)
	if len(targetRows) != 1 {
		t.Fatalf("expected one target row, got %d", len(targetRows))
	}
	m := targetRows[0]
	if m.TrxID != tx.ID {
		t.Errorf("expected transaction id %d, got %d", tx.ID, m.TrxID)
	}
	if !m.SourceAccountID.Valid || m.SourceAccountID.Int64 != source.ID {
		t.Errorf("expected source account %d, got %+v", source.ID, m.SourceAccountID)
	}
	if !m.SourceAccountName.Valid || m.SourceAccountName.String != "Checking" {
		t.Errorf("expected source account name Checking, got %+v", m.SourceAccountName)
	}
	if m.TrxLineID.Valid || m.DestAccountID.Valid || m.CategoryID.Valid {
		t.Error("mirror row should not carry line, dest-account, or category columns")
	}
	if m.Income {
		t.Error("mirror row should not be income")
	}
	if m.LineOutflow.Valid || m.LineInflow.Valid {
		t.Error("mirror row should have null per-line amounts")
	}
	if m.Outflow != 0 || m.Inflow != 3000 {
		t.Errorf("expected mirror totals out 0 in 3000, got out %d in %d", m.Outflow, m.Inflow)
	}
	if m.AccountID != target.ID || !m.AccountName.Valid || m.AccountName.String != "Savings" {
		t.Errorf("unexpected mirror account fields: id %d name %+v", m.AccountID, m.AccountName)
	}
	if !m.Date.Time.Equal(mustTime(t, 2026, 1, 1).Time) {
		t.Errorf("unexpected mirror date: %v", m.Date.Time)
	}
	if m.PayeeName.String != "Bank" {
		t.Errorf("unexpected mirror payee: %q", m.PayeeName.String)
	}
}

func TestListAccountTransactionsTransferIn(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)

	budget := newBudget(t, queries, ctx, "testBudget")
	source := newAccount(t, queries, ctx, budget.ID, "Checking")
	target := newAccount(t, queries, ctx, budget.ID, "Savings")
	payee := newPayee(t, queries, ctx, budget.ID, "Bank")
	tx := newTrx(t, queries, ctx, target, payee, mustTime(t, 2026, 1, 1), 0, 500, "moved in")
	newTransfer(t, queries, ctx, tx, source, 0, 500)

	targetRows := listAccountTransactions(t, queries, ctx, budget.ID, target.ID)
	if len(targetRows) != 1 {
		t.Fatalf("expected one target row, got %d", len(targetRows))
	}
	if !targetRows[0].DestAccountID.Valid || targetRows[0].DestAccountID.Int64 != source.ID {
		t.Errorf("expected dest account %d, got %+v", source.ID, targetRows[0].DestAccountID)
	}
	if !targetRows[0].LineInflow.Valid || targetRows[0].LineInflow.Int64 != 500 {
		t.Errorf("expected line inflow 500, got %+v", targetRows[0].LineInflow)
	}
	if targetRows[0].Outflow != 0 || targetRows[0].Inflow != 500 {
		t.Errorf("unexpected target totals: out %d in %d", targetRows[0].Outflow, targetRows[0].Inflow)
	}

	sourceRows := listAccountTransactions(t, queries, ctx, budget.ID, source.ID)
	if len(sourceRows) != 1 {
		t.Fatalf("expected one source row, got %d", len(sourceRows))
	}
	if !sourceRows[0].SourceAccountID.Valid || sourceRows[0].SourceAccountID.Int64 != target.ID {
		t.Errorf("expected source account %d, got %+v", target.ID, sourceRows[0].SourceAccountID)
	}
	if sourceRows[0].Outflow != 500 || sourceRows[0].Inflow != 0 {
		t.Errorf("expected mirror totals out 500 in 0, got out %d in %d", sourceRows[0].Outflow, sourceRows[0].Inflow)
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
	tx := newTrx(t, queries, ctx, source, payee, mustTime(t, 2026, 1, 1), 3000, 0, "")
	newCategoryLine(t, queries, ctx, tx, category, 1000, 0)
	newTransfer(t, queries, ctx, tx, target, 2000, 0)

	sourceRows := listAccountTransactions(t, queries, ctx, budget.ID, source.ID)
	if len(sourceRows) != 2 {
		t.Fatalf("expected two source rows, got %d", len(sourceRows))
	}
	if sourceRows[0].CategoryID.Int64 != category.ID || sourceRows[0].LineOutflow.Int64 != 1000 {
		t.Errorf("expected first line to be the category, got %+v %+v", sourceRows[0].CategoryID, sourceRows[0].LineOutflow)
	}
	if sourceRows[1].DestAccountID.Int64 != target.ID || sourceRows[1].LineOutflow.Int64 != 2000 {
		t.Errorf("expected second line to be the transfer, got %+v %+v", sourceRows[1].DestAccountID, sourceRows[1].LineOutflow)
	}

	targetRows := listAccountTransactions(t, queries, ctx, budget.ID, target.ID)
	if len(targetRows) != 1 {
		t.Fatalf("expected one target row, got %d", len(targetRows))
	}
	if targetRows[0].Inflow != 2000 {
		t.Errorf("expected mirror inflow 2000, got %d", targetRows[0].Inflow)
	}
}

func TestListAccountTransactionsReconciled(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)

	budget := newBudget(t, queries, ctx, "testBudget")
	account := newAccount(t, queries, ctx, budget.ID, "Checking")
	payee := newPayee(t, queries, ctx, budget.ID, "Cafe")
	newTrx(t, queries, ctx, account, payee, mustTime(t, 2026, 1, 1), 1000, 0, "")
	newTrx(t, queries, ctx, account, payee, mustTime(t, 2026, 2, 1), 2000, 0, "")

	if _, err := queries.ReconcileAccountTransactions(ctx, data.ReconcileAccountTransactionsParams{
		BudgetID:  budget.ID,
		AccountID: account.ID,
		Date:      mustTime(t, 2026, 1, 31),
	}); err != nil {
		t.Fatal(err)
	}

	rows := listAccountTransactions(t, queries, ctx, budget.ID, account.ID)
	if len(rows) != 2 {
		t.Fatalf("expected two rows, got %d", len(rows))
	}
	byID := map[int64]data.ListAccountTransactionsRow{}
	for _, r := range rows {
		byID[r.TrxID] = r
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
	tx := newTrx(t, queries, ctx, source, payee, mustTime(t, 2026, 1, 1), 3000, 0, "")
	newTransfer(t, queries, ctx, tx, target, 3000, 0)

	if _, err := queries.ReconcileAccountTransactions(ctx, data.ReconcileAccountTransactionsParams{
		BudgetID:  budget.ID,
		AccountID: target.ID,
		Date:      mustTime(t, 2026, 2, 1),
	}); err != nil {
		t.Fatal(err)
	}

	sourceRows := listAccountTransactions(t, queries, ctx, budget.ID, source.ID)
	if sourceRows[0].Reconciled {
		t.Error("reconciling the target should not reconcile the source's view")
	}
	targetRows := listAccountTransactions(t, queries, ctx, budget.ID, target.ID)
	if !targetRows[0].Reconciled {
		t.Error("expected the mirror row to be reconciled in the target")
	}

	if _, err := queries.ReconcileAccountTransactions(ctx, data.ReconcileAccountTransactionsParams{
		BudgetID:  budget.ID,
		AccountID: source.ID,
		Date:      mustTime(t, 2026, 2, 1),
	}); err != nil {
		t.Fatal(err)
	}

	sourceRows = listAccountTransactions(t, queries, ctx, budget.ID, source.ID)
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
	tx1 := newTrx(t, queries, ctx, account1, payee, mustTime(t, 2026, 1, 1), 1000, 0, "")
	tx2 := newTrx(t, queries, ctx, account2, payee, mustTime(t, 2026, 1, 2), 2000, 0, "")

	rows1 := listAccountTransactions(t, queries, ctx, budget.ID, account1.ID)
	if len(rows1) != 1 || rows1[0].TrxID != tx1.ID {
		t.Errorf("account1 should only contain its own transaction, got %+v", rows1)
	}
	rows2 := listAccountTransactions(t, queries, ctx, budget.ID, account2.ID)
	if len(rows2) != 1 || rows2[0].TrxID != tx2.ID {
		t.Errorf("account2 should only contain its own transaction, got %+v", rows2)
	}
}

func TestListAccountTransactionsOrdering(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)

	budget := newBudget(t, queries, ctx, "testBudget")
	account := newAccount(t, queries, ctx, budget.ID, "Checking")
	payee := newPayee(t, queries, ctx, budget.ID, "Cafe")
	t1 := newTrx(t, queries, ctx, account, payee, mustTime(t, 2026, 1, 3), 1000, 0, "")
	t2 := newTrx(t, queries, ctx, account, payee, mustTime(t, 2026, 1, 1), 2000, 0, "")
	t3 := newTrx(t, queries, ctx, account, payee, mustTime(t, 2026, 1, 2), 3000, 0, "")

	rows := listAccountTransactions(t, queries, ctx, budget.ID, account.ID)
	if len(rows) != 3 {
		t.Fatalf("expected three rows, got %d", len(rows))
	}
	want := []int64{t1.ID, t3.ID, t2.ID}
	for i, id := range want {
		if rows[i].TrxID != id {
			t.Errorf("expected row %d to be transaction %d, got %d", i, id, rows[i].TrxID)
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
	t1 := newTrx(t, queries, ctx, account, zeta, mustTime(t, 2026, 1, 1), 1000, 0, "")
	t2 := newTrx(t, queries, ctx, account, alpha, mustTime(t, 2026, 1, 1), 2000, 0, "")

	rows := listAccountTransactions(t, queries, ctx, budget.ID, account.ID)
	if len(rows) != 2 {
		t.Fatalf("expected two rows, got %d", len(rows))
	}
	if rows[0].TrxID != t2.ID || !rows[0].PayeeName.Valid || rows[0].PayeeName.String != "Alpha" {
		t.Errorf("expected Alpha (payee name, id %d) first, got %+v", alpha.ID, rows[0])
	}
	if rows[1].TrxID != t1.ID || !rows[1].PayeeName.Valid || rows[1].PayeeName.String != "Zeta" {
		t.Errorf("expected Zeta second, got %+v", rows[1])
	}
}

func TestListAccountTransactionsSameDateSamePayeeOrderByTrxID(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)

	budget := newBudget(t, queries, ctx, "testBudget")
	account := newAccount(t, queries, ctx, budget.ID, "Checking")
	payee := newPayee(t, queries, ctx, budget.ID, "Cafe")
	t1 := newTrx(t, queries, ctx, account, payee, mustTime(t, 2026, 1, 1), 1000, 0, "")
	t2 := newTrx(t, queries, ctx, account, payee, mustTime(t, 2026, 1, 1), 2000, 0, "")

	rows := listAccountTransactions(t, queries, ctx, budget.ID, account.ID)
	if len(rows) != 2 {
		t.Fatalf("expected two rows, got %d", len(rows))
	}
	if rows[0].TrxID != t1.ID || rows[1].TrxID != t2.ID {
		t.Errorf("expected insertion order t1 then t2, got %d then %d", rows[0].TrxID, rows[1].TrxID)
	}
}