package data_test

import (
	"testing"
	"time"

	"samuellando.com/YNAFB/data"
	"samuellando.com/YNAFB/internal/db/types"
)

func TestCreateGoal(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	category := newCategory(t, queries, ctx, budget.ID, "testcategory")
	start := mustTime(t, 2026, 1, 1)
	goal, err := queries.CreateGoal(ctx, data.CreateGoalParams{
		Budget:   budget.ID,
		Type:     "monthly",
		Start:    start,
		End:      types.NullUnixTime{},
		Category: category.ID,
		Amount:   5000,
	})
	if err != nil {
		t.Fatal(err)
	}
	if goal.Budget != budget.ID {
		t.Error("budget id does not match")
	}
	if goal.Type != "monthly" {
		t.Error("goal type does not match")
	}
	if goal.Start.Unix() != start.Unix() {
		t.Error("goal start does not match")
	}
	if goal.End.Valid {
		t.Error("goal end should be null")
	}
	if goal.Category != category.ID {
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
	category := newCategory(t, queries, ctx, budget.ID, "testcategory")
	_, err := queries.CreateGoal(ctx, data.CreateGoalParams{
		Budget:   budget.ID,
		Type:     "bogus",
		Start:    mustTime(t, 2026, 1, 1),
		End:      types.NullUnixTime{},
		Category: category.ID,
		Amount:   5000,
	})
	if err == nil {
		t.Error("Unknown goal type should raise an error")
	}
}

func TestCreateGoalSaveRequiresEnd(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	category := newCategory(t, queries, ctx, budget.ID, "testcategory")
	_, err := queries.CreateGoal(ctx, data.CreateGoalParams{
		Budget:   budget.ID,
		Type:     "save",
		Start:    mustTime(t, 2026, 1, 1),
		End:      types.NullUnixTime{},
		Category: category.ID,
		Amount:   5000,
	})
	if err == nil {
		t.Error("Save goals should require an end month")
	}
}

func TestCreateGoalEndBeforeStart(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	category := newCategory(t, queries, ctx, budget.ID, "testcategory")
	_, err := queries.CreateGoal(ctx, data.CreateGoalParams{
		Budget:   budget.ID,
		Type:     "monthly",
		Start:    mustTime(t, 2026, 9, 1),
		End:      types.NullUnixTime{Time: time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC), Valid: true},
		Category: category.ID,
		Amount:   5000,
	})
	if err == nil {
		t.Error("End month before start month should raise an error")
	}
}

func TestCreateGoalDuplicateBudgetCategory(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	category := newCategory(t, queries, ctx, budget.ID, "testcategory")
	params := data.CreateGoalParams{
		Budget:   budget.ID,
		Type:     "monthly",
		Start:    mustTime(t, 2026, 1, 1),
		End:      types.NullUnixTime{},
		Category: category.ID,
		Amount:   5000,
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
	category := newCategory(t, queries, ctx, budget.ID, "testcategory")
	_, err := queries.CreateGoal(ctx, data.CreateGoalParams{
		Budget:   budget.ID,
		Type:     "monthly",
		Start:    mustTime(t, 2026, 1, 1),
		End:      types.NullUnixTime{},
		Category: category.ID,
		Amount:   0,
	})
	if err == nil {
		t.Error("A goal with a zero amount should be rejected")
	}
}

func TestCreateGoalNegativeAmountRejected(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	category := newCategory(t, queries, ctx, budget.ID, "testcategory")
	_, err := queries.CreateGoal(ctx, data.CreateGoalParams{
		Budget:   budget.ID,
		Type:     "monthly",
		Start:    mustTime(t, 2026, 1, 1),
		End:      types.NullUnixTime{},
		Category: category.ID,
		Amount:   -5000,
	})
	if err == nil {
		t.Error("A goal with a negative amount should be rejected")
	}
}

func TestCreateGoalNonExistingBudget(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	category := newCategory(t, queries, ctx, budget.ID, "testcategory")
	_, err := queries.CreateGoal(ctx, data.CreateGoalParams{
		Budget:   99,
		Type:     "monthly",
		Start:    mustTime(t, 2026, 1, 1),
		End:      types.NullUnixTime{},
		Category: category.ID,
		Amount:   5000,
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
		Budget:   budget.ID,
		Type:     "monthly",
		Start:    mustTime(t, 2026, 1, 1),
		End:      types.NullUnixTime{},
		Category: 99,
		Amount:   5000,
	})
	if err == nil {
		t.Error("Non existing category should raise an error")
	}
}

func TestDeleteBudgetCascadesGoals(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	category := newCategory(t, queries, ctx, budget.ID, "testcategory")
	goal, err := queries.CreateGoal(ctx, data.CreateGoalParams{
		Budget:   budget.ID,
		Type:     "monthly",
		Start:    mustTime(t, 2026, 1, 1),
		End:      types.NullUnixTime{},
		Category: category.ID,
		Amount:   5000,
	})
	if err != nil {
		t.Fatal(err)
	}
	goals, err := queries.ListGoals(ctx, budget.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(goals) != 1 {
		t.Fatal("There should be one goal before")
	}
	err = queries.DeleteBudget(ctx, budget.ID)
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
	category := newCategory(t, queries, ctx, budget.ID, "testcategory")
	goal, err := queries.CreateGoal(ctx, data.CreateGoalParams{
		Budget:   budget.ID,
		Type:     "monthly",
		Start:    mustTime(t, 2026, 1, 1),
		End:      types.NullUnixTime{},
		Category: category.ID,
		Amount:   5000,
	})
	if err != nil {
		t.Fatal(err)
	}
	goals, err := queries.ListGoals(ctx, budget.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(goals) != 1 {
		t.Fatal("There should be one goal before")
	}
	err = queries.DeleteCategory(ctx, category.ID)
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
	category := newCategory(t, queries, ctx, budget.ID, "testcategory")
	if _, err := queries.CreateGoal(ctx, data.CreateGoalParams{
		Budget:   budget.ID,
		Type:     "monthly",
		Start:    mustTime(t, 2026, 1, 1),
		End:      types.NullUnixTime{},
		Category: category.ID,
		Amount:   5000,
	}); err != nil {
		t.Fatal(err)
	}
	_, err := queries.UpdateGoal(ctx, data.UpdateGoalParams{
		Type:     "monthly",
		Start:    mustTime(t, 2026, 1, 1),
		End:      types.NullUnixTime{},
		Amount:   0,
		Budget:   budget.ID,
		Category: category.ID,
	})
	if err == nil {
		t.Error("Updating a goal to a zero amount should be rejected")
	}
}

func TestUpdateGoal(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	category := newCategory(t, queries, ctx, budget.ID, "testcategory")
	goal, err := queries.CreateGoal(ctx, data.CreateGoalParams{
		Budget:   budget.ID,
		Type:     "monthly",
		Start:    mustTime(t, 2026, 1, 1),
		End:      types.NullUnixTime{},
		Category: category.ID,
		Amount:   5000,
	})
	if err != nil {
		t.Fatal(err)
	}
	start := mustTime(t, 2026, 3, 1)
	n, err := queries.UpdateGoal(ctx, data.UpdateGoalParams{
		Type:     "refill",
		Start:    start,
		End:      types.NullUnixTime{},
		Amount:   8000,
		Budget:   budget.ID,
		Category: category.ID,
	})
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Error("The number of affected rows should be 1")
	}
	updated, err := queries.GetGoalByCategory(ctx, data.GetGoalByCategoryParams{
		Budget:   budget.ID,
		Category: category.ID,
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
	if updated.Start.Unix() != start.Unix() {
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
	category := newCategory(t, queries, ctx, budget.ID, "testcategory")
	goal, err := queries.CreateGoal(ctx, data.CreateGoalParams{
		Budget:   budget.ID,
		Type:     "monthly",
		Start:    mustTime(t, 2026, 1, 1),
		End:      types.NullUnixTime{},
		Category: category.ID,
		Amount:   5000,
	})
	if err != nil {
		t.Fatal(err)
	}
	goals, err := queries.ListGoals(ctx, budget.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(goals) != 1 {
		t.Fatal("There should be one goal before")
	}
	err = queries.DeleteGoal(ctx, goal.ID)
	if err != nil {
		t.Fatal(err)
	}
	goals, err = queries.ListGoals(ctx, budget.ID)
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
	category := newCategory(t, queries, ctx, budget.ID, "testcategory")
	start := mustTime(t, 2026, 1, 1)
	goal, err := queries.CreateGoal(ctx, data.CreateGoalParams{
		Budget:   budget.ID,
		Type:     "save",
		Start:    start,
		End:      types.NullUnixTime{Time: time.Date(2026, 12, 1, 0, 0, 0, 0, time.UTC), Valid: true},
		Category: category.ID,
		Amount:   5000,
	})
	if err != nil {
		t.Fatal(err)
	}
	goals, err := queries.ListGoals(ctx, budget.ID)
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
	if goals[0].Start.Unix() != start.Unix() {
		t.Error("goal start does not match")
	}
	if !goals[0].End.Valid || goals[0].End.Time.Unix() != time.Date(2026, 12, 1, 0, 0, 0, 0, time.UTC).Unix() {
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
	category := newCategory(t, queries, ctx, budget.ID, "testcategory")
	goal, err := queries.CreateGoal(ctx, data.CreateGoalParams{
		Budget:   budget.ID,
		Type:     "monthly",
		Start:    mustTime(t, 2026, 1, 1),
		End:      types.NullUnixTime{},
		Category: category.ID,
		Amount:   5000,
	})
	if err != nil {
		t.Fatal(err)
	}
	categoryGoal, err := queries.GetGoalByCategory(ctx, data.GetGoalByCategoryParams{
		Budget:   budget.ID,
		Category: category.ID,
	})
	if err != nil {
		t.Fatal(err)
	}
	if goal.ID != categoryGoal.ID {
		t.Error("getting goal by category, id does not match")
	}
	if categoryGoal.Category != category.ID {
		t.Error("getting goal by category, category id does not match")
	}
}

func TestGetGoalByCategoryDoesNotExist(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	category := newCategory(t, queries, ctx, budget.ID, "testcategory")
	_, err := queries.GetGoalByCategory(ctx, data.GetGoalByCategoryParams{
		Budget:   budget.ID,
		Category: category.ID,
	})
	if err == nil {
		t.Fatal("Getting non existent goal by category should fail")
	}
}
