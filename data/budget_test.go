package data_test

import (
	"testing"

	"samuellando.com/YNAFB/data"
)

func TestCreateBudget(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	if budget.Name != "testBudget" {
		t.Error("budget name does not match")
	}
	if budget.ID != 1 {
		t.Error("ID of the first budget should be 1")
	}
}

func TestCreateBudgetEmptyName(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	_, err := queries.CreateBudget(ctx, "")
	if err == nil {
		t.Error("Empty budget name should raise an error")
	}
}

func TestCreateBudgetDuplicateName(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	newBudget(t, queries, ctx, "testBudget")
	_, err := queries.CreateBudget(ctx, "testBudget")
	if err == nil {
		t.Error("Duplicate budget name should raise an error")
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
	newNameBudget, err := queries.GetBudgetByName(ctx, "newName")
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
	budgets, err := queries.ListBudgets(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(budgets) != 1 {
		t.Fatal("There should be one budget before")
	}
	err = queries.DeleteBudget(ctx, budget.ID)
	if err != nil {
		t.Fatal(err)
	}
	budgets, err = queries.ListBudgets(ctx)
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
	newBudget(t, queries, ctx, "budgetB")
	newBudget(t, queries, ctx, "budgetA")
	budgets, err := queries.ListBudgets(ctx)
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

func TestGetBudgetByName(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	nameBudget, err := queries.GetBudgetByName(ctx, "testBudget")
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
	_, err := queries.GetBudgetByName(ctx, "testBudget")
	if err == nil {
		t.Fatal("Getting non existent budget by name should fail")
	}
}
