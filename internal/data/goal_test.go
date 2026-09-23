package data_test

import (
	"database/sql"
	"testing"
	"time"

	"samuellando.com/YNAFB/internal/data"
	"samuellando.com/YNAFB/internal/db/types"
)

func TestCreateGoal(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	category := newCategory(t, queries, ctx, budget, "testcategory")
	start := mustTime(t, 2026, 1, 1)
	goal, err := queries.CreateGoal(ctx, data.CreateGoalParams{
		LoginID:    budget.LoginID,
		BudgetID:   budget.ID,
		Type:       "monthly",
		StartDate:  start,
		EndDate:    types.NullUnixTime{},
		CategoryID: category.ID,
		Amount:     5000,
	})
	if err != nil {
		t.Fatal(err)
	}
	if goal.BudgetID != budget.ID {
		t.Error("budget id does not match")
	}
	if goal.Type != "monthly" {
		t.Error("goal type does not match")
	}
	if goal.StartDate.Unix() != start.Unix() {
		t.Error("goal start does not match")
	}
	if goal.EndDate.Valid {
		t.Error("goal end should be null")
	}
	if goal.CategoryID != category.ID {
		t.Error("goal category does not match")
	}
	if goal.Amount != 5000 {
		t.Error("goal amount does not match")
	}
	if goal.ID != 1 {
		t.Error("ID of the first goal should be 1")
	}
}

func TestCreateGoalInvalidType(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	category := newCategory(t, queries, ctx, budget, "testcategory")
	_, err := queries.CreateGoal(ctx, data.CreateGoalParams{
		LoginID:    budget.LoginID,
		BudgetID:   budget.ID,
		Type:       "bogus",
		StartDate:  mustTime(t, 2026, 1, 1),
		EndDate:    types.NullUnixTime{},
		CategoryID: category.ID,
		Amount:     5000,
	})
	if err == nil {
		t.Error("Unknown goal type should raise an error")
	}
}

func TestCreateGoalSaveRequiresEnd(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	category := newCategory(t, queries, ctx, budget, "testcategory")
	_, err := queries.CreateGoal(ctx, data.CreateGoalParams{
		LoginID:    budget.LoginID,
		BudgetID:   budget.ID,
		Type:       "save",
		StartDate:  mustTime(t, 2026, 1, 1),
		EndDate:    types.NullUnixTime{},
		CategoryID: category.ID,
		Amount:     5000,
	})
	if err == nil {
		t.Error("Save goals should require an end month")
	}
}

func TestCreateGoalEndBeforeStart(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	category := newCategory(t, queries, ctx, budget, "testcategory")
	_, err := queries.CreateGoal(ctx, data.CreateGoalParams{
		LoginID:    budget.LoginID,
		BudgetID:   budget.ID,
		Type:       "monthly",
		StartDate:  mustTime(t, 2026, 9, 1),
		EndDate:    types.NullUnixTime{Time: time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC), Valid: true},
		CategoryID: category.ID,
		Amount:     5000,
	})
	if err == nil {
		t.Error("End month before start month should raise an error")
	}
}

func TestCreateGoalDuplicateBudgetCategory(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	category := newCategory(t, queries, ctx, budget, "testcategory")
	params := data.CreateGoalParams{
		LoginID:    budget.LoginID,
		BudgetID:   budget.ID,
		Type:       "monthly",
		StartDate:  mustTime(t, 2026, 1, 1),
		EndDate:    types.NullUnixTime{},
		CategoryID: category.ID,
		Amount:     5000,
	}
	if _, err := queries.CreateGoal(ctx, params); err != nil {
		t.Fatal(err)
	}
	_, err := queries.CreateGoal(ctx, params)
	if err == nil {
		t.Error("Duplicate goal for the same budget and category should raise an error")
	}
}

func TestCreateGoalZeroAmountRejected(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	category := newCategory(t, queries, ctx, budget, "testcategory")
	_, err := queries.CreateGoal(ctx, data.CreateGoalParams{
		LoginID:    budget.LoginID,
		BudgetID:   budget.ID,
		Type:       "monthly",
		StartDate:  mustTime(t, 2026, 1, 1),
		EndDate:    types.NullUnixTime{},
		CategoryID: category.ID,
		Amount:     0,
	})
	if err == nil {
		t.Error("A goal with a zero amount should be rejected")
	}
}

