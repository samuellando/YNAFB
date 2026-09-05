package data_test

import (
	"database/sql"
	"testing"
	"time"

	"samuellando.com/YNAFB/data"
	"samuellando.com/YNAFB/internal/db/types"
)

func monthRelative(t *testing.T, offset int) types.UnixTime {
	t.Helper()
	now := time.Now().UTC()
	first := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	return types.UnixTime{Time: first.AddDate(0, offset, 0)}
}

func TestListBudgetMonthCategoriesEmpty(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	rows, err := queries.ListBudgetMonthCategories(ctx, data.ListBudgetMonthCategoriesParams{
		BudgetID: budget.ID,
		Month:    monthRelative(t, 1),
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 0 {
		t.Fatalf("expected no categories, got %d", len(rows))
	}
}

func TestListBudgetMonthCategoriesNoActivity(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	category := newCategory(t, queries, ctx, budget.ID, "testcategory")
	rows, err := queries.ListBudgetMonthCategories(ctx, data.ListBudgetMonthCategoriesParams{
		BudgetID: budget.ID,
		Month:    monthRelative(t, 1),
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected one category, got %d", len(rows))
	}
	if rows[0].ID != category.ID {
		t.Error("category id does not match")
	}
	if rows[0].CategoryID != category.ID {
		t.Error("category id does not match")
	}
	if rows[0].CategoryName != "testcategory" {
		t.Error("category name does not match")
	}
	if rows[0].CategoryGroupName.Valid {
		t.Error("ungrouped category should have an invalid group name")
	}
	if rows[0].Allocated != 0 {
		t.Errorf("expected allocated 0, got %d", rows[0].Allocated)
	}
	if rows[0].Spent != 0 {
		t.Errorf("expected spent 0, got %d", rows[0].Spent)
	}
	if rows[0].Available != 0 {
		t.Errorf("expected available 0, got %d", rows[0].Available)
	}
}

func TestListBudgetMonthCategoriesAllocated(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	category := newCategory(t, queries, ctx, budget.ID, "testcategory")
	month := monthRelative(t, 2)
	newAllocation(t, queries, ctx, budget.ID, category.ID, month, 5000)
	rows, err := queries.ListBudgetMonthCategories(ctx, data.ListBudgetMonthCategoriesParams{
		BudgetID: budget.ID,
		Month:    month,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected one category, got %d", len(rows))
	}
	if rows[0].Allocated != 5000 {
		t.Errorf("expected allocated 5000, got %d", rows[0].Allocated)
	}
	if rows[0].Spent != 0 {
		t.Errorf("expected spent 0, got %d", rows[0].Spent)
	}
	if rows[0].Available != 5000 {
		t.Errorf("expected available 5000, got %d", rows[0].Available)
	}
}

func TestListBudgetMonthCategoriesSpent(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	account := newAccount(t, queries, ctx, budget.ID, "testaccount")
	payee := newPayee(t, queries, ctx, budget.ID, "testpayee")
	category := newCategory(t, queries, ctx, budget.ID, "testcategory")
	month := monthRelative(t, 2)
	transaction := newTrx(t, queries, ctx, account, payee, month, 1000, 0, "")
	newCategoryLine(t, queries, ctx, transaction, category, 1000, 0)
	rows, err := queries.ListBudgetMonthCategories(ctx, data.ListBudgetMonthCategoriesParams{
		BudgetID: budget.ID,
		Month:    month,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected one category, got %d", len(rows))
	}
	if rows[0].Allocated != 0 {
		t.Errorf("expected allocated 0, got %d", rows[0].Allocated)
	}
	if rows[0].Spent != 1000 {
		t.Errorf("expected spent 1000, got %d", rows[0].Spent)
	}
	if rows[0].Available != -1000 {
		t.Errorf("expected available -1000, got %d", rows[0].Available)
	}
}

func TestListBudgetMonthCategoriesAllocatedAndSpent(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	account := newAccount(t, queries, ctx, budget.ID, "testaccount")
	payee := newPayee(t, queries, ctx, budget.ID, "testpayee")
	category := newCategory(t, queries, ctx, budget.ID, "testcategory")
	month := monthRelative(t, 2)
	newAllocation(t, queries, ctx, budget.ID, category.ID, month, 5000)
	transaction := newTrx(t, queries, ctx, account, payee, month, 1000, 0, "")
	newCategoryLine(t, queries, ctx, transaction, category, 1000, 0)
	rows, err := queries.ListBudgetMonthCategories(ctx, data.ListBudgetMonthCategoriesParams{
		BudgetID: budget.ID,
		Month:    month,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected one category, got %d", len(rows))
	}
	if rows[0].Allocated != 5000 {
		t.Errorf("expected allocated 5000, got %d", rows[0].Allocated)
	}
	if rows[0].Spent != 1000 {
		t.Errorf("expected spent 1000, got %d", rows[0].Spent)
	}
	if rows[0].Available != 4000 {
		t.Errorf("expected available 4000, got %d", rows[0].Available)
	}
}

func TestListBudgetMonthCategoriesSpendSummedAcrossTransactions(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	account := newAccount(t, queries, ctx, budget.ID, "testaccount")
	payee := newPayee(t, queries, ctx, budget.ID, "testpayee")
	category := newCategory(t, queries, ctx, budget.ID, "testcategory")
	month := monthRelative(t, 2)
	first := newTrx(t, queries, ctx, account, payee, month, 1000, 0, "")
	newCategoryLine(t, queries, ctx, first, category, 1000, 0)
	second := newTrx(t, queries, ctx, account, payee, month, 500, 0, "")
	newCategoryLine(t, queries, ctx, second, category, 500, 0)
	rows, err := queries.ListBudgetMonthCategories(ctx, data.ListBudgetMonthCategoriesParams{
		BudgetID: budget.ID,
		Month:    month,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected one category, got %d", len(rows))
	}
	if rows[0].Spent != 1500 {
		t.Errorf("expected spent 1500, got %d", rows[0].Spent)
	}
}

func TestListBudgetMonthCategoriesCarryForwardAvailable(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	account := newAccount(t, queries, ctx, budget.ID, "testaccount")
	payee := newPayee(t, queries, ctx, budget.ID, "testpayee")
	category := newCategory(t, queries, ctx, budget.ID, "testcategory")
	jan := monthRelative(t, -2)
	feb := monthRelative(t, -1)
	newAllocation(t, queries, ctx, budget.ID, category.ID, jan, 5000)
	transaction := newTrx(t, queries, ctx, account, payee, jan, 1000, 0, "")
	newCategoryLine(t, queries, ctx, transaction, category, 1000, 0)
	rows, err := queries.ListBudgetMonthCategories(ctx, data.ListBudgetMonthCategoriesParams{
		BudgetID: budget.ID,
		Month:    feb,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected one category, got %d", len(rows))
	}
	if rows[0].Allocated != 0 {
		t.Errorf("expected allocated 0, got %d", rows[0].Allocated)
	}
	if rows[0].Spent != 0 {
		t.Errorf("expected spent 0, got %d", rows[0].Spent)
	}
	if rows[0].Available != 4000 {
		t.Errorf("expected available 4000 carried forward, got %d", rows[0].Available)
	}
}

func TestListBudgetMonthCategoriesCarryForwardFloorsAtZero(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	account := newAccount(t, queries, ctx, budget.ID, "testaccount")
	payee := newPayee(t, queries, ctx, budget.ID, "testpayee")
	category := newCategory(t, queries, ctx, budget.ID, "testcategory")
	jan := monthRelative(t, -2)
	feb := monthRelative(t, -1)
	newAllocation(t, queries, ctx, budget.ID, category.ID, jan, 1000)
	transaction := newTrx(t, queries, ctx, account, payee, jan, 5000, 0, "")
	newCategoryLine(t, queries, ctx, transaction, category, 5000, 0)
	rows, err := queries.ListBudgetMonthCategories(ctx, data.ListBudgetMonthCategoriesParams{
		BudgetID: budget.ID,
		Month:    feb,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected one category, got %d", len(rows))
	}
	if rows[0].Available != 0 {
		t.Errorf("expected available floored at 0, got %d", rows[0].Available)
	}
}

func TestListBudgetMonthCategoriesCarryForwardMostRecentPriorMonth(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	category := newCategory(t, queries, ctx, budget.ID, "testcategory")
	jan := monthRelative(t, -3)
	feb := monthRelative(t, -2)
	mar := monthRelative(t, -1)
	newAllocation(t, queries, ctx, budget.ID, category.ID, jan, 2000)
	newAllocation(t, queries, ctx, budget.ID, category.ID, mar, 3000)
	rows, err := queries.ListBudgetMonthCategories(ctx, data.ListBudgetMonthCategoriesParams{
		BudgetID: budget.ID,
		Month:    feb,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected one category, got %d", len(rows))
	}
	if rows[0].Available != 2000 {
		t.Errorf("expected available 2000 from the most recent prior month, got %d", rows[0].Available)
	}
}

func TestListBudgetMonthCategoriesPastMonthCumulativeAvailable(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	account := newAccount(t, queries, ctx, budget.ID, "testaccount")
	payee := newPayee(t, queries, ctx, budget.ID, "testpayee")
	category := newCategory(t, queries, ctx, budget.ID, "testcategory")
	jan := monthRelative(t, -2)
	feb := monthRelative(t, -1)
	newAllocation(t, queries, ctx, budget.ID, category.ID, jan, 5000)
	janTransaction := newTrx(t, queries, ctx, account, payee, jan, 1000, 0, "")
	newCategoryLine(t, queries, ctx, janTransaction, category, 1000, 0)
	newAllocation(t, queries, ctx, budget.ID, category.ID, feb, 2000)
	rows, err := queries.ListBudgetMonthCategories(ctx, data.ListBudgetMonthCategoriesParams{
		BudgetID: budget.ID,
		Month:    feb,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected one category, got %d", len(rows))
	}
	if rows[0].Allocated != 2000 {
		t.Errorf("expected allocated 2000, got %d", rows[0].Allocated)
	}
	if rows[0].Spent != 0 {
		t.Errorf("expected spent 0, got %d", rows[0].Spent)
	}
	if rows[0].Available != 6000 {
		t.Errorf("expected cumulative available 6000, got %d", rows[0].Available)
	}
}

func TestListBudgetMonthCategoriesOrderingAndGroups(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	alpha := newCategory(t, queries, ctx, budget.ID, "alpha")
	groceries := newCategoryGroup(t, queries, ctx, budget.ID, "groceries")
	bills := newCategoryGroup(t, queries, ctx, budget.ID, "bills")
	banana := newCategory(t, queries, ctx, budget.ID, "banana")
	apple := newCategory(t, queries, ctx, budget.ID, "apple")
	rent := newCategory(t, queries, ctx, budget.ID, "rent")
	if _, err := queries.UpdateCategory(ctx, data.UpdateCategoryParams{
		Name:            banana.Name,
		CategoryGroupID: sql.NullInt64{Int64: groceries, Valid: true},
		ID:              banana.ID,
		BudgetID:        budget.ID,
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := queries.UpdateCategory(ctx, data.UpdateCategoryParams{
		Name:            apple.Name,
		CategoryGroupID: sql.NullInt64{Int64: groceries, Valid: true},
		ID:              apple.ID,
		BudgetID:        budget.ID,
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := queries.UpdateCategory(ctx, data.UpdateCategoryParams{
		Name:            rent.Name,
		CategoryGroupID: sql.NullInt64{Int64: bills, Valid: true},
		ID:              rent.ID,
		BudgetID:        budget.ID,
	}); err != nil {
		t.Fatal(err)
	}
	rows, err := queries.ListBudgetMonthCategories(ctx, data.ListBudgetMonthCategoriesParams{
		BudgetID: budget.ID,
		Month:    monthRelative(t, 1),
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 4 {
		t.Fatalf("expected four categories, got %d", len(rows))
	}
	byName := map[string]data.ListBudgetMonthCategoriesRow{}
	for _, r := range rows {
		byName[r.CategoryName] = r
	}
	if r := byName["alpha"]; r.ID != alpha.ID || r.CategoryGroupName.Valid {
		t.Errorf("alpha should be ungrouped with id %d, got id %d valid %v", alpha.ID, r.ID, r.CategoryGroupName.Valid)
	}
	if r := byName["apple"]; r.ID != apple.ID || !r.CategoryGroupName.Valid || r.CategoryGroupName.String != "groceries" {
		t.Errorf("apple should be in groceries with id %d, got id %d group %v", apple.ID, r.ID, r.CategoryGroupName)
	}
	if r := byName["banana"]; r.ID != banana.ID || !r.CategoryGroupName.Valid || r.CategoryGroupName.String != "groceries" {
		t.Errorf("banana should be in groceries with id %d, got id %d group %v", banana.ID, r.ID, r.CategoryGroupName)
	}
	if r := byName["rent"]; r.ID != rent.ID || !r.CategoryGroupName.Valid || r.CategoryGroupName.String != "bills" {
		t.Errorf("rent should be in bills with id %d, got id %d group %v", rent.ID, r.ID, r.CategoryGroupName)
	}
	order := []string{}
	for _, r := range rows {
		order = append(order, r.CategoryName)
	}
	expected := []string{"alpha", "rent", "apple", "banana"}
	for i, name := range expected {
		if order[i] != name {
			t.Errorf("expected order %v, got %v", expected, order)
			break
		}
	}
}

func TestListBudgetMonthCategoriesScopedToBudget(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget1 := newBudget(t, queries, ctx, "testBudget1")
	budget2 := newBudget(t, queries, ctx, "testBudget2")
	category1 := newCategory(t, queries, ctx, budget1.ID, "shared")
	category2 := newCategory(t, queries, ctx, budget2.ID, "shared")
	rows1, err := queries.ListBudgetMonthCategories(ctx, data.ListBudgetMonthCategoriesParams{
		BudgetID: budget1.ID,
		Month:    monthRelative(t, 1),
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(rows1) != 1 {
		t.Fatalf("expected one category for budget1, got %d", len(rows1))
	}
	if rows1[0].ID != category1.ID {
		t.Error("budget1 returned the wrong category")
	}
	rows2, err := queries.ListBudgetMonthCategories(ctx, data.ListBudgetMonthCategoriesParams{
		BudgetID: budget2.ID,
		Month:    monthRelative(t, 1),
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(rows2) != 1 {
		t.Fatalf("expected one category for budget2, got %d", len(rows2))
	}
	if rows2[0].ID != category2.ID {
		t.Error("budget2 returned the wrong category")
	}
}

func TestListBudgetMonthCategoriesRefundReducesSpent(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	account := newAccount(t, queries, ctx, budget.ID, "testaccount")
	payee := newPayee(t, queries, ctx, budget.ID, "testpayee")
	category := newCategory(t, queries, ctx, budget.ID, "testcategory")
	month := monthRelative(t, 2)
	newAllocation(t, queries, ctx, budget.ID, category.ID, month, 5000)
	spend := newTrx(t, queries, ctx, account, payee, month, 1000, 0, "")
	newCategoryLine(t, queries, ctx, spend, category, 1000, 0)
	refund := newTrx(t, queries, ctx, account, payee, month, 0, 200, "")
	newCategoryLine(t, queries, ctx, refund, category, 0, 200)
	rows, err := queries.ListBudgetMonthCategories(ctx, data.ListBudgetMonthCategoriesParams{
		BudgetID: budget.ID,
		Month:    month,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected one category, got %d", len(rows))
	}
	if rows[0].Allocated != 5000 {
		t.Errorf("expected allocated 5000, got %d", rows[0].Allocated)
	}
	if rows[0].Spent != 800 {
		t.Errorf("expected spent 800, got %d", rows[0].Spent)
	}
	if rows[0].Available != 4200 {
		t.Errorf("expected available 4200, got %d", rows[0].Available)
	}
}

func TestListBudgetMonthCategoriesIncomeAndTransferExcludedFromSpent(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	account := newAccount(t, queries, ctx, budget.ID, "testaccount")
	otherAccount := newAccount(t, queries, ctx, budget.ID, "otheraccount")
	payee := newPayee(t, queries, ctx, budget.ID, "testpayee")
	category := newCategory(t, queries, ctx, budget.ID, "testcategory")
	month := monthRelative(t, 2)
	spend := newTrx(t, queries, ctx, account, payee, month, 1000, 0, "")
	newCategoryLine(t, queries, ctx, spend, category, 1000, 0)
	income := newTrx(t, queries, ctx, account, payee, month, 0, 5000, "")
	newIncomeLine(t, queries, ctx, income, 5000)
	transfer := newTrx(t, queries, ctx, account, payee, month, 2000, 0, "")
	newTransfer(t, queries, ctx, transfer, otherAccount, 2000, 0)
	rows, err := queries.ListBudgetMonthCategories(ctx, data.ListBudgetMonthCategoriesParams{
		BudgetID: budget.ID,
		Month:    month,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected one category, got %d", len(rows))
	}
	if rows[0].Allocated != 0 {
		t.Errorf("expected allocated 0, got %d", rows[0].Allocated)
	}
	if rows[0].Spent != 1000 {
		t.Errorf("expected spent 1000, got %d", rows[0].Spent)
	}
	if rows[0].Available != -1000 {
		t.Errorf("expected available -1000, got %d", rows[0].Available)
	}
}

func TestListBudgetMonthCategoriesFutureMonthResetsAvailable(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	category := newCategory(t, queries, ctx, budget.ID, "testcategory")
	past := monthRelative(t, -2)
	future := monthRelative(t, 2)
	newAllocation(t, queries, ctx, budget.ID, category.ID, past, 5000)
	rows, err := queries.ListBudgetMonthCategories(ctx, data.ListBudgetMonthCategoriesParams{
		BudgetID: budget.ID,
		Month:    future,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected one category, got %d", len(rows))
	}
	if rows[0].Allocated != 0 {
		t.Errorf("expected allocated 0, got %d", rows[0].Allocated)
	}
	if rows[0].Spent != 0 {
		t.Errorf("expected spent 0, got %d", rows[0].Spent)
	}
	if rows[0].Available != 0 {
		t.Errorf("expected available 0, got %d", rows[0].Available)
	}
}

func TestListBudgetMonthCategoriesNegativeAllocation(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	category := newCategory(t, queries, ctx, budget.ID, "testcategory")
	jan := monthRelative(t, -2)
	feb := monthRelative(t, -1)
	newAllocation(t, queries, ctx, budget.ID, category.ID, jan, -1000)
	rows, err := queries.ListBudgetMonthCategories(ctx, data.ListBudgetMonthCategoriesParams{
		BudgetID: budget.ID,
		Month:    jan,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected one category, got %d", len(rows))
	}
	if rows[0].Allocated != -1000 {
		t.Errorf("expected allocated -1000, got %d", rows[0].Allocated)
	}
	if rows[0].Available != -1000 {
		t.Errorf("expected available -1000, got %d", rows[0].Available)
	}
	carried, err := queries.ListBudgetMonthCategories(ctx, data.ListBudgetMonthCategoriesParams{
		BudgetID: budget.ID,
		Month:    feb,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(carried) != 1 {
		t.Fatalf("expected one category, got %d", len(carried))
	}
	if carried[0].Available != 0 {
		t.Errorf("expected carried forward available floored at 0, got %d", carried[0].Available)
	}
}

func TestListBudgetMonthCategoriesSiblingActivityCreatesViewRow(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	categoryA := newCategory(t, queries, ctx, budget.ID, "categoryA")
	categoryB := newCategory(t, queries, ctx, budget.ID, "categoryB")
	month := monthRelative(t, 2)
	newAllocation(t, queries, ctx, budget.ID, categoryA.ID, month, 5000)
	rows, err := queries.ListBudgetMonthCategories(ctx, data.ListBudgetMonthCategoriesParams{
		BudgetID: budget.ID,
		Month:    month,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 2 {
		t.Fatalf("expected two categories, got %d", len(rows))
	}
	byName := map[string]data.ListBudgetMonthCategoriesRow{}
	for _, r := range rows {
		byName[r.CategoryName] = r
	}
	if r := byName["categoryA"]; r.Allocated != 5000 || r.Available != 5000 {
		t.Errorf("expected categoryA allocated/available 5000/5000, got %d/%d", r.Allocated, r.Available)
	}
	if r := byName["categoryB"]; r.ID != categoryB.ID || r.Allocated != 0 || r.Spent != 0 || r.Available != 0 {
		t.Errorf("expected categoryB zeroed with id %d, got id %d allocated %d spent %d available %d", categoryB.ID, r.ID, r.Allocated, r.Spent, r.Available)
	}
}