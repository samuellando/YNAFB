package api_test

import (
	"net/http"
	"strconv"
	"testing"

	"samuellando.com/YNAFB/internal/domain"
	"samuellando.com/YNAFB/internal/http/api"
)

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

// Allocations should substract from teh ready to assign in month and all future months
// It should show in available for the month and future months up to the current month
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