func TestCreateGoalNegativeAmountRejected(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	category := newCategory(t, queries, ctx, budget, "testcategory")
	_, err := queries.CreateGoal(ctx, data.CreateGoalParams{
		LoginID:    budget.LoginID,
		BudgetID:   budget.ID,
		Type:       "monthly",
		StartDate:  mustTime(t, 2026, 1, 1),
		EndDate:    types.NullUnixTime{},
		CategoryID: category.ID,
		Amount:     -5000,
	})
	if err == nil {
		t.Error("A goal with a negative amount should be rejected")
	}
}

func TestCreateGoalNonExistingBudget(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	category := newCategory(t, queries, ctx, budget, "testcategory")
	_, err := queries.CreateGoal(ctx, data.CreateGoalParams{
		BudgetID:   99,
		Type:       "monthly",
		StartDate:  mustTime(t, 2026, 1, 1),
		EndDate:    types.NullUnixTime{},
		CategoryID: category.ID,
		Amount:     5000,
	})
	if err == nil {
		t.Error("Non existing budget should raise an error")
	}
}

func TestCreateGoalNonExistingCategory(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	_, err := queries.CreateGoal(ctx, data.CreateGoalParams{
		LoginID:    budget.LoginID,
		BudgetID:   budget.ID,
		Type:       "monthly",
		StartDate:  mustTime(t, 2026, 1, 1),
		EndDate:    types.NullUnixTime{},
		CategoryID: 99,
		Amount:     5000,
	})
	if err == nil {
		t.Error("Non existing category should raise an error")
	}
}

func TestDeleteBudgetCascadesGoals(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	category := newCategory(t, queries, ctx, budget, "testcategory")
	goal := newGoal(t, queries, ctx, budget, category.ID, "monthly", mustTime(t, 2026, 1, 1), types.NullUnixTime{}, 5000)
	goals, err := queries.ListGoals(ctx, data.ListGoalsParams{LoginID: budget.LoginID, BudgetID: budget.ID})
	if err != nil {
		t.Fatal(err)
	}
	if len(goals) != 1 {
		t.Fatal("There should be one goal before")
	}
	err = queries.DeleteBudget(ctx, data.DeleteBudgetParams{LoginID: budget.LoginID, ID: budget.ID})
	if err != nil {
		t.Fatal(err)
	}
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM goal WHERE id = ?`, goal.ID).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatal("Deleting a budget should cascade delete its goals")
	}
}

func TestDeleteCategoryCascadesGoals(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	category := newCategory(t, queries, ctx, budget, "testcategory")
	goal := newGoal(t, queries, ctx, budget, category.ID, "monthly", mustTime(t, 2026, 1, 1), types.NullUnixTime{}, 5000)
	goals, err := queries.ListGoals(ctx, data.ListGoalsParams{LoginID: budget.LoginID, BudgetID: budget.ID})
	if err != nil {
		t.Fatal(err)
	}
	if len(goals) != 1 {
		t.Fatal("There should be one goal before")
	}
	err = queries.DeleteCategory(ctx, data.DeleteCategoryParams{ID: category.ID, BudgetID: budget.ID, LoginID: budget.LoginID})
	if err != nil {
		t.Fatal(err)
	}
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM goal WHERE id = ?`, goal.ID).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatal("Deleting a category should cascade delete its goals")
	}
}

func TestUpdateGoalZeroAmountRejected(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	category := newCategory(t, queries, ctx, budget, "testcategory")
	newGoal(t, queries, ctx, budget, category.ID, "monthly", mustTime(t, 2026, 1, 1), types.NullUnixTime{}, 5000)
	_, err := queries.UpdateGoal(ctx, data.UpdateGoalParams{
		LoginID:    budget.LoginID,
		Type:       "monthly",
		StartDate:  mustTime(t, 2026, 1, 1),
		EndDate:    types.NullUnixTime{},
		Amount:     0,
		BudgetID:   budget.ID,
		CategoryID: category.ID,
	})
	if err == nil {
		t.Error("Updating a goal to a zero amount should be rejected")
	}
}

