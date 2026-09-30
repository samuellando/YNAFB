package api_test

import (
	"net/http"
	"strconv"
	"testing"

	"samuellando.com/YNAFB/internal/domain"
	"samuellando.com/YNAFB/internal/http/api"
)

func assertGoal(t *testing.T, m api.BudgetMonth, name string, expected api.BudgetMonthGoal) {
	t.Helper()
	cat := findBudgetCat(t, m, name)
	if cat.Goal == nil {
		t.Fatalf("expected goal on category %q, got nil", name)
	}
	got := *cat.Goal
	if got.Type != expected.Type {
		t.Fatalf("category %q: expect goal type %q, got %q", name, expected.Type, got.Type)
	}
	if got.Amount != expected.Amount {
		t.Fatalf("category %q: expect goal amount %d, got %d", name, expected.Amount, got.Amount)
	}
	if got.Allocated != expected.Allocated {
		t.Fatalf("category %q: expect goal allocated %d, got %d", name, expected.Allocated, got.Allocated)
	}
	if got.AmountForMonth != expected.AmountForMonth {
		t.Fatalf("category %q: expect goal amountForMonth %d, got %d", name, expected.AmountForMonth, got.AmountForMonth)
	}
	if got.Gap != expected.Gap {
		t.Fatalf("category %q: expect goal gap %d, got %d", name, expected.Gap, got.Gap)
	}
}

func assertNoGoal(t *testing.T, m api.BudgetMonth, name string) {
	t.Helper()
	cat := findBudgetCat(t, m, name)
	if cat.Goal != nil {
		t.Fatalf("expected no goal on category %q, got %+v", name, *cat.Goal)
	}
}

func fetchBudgetMonth(t *testing.T, ts *testServer, budgetID int, month string) api.BudgetMonth {
	t.Helper()
	w := ts.doReq(t, "GET", "/api/v1/budget/"+strconv.Itoa(budgetID)+"/"+month, nil, http.StatusOK)
	var m api.BudgetMonth
	decodeJSON(t, w, &m)
	return m
}

func assertBudgetMonthSummary(t *testing.T, m api.BudgetMonth, expected api.BudgetMonthSummary) {
	t.Helper()
	if m.Summary != expected {
		t.Fatalf("expect %+v\ngot %+v", expected, m.Summary)
	}
}

func findBudgetCat(t *testing.T, m api.BudgetMonth, name string) api.BudgetMonthCategory {
	t.Helper()
	for _, c := range m.Categories {
		if c.CategoryName == name {
			return c
		}
	}
	t.Fatalf("category %q not found in %+v", name, m.Categories)
	return api.BudgetMonthCategory{}
}

func assertBudgetMonthCategory(t *testing.T, m api.BudgetMonth, name string, expected api.BudgetMonthCategory) {
	t.Helper()
	cat := findBudgetCat(t, m, name)
	if cat.Allocated != expected.Allocated {
		t.Fatalf("expect allocated %d\ngot %d", expected.Allocated, cat.Allocated)
	}
	if cat.Spent != expected.Spent {
		t.Fatalf("expect spent %d\ngot %d", expected.Spent, cat.Spent)
	}
	if cat.Available != expected.Available {
		t.Fatalf("expect available %d\ngot %d", expected.Available, cat.Available)
	}
}

func calcSetup(t *testing.T) (*testServer, *domain.Budget, *domain.Account, *domain.Payee) {
	ts := setupTestServer(t)
	budget, err := ts.domain.CreateBudget(ts.ctx, int(ts.loginID), "Cacl Budget")
	if err != nil {
		t.Fatal(err)
	}
	account, err := budget.CreateAccount(ts.ctx, "Chequing")
	if err != nil {
		t.Fatal(err)
	}
	payee, err := budget.CreatePayee(ts.ctx, "Payee")
	if err != nil {
		t.Fatal(err)
	}
	return ts, budget, account, payee
}

