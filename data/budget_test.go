package data_test

import (
	"testing"

	"samuellando.com/YNAFB/data"
)

func TestCreateBudget(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	login := newLogin(t, queries, ctx, "user")
	budget, err := queries.CreateBudget(ctx, data.CreateBudgetParams{LoginID: login.ID, Name: "testBudget"})
	if err != nil {
		t.Fatal(err)
	}
	if budget.Name != "testBudget" {
		t.Error("budget name does not match")
	}
	if budget.LoginID != login.ID {
		t.Error("budget login id does not match")
	}
	if budget.ID != 1 {
		t.Error("ID of the first budget should be 1")
	}
}

func TestCreateBudgetEmptyName(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	login := newLogin(t, queries, ctx, "user")
	_, err := queries.CreateBudget(ctx, data.CreateBudgetParams{LoginID: login.ID, Name: ""})
	if err == nil {
		t.Error("Empty budget name should raise an error")
	}
}

func TestCreateBudgetDuplicateName(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	login := newLogin(t, queries, ctx, "user")
	if _, err := queries.CreateBudget(ctx, data.CreateBudgetParams{LoginID: login.ID, Name: "testBudget"}); err != nil {
		t.Fatal(err)
	}
	_, err := queries.CreateBudget(ctx, data.CreateBudgetParams{LoginID: login.ID, Name: "testBudget"})
	if err == nil {
		t.Error("Duplicate budget name for the same login should raise an error")
	}
}

func TestCreateBudgetSameNameAcrossLogins(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	login1 := newLogin(t, queries, ctx, "user1")
	login2 := newLogin(t, queries, ctx, "user2")
	if _, err := queries.CreateBudget(ctx, data.CreateBudgetParams{LoginID: login1.ID, Name: "testBudget"}); err != nil {
		t.Fatal(err)
	}
	if _, err := queries.CreateBudget(ctx, data.CreateBudgetParams{LoginID: login2.ID, Name: "testBudget"}); err != nil {
		t.Error("Duplicate budget names across logins should be allowed")
	}
}

func TestUpdateBudget(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	n, err := queries.UpdateBudget(ctx, data.UpdateBudgetParams{
		Name: "newName",
		ID:   budget.ID,
	})
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Error("The number of affected rows should be 1")
	}
	newNameBudget, err := queries.GetBudgetByName(ctx, data.GetBudgetByNameParams{LoginID: budget.LoginID, Name: "newName"})
	if err != nil {
		t.Fatal(err)
	}
	if budget.ID != newNameBudget.ID {
		t.Error("ID changed on update")
	}
}

func TestUpdateBudgetNonexistent(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	n, err := queries.UpdateBudget(ctx, data.UpdateBudgetParams{
		Name: "newName",
		ID:   1,
	})
	if err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Error("Updating a non existent budget should affect 0 rows")
	}
}

func TestDeleteBudget(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	budgets, err := queries.ListBudgets(ctx, data.ListBudgetsParams{LoginID: budget.LoginID})
	if err != nil {
		t.Fatal(err)
	}
	if len(budgets) != 1 {
		t.Fatal("There should be one budget before")
	}
	err = queries.DeleteBudget(ctx, data.DeleteBudgetParams{ID: budget.ID})
	if err != nil {
		t.Fatal(err)
	}
	budgets, err = queries.ListBudgets(ctx, data.ListBudgetsParams{LoginID: budget.LoginID})
	if err != nil {
		t.Fatal(err)
	}
	if len(budgets) != 0 {
		t.Fatal("There should be no budget after")
	}
}

func TestListBudgets(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	login := newLogin(t, queries, ctx, "user")
	newBudgetForLogin(t, queries, ctx, login.ID, "budgetB")
	newBudgetForLogin(t, queries, ctx, login.ID, "budgetA")
	budgets, err := queries.ListBudgets(ctx, data.ListBudgetsParams{LoginID: login.ID})
	if err != nil {
		t.Fatal(err)
	}
	if len(budgets) != 2 {
		t.Fatal("There should be two budgets")
	}
	if budgets[0].Name != "budgetB" {
		t.Error("First budget should be budgetB (ordered by id)")
	}
	if budgets[1].Name != "budgetA" {
		t.Error("Second budget should be budgetA (ordered by id)")
	}
}

