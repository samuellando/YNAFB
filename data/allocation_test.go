package data_test

import (
	"database/sql"
	"testing"

	"samuellando.com/YNAFB/data"
)

func TestCreateAllocation(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	category := newCategory(t, queries, ctx, budget, "testcategory")
	month := mustTime(t, 2026, 1, 1)
	allocation, err := queries.CreateAllocation(ctx, data.CreateAllocationParams{
		BudgetID:   budget.ID,
		CategoryID: category.ID,
		Month:      month,
		Amount:     5000,
		LoginID:    budget.LoginID,
	})
	if err != nil {
		t.Fatal(err)
	}
	if allocation.BudgetID != budget.ID {
		t.Error("budget id does not match")
	}
	if allocation.CategoryID != category.ID {
		t.Error("category id does not match")
	}
	if allocation.Month.Unix() != month.Unix() {
		t.Error("month does not match")
	}
	if allocation.Amount != 5000 {
		t.Error("amount does not match")
	}
	if allocation.ID != 1 {
		t.Error("ID of the first allocation should be 1")
	}
}

func TestCreateAllocationNonExistingBudget(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	category := newCategory(t, queries, ctx, budget, "testcategory")
	_, err := queries.CreateAllocation(ctx, data.CreateAllocationParams{
		BudgetID:   99,
		CategoryID: category.ID,
		Month:      mustTime(t, 2026, 1, 1),
		Amount:     5000,
		LoginID:    budget.LoginID,
	})
	if err == nil {
		t.Error("Non existing budget should raise an error")
	}
}

func TestCreateAllocationNonExistingCategory(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	_, err := queries.CreateAllocation(ctx, data.CreateAllocationParams{
		BudgetID:   budget.ID,
		CategoryID: 99,
		Month:      mustTime(t, 2026, 1, 1),
		Amount:     5000,
		LoginID:    budget.LoginID,
	})
	if err == nil {
		t.Error("Non existing category should raise an error")
	}
}

func TestDeleteBudgetCascadesAllocations(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	category := newCategory(t, queries, ctx, budget, "testcategory")
	newAllocation(t, queries, ctx, budget.LoginID, budget.ID, category.ID, mustTime(t, 2026, 1, 1), 5000)
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM allocation`).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatal("There should be one allocation before")
	}
	err := queries.DeleteBudget(ctx, data.DeleteBudgetParams{LoginID: budget.LoginID, ID: budget.ID})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`SELECT COUNT(*) FROM allocation`).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatal("Deleting a budget should cascade delete its allocations")
	}
}

func TestDeleteCategoryCascadesAllocations(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	category := newCategory(t, queries, ctx, budget, "testcategory")
	newAllocation(t, queries, ctx, budget.LoginID, budget.ID, category.ID, mustTime(t, 2026, 1, 1), 5000)
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM allocation`).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatal("There should be one allocation before")
	}
	err := queries.DeleteCategory(ctx, data.DeleteCategoryParams{ID: category.ID, BudgetID: budget.ID, LoginID: budget.LoginID})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`SELECT COUNT(*) FROM allocation`).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatal("Deleting a category should cascade delete its allocations")
	}
}

func TestCreateAllocationDuplicateBudgetCategoryMonth(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	category := newCategory(t, queries, ctx, budget, "testcategory")
	params := data.CreateAllocationParams{
		BudgetID:   budget.ID,
		CategoryID: category.ID,
		Month:      mustTime(t, 2026, 1, 1),
		Amount:     5000,
		LoginID:    budget.LoginID,
	}
	if _, err := queries.CreateAllocation(ctx, params); err != nil {
		t.Fatal(err)
	}
	_, err := queries.CreateAllocation(ctx, params)
	if err == nil {
		t.Error("Duplicate allocation for the same budget, category and month should raise an error")
	}
}

func TestUpdateAllocation(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	category := newCategory(t, queries, ctx, budget, "testcategory")
	month := mustTime(t, 2026, 1, 1)
	allocation := newAllocation(t, queries, ctx, budget.LoginID, budget.ID, category.ID, month, 5000)
	updated, err := queries.UpdateAllocation(ctx, data.UpdateAllocationParams{
		Amount:     8000,
		BudgetID:   budget.ID,
		CategoryID: category.ID,
		Month:      month,
		LoginID:    budget.LoginID,
	})
	if err != nil {
		t.Fatal(err)
	}
	if updated.ID != allocation.ID {
		t.Error("ID changed on update")
	}
	if updated.Amount != 8000 {
		t.Error("allocation amount was not updated")
	}
}

func TestScopingAllocationCreateScopedByLogin(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budgetA := newBudget(t, queries, ctx, "budgetA")
	budgetB := newBudget(t, queries, ctx, "budgetB")
	categoryB := newCategory(t, queries, ctx, budgetB, "catB")

	_, err := queries.CreateAllocation(ctx, data.CreateAllocationParams{
		BudgetID:   budgetB.ID,
		CategoryID: categoryB.ID,
		Month:      mustTime(t, 2026, 1, 1),
		Amount:     100,
		LoginID:    budgetA.LoginID,
	})
	if err == nil {
		t.Fatal("expected creating an allocation for another login's budget to fail")
	}
}

func TestScopingUpdateAllocationScopedByLogin(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budgetA := newBudget(t, queries, ctx, "budgetA")
	budgetB := newBudget(t, queries, ctx, "budgetB")
	categoryB := newCategory(t, queries, ctx, budgetB, "catB")
	newAllocation(t, queries, ctx, budgetB.LoginID, budgetB.ID, categoryB.ID, mustTime(t, 2026, 1, 1), 100)

	_, err := queries.UpdateAllocation(ctx, data.UpdateAllocationParams{
		Amount:     500,
		CategoryID: categoryB.ID,
		Month:      mustTime(t, 2026, 1, 1),
		BudgetID:   budgetB.ID,
		LoginID:    budgetA.LoginID,
	})
	if err != sql.ErrNoRows {
		t.Fatalf("expected sql.ErrNoRows when updating across logins, got %v", err)
	}
}