func TestUpdateGoal(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	category := newCategory(t, queries, ctx, budget, "testcategory")
	goal := newGoal(t, queries, ctx, budget, category.ID, "monthly", mustTime(t, 2026, 1, 1), types.NullUnixTime{}, 5000)
	start := mustTime(t, 2026, 3, 1)
	updated, err := queries.UpdateGoal(ctx, data.UpdateGoalParams{
		LoginID:    budget.LoginID,
		Type:       "refill",
		StartDate:  start,
		EndDate:    types.NullUnixTime{},
		Amount:     8000,
		BudgetID:   budget.ID,
		CategoryID: category.ID,
	})
	if err != nil {
		t.Fatal(err)
	}
	if updated.ID != goal.ID {
		t.Error("ID changed on update")
	}
	if updated.Type != "refill" {
		t.Error("goal type was not updated")
	}
	if updated.StartDate.Unix() != start.Unix() {
		t.Error("goal start was not updated")
	}
	if updated.Amount != 8000 {
		t.Error("goal amount was not updated")
	}
}

func TestDeleteGoal(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	category := newCategory(t, queries, ctx, budget, "testcategory")
	_ = newGoal(t, queries, ctx, budget, category.ID, "monthly", mustTime(t, 2026, 1, 1), types.NullUnixTime{}, 5000)
	goals, err := queries.ListGoals(ctx, data.ListGoalsParams{LoginID: budget.LoginID, BudgetID: budget.ID})
	if err != nil {
		t.Fatal(err)
	}
	if len(goals) != 1 {
		t.Fatal("There should be one goal before")
	}
	err = queries.DeleteGoal(ctx, data.DeleteGoalParams{CategoryID: category.ID, BudgetID: budget.ID, LoginID: budget.LoginID})
	if err != nil {
		t.Fatal(err)
	}
	goals, err = queries.ListGoals(ctx, data.ListGoalsParams{LoginID: budget.LoginID, BudgetID: budget.ID})
	if err != nil {
		t.Fatal(err)
	}
	if len(goals) != 0 {
		t.Fatal("There should be no goal after")
	}
}

func TestListGoals(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	category := newCategory(t, queries, ctx, budget, "testcategory")
	start := mustTime(t, 2026, 1, 1)
	goal := newGoal(t, queries, ctx, budget, category.ID, "save", start, types.NullUnixTime{Time: time.Date(2026, 12, 1, 0, 0, 0, 0, time.UTC), Valid: true}, 5000)
	goals, err := queries.ListGoals(ctx, data.ListGoalsParams{LoginID: budget.LoginID, BudgetID: budget.ID})
	if err != nil {
		t.Fatal(err)
	}
	if len(goals) != 1 {
		t.Fatal("There should be one goal")
	}
	if goals[0].ID != goal.ID {
		t.Error("goal id does not match")
	}
	if goals[0].Type != "save" {
		t.Error("goal type does not match")
	}
	if goals[0].StartDate.Unix() != start.Unix() {
		t.Error("goal start does not match")
	}
	if !goals[0].EndDate.Valid || goals[0].EndDate.Time.Unix() != time.Date(2026, 12, 1, 0, 0, 0, 0, time.UTC).Unix() {
		t.Error("goal end does not match")
	}
	if goals[0].CategoryID != category.ID {
		t.Error("goal category id does not match")
	}
	if goals[0].CategoryName != "testcategory" {
		t.Error("goal category name does not match")
	}
	if goals[0].Amount != 5000 {
		t.Error("goal amount does not match")
	}
}

func TestGetGoalByCategory(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	category := newCategory(t, queries, ctx, budget, "testcategory")
	goal := newGoal(t, queries, ctx, budget, category.ID, "monthly", mustTime(t, 2026, 1, 1), types.NullUnixTime{}, 5000)
	categoryGoal, err := queries.GetGoalByCategory(ctx, data.GetGoalByCategoryParams{
		LoginID:    budget.LoginID,
		BudgetID:   budget.ID,
		CategoryID: category.ID,
	})
	if err != nil {
		t.Fatal(err)
	}
	if goal.ID != categoryGoal.ID {
		t.Error("getting goal by category, id does not match")
	}
	if categoryGoal.CategoryID != category.ID {
		t.Error("getting goal by category, category id does not match")
	}
}

func TestGetGoalByCategoryDoesNotExist(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	category := newCategory(t, queries, ctx, budget, "testcategory")
	_, err := queries.GetGoalByCategory(ctx, data.GetGoalByCategoryParams{
		LoginID:    budget.LoginID,
		BudgetID:   budget.ID,
		CategoryID: category.ID,
	})
	if err == nil {
		t.Fatal("Getting non existent goal by category should fail")
	}
}