func TestListBudgetsScopedToLogin(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	login1 := newLogin(t, queries, ctx, "user1")
	login2 := newLogin(t, queries, ctx, "user2")
	newBudgetForLogin(t, queries, ctx, login1.ID, "budget1")
	newBudgetForLogin(t, queries, ctx, login2.ID, "budget2")
	budgets, err := queries.ListBudgets(ctx, data.ListBudgetsParams{LoginID: login1.ID})
	if err != nil {
		t.Fatal(err)
	}
	if len(budgets) != 1 {
		t.Fatalf("There should be one budget for login1, got %d", len(budgets))
	}
	if budgets[0].Name != "budget1" {
		t.Error("login1 returned the wrong budget")
	}
}

func TestGetBudgetByName(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	nameBudget, err := queries.GetBudgetByName(ctx, data.GetBudgetByNameParams{LoginID: budget.LoginID, Name: "testBudget"})
	if err != nil {
		t.Fatal(err)
	}
	if budget.ID != nameBudget.ID {
		t.Error("getting budget by name, id does not match")
	}
	if nameBudget.Name != "testBudget" {
		t.Error("getting budget by name, name does not match")
	}
}

func TestGetBudgetByNameDoesNotExist(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	login := newLogin(t, queries, ctx, "user")
	_, err := queries.GetBudgetByName(ctx, data.GetBudgetByNameParams{LoginID: login.ID, Name: "testBudget"})
	if err == nil {
		t.Fatal("Getting non existent budget by name should fail")
	}
}

func TestListBudgetActivityMonthsEmpty(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	months, err := queries.ListBudgetActivityMonths(ctx, data.ListBudgetActivityMonthsParams{BudgetID: budget.ID})
	if err != nil {
		t.Fatal(err)
	}
	if len(months) != 0 {
		t.Fatal("A budget with no transactions or allocations should have no activity months")
	}
}

func TestListBudgetActivityMonthsFromTransactions(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	account := newAccount(t, queries, ctx, budget.ID, "testaccount")
	payee := newPayee(t, queries, ctx, budget.ID, "testpayee")
	newTrx(t, queries, ctx, account, payee, mustTime(t, 2026, 1, 15), 1000, 0, "")
	newTrx(t, queries, ctx, account, payee, mustTime(t, 2026, 3, 1), 2000, 0, "")
	newTrx(t, queries, ctx, account, payee, mustTime(t, 2026, 1, 31), 500, 0, "")
	months, err := queries.ListBudgetActivityMonths(ctx, data.ListBudgetActivityMonthsParams{BudgetID: budget.ID})
	if err != nil {
		t.Fatal(err)
	}
	if len(months) != 2 {
		t.Fatalf("There should be two distinct months, got %d", len(months))
	}
	if months[0].Unix() != mustTime(t, 2026, 1, 1).Unix() {
		t.Error("first month should be January")
	}
	if months[1].Unix() != mustTime(t, 2026, 3, 1).Unix() {
		t.Error("second month should be March")
	}
}

func TestListBudgetActivityMonthsFromAllocations(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	category := newCategory(t, queries, ctx, budget.ID, "testcategory")
	jan := mustTime(t, 2026, 1, 1)
	feb := mustTime(t, 2026, 2, 1)
	newAllocation(t, queries, ctx, budget.ID, category.ID, jan, 5000)
	newAllocation(t, queries, ctx, budget.ID, category.ID, feb, 8000)
	months, err := queries.ListBudgetActivityMonths(ctx, data.ListBudgetActivityMonthsParams{BudgetID: budget.ID})
	if err != nil {
		t.Fatal(err)
	}
	if len(months) != 2 {
		t.Fatalf("There should be two months, got %d", len(months))
	}
	if months[0].Unix() != jan.Unix() {
		t.Error("first month should be January")
	}
	if months[1].Unix() != feb.Unix() {
		t.Error("second month should be February")
	}
}

func TestListBudgetActivityMonthsUnionDedupeOrder(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	account := newAccount(t, queries, ctx, budget.ID, "testaccount")
	payee := newPayee(t, queries, ctx, budget.ID, "testpayee")
	category := newCategory(t, queries, ctx, budget.ID, "testcategory")
	jan := mustTime(t, 2026, 1, 1)
	feb := mustTime(t, 2026, 2, 1)
	mar := mustTime(t, 2026, 3, 1)
	newTrx(t, queries, ctx, account, payee, jan, 1000, 0, "")
	newTrx(t, queries, ctx, account, payee, feb, 2000, 0, "")
	newAllocation(t, queries, ctx, budget.ID, category.ID, feb, 5000)
	newAllocation(t, queries, ctx, budget.ID, category.ID, mar, 8000)
	months, err := queries.ListBudgetActivityMonths(ctx, data.ListBudgetActivityMonthsParams{BudgetID: budget.ID})
	if err != nil {
		t.Fatal(err)
	}
	if len(months) != 3 {
		t.Fatalf("There should be three distinct months, got %d", len(months))
	}
	if months[0].Unix() != jan.Unix() {
		t.Error("first month should be January")
	}
	if months[1].Unix() != feb.Unix() {
		t.Error("second month should be February")
	}
	if months[2].Unix() != mar.Unix() {
		t.Error("third month should be March")
	}
}

func TestListBudgetActivityMonthsExcludesOtherBudgets(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	otherBudget := newBudget(t, queries, ctx, "otherBudget")
	account := newAccount(t, queries, ctx, budget.ID, "testaccount")
	payee := newPayee(t, queries, ctx, budget.ID, "testpayee")
	otherAccount := newAccount(t, queries, ctx, otherBudget.ID, "otheraccount")
	otherPayee := newPayee(t, queries, ctx, otherBudget.ID, "otherpayee")
	otherCategory := newCategory(t, queries, ctx, otherBudget.ID, "othercategory")
	newTrx(t, queries, ctx, account, payee, mustTime(t, 2026, 1, 1), 1000, 0, "")
	newTrx(t, queries, ctx, otherAccount, otherPayee, mustTime(t, 2026, 2, 1), 1000, 0, "")
	newAllocation(t, queries, ctx, otherBudget.ID, otherCategory.ID, mustTime(t, 2026, 3, 1), 5000)
	months, err := queries.ListBudgetActivityMonths(ctx, data.ListBudgetActivityMonthsParams{BudgetID: budget.ID})
	if err != nil {
		t.Fatal(err)
	}
	if len(months) != 1 {
		t.Fatalf("There should be one month for this budget, got %d", len(months))
	}
	if months[0].Unix() != mustTime(t, 2026, 1, 1).Unix() {
		t.Error("month should be January")
	}
}

func TestListBudgetActivityMonthsNormalizesToStartOfMonth(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	account := newAccount(t, queries, ctx, budget.ID, "testaccount")
	payee := newPayee(t, queries, ctx, budget.ID, "testpayee")
	newTrx(t, queries, ctx, account, payee, mustTime(t, 2026, 2, 15), 1000, 0, "")
	months, err := queries.ListBudgetActivityMonths(ctx, data.ListBudgetActivityMonthsParams{BudgetID: budget.ID})
	if err != nil {
		t.Fatal(err)
	}
	if len(months) != 1 {
		t.Fatalf("There should be one month, got %d", len(months))
	}
	if months[0].Unix() != mustTime(t, 2026, 2, 1).Unix() {
		t.Error("A mid-month transaction should be normalized to the start of its month")
	}
}