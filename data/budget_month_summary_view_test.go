package data_test

import (
	"database/sql"
	"testing"

	"samuellando.com/YNAFB/data"
)

func TestGetBudgetMonthSummaryNoActivity(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	summary, err := queries.GetBudgetMonthSummary(ctx, data.GetBudgetMonthSummaryParams{LoginID: budget.LoginID, ID: budget.ID, Month: monthRelative(t, 2)})
	if err != nil {
		t.Fatal(err)
	}
	if summary.ReadyToAssign != 0 {
		t.Errorf("expected ready to assign 0, got %d", summary.ReadyToAssign)
	}
	if summary.Income != 0 {
		t.Errorf("expected income 0, got %d", summary.Income)
	}
	if summary.Allocated != 0 {
		t.Errorf("expected allocated 0, got %d", summary.Allocated)
	}
	if summary.Spent != 0 {
		t.Errorf("expected spent 0, got %d", summary.Spent)
	}
	if summary.Available != 0 {
		t.Errorf("expected available 0, got %d", summary.Available)
	}
	if summary.Uncategorized != 0 {
		t.Errorf("expected uncategorized 0, got %d", summary.Uncategorized)
	}
}

func TestGetBudgetMonthSummaryIncomeAndAllocation(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	account := newAccount(t, queries, ctx, budget, "testaccount")
	payee := newPayee(t, queries, ctx, budget, "testpayee")
	category := newCategory(t, queries, ctx, budget, "testcategory")
	month := monthRelative(t, 2)
	newAllocation(t, queries, ctx, budget.LoginID, budget.ID, category.ID, month, 5000)
	income := newTrx(t, queries, ctx, budget, account, payee, month, 0, 8000, "")
	newIncomeLine(t, queries, ctx, budget, income, 8000)
	summary, err := queries.GetBudgetMonthSummary(ctx, data.GetBudgetMonthSummaryParams{LoginID: budget.LoginID, ID: budget.ID, Month: month})
	if err != nil {
		t.Fatal(err)
	}
	if summary.ReadyToAssign != 3000 {
		t.Errorf("expected ready to assign 3000, got %d", summary.ReadyToAssign)
	}
	if summary.Income != 8000 {
		t.Errorf("expected income 8000, got %d", summary.Income)
	}
	if summary.Allocated != 5000 {
		t.Errorf("expected allocated 5000, got %d", summary.Allocated)
	}
	if summary.Spent != 0 {
		t.Errorf("expected spent 0, got %d", summary.Spent)
	}
	if summary.Available != 5000 {
		t.Errorf("expected available 5000, got %d", summary.Available)
	}
	if summary.Uncategorized != 0 {
		t.Errorf("expected uncategorized 0, got %d", summary.Uncategorized)
	}
}

func TestGetBudgetMonthSummarySpent(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	account := newAccount(t, queries, ctx, budget, "testaccount")
	payee := newPayee(t, queries, ctx, budget, "testpayee")
	category := newCategory(t, queries, ctx, budget, "testcategory")
	month := monthRelative(t, 2)
	newAllocation(t, queries, ctx, budget.LoginID, budget.ID, category.ID, month, 5000)
	income := newTrx(t, queries, ctx, budget, account, payee, month, 0, 8000, "")
	newIncomeLine(t, queries, ctx, budget, income, 8000)
	spend := newTrx(t, queries, ctx, budget, account, payee, month, 1000, 0, "")
	newCategoryLine(t, queries, ctx, budget, spend, category, 1000, 0)
	summary, err := queries.GetBudgetMonthSummary(ctx, data.GetBudgetMonthSummaryParams{LoginID: budget.LoginID, ID: budget.ID, Month: month})
	if err != nil {
		t.Fatal(err)
	}
	if summary.ReadyToAssign != 3000 {
		t.Errorf("expected ready to assign 3000, got %d", summary.ReadyToAssign)
	}
	if summary.Income != 8000 {
		t.Errorf("expected income 8000, got %d", summary.Income)
	}
	if summary.Allocated != 5000 {
		t.Errorf("expected allocated 5000, got %d", summary.Allocated)
	}
	if summary.Spent != 1000 {
		t.Errorf("expected spent 1000, got %d", summary.Spent)
	}
	if summary.Available != 4000 {
		t.Errorf("expected available 4000, got %d", summary.Available)
	}
	if summary.Uncategorized != 0 {
		t.Errorf("expected uncategorized 0, got %d", summary.Uncategorized)
	}
}

func TestGetBudgetMonthSummaryUncategorized(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	account := newAccount(t, queries, ctx, budget, "testaccount")
	payee := newPayee(t, queries, ctx, budget, "testpayee")
	category := newCategory(t, queries, ctx, budget, "testcategory")
	month := monthRelative(t, 2)
	newAllocation(t, queries, ctx, budget.LoginID, budget.ID, category.ID, month, 5000)
	income := newTrx(t, queries, ctx, budget, account, payee, month, 0, 8000, "")
	newIncomeLine(t, queries, ctx, budget, income, 8000)
	newTrx(t, queries, ctx, budget, account, payee, month, 1000, 0, "")
	summary, err := queries.GetBudgetMonthSummary(ctx, data.GetBudgetMonthSummaryParams{LoginID: budget.LoginID, ID: budget.ID, Month: month})
	if err != nil {
		t.Fatal(err)
	}
	if summary.ReadyToAssign != 2000 {
		t.Errorf("expected ready to assign 2000, got %d", summary.ReadyToAssign)
	}
	if summary.Income != 8000 {
		t.Errorf("expected income 8000, got %d", summary.Income)
	}
	if summary.Allocated != 5000 {
		t.Errorf("expected allocated 5000, got %d", summary.Allocated)
	}
	if summary.Spent != 0 {
		t.Errorf("expected spent 0, got %d", summary.Spent)
	}
	if summary.Available != 5000 {
		t.Errorf("expected available 5000, got %d", summary.Available)
	}
	if summary.Uncategorized != 1000 {
		t.Errorf("expected uncategorized 1000, got %d", summary.Uncategorized)
	}
}

func TestGetBudgetMonthSummaryIncomeOnly(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	account := newAccount(t, queries, ctx, budget, "testaccount")
	payee := newPayee(t, queries, ctx, budget, "testpayee")
	month := monthRelative(t, 2)
	income := newTrx(t, queries, ctx, budget, account, payee, month, 0, 8000, "")
	newIncomeLine(t, queries, ctx, budget, income, 8000)
	summary, err := queries.GetBudgetMonthSummary(ctx, data.GetBudgetMonthSummaryParams{LoginID: budget.LoginID, ID: budget.ID, Month: month})
	if err != nil {
		t.Fatal(err)
	}
	if summary.ReadyToAssign != 8000 {
		t.Errorf("expected ready to assign 8000, got %d", summary.ReadyToAssign)
	}
	if summary.Income != 8000 {
		t.Errorf("expected income 8000, got %d", summary.Income)
	}
	if summary.Allocated != 0 {
		t.Errorf("expected allocated 0, got %d", summary.Allocated)
	}
	if summary.Spent != 0 {
		t.Errorf("expected spent 0, got %d", summary.Spent)
	}
	if summary.Available != 0 {
		t.Errorf("expected available 0, got %d", summary.Available)
	}
	if summary.Uncategorized != 0 {
		t.Errorf("expected uncategorized 0, got %d", summary.Uncategorized)
	}
}

func TestGetBudgetMonthSummaryGapMonthCarriesForward(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	account := newAccount(t, queries, ctx, budget, "testaccount")
	payee := newPayee(t, queries, ctx, budget, "testpayee")
	category := newCategory(t, queries, ctx, budget, "testcategory")
	jan := monthRelative(t, -2)
	feb := monthRelative(t, -1)
	newAllocation(t, queries, ctx, budget.LoginID, budget.ID, category.ID, jan, 5000)
	income := newTrx(t, queries, ctx, budget, account, payee, jan, 0, 8000, "")
	newIncomeLine(t, queries, ctx, budget, income, 8000)
	summary, err := queries.GetBudgetMonthSummary(ctx, data.GetBudgetMonthSummaryParams{LoginID: budget.LoginID, ID: budget.ID, Month: feb})
	if err != nil {
		t.Fatal(err)
	}
	if summary.ReadyToAssign != 3000 {
		t.Errorf("expected ready to assign 3000 carried forward, got %d", summary.ReadyToAssign)
	}
	if summary.Income != 0 {
		t.Errorf("expected income 0, got %d", summary.Income)
	}
	if summary.Allocated != 0 {
		t.Errorf("expected allocated 0, got %d", summary.Allocated)
	}
	if summary.Spent != 0 {
		t.Errorf("expected spent 0, got %d", summary.Spent)
	}
	if summary.Available != 5000 {
		t.Errorf("expected available 5000 carried forward, got %d", summary.Available)
	}
	if summary.Uncategorized != 0 {
		t.Errorf("expected uncategorized 0, got %d", summary.Uncategorized)
	}
}

