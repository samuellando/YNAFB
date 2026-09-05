package data_test

import "testing"

func TestGetBudgetMonthSummaryNoActivity(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	newBudget(t, queries, ctx, "testBudget")
	summary, err := queries.GetBudgetMonthSummary(ctx, monthRelative(t, 2))
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
	account := newAccount(t, queries, ctx, budget.ID, "testaccount")
	payee := newPayee(t, queries, ctx, budget.ID, "testpayee")
	category := newCategory(t, queries, ctx, budget.ID, "testcategory")
	month := monthRelative(t, 2)
	newAllocation(t, queries, ctx, budget.ID, category.ID, month, 5000)
	income := newTransaction(t, queries, ctx, account.ID, payee.ID, month, 0, 8000, "")
	newIncomeLine(t, queries, ctx, income.ID, 8000)
	summary, err := queries.GetBudgetMonthSummary(ctx, month)
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
	account := newAccount(t, queries, ctx, budget.ID, "testaccount")
	payee := newPayee(t, queries, ctx, budget.ID, "testpayee")
	category := newCategory(t, queries, ctx, budget.ID, "testcategory")
	month := monthRelative(t, 2)
	newAllocation(t, queries, ctx, budget.ID, category.ID, month, 5000)
	income := newTransaction(t, queries, ctx, account.ID, payee.ID, month, 0, 8000, "")
	newIncomeLine(t, queries, ctx, income.ID, 8000)
	spend := newTransaction(t, queries, ctx, account.ID, payee.ID, month, 1000, 0, "")
	newCategoryLine(t, queries, ctx, spend.ID, category.ID, 1000, 0)
	summary, err := queries.GetBudgetMonthSummary(ctx, month)
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
	account := newAccount(t, queries, ctx, budget.ID, "testaccount")
	payee := newPayee(t, queries, ctx, budget.ID, "testpayee")
	category := newCategory(t, queries, ctx, budget.ID, "testcategory")
	month := monthRelative(t, 2)
	newAllocation(t, queries, ctx, budget.ID, category.ID, month, 5000)
	income := newTransaction(t, queries, ctx, account.ID, payee.ID, month, 0, 8000, "")
	newIncomeLine(t, queries, ctx, income.ID, 8000)
	newTransaction(t, queries, ctx, account.ID, payee.ID, month, 1000, 0, "")
	summary, err := queries.GetBudgetMonthSummary(ctx, month)
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
	account := newAccount(t, queries, ctx, budget.ID, "testaccount")
	payee := newPayee(t, queries, ctx, budget.ID, "testpayee")
	month := monthRelative(t, 2)
	income := newTransaction(t, queries, ctx, account.ID, payee.ID, month, 0, 8000, "")
	newIncomeLine(t, queries, ctx, income.ID, 8000)
	summary, err := queries.GetBudgetMonthSummary(ctx, month)
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
	account := newAccount(t, queries, ctx, budget.ID, "testaccount")
	payee := newPayee(t, queries, ctx, budget.ID, "testpayee")
	category := newCategory(t, queries, ctx, budget.ID, "testcategory")
	jan := monthRelative(t, -2)
	feb := monthRelative(t, -1)
	newAllocation(t, queries, ctx, budget.ID, category.ID, jan, 5000)
	income := newTransaction(t, queries, ctx, account.ID, payee.ID, jan, 0, 8000, "")
	newIncomeLine(t, queries, ctx, income.ID, 8000)
	summary, err := queries.GetBudgetMonthSummary(ctx, feb)
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
	account := newAccount(t, queries, ctx, budget.ID, "testaccount")
	payee := newPayee(t, queries, ctx, budget.ID, "testpayee")
	category := newCategory(t, queries, ctx, budget.ID, "testcategory")
	jan := monthRelative(t, -2)
	feb := monthRelative(t, -1)
	newAllocation(t, queries, ctx, budget.ID, category.ID, jan, 5000)
	janSpend := newTransaction(t, queries, ctx, account.ID, payee.ID, jan, 1000, 0, "")
	newCategoryLine(t, queries, ctx, janSpend.ID, category.ID, 1000, 0)
	newAllocation(t, queries, ctx, budget.ID, category.ID, feb, 2000)
	summary, err := queries.GetBudgetMonthSummary(ctx, feb)
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
	category := newCategory(t, queries, ctx, budget.ID, "testcategory")
	past := monthRelative(t, -2)
	future := monthRelative(t, 2)
	newAllocation(t, queries, ctx, budget.ID, category.ID, past, 5000)
	newAllocation(t, queries, ctx, budget.ID, category.ID, future, 3000)
	summary, err := queries.GetBudgetMonthSummary(ctx, future)
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
	account := newAccount(t, queries, ctx, budget.ID, "testaccount")
	payee := newPayee(t, queries, ctx, budget.ID, "testpayee")
	category := newCategory(t, queries, ctx, budget.ID, "testcategory")
	month := monthRelative(t, 2)
	newAllocation(t, queries, ctx, budget.ID, category.ID, month, 5000)
	income := newTransaction(t, queries, ctx, account.ID, payee.ID, month, 0, 8000, "")
	newIncomeLine(t, queries, ctx, income.ID, 8000)
	overspend := newTransaction(t, queries, ctx, account.ID, payee.ID, month, 6000, 0, "")
	newCategoryLine(t, queries, ctx, overspend.ID, category.ID, 6000, 0)
	summary, err := queries.GetBudgetMonthSummary(ctx, month)
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
	account := newAccount(t, queries, ctx, budget.ID, "testaccount")
	payee := newPayee(t, queries, ctx, budget.ID, "testpayee")
	category := newCategory(t, queries, ctx, budget.ID, "testcategory")
	jan := monthRelative(t, -2)
	feb := monthRelative(t, -1)
	newAllocation(t, queries, ctx, budget.ID, category.ID, jan, 5000)
	income := newTransaction(t, queries, ctx, account.ID, payee.ID, jan, 0, 8000, "")
	newIncomeLine(t, queries, ctx, income.ID, 8000)
	overspend := newTransaction(t, queries, ctx, account.ID, payee.ID, jan, 6000, 0, "")
	newCategoryLine(t, queries, ctx, overspend.ID, category.ID, 6000, 0)
	summary, err := queries.GetBudgetMonthSummary(ctx, feb)
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