package data_test

import (
	"testing"

	"samuellando.com/YNAFB/data"
	"samuellando.com/YNAFB/internal/db/types"
)

func TestListGoalsValuesMonthlyNoAllocation(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	category := newCategory(t, queries, ctx, budget.ID, "testcategory")
	start := monthRelative(t, -2)
	newGoal(t, queries, ctx, budget.ID, category.ID, "monthly", start, types.NullUnixTime{}, 5000)
	month := monthRelative(t, 2)
	goals, err := queries.ListGoalsValues(ctx, data.ListGoalsValuesParams{
		BudgetID: budget.ID,
		Month:    month,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(goals) != 1 {
		t.Fatalf("expected one goal, got %d", len(goals))
	}
	g := goals[0]
	if g.Type != "monthly" {
		t.Errorf("expected type monthly, got %q", g.Type)
	}
	if g.CategoryID != category.ID {
		t.Errorf("expected category %d, got %d", category.ID, g.CategoryID)
	}
	if g.StartDate.Unix() != start.Unix() {
		t.Error("goal start does not match")
	}
	if g.EndDate.Valid {
		t.Error("goal end should be null")
	}
	if g.Amount != 5000 {
		t.Errorf("expected amount 5000, got %d", g.Amount)
	}
	if g.Allocated != 0 {
		t.Errorf("expected allocated 0, got %d", g.Allocated)
	}
	if g.AmountForMonth != 5000 {
		t.Errorf("expected amount for month 5000, got %d", g.AmountForMonth)
	}
	if g.Gap != -5000 {
		t.Errorf("expected gap -5000, got %d", g.Gap)
	}
}

func TestListGoalsValuesMonthlyFullyAllocated(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	category := newCategory(t, queries, ctx, budget.ID, "testcategory")
	start := monthRelative(t, -2)
	newGoal(t, queries, ctx, budget.ID, category.ID, "monthly", start, types.NullUnixTime{}, 5000)
	month := monthRelative(t, 2)
	newAllocation(t, queries, ctx, budget.ID, category.ID, month, 5000)
	goals, err := queries.ListGoalsValues(ctx, data.ListGoalsValuesParams{
		BudgetID: budget.ID,
		Month:    month,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(goals) != 1 {
		t.Fatalf("expected one goal, got %d", len(goals))
	}
	g := goals[0]
	if g.Allocated != 0 {
		t.Errorf("expected allocated 0 (excludes current month), got %d", g.Allocated)
	}
	if g.AmountForMonth != 5000 {
		t.Errorf("expected amount for month 5000, got %d", g.AmountForMonth)
	}
	if g.Gap != 0 {
		t.Errorf("expected gap 0, got %d", g.Gap)
	}
}

func TestListGoalsValuesRefillWithCarryOver(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	category := newCategory(t, queries, ctx, budget.ID, "testcategory")
	start := monthRelative(t, -3)
	newGoal(t, queries, ctx, budget.ID, category.ID, "refill", start, types.NullUnixTime{}, 10000)
	newAllocation(t, queries, ctx, budget.ID, category.ID, monthRelative(t, -2), 4000)
	month := monthRelative(t, -1)
	goals, err := queries.ListGoalsValues(ctx, data.ListGoalsValuesParams{
		BudgetID: budget.ID,
		Month:    month,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(goals) != 1 {
		t.Fatalf("expected one goal, got %d", len(goals))
	}
	g := goals[0]
	if g.Allocated != 4000 {
		t.Errorf("expected allocated 4000, got %d", g.Allocated)
	}
	if g.AmountForMonth != 6000 {
		t.Errorf("expected amount for month 6000, got %d", g.AmountForMonth)
	}
	if g.Gap != -6000 {
		t.Errorf("expected gap -6000, got %d", g.Gap)
	}
}

func TestListGoalsValuesRefillCarryOverFlooredAtZero(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	account := newAccount(t, queries, ctx, budget.ID, "testaccount")
	payee := newPayee(t, queries, ctx, budget.ID, "testpayee")
	category := newCategory(t, queries, ctx, budget.ID, "testcategory")
	start := monthRelative(t, -3)
	newGoal(t, queries, ctx, budget.ID, category.ID, "refill", start, types.NullUnixTime{}, 10000)
	jan := monthRelative(t, -2)
	newAllocation(t, queries, ctx, budget.ID, category.ID, jan, 3000)
	transaction := newTrx(t, queries, ctx, account, payee, jan, 5000, 0, "")
	newCategoryLine(t, queries, ctx, transaction, category, 5000, 0)
	month := monthRelative(t, -1)
	goals, err := queries.ListGoalsValues(ctx, data.ListGoalsValuesParams{
		BudgetID: budget.ID,
		Month:    month,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(goals) != 1 {
		t.Fatalf("expected one goal, got %d", len(goals))
	}
	g := goals[0]
	if g.AmountForMonth != 10000 {
		t.Errorf("expected amount for month 10000 with carry over floored at 0, got %d", g.AmountForMonth)
	}
	if g.Gap != -10000 {
		t.Errorf("expected gap -10000, got %d", g.Gap)
	}
}

func TestListGoalsValuesSave(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	category := newCategory(t, queries, ctx, budget.ID, "testcategory")
	start := monthRelative(t, -2)
	end := monthRelative(t, 2)
	newGoal(t, queries, ctx, budget.ID, category.ID, "save", start, types.NullUnixTime{Time: end.Time, Valid: true}, 12000)
	newAllocation(t, queries, ctx, budget.ID, category.ID, start, 6000)
	month := monthRelative(t, -1)
	goals, err := queries.ListGoalsValues(ctx, data.ListGoalsValuesParams{
		BudgetID: budget.ID,
		Month:    month,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(goals) != 1 {
		t.Fatalf("expected one goal, got %d", len(goals))
	}
	g := goals[0]
	if !g.EndDate.Valid || g.EndDate.Time.Unix() != end.Unix() {
		t.Error("goal end does not match")
	}
	if g.Allocated != 6000 {
		t.Errorf("expected allocated 6000, got %d", g.Allocated)
	}
	if g.AmountForMonth != 2000 {
		t.Errorf("expected amount for month 2000, got %d", g.AmountForMonth)
	}
	if g.Gap != -2000 {
		t.Errorf("expected gap -2000, got %d", g.Gap)
	}
}

func TestListGoalsValuesSaveGapZero(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	category := newCategory(t, queries, ctx, budget.ID, "testcategory")
	start := monthRelative(t, -2)
	end := monthRelative(t, 2)
	newGoal(t, queries, ctx, budget.ID, category.ID, "save", start, types.NullUnixTime{Time: end.Time, Valid: true}, 12000)
	newAllocation(t, queries, ctx, budget.ID, category.ID, start, 6000)
	month := monthRelative(t, -1)
	newAllocation(t, queries, ctx, budget.ID, category.ID, month, 2000)
	goals, err := queries.ListGoalsValues(ctx, data.ListGoalsValuesParams{
		BudgetID: budget.ID,
		Month:    month,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(goals) != 1 {
		t.Fatalf("expected one goal, got %d", len(goals))
	}
	g := goals[0]
	if g.Allocated != 6000 {
		t.Errorf("expected allocated 6000 (excludes current month), got %d", g.Allocated)
	}
	if g.AmountForMonth != 2000 {
		t.Errorf("expected amount for month 2000, got %d", g.AmountForMonth)
	}
	if g.Gap != 0 {
		t.Errorf("expected gap 0, got %d", g.Gap)
	}
}

func TestListGoalsValuesSaveFullyFunded(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	category := newCategory(t, queries, ctx, budget.ID, "testcategory")
	start := monthRelative(t, -2)
	end := monthRelative(t, 2)
	newGoal(t, queries, ctx, budget.ID, category.ID, "save", start, types.NullUnixTime{Time: end.Time, Valid: true}, 6000)
	newAllocation(t, queries, ctx, budget.ID, category.ID, start, 6000)
	month := monthRelative(t, -1)
	goals, err := queries.ListGoalsValues(ctx, data.ListGoalsValuesParams{
		BudgetID: budget.ID,
		Month:    month,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(goals) != 1 {
		t.Fatalf("expected one goal, got %d", len(goals))
	}
	g := goals[0]
	if g.AmountForMonth != 0 {
		t.Errorf("expected amount for month 0, got %d", g.AmountForMonth)
	}
	if g.Gap != 0 {
		t.Errorf("expected gap 0, got %d", g.Gap)
	}
}

func TestListGoalsValuesExcludesInactive(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	month := monthRelative(t, 0)
	notStarted := newCategory(t, queries, ctx, budget.ID, "notStarted")
	ended := newCategory(t, queries, ctx, budget.ID, "ended")
	endingThisMonth := newCategory(t, queries, ctx, budget.ID, "endingThisMonth")
	activeCategory := newCategory(t, queries, ctx, budget.ID, "activeCategory")
	newGoal(t, queries, ctx, budget.ID, notStarted.ID, "monthly", monthRelative(t, 2), types.NullUnixTime{}, 1000)
	newGoal(t, queries, ctx, budget.ID, ended.ID, "monthly", monthRelative(t, -3), types.NullUnixTime{Time: monthRelative(t, -1).Time, Valid: true}, 1000)
	newGoal(t, queries, ctx, budget.ID, endingThisMonth.ID, "monthly", monthRelative(t, -3), types.NullUnixTime{Time: month.Time, Valid: true}, 1000)
	active := newGoal(t, queries, ctx, budget.ID, activeCategory.ID, "monthly", monthRelative(t, -3), types.NullUnixTime{}, 1000)
	goals, err := queries.ListGoalsValues(ctx, data.ListGoalsValuesParams{
		BudgetID: budget.ID,
		Month:    month,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(goals) != 1 {
		t.Fatalf("expected one active goal, got %d", len(goals))
	}
	if goals[0].ID != active.ID {
		t.Error("the returned goal is not the active one")
	}
}

func TestListGoalsValuesScopedToBudget(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget1 := newBudget(t, queries, ctx, "testBudget1")
	budget2 := newBudget(t, queries, ctx, "testBudget2")
	category1 := newCategory(t, queries, ctx, budget1.ID, "shared")
	category2 := newCategory(t, queries, ctx, budget2.ID, "shared")
	start := monthRelative(t, -2)
	goal1 := newGoal(t, queries, ctx, budget1.ID, category1.ID, "monthly", start, types.NullUnixTime{}, 1000)
	newGoal(t, queries, ctx, budget2.ID, category2.ID, "monthly", start, types.NullUnixTime{}, 1000)
	goals, err := queries.ListGoalsValues(ctx, data.ListGoalsValuesParams{
		BudgetID: budget1.ID,
		Month:    monthRelative(t, 1),
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(goals) != 1 {
		t.Fatalf("expected one goal for budget1, got %d", len(goals))
	}
	if goals[0].ID != goal1.ID {
		t.Error("budget1 returned the wrong goal")
	}
}

func TestListGoalsValuesRefillOverfunded(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	category := newCategory(t, queries, ctx, budget.ID, "testcategory")
	start := monthRelative(t, -3)
	newGoal(t, queries, ctx, budget.ID, category.ID, "refill", start, types.NullUnixTime{}, 5000)
	newAllocation(t, queries, ctx, budget.ID, category.ID, monthRelative(t, -2), 8000)
	month := monthRelative(t, -1)
	goals, err := queries.ListGoalsValues(ctx, data.ListGoalsValuesParams{
		BudgetID: budget.ID,
		Month:    month,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(goals) != 1 {
		t.Fatalf("expected one goal, got %d", len(goals))
	}
	g := goals[0]
	if g.AmountForMonth != -3000 {
		t.Errorf("expected amount for month -3000, got %d", g.AmountForMonth)
	}
	if g.Gap != 3000 {
		t.Errorf("expected gap 3000, got %d", g.Gap)
	}
}