func TestScopingGoalScopedByLogin(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budgetA := newBudget(t, queries, ctx, "budgetA")
	budgetB := newBudget(t, queries, ctx, "budgetB")
	categoryB := newCategory(t, queries, ctx, budgetB, "catB")
	newGoal(t, queries, ctx, budgetB, categoryB.ID, "monthly", mustTime(t, 2026, 1, 1), types.NullUnixTime{}, 1000)

	rows, err := queries.ListGoals(ctx, data.ListGoalsParams{
		LoginID:  budgetA.LoginID,
		BudgetID: budgetB.ID,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 0 {
		t.Fatalf("expected 0 goals for another login's budget, got %d", len(rows))
	}
}

func TestScopingCreateGoalScopedByLogin(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budgetA := newBudget(t, queries, ctx, "budgetA")
	budgetB := newBudget(t, queries, ctx, "budgetB")
	categoryB := newCategory(t, queries, ctx, budgetB, "catB")

	_, err := queries.CreateGoal(ctx, data.CreateGoalParams{
		LoginID:    budgetA.LoginID,
		BudgetID:   budgetB.ID,
		Type:       "monthly",
		StartDate:  mustTime(t, 2026, 1, 1),
		EndDate:    types.NullUnixTime{},
		CategoryID: categoryB.ID,
		Amount:     1000,
	})
	if err == nil {
		t.Fatal("expected creating a goal for another login's budget to fail")
	}
}

func TestScopingDeleteGoalScopedByLogin(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budgetA := newBudget(t, queries, ctx, "budgetA")
	budgetB := newBudget(t, queries, ctx, "budgetB")
	categoryB := newCategory(t, queries, ctx, budgetB, "catB")
	_ = newGoal(t, queries, ctx, budgetB, categoryB.ID, "monthly", mustTime(t, 2026, 1, 1), types.NullUnixTime{}, 1000)

	err := queries.DeleteGoal(ctx, data.DeleteGoalParams{
		CategoryID: categoryB.ID,
		BudgetID:   budgetB.ID,
		LoginID:    budgetA.LoginID,
	})
	if err != nil {
		t.Fatal(err)
	}
	goals, err := queries.ListGoals(ctx, data.ListGoalsParams{LoginID: budgetB.LoginID, BudgetID: budgetB.ID})
	if err != nil {
		t.Fatal(err)
	}
	if len(goals) != 1 {
		t.Fatalf("expected goalB to survive a cross-login delete, got %d goals", len(goals))
	}
}

func TestScopingGetGoalByCategoryScopedByLogin(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budgetA := newBudget(t, queries, ctx, "budgetA")
	budgetB := newBudget(t, queries, ctx, "budgetB")
	categoryB := newCategory(t, queries, ctx, budgetB, "catB")
	newGoal(t, queries, ctx, budgetB, categoryB.ID, "monthly", mustTime(t, 2026, 1, 1), types.NullUnixTime{}, 1000)

	_, err := queries.GetGoalByCategory(ctx, data.GetGoalByCategoryParams{
		LoginID:    budgetA.LoginID,
		BudgetID:   budgetB.ID,
		CategoryID: categoryB.ID,
	})
	if err != sql.ErrNoRows {
		t.Fatalf("expected sql.ErrNoRows for another login's goal, got %v", err)
	}
}

func TestScopingUpdateGoalScopedByLogin(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budgetA := newBudget(t, queries, ctx, "budgetA")
	budgetB := newBudget(t, queries, ctx, "budgetB")
	categoryB := newCategory(t, queries, ctx, budgetB, "catB")
	newGoal(t, queries, ctx, budgetB, categoryB.ID, "monthly", mustTime(t, 2026, 1, 1), types.NullUnixTime{}, 1000)

	_, err := queries.UpdateGoal(ctx, data.UpdateGoalParams{
		Type:       "monthly",
		StartDate:  mustTime(t, 2026, 1, 1),
		EndDate:    types.NullUnixTime{},
		Amount:     2000,
		BudgetID:   budgetB.ID,
		LoginID:    budgetA.LoginID,
		CategoryID: categoryB.ID,
	})
	if err != sql.ErrNoRows {
		t.Fatalf("expected sql.ErrNoRows when updating across logins, got %v", err)
	}
}
