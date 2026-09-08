package data_test

import (
	"database/sql"
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
	updated, err := queries.UpdateBudget(ctx, data.UpdateBudgetParams{
		Name:    "newName",
		LoginID: budget.LoginID,
		ID:      budget.ID,
	})
	if err != nil {
		t.Fatal(err)
	}
	if updated.ID != budget.ID {
		t.Error("ID changed on update")
	}
	if updated.Name != "newName" {
		t.Error("budget name was not updated")
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
	_, err := queries.UpdateBudget(ctx, data.UpdateBudgetParams{
		Name:    "newName",
		LoginID: 1,
		ID:      1,
	})
	if err != sql.ErrNoRows {
		t.Fatalf("Updating a non existent budget should return sql.ErrNoRows, got %v", err)
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
	err = queries.DeleteBudget(ctx, data.DeleteBudgetParams{LoginID: budget.LoginID, ID: budget.ID})
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

func TestScopingGetBudgetByNameScopedByLogin(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budgetA := newBudget(t, queries, ctx, "budgetA")
	budgetB := newBudget(t, queries, ctx, "budgetB")

	_, err := queries.GetBudgetByName(ctx, data.GetBudgetByNameParams{
		LoginID: budgetA.LoginID,
		Name:    budgetB.Name,
	})
	if err != sql.ErrNoRows {
		t.Fatalf("expected sql.ErrNoRows for another login's budget, got %v", err)
	}
}

func TestScopingUpdateBudgetScopedByLogin(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budgetA := newBudget(t, queries, ctx, "budgetA")
	budgetB := newBudget(t, queries, ctx, "budgetB")

	_, err := queries.UpdateBudget(ctx, data.UpdateBudgetParams{
		Name:    "hacked",
		LoginID: budgetA.LoginID,
		ID:      budgetB.ID,
	})
	if err != sql.ErrNoRows {
		t.Fatalf("expected sql.ErrNoRows when updating across logins, got %v", err)
	}
}

func TestScopingDeleteBudgetScopedByLogin(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budgetA := newBudget(t, queries, ctx, "budgetA")
	budgetB := newBudget(t, queries, ctx, "budgetB")

	err := queries.DeleteBudget(ctx, data.DeleteBudgetParams{
		LoginID: budgetA.LoginID,
		ID:      budgetB.ID,
	})
	if err != nil {
		t.Fatal(err)
	}
	budgets, err := queries.ListBudgets(ctx, data.ListBudgetsParams{LoginID: budgetB.LoginID})
	if err != nil {
		t.Fatal(err)
	}
	if len(budgets) != 1 {
		t.Fatalf("expected budgetB to survive a cross-login delete, got %d budgets", len(budgets))
	}
}