func TestGetBudgetMonthSummaryCumulativeAvailablePastMonth(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	account := newAccount(t, queries, ctx, budget, "testaccount")
	payee := newPayee(t, queries, ctx, budget, "testpayee")
	category := newCategory(t, queries, ctx, budget, "testcategory")
	jan := monthRelative(t, -2)
	feb := monthRelative(t, -1)
	newAllocation(t, queries, ctx, budget.LoginID, budget.ID, category.ID, jan, 5000)
	janSpend := newTrx(t, queries, ctx, budget, account, payee, jan, 1000, 0, "")
	newCategoryLine(t, queries, ctx, budget, janSpend, category, 1000, 0)
	newAllocation(t, queries, ctx, budget.LoginID, budget.ID, category.ID, feb, 2000)
	summary, err := queries.GetBudgetMonthSummary(ctx, data.GetBudgetMonthSummaryParams{LoginID: budget.LoginID, ID: budget.ID, Month: feb})
	if err != nil {
		t.Fatal(err)
	}
	if summary.ReadyToAssign != -7000 {
		t.Errorf("expected ready to assign -7000, got %d", summary.ReadyToAssign)
	}
	if summary.Income != 0 {
		t.Errorf("expected income 0, got %d", summary.Income)
	}
	if summary.Allocated != 2000 {
		t.Errorf("expected allocated 2000, got %d", summary.Allocated)
	}
	if summary.Spent != 0 {
		t.Errorf("expected spent 0, got %d", summary.Spent)
	}
	if summary.Available != 6000 {
		t.Errorf("expected cumulative available 6000, got %d", summary.Available)
	}
	if summary.Uncategorized != 0 {
		t.Errorf("expected uncategorized 0, got %d", summary.Uncategorized)
	}
}

func TestGetBudgetMonthSummaryFutureMonthStartsFresh(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	category := newCategory(t, queries, ctx, budget, "testcategory")
	past := monthRelative(t, -2)
	future := monthRelative(t, 2)
	newAllocation(t, queries, ctx, budget.LoginID, budget.ID, category.ID, past, 5000)
	newAllocation(t, queries, ctx, budget.LoginID, budget.ID, category.ID, future, 3000)
	summary, err := queries.GetBudgetMonthSummary(ctx, data.GetBudgetMonthSummaryParams{LoginID: budget.LoginID, ID: budget.ID, Month: future})
	if err != nil {
		t.Fatal(err)
	}
	if summary.ReadyToAssign != -3000 {
		t.Errorf("expected ready to assign -3000, got %d", summary.ReadyToAssign)
	}
	if summary.Income != 0 {
		t.Errorf("expected income 0, got %d", summary.Income)
	}
	if summary.Allocated != 3000 {
		t.Errorf("expected allocated 3000, got %d", summary.Allocated)
	}
	if summary.Spent != 0 {
		t.Errorf("expected spent 0, got %d", summary.Spent)
	}
	if summary.Available != 3000 {
		t.Errorf("expected available 3000, got %d", summary.Available)
	}
	if summary.Uncategorized != 0 {
		t.Errorf("expected uncategorized 0, got %d", summary.Uncategorized)
	}
}

func TestGetBudgetMonthSummaryOverspendDeductsFromReadyToAssign(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	account := newAccount(t, queries, ctx, budget, "testaccount")
	payee := newPayee(t, queries, ctx, budget, "testpayee")
	category := newCategory(t, queries, ctx, budget, "testcategory")
	month := monthRelative(t, 2)
	newAllocation(t, queries, ctx, budget.LoginID, budget.ID, category.ID, month, 5000)
	income := newTrx(t, queries, ctx, budget, account, payee, month, 0, 8000, "")
	newIncomeLine(t, queries, ctx, budget, income, 8000)
	overspend := newTrx(t, queries, ctx, budget, account, payee, month, 6000, 0, "")
	newCategoryLine(t, queries, ctx, budget, overspend, category, 6000, 0)
	summary, err := queries.GetBudgetMonthSummary(ctx, data.GetBudgetMonthSummaryParams{LoginID: budget.LoginID, ID: budget.ID, Month: month})
	if err != nil {
		t.Fatal(err)
	}
	if summary.ReadyToAssign != 2000 {
		t.Errorf("expected ready to assign 2000 with overspend deducted, got %d", summary.ReadyToAssign)
	}
	if summary.Income != 8000 {
		t.Errorf("expected income 8000, got %d", summary.Income)
	}
	if summary.Allocated != 5000 {
		t.Errorf("expected allocated 5000, got %d", summary.Allocated)
	}
	if summary.Spent != 6000 {
		t.Errorf("expected spent 6000, got %d", summary.Spent)
	}
	if summary.Available != -1000 {
		t.Errorf("expected available -1000, got %d", summary.Available)
	}
	if summary.Uncategorized != 0 {
		t.Errorf("expected uncategorized 0, got %d", summary.Uncategorized)
	}
}

func TestGetBudgetMonthSummaryOverspendCarriesForward(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	account := newAccount(t, queries, ctx, budget, "testaccount")
	payee := newPayee(t, queries, ctx, budget, "testpayee")
	category := newCategory(t, queries, ctx, budget, "testcategory")
	jan := monthRelative(t, -2)
	feb := monthRelative(t, -1)
	newAllocation(t, queries, ctx, budget.LoginID, budget.ID, category.ID, jan, 5000)
	income := newTrx(t, queries, ctx, budget, account, payee, jan, 0, 8000, "")
	newIncomeLine(t, queries, ctx, budget, income, 8000)
	overspend := newTrx(t, queries, ctx, budget, account, payee, jan, 6000, 0, "")
	newCategoryLine(t, queries, ctx, budget, overspend, category, 6000, 0)
	summary, err := queries.GetBudgetMonthSummary(ctx, data.GetBudgetMonthSummaryParams{LoginID: budget.LoginID, ID: budget.ID, Month: feb})
	if err != nil {
		t.Fatal(err)
	}
	if summary.ReadyToAssign != 2000 {
		t.Errorf("expected ready to assign 2000 carried forward, got %d", summary.ReadyToAssign)
	}
	if summary.Income != 0 {
		t.Errorf("expected income 0, got %d", summary.Income)
	}
	if summary.Allocated != 0 {
		t.Errorf("expected allocated 0, got %d", summary.Allocated)
	}
	if summary.Spent != 0 {
		t.Errorf("expected spent 0, got %d", summary.Spent)
	}
	if summary.Available != -1000 {
		t.Errorf("expected available -1000 carried forward, got %d", summary.Available)
	}
	if summary.Uncategorized != 0 {
		t.Errorf("expected uncategorized 0, got %d", summary.Uncategorized)
	}
}

func TestGetBudgetMonthSummaryScopedToBudget(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget1 := newBudget(t, queries, ctx, "budget1")
	budget2 := newBudget(t, queries, ctx, "budget2")
	account1 := newAccount(t, queries, ctx, budget1, "account1")
	payee1 := newPayee(t, queries, ctx, budget1, "payee1")
	account2 := newAccount(t, queries, ctx, budget2, "account2")
	payee2 := newPayee(t, queries, ctx, budget2, "payee2")
	month := monthRelative(t, 2)
	income1 := newTrx(t, queries, ctx, budget1, account1, payee1, month, 0, 8000, "")
	newIncomeLine(t, queries, ctx, budget1, income1, 8000)
	income2 := newTrx(t, queries, ctx, budget2, account2, payee2, month, 0, 5000, "")
	newIncomeLine(t, queries, ctx, budget2, income2, 5000)
	summary, err := queries.GetBudgetMonthSummary(ctx, data.GetBudgetMonthSummaryParams{LoginID: budget1.LoginID, ID: budget1.ID, Month: month})
	if err != nil {
		t.Fatal(err)
	}
	if summary.Income != 8000 {
		t.Errorf("expected income 8000 for budget1, got %d", summary.Income)
	}
}

func TestScopingListBudgetMonthSummaryScopedByLogin(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budgetA := newBudget(t, queries, ctx, "budgetA")
	budgetB := newBudget(t, queries, ctx, "budgetB")

	_, err := queries.GetBudgetMonthSummary(ctx, data.GetBudgetMonthSummaryParams{
		LoginID: budgetA.LoginID,
		ID:      budgetB.ID,
		Month:   monthRelative(t, 1),
	})
	if err != sql.ErrNoRows {
		t.Fatalf("expected sql.ErrNoRows for another login's budget summary, got %v", err)
	}
}