// Values should all be zero when there is no zero
func TestCalcEmptyBudget(t *testing.T) {
	ts, budget, _, _ := calcSetup(t)
	budget.CreateCategory(ts.ctx, "Groceries", nil)

	m := fetchBudgetMonth(t, ts, budget.ID(), sRelMonth(0))
	assertBudgetMonthSummary(t, m, api.BudgetMonthSummary{
		Allocated:     0,
		Income:        0,
		ReadyToAssign: 0,
		Spent:         0,
		Uncategorized: 0,
	})
	assertBudgetMonthCategory(t, m, "Groceries", api.BudgetMonthCategory{
		Allocated: 0,
		Spent:     0,
		Available: 0,
	})
}

// Uncategorized transactions should count as absolute values
func TestCalcUncategorized(t *testing.T) {
	ts, budget, account, payee := calcSetup(t)
	account.CreateTransaction(ts.ctx, payee, relMonth(0), 0, 100, "")
	account.CreateTransaction(ts.ctx, payee, relMonth(0), 123, 0, "")

	m := fetchBudgetMonth(t, ts, budget.ID(), sRelMonth(0))
	assertBudgetMonthSummary(t, m, api.BudgetMonthSummary{
		Allocated:     0,
		Income:        0,
		ReadyToAssign: 0,
		Spent:         0,
		Uncategorized: 223,
	})
}

// Test that categories are sorted by group and alphabetically
func TestCalcCatgorySorting(t *testing.T) {
	ts, budget, _, _ := calcSetup(t)
	groupB, err := budget.CreateCategoryGroup(ts.ctx, "B Group")
	if err != nil {
		t.Fatal(err)
	}
	groupA, err := budget.CreateCategoryGroup(ts.ctx, "A Group")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := budget.CreateCategory(ts.ctx, "Zebra", nil); err != nil {
		t.Fatal(err)
	}
	if _, err := budget.CreateCategory(ts.ctx, "Banana", groupA); err != nil {
		t.Fatal(err)
	}
	if _, err := budget.CreateCategory(ts.ctx, "Mango", groupB); err != nil {
		t.Fatal(err)
	}
	if _, err := budget.CreateCategory(ts.ctx, "Apple", groupB); err != nil {
		t.Fatal(err)
	}

	m := fetchBudgetMonth(t, ts, budget.ID(), sRelMonth(0))
	wantOrder := []string{"Zebra", "Banana", "Apple", "Mango"}
	if len(m.Categories) != len(wantOrder) {
		t.Fatalf("expect %d categories, got %+v", len(wantOrder), m.Categories)
	}
	for i, want := range wantOrder {
		if m.Categories[i].CategoryName != want {
			names := make([]string, len(m.Categories))
			for j, c := range m.Categories {
				names[j] = c.CategoryName
			}
			t.Fatalf("expect order %v, got %v", wantOrder, names)
		}
	}
	// Nil group sorts first with no group name; grouped cats carry theirs.
	if findBudgetCat(t, m, "Zebra").CategoryGroupName != nil {
		t.Fatalf("expect nil group name for ungrouped category")
	}
	if got := findBudgetCat(t, m, "Banana").CategoryGroupName; got == nil || *got != "A Group" {
		t.Fatalf("expect group %q, got %+v", "A Group", got)
	}
	if got := findBudgetCat(t, m, "Apple").CategoryGroupName; got == nil || *got != "B Group" {
		t.Fatalf("expect group %q, got %+v", "B Group", got)
	}
}

// Income in a month should display. it should get added to ready to assign for the month
// And future months.
func TestCalcIncomeInMonth(t *testing.T) {
	ts, budget, account, payee := calcSetup(t)
	budget.CreateCategory(ts.ctx, "Groceries", nil)
	trx, _ := account.CreateTransaction(ts.ctx, payee, relMonth(-1), 0, 100, "")
	trx.AddLine(ts.ctx, nil, nil, true, 0, 100)

	// Before income
	m := fetchBudgetMonth(t, ts, budget.ID(), sRelMonth(-2))
	assertBudgetMonthSummary(t, m, api.BudgetMonthSummary{
		Income:        0,
		ReadyToAssign: 0,
	})
	// income month
	m = fetchBudgetMonth(t, ts, budget.ID(), sRelMonth(-1))
	assertBudgetMonthSummary(t, m, api.BudgetMonthSummary{
		Income:        100,
		ReadyToAssign: 100,
	})
	// current month
	m = fetchBudgetMonth(t, ts, budget.ID(), sRelMonth(0))
	assertBudgetMonthSummary(t, m, api.BudgetMonthSummary{
		Income:        0,
		ReadyToAssign: 100,
	})
	// future month
	m = fetchBudgetMonth(t, ts, budget.ID(), sRelMonth(1))
	assertBudgetMonthSummary(t, m, api.BudgetMonthSummary{
		Income:        0,
		ReadyToAssign: 100,
	})
}

// Spending should show up in spending and available
// It should substract from ready to assign for future months, but not current
func TestCalcSpentInMonth(t *testing.T) {
	ts, budget, account, payee := calcSetup(t)
	cat, err := budget.CreateCategory(ts.ctx, "Groceries", nil)
	if err != nil {
		t.Fatal(err)
	}
	trx, _ := account.CreateTransaction(ts.ctx, payee, relMonth(-1), 100, 0, "")
	trx.AddLine(ts.ctx, nil, cat, false, 100, 0)

	// before spending
	m := fetchBudgetMonth(t, ts, budget.ID(), sRelMonth(-2))
	assertBudgetMonthSummary(t, m, api.BudgetMonthSummary{
		Spent:         0,
		ReadyToAssign: 0,
	})
	assertBudgetMonthCategory(t, m, "Groceries", api.BudgetMonthCategory{
		Spent:     0,
		Available: 0,
	})
	// month with spending
	m = fetchBudgetMonth(t, ts, budget.ID(), sRelMonth(-1))
	assertBudgetMonthSummary(t, m, api.BudgetMonthSummary{
		Spent:         100,
		ReadyToAssign: 0,
	})
	assertBudgetMonthCategory(t, m, "Groceries", api.BudgetMonthCategory{
		Spent:     100,
		Available: -100,
	})
	// Current month
	m = fetchBudgetMonth(t, ts, budget.ID(), sRelMonth(0))
	assertBudgetMonthSummary(t, m, api.BudgetMonthSummary{
		Spent:         0,
		ReadyToAssign: -100,
	})
	assertBudgetMonthCategory(t, m, "Groceries", api.BudgetMonthCategory{
		Spent:     0,
		Available: 0,
	})
	// future month
	m = fetchBudgetMonth(t, ts, budget.ID(), sRelMonth(1))
	assertBudgetMonthSummary(t, m, api.BudgetMonthSummary{
		Spent:         0,
		ReadyToAssign: -100,
	})
	assertBudgetMonthCategory(t, m, "Groceries", api.BudgetMonthCategory{
		Spent:     0,
		Available: 0,
	})
}

// Allocations should substract from the ready to assign in month and all future months
// It should show in available for the month and future months up to the current month
func TestCalcAllocatedInMonth(t *testing.T) {
	ts, budget, _, _ := calcSetup(t)
	cat, err := budget.CreateCategory(ts.ctx, "Groceries", nil)
	if err != nil {
		t.Fatal(err)
	}
	cat.SetAllocation(ts.ctx, relMonth(-1), 100)

	// before allocation
	m := fetchBudgetMonth(t, ts, budget.ID(), sRelMonth(-2))
	assertBudgetMonthSummary(t, m, api.BudgetMonthSummary{
		Allocated:     0,
		ReadyToAssign: 0,
	})
	assertBudgetMonthCategory(t, m, "Groceries", api.BudgetMonthCategory{
		Allocated: 0,
		Available: 0,
	})
	// month with allocation
	m = fetchBudgetMonth(t, ts, budget.ID(), sRelMonth(-1))
	assertBudgetMonthSummary(t, m, api.BudgetMonthSummary{
		Allocated:     100,
		ReadyToAssign: -100,
	})
	assertBudgetMonthCategory(t, m, "Groceries", api.BudgetMonthCategory{
		Allocated: 100,
		Available: 100,
	})
	// Current month
	m = fetchBudgetMonth(t, ts, budget.ID(), sRelMonth(0))
	assertBudgetMonthSummary(t, m, api.BudgetMonthSummary{
		Allocated:     0,
		ReadyToAssign: -100,
	})
	assertBudgetMonthCategory(t, m, "Groceries", api.BudgetMonthCategory{
		Allocated: 0,
		Available: 100,
	})
	// future month
	m = fetchBudgetMonth(t, ts, budget.ID(), sRelMonth(1))
	assertBudgetMonthSummary(t, m, api.BudgetMonthSummary{
		Allocated:     0,
		ReadyToAssign: -100,
	})
	assertBudgetMonthCategory(t, m, "Groceries", api.BudgetMonthCategory{
		Allocated: 0,
		Available: 0,
	})
}

// Positive avaliable amounts should carry forward, up to current month
func TestCalcAvailableCarriesPositiveOnly(t *testing.T) {
	ts, budget, account, payee := calcSetup(t)
	cat, err := budget.CreateCategory(ts.ctx, "Groceries", nil)
	if err != nil {
		t.Fatal(err)
	}
	// User overspends 100 3 months ago
	trx1, _ := account.CreateTransaction(ts.ctx, payee, relMonth(-3), 100, 0, "")
	trx1.AddLine(ts.ctx, nil, cat, false, 100, 0)
	// User spends another 50 in month -1, overallocates and has income
	trx2, _ := account.CreateTransaction(ts.ctx, payee, relMonth(-1), 50, 0, "")
	trx2.AddLine(ts.ctx, nil, cat, false, 50, 0)
	trx3, _ := account.CreateTransaction(ts.ctx, payee, relMonth(-1), 0, 1000, "")
	trx3.AddLine(ts.ctx, nil, nil, true, 0, 1000)
	cat.SetAllocation(ts.ctx, relMonth(-1), 200)
	// User allocates more in current month
	cat.SetAllocation(ts.ctx, relMonth(0), 200)
	// User allocates more in future month
	cat.SetAllocation(ts.ctx, relMonth(1), 123)

	// 3 months ago
	m := fetchBudgetMonth(t, ts, budget.ID(), sRelMonth(-3))
	assertBudgetMonthSummary(t, m, api.BudgetMonthSummary{
		Allocated:     0,
		Income:        0,
		Spent:         100,
		ReadyToAssign: 0,
	})
	assertBudgetMonthCategory(t, m, "Groceries", api.BudgetMonthCategory{
		Allocated: 0,
		Spent:     100,
		Available: -100,
	})
	// 2 months ago
	m = fetchBudgetMonth(t, ts, budget.ID(), sRelMonth(-2))
	assertBudgetMonthSummary(t, m, api.BudgetMonthSummary{
		Allocated:     0,
		Income:        0,
		Spent:         0,
		ReadyToAssign: -100,
	})
	assertBudgetMonthCategory(t, m, "Groceries", api.BudgetMonthCategory{
		Allocated: 0,
		Spent:     0,
		Available: 0,
	})
	// month with allocation
	m = fetchBudgetMonth(t, ts, budget.ID(), sRelMonth(-1))
	assertBudgetMonthSummary(t, m, api.BudgetMonthSummary{
		Allocated:     200,
		Income:        1000,
		Spent:         50,
		ReadyToAssign: 700,
	})
	assertBudgetMonthCategory(t, m, "Groceries", api.BudgetMonthCategory{
		Allocated: 200,
		Spent:     50,
		Available: 150,
	})
	// Current month
	m = fetchBudgetMonth(t, ts, budget.ID(), sRelMonth(0))
	assertBudgetMonthSummary(t, m, api.BudgetMonthSummary{
		Allocated:     200,
		Income:        0,
		Spent:         0,
		ReadyToAssign: 500,
	})
	assertBudgetMonthCategory(t, m, "Groceries", api.BudgetMonthCategory{
		Allocated: 200,
		Spent:     0,
		Available: 350,
	})
	// future month
	m = fetchBudgetMonth(t, ts, budget.ID(), sRelMonth(1))
	assertBudgetMonthSummary(t, m, api.BudgetMonthSummary{
		Allocated:     123,
		Income:        0,
		Spent:         0,
		ReadyToAssign: 377,
	})
	assertBudgetMonthCategory(t, m, "Groceries", api.BudgetMonthCategory{
		Allocated: 123,
		Spent:     0,
		Available: 123,
	})
	// future future month
	m = fetchBudgetMonth(t, ts, budget.ID(), sRelMonth(2))
	assertBudgetMonthSummary(t, m, api.BudgetMonthSummary{
		Allocated:     0,
		Income:        0,
		Spent:         0,
		ReadyToAssign: 377,
	})
	assertBudgetMonthCategory(t, m, "Groceries", api.BudgetMonthCategory{
		Allocated: 0,
		Spent:     0,
		Available: 0,
	})
}

// Negative allocations should net zero in ready to assign
func TestCalcNegativeAllocations(t *testing.T) {
	ts, budget, account, payee := calcSetup(t)
	cat, err := budget.CreateCategory(ts.ctx, "Groceries", nil)
	if err != nil {
		t.Fatal(err)
	}
	// Add income
	trx, _ := account.CreateTransaction(ts.ctx, payee, relMonth(-3), 0, 1000, "")
	trx.AddLine(ts.ctx, nil, nil, true, 0, 1000)
	// Allocations
	cat.SetAllocation(ts.ctx, relMonth(-2), 200)
	cat.SetAllocation(ts.ctx, relMonth(-1), -300)

	// 3 months ago
	m := fetchBudgetMonth(t, ts, budget.ID(), sRelMonth(-3))
	assertBudgetMonthSummary(t, m, api.BudgetMonthSummary{
		Allocated:     0,
		Income:        1000,
		Spent:         0,
		ReadyToAssign: 1000,
	})
	assertBudgetMonthCategory(t, m, "Groceries", api.BudgetMonthCategory{
		Allocated: 0,
		Spent:     0,
		Available: 0,
	})
	// 2 months ago
	m = fetchBudgetMonth(t, ts, budget.ID(), sRelMonth(-2))
	assertBudgetMonthSummary(t, m, api.BudgetMonthSummary{
		Allocated:     200,
		Income:        0,
		Spent:         0,
		ReadyToAssign: 800,
	})
	assertBudgetMonthCategory(t, m, "Groceries", api.BudgetMonthCategory{
		Allocated: 200,
		Spent:     0,
		Available: 200,
	})
	// month with allocation
	m = fetchBudgetMonth(t, ts, budget.ID(), sRelMonth(-1))
	assertBudgetMonthSummary(t, m, api.BudgetMonthSummary{
		Allocated:     -300,
		Income:        0,
		Spent:         0,
		ReadyToAssign: 1100,
	})
	assertBudgetMonthCategory(t, m, "Groceries", api.BudgetMonthCategory{
		Allocated: -300,
		Spent:     0,
		Available: -100,
	})
	// Current month
	m = fetchBudgetMonth(t, ts, budget.ID(), sRelMonth(0))
	assertBudgetMonthSummary(t, m, api.BudgetMonthSummary{
		Allocated:     0,
		Income:        0,
		Spent:         0,
		ReadyToAssign: 1000,
	})
	assertBudgetMonthCategory(t, m, "Groceries", api.BudgetMonthCategory{
		Allocated: 0,
		Spent:     0,
		Available: 0,
	})
}

// Negative allocations on spend should only reduce RTA by spend
func TestNegativeAllocationOnSpend(t *testing.T) {
	ts, budget, account, payee := calcSetup(t)
	cat, err := budget.CreateCategory(ts.ctx, "Groceries", nil)
	if err != nil {
		t.Fatal(err)
	}
	// Add income
	trx, _ := account.CreateTransaction(ts.ctx, payee, relMonth(-3), 0, 1000, "")
	trx.AddLine(ts.ctx, nil, nil, true, 0, 1000)
	trx2, _ := account.CreateTransaction(ts.ctx, payee, relMonth(-2), 100, 0, "")
	trx2.AddLine(ts.ctx, nil, cat, false, 100, 0)
	// Allocations
	cat.SetAllocation(ts.ctx, relMonth(-2), -300)

	// 3 months ago
	m := fetchBudgetMonth(t, ts, budget.ID(), sRelMonth(-3))
	assertBudgetMonthSummary(t, m, api.BudgetMonthSummary{
		Allocated:     0,
		Income:        1000,
		Spent:         0,
		ReadyToAssign: 1000,
	})
	assertBudgetMonthCategory(t, m, "Groceries", api.BudgetMonthCategory{
		Allocated: 0,
		Spent:     0,
		Available: 0,
	})
	// 2 months ago
	m = fetchBudgetMonth(t, ts, budget.ID(), sRelMonth(-2))
	assertBudgetMonthSummary(t, m, api.BudgetMonthSummary{
		Allocated:     -300,
		Income:        0,
		Spent:         100,
		ReadyToAssign: 1300,
	})
	assertBudgetMonthCategory(t, m, "Groceries", api.BudgetMonthCategory{
		Allocated: -300,
		Spent:     100,
		Available: -400,
	})
	// month with allocation
	m = fetchBudgetMonth(t, ts, budget.ID(), sRelMonth(-1))
	assertBudgetMonthSummary(t, m, api.BudgetMonthSummary{
		Allocated:     0,
		Income:        0,
		Spent:         0,
		ReadyToAssign: 900,
	})
	assertBudgetMonthCategory(t, m, "Groceries", api.BudgetMonthCategory{
		Allocated: 0,
		Spent:     0,
		Available: 0,
	})
	// Current month
	m = fetchBudgetMonth(t, ts, budget.ID(), sRelMonth(0))
	assertBudgetMonthSummary(t, m, api.BudgetMonthSummary{
		Allocated:     0,
		Income:        0,
		Spent:         0,
		ReadyToAssign: 900,
	})
	assertBudgetMonthCategory(t, m, "Groceries", api.BudgetMonthCategory{
		Allocated: 0,
		Spent:     0,
		Available: 0,
	})
}

// Test a monthly allocation goal
// Monthly needs the full amount every month regardless of prior allocations.
func TestCalcMonthlyAllocationGoal(t *testing.T) {
	ts, budget, _, _ := calcSetup(t)
	cat, err := budget.CreateCategory(ts.ctx, "Groceries", nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := cat.CreateGoal(ts.ctx, "monthly", relMonth(-1), nil, 500); err != nil {
		t.Fatal(err)
	}
	if _, err := cat.SetAllocation(ts.ctx, relMonth(-2), 200); err != nil {
		t.Fatal(err)
	}
	if _, err := cat.SetAllocation(ts.ctx, relMonth(-1), 500); err != nil {
		t.Fatal(err)
	}
	if _, err := cat.SetAllocation(ts.ctx, relMonth(0), 200); err != nil {
		t.Fatal(err)
	}

	// Before the goal starts there is no goal on the category.
	m := fetchBudgetMonth(t, ts, budget.ID(), sRelMonth(-2))
	assertNoGoal(t, m, "Groceries")

	// Goal month: need the full amount, gap is what is still missing.
	m = fetchBudgetMonth(t, ts, budget.ID(), sRelMonth(-1))
	assertGoal(t, m, "Groceries", api.BudgetMonthGoal{
		Type:           api.BudgetMonthGoalTypeMonthly,
		Amount:         500,
		Allocated:      500,
		AmountForMonth: 500,
		Gap:            0,
	})

	// Monthly goals persist with no end date; nothing allocated this month.
	m = fetchBudgetMonth(t, ts, budget.ID(), sRelMonth(0))
	assertGoal(t, m, "Groceries", api.BudgetMonthGoal{
		Type:           api.BudgetMonthGoalTypeMonthly,
		Amount:         500,
		Allocated:      700,
		AmountForMonth: 500,
		Gap:            300,
	})
}

// Test a monthly refill goal.
// Accounts for the available amount at the begining of the month
func TestCalcMonthlyRefillGoal(t *testing.T) {
	ts, budget, _, _ := calcSetup(t)
	cat, err := budget.CreateCategory(ts.ctx, "Groceries", nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := cat.CreateGoal(ts.ctx, "refill", relMonth(-1), nil, 400); err != nil {
		t.Fatal(err)
	}
	if _, err := cat.SetAllocation(ts.ctx, relMonth(-2), 400); err != nil {
		t.Fatal(err)
	}
	if _, err := cat.SetAllocation(ts.ctx, relMonth(0), 400); err != nil {
		t.Fatal(err)
	}
	// Before the goal starts there is no goal on the category.
	m := fetchBudgetMonth(t, ts, budget.ID(), sRelMonth(-2))
	assertNoGoal(t, m, "Groceries")
	// First month of goal, available is carried
	m = fetchBudgetMonth(t, ts, budget.ID(), sRelMonth(-1))
	assertGoal(t, m, "Groceries", api.BudgetMonthGoal{
		Type:           api.BudgetMonthGoalTypeRefill,
		Amount:         400,
		Allocated:      0,
		AmountForMonth: 0,
		Gap:            0,
	})
	// First allocation, overallocated
	m = fetchBudgetMonth(t, ts, budget.ID(), sRelMonth(0))
	assertGoal(t, m, "Groceries", api.BudgetMonthGoal{
		Type:           api.BudgetMonthGoalTypeRefill,
		Amount:         400,
		Allocated:      400,
		AmountForMonth: 0,
		Gap:            -400,
	})
	// still overallocated
	m = fetchBudgetMonth(t, ts, budget.ID(), sRelMonth(1))
	assertGoal(t, m, "Groceries", api.BudgetMonthGoal{
		Type:           api.BudgetMonthGoalTypeRefill,
		Amount:         400,
		Allocated:      400,
		AmountForMonth: -400,
		Gap:            -400,
	})
}

// Test a save goal
func TestCalcMonthlySave(t *testing.T) {
	ts, budget, _, _ := calcSetup(t)
	cat, err := budget.CreateCategory(ts.ctx, "Groceries", nil)
	if err != nil {
		t.Fatal(err)
	}
	end := relMonth(0)
	if _, err := cat.CreateGoal(ts.ctx, "save", relMonth(-2), &end, 900); err != nil {
		t.Fatal(err)
	}
	if _, err := cat.SetAllocation(ts.ctx, relMonth(-2), 100); err != nil {
		t.Fatal(err)
	}
	if _, err := cat.SetAllocation(ts.ctx, relMonth(-1), 100); err != nil {
		t.Fatal(err)
	}

	// Before the goal starts there is no goal on the category.
	m := fetchBudgetMonth(t, ts, budget.ID(), sRelMonth(-3))
	assertNoGoal(t, m, "Groceries")

	// Needed spreads the remainder over the months left, inclusive of end.
	// Month -2: (900-100+100)/3 = 300, gap 300-100 = 200.
	m = fetchBudgetMonth(t, ts, budget.ID(), sRelMonth(-2))
	assertGoal(t, m, "Groceries", api.BudgetMonthGoal{
		Type:           api.BudgetMonthGoalTypeSave,
		Amount:         900,
		Allocated:      100,
		AmountForMonth: 300,
		Gap:            200,
	})

	// Month -1: (900-200+100)/2 = 400, gap 400-100 = 300.
	m = fetchBudgetMonth(t, ts, budget.ID(), sRelMonth(-1))
	assertGoal(t, m, "Groceries", api.BudgetMonthGoal{
		Type:           api.BudgetMonthGoalTypeSave,
		Amount:         900,
		Allocated:      200,
		AmountForMonth: 400,
		Gap:            300,
	})

	// Final month: (900-200+0)/1 = 700, gap 700.
	m = fetchBudgetMonth(t, ts, budget.ID(), sRelMonth(0))
	assertGoal(t, m, "Groceries", api.BudgetMonthGoal{
		Type:           api.BudgetMonthGoalTypeSave,
		Amount:         900,
		Allocated:      200,
		AmountForMonth: 700,
		Gap:            700,
	})

	// After the end month the goal is gone.
	m = fetchBudgetMonth(t, ts, budget.ID(), sRelMonth(1))
	assertNoGoal(t, m, "Groceries")
}
