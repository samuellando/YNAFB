package data_test

import (
	"database/sql"
	"testing"

	"samuellando.com/YNAFB/internal/data"
)

func TestSetAllocation(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	category := newCategory(t, queries, ctx, budget, "testcategory")
	month := mustTime(t, 2026, 1, 1)
	allocation, err := queries.SetAllocation(ctx, data.SetAllocationParams{
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

func TestSetAllocationNonExistingBudget(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	category := newCategory(t, queries, ctx, budget, "testcategory")
	_, err := queries.SetAllocation(ctx, data.SetAllocationParams{
		BudgetID:   99,
		CategoryID: category.ID,
		Month:      mustTime(t, 2026, 1, 1),
		Amount:     5000,
		LoginID:    budget.LoginID,
	})
	if err != sql.ErrNoRows {
		t.Errorf("Non existing budget should return sql.ErrNoRows, got %v", err)
	}
}

func TestSetAllocationNonExistingCategory(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	_, err := queries.SetAllocation(ctx, data.SetAllocationParams{
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

func TestSetAllocationReplacesDuplicateBudgetCategoryMonth(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	category := newCategory(t, queries, ctx, budget, "testcategory")
	params := data.SetAllocationParams{
		BudgetID:   budget.ID,
		CategoryID: category.ID,
		Month:      mustTime(t, 2026, 1, 1),
		Amount:     5000,
		LoginID:    budget.LoginID,
	}
	first, err := queries.SetAllocation(ctx, params)
	if err != nil {
		t.Fatal(err)
	}
	second, err := queries.SetAllocation(ctx, params)
	if err != nil {
		t.Fatalf("Setting the same budget, category and month again should replace instead of erroring: %v", err)
	}
	if second.ID == first.ID {
		t.Error("Replacing an allocation should assign a new ID")
	}
	if second.Amount != first.Amount {
		t.Error("replaced allocation amount does not match")
	}
}

func TestSetAllocationUpdatesAmount(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	category := newCategory(t, queries, ctx, budget, "testcategory")
	month := mustTime(t, 2026, 1, 1)
	allocation := newAllocation(t, queries, ctx, budget.LoginID, budget.ID, category.ID, month, 5000)
	updated, err := queries.SetAllocation(ctx, data.SetAllocationParams{
		Amount:     8000,
		BudgetID:   budget.ID,
		CategoryID: category.ID,
		Month:      month,
		LoginID:    budget.LoginID,
	})
	if err != nil {
		t.Fatal(err)
	}
	if updated.ID == allocation.ID {
		t.Error("replacing an allocation should assign a new ID")
	}
	if updated.Amount != 8000 {
		t.Error("allocation amount was not updated")
	}
}

func TestScopingAllocationSetScopedByLogin(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budgetA := newBudget(t, queries, ctx, "budgetA")
	budgetB := newBudget(t, queries, ctx, "budgetB")
	categoryB := newCategory(t, queries, ctx, budgetB, "catB")

	_, err := queries.SetAllocation(ctx, data.SetAllocationParams{
		BudgetID:   budgetB.ID,
		CategoryID: categoryB.ID,
		Month:      mustTime(t, 2026, 1, 1),
		Amount:     100,
		LoginID:    budgetA.LoginID,
	})
	if err != sql.ErrNoRows {
		t.Fatalf("expected sql.ErrNoRows when setting an allocation for another login's budget, got %v", err)
	}
}

func TestListAllocations(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	groupID := newCategoryGroup(t, queries, ctx, budget, "testgroup")
	grouped, err := queries.CreateCategory(ctx, data.CreateCategoryParams{
		LoginID:         budget.LoginID,
		BudgetID:        budget.ID,
		Name:            "grouped",
		CategoryGroupID: sql.NullInt64{Int64: groupID, Valid: true},
	})
	if err != nil {
		t.Fatal(err)
	}
	ungrouped := newCategory(t, queries, ctx, budget, "ungrouped")
	month := mustTime(t, 2026, 1, 1)
	newAllocation(t, queries, ctx, budget.LoginID, budget.ID, grouped.ID, month, 5000)
	newAllocation(t, queries, ctx, budget.LoginID, budget.ID, ungrouped.ID, month, 3000)

	rows, err := queries.ListAllocations(ctx, data.ListAllocationsParams{LoginID: budget.LoginID, BudgetID: budget.ID})
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 2 {
		t.Fatalf("There should be two allocations, got %d", len(rows))
	}
	byName := map[string]data.ListAllocationsRow{}
	for _, row := range rows {
		byName[row.CategoryName] = row
	}
	groupedRow := byName["grouped"]
	if groupedRow.Amount != 5000 {
		t.Error("grouped allocation amount does not match")
	}
	if groupedRow.CategoryID != grouped.ID {
		t.Error("grouped allocation category id does not match")
	}
	if !groupedRow.CategoryGroupID.Valid || groupedRow.CategoryGroupID.Int64 != groupID {
		t.Error("grouped allocation category group id does not match")
	}
	if groupedRow.CategoryGroupName.String != "testgroup" {
		t.Error("grouped allocation category group name does not match")
	}
	ungroupedRow := byName["ungrouped"]
	if ungroupedRow.Amount != 3000 {
		t.Error("ungrouped allocation amount does not match")
	}
	if ungroupedRow.CategoryGroupID.Valid {
		t.Error("ungrouped allocation category group should be null")
	}
}

func TestListAllocationsScopedToBudget(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget1 := newBudget(t, queries, ctx, "budget1")
	budget2 := newBudget(t, queries, ctx, "budget2")
	category1 := newCategory(t, queries, ctx, budget1, "cat")
	category2 := newCategory(t, queries, ctx, budget2, "cat")
	month := mustTime(t, 2026, 1, 1)
	newAllocation(t, queries, ctx, budget1.LoginID, budget1.ID, category1.ID, month, 1000)
	newAllocation(t, queries, ctx, budget2.LoginID, budget2.ID, category2.ID, month, 2000)

	rows, err := queries.ListAllocations(ctx, data.ListAllocationsParams{LoginID: budget1.LoginID, BudgetID: budget1.ID})
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 {
		t.Fatalf("There should be one allocation for budget1, got %d", len(rows))
	}
	if rows[0].Amount != 1000 {
		t.Error("budget1 returned the wrong allocation")
	}
}

func TestScopingListAllocationsScopedByLogin(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budgetA := newBudget(t, queries, ctx, "budgetA")
	budgetB := newBudget(t, queries, ctx, "budgetB")
	categoryB := newCategory(t, queries, ctx, budgetB, "catB")
	newAllocation(t, queries, ctx, budgetB.LoginID, budgetB.ID, categoryB.ID, mustTime(t, 2026, 1, 1), 1000)

	rows, err := queries.ListAllocations(ctx, data.ListAllocationsParams{
		LoginID:  budgetA.LoginID,
		BudgetID: budgetB.ID,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 0 {
		t.Fatalf("expected 0 allocations for another login's budget, got %d", len(rows))
	}
}
