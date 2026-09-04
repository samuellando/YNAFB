package data_test

import (
	"testing"

	"samuellando.com/YNAFB/data"
)

func TestCreateAllocation(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget, err := queries.CreateBudget(ctx, "testBudget")
	if err != nil {
		t.Fatal(err)
	}
	category, err := queries.CreateCategory(ctx, data.CreateCategoryParams{
		Budget: budget.ID,
		Name:   "testcategory",
	})
	if err != nil {
		t.Fatal(err)
	}
	month := mustTime(t, 2026, 1, 1)
	allocation, err := queries.CreateAllocation(ctx, data.CreateAllocationParams{
		Budget:   budget.ID,
		Category: category.ID,
		Month:    month,
		Amount:   5000,
	})
	if err != nil {
		t.Fatal(err)
	}
	if allocation.Budget != budget.ID {
		t.Error("budget id does not match")
	}
	if allocation.Category != category.ID {
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
	budget, err := queries.CreateBudget(ctx, "testBudget")
	if err != nil {
		t.Fatal(err)
	}
	category, err := queries.CreateCategory(ctx, data.CreateCategoryParams{
		Budget: budget.ID,
		Name:   "testcategory",
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = queries.CreateAllocation(ctx, data.CreateAllocationParams{
		Budget:   99,
		Category: category.ID,
		Month:    mustTime(t, 2026, 1, 1),
		Amount:   5000,
	})
	if err == nil {
		t.Error("Non existing budget should raise an error")
	}
}

func TestCreateAllocationNonExistingCategory(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget, err := queries.CreateBudget(ctx, "testBudget")
	if err != nil {
		t.Fatal(err)
	}
	_, err = queries.CreateAllocation(ctx, data.CreateAllocationParams{
		Budget:   budget.ID,
		Category: 99,
		Month:    mustTime(t, 2026, 1, 1),
		Amount:   5000,
	})
	if err == nil {
		t.Error("Non existing category should raise an error")
	}
}

func TestDeleteBudgetCascadesAllocations(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget, err := queries.CreateBudget(ctx, "testBudget")
	if err != nil {
		t.Fatal(err)
	}
	category, err := queries.CreateCategory(ctx, data.CreateCategoryParams{
		Budget: budget.ID,
		Name:   "testcategory",
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := queries.CreateAllocation(ctx, data.CreateAllocationParams{
		Budget:   budget.ID,
		Category: category.ID,
		Month:    mustTime(t, 2026, 1, 1),
		Amount:   5000,
	}); err != nil {
		t.Fatal(err)
	}
	allocations, err := queries.ListAllocations(ctx, budget.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(allocations) != 1 {
		t.Fatal("There should be one allocation before")
	}
	err = queries.DeleteBudget(ctx, budget.ID)
	if err != nil {
		t.Fatal(err)
	}
	var count int
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
	budget, err := queries.CreateBudget(ctx, "testBudget")
	if err != nil {
		t.Fatal(err)
	}
	category, err := queries.CreateCategory(ctx, data.CreateCategoryParams{
		Budget: budget.ID,
		Name:   "testcategory",
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := queries.CreateAllocation(ctx, data.CreateAllocationParams{
		Budget:   budget.ID,
		Category: category.ID,
		Month:    mustTime(t, 2026, 1, 1),
		Amount:   5000,
	}); err != nil {
		t.Fatal(err)
	}
	allocations, err := queries.ListAllocations(ctx, budget.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(allocations) != 1 {
		t.Fatal("There should be one allocation before")
	}
	err = queries.DeleteCategory(ctx, category.ID)
	if err != nil {
		t.Fatal(err)
	}
	var count int
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
	budget, err := queries.CreateBudget(ctx, "testBudget")
	if err != nil {
		t.Fatal(err)
	}
	category, err := queries.CreateCategory(ctx, data.CreateCategoryParams{
		Budget: budget.ID,
		Name:   "testcategory",
	})
	if err != nil {
		t.Fatal(err)
	}
	params := data.CreateAllocationParams{
		Budget:   budget.ID,
		Category: category.ID,
		Month:    mustTime(t, 2026, 1, 1),
		Amount:   5000,
	}
	if _, err := queries.CreateAllocation(ctx, params); err != nil {
		t.Fatal(err)
	}
	_, err = queries.CreateAllocation(ctx, params)
	if err == nil {
		t.Error("Duplicate allocation for the same budget, category and month should raise an error")
	}
}

func TestUpdateAllocation(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget, err := queries.CreateBudget(ctx, "testBudget")
	if err != nil {
		t.Fatal(err)
	}
	category, err := queries.CreateCategory(ctx, data.CreateCategoryParams{
		Budget: budget.ID,
		Name:   "testcategory",
	})
	if err != nil {
		t.Fatal(err)
	}
	month := mustTime(t, 2026, 1, 1)
	allocation, err := queries.CreateAllocation(ctx, data.CreateAllocationParams{
		Budget:   budget.ID,
		Category: category.ID,
		Month:    month,
		Amount:   5000,
	})
	if err != nil {
		t.Fatal(err)
	}
	n, err := queries.UpdateAllocation(ctx, data.UpdateAllocationParams{
		Amount:   8000,
		Budget:   budget.ID,
		Category: category.ID,
		Month:    month,
	})
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Error("The number of affected rows should be 1")
	}
	allocations, err := queries.ListAllocations(ctx, budget.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(allocations) != 1 {
		t.Fatal("There should be one allocation")
	}
	if allocations[0].ID != allocation.ID {
		t.Error("ID changed on update")
	}
	if allocations[0].Amount != 8000 {
		t.Error("allocation amount was not updated")
	}
}

func TestDeleteAllocationByCategoryAndMonth(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget, err := queries.CreateBudget(ctx, "testBudget")
	if err != nil {
		t.Fatal(err)
	}
	category, err := queries.CreateCategory(ctx, data.CreateCategoryParams{
		Budget: budget.ID,
		Name:   "testcategory",
	})
	if err != nil {
		t.Fatal(err)
	}
	month := mustTime(t, 2026, 1, 1)
	if _, err := queries.CreateAllocation(ctx, data.CreateAllocationParams{
		Budget:   budget.ID,
		Category: category.ID,
		Month:    month,
		Amount:   5000,
	}); err != nil {
		t.Fatal(err)
	}
	allocations, err := queries.ListAllocations(ctx, budget.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(allocations) != 1 {
		t.Fatal("There should be one allocation before")
	}
	err = queries.DeleteAllocationByCategoryAndMonth(ctx, data.DeleteAllocationByCategoryAndMonthParams{
		Budget:   budget.ID,
		Category: category.ID,
		Month:    month,
	})
	if err != nil {
		t.Fatal(err)
	}
	allocations, err = queries.ListAllocations(ctx, budget.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(allocations) != 0 {
		t.Fatal("There should be no allocation after")
	}
}

func TestListAllocations(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget, err := queries.CreateBudget(ctx, "testBudget")
	if err != nil {
		t.Fatal(err)
	}
	category, err := queries.CreateCategory(ctx, data.CreateCategoryParams{
		Budget: budget.ID,
		Name:   "testcategory",
	})
	if err != nil {
		t.Fatal(err)
	}
	jan := mustTime(t, 2026, 1, 1)
	feb := mustTime(t, 2026, 2, 1)
	allocation, err := queries.CreateAllocation(ctx, data.CreateAllocationParams{
		Budget:   budget.ID,
		Category: category.ID,
		Month:    jan,
		Amount:   5000,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := queries.CreateAllocation(ctx, data.CreateAllocationParams{
		Budget:   budget.ID,
		Category: category.ID,
		Month:    feb,
		Amount:   8000,
	}); err != nil {
		t.Fatal(err)
	}
	allocations, err := queries.ListAllocations(ctx, budget.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(allocations) != 2 {
		t.Fatal("There should be two allocations")
	}
	if allocations[0].ID != allocation.ID {
		t.Error("First allocation should be the January one (ordered by month)")
	}
	if allocations[0].Month.Unix() != jan.Unix() {
		t.Error("first allocation month does not match")
	}
	if allocations[0].CategoryID != category.ID {
		t.Error("first allocation category id does not match")
	}
	if allocations[0].CategoryName != "testcategory" {
		t.Error("first allocation category name does not match")
	}
	if allocations[0].Amount != 5000 {
		t.Error("first allocation amount does not match")
	}
	if allocations[1].Amount != 8000 {
		t.Error("second allocation amount does not match")
	}
}
