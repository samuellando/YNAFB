package api_test

import (
	"net/http"
	"strconv"
	"testing"

	"samuellando.com/YNAFB/internal/http/api"
)

func TestGetBudget(t *testing.T) {
	ts := setupTestServer(t)
	budget := ts.newBudget(t, "Home Budget")
	w := ts.doReq(t, "GET", "/api/v1/budget", nil, http.StatusOK)
	var budgets []api.Budget
	decodeJSON(t, w, &budgets)
	if len(budgets) != 1 {
		t.Fatalf("budgets = %+v, want 1 entry", budgets)
	}
	if budgets[0].Id != int(budget.ID) || budgets[0].Name != "Home Budget" {
		t.Fatalf("budgets[0] = %+v, want id %d name %q", budgets[0], budget.ID, "Home Budget")
	}
}

func TestPostBudget(t *testing.T) {
	ts := setupTestServer(t)
	w := ts.doReq(t, "POST", "/api/v1/budget", map[string]any{"name": "New Budget"}, http.StatusCreated)
	var budget api.Budget
	decodeJSON(t, w, &budget)
	if budget.Id == 0 {
		t.Fatalf("budget = %+v, want a non-zero id", budget)
	}
	if budget.Name != "New Budget" {
		t.Fatalf("budget name = %q, want %q", budget.Name, "New Budget")
	}
}

func TestGetBudgetBudgetId(t *testing.T) {
	ts := setupTestServer(t)
	budget := ts.newBudget(t, "Home Budget")
	w := ts.doReq(t, "GET", "/api/v1/budget/"+strconv.Itoa(int(budget.ID)), nil, http.StatusOK)
	var month api.BudgetMonth
	decodeJSON(t, w, &month)
	if month.Month == "" {
		t.Fatalf("month = %+v, want a non-empty month", month)
	}
}

func TestPutBudgetBudgetId(t *testing.T) {
	ts := setupTestServer(t)
	budget := ts.newBudget(t, "Home Budget")
	w := ts.doReq(t, "PUT", "/api/v1/budget/"+strconv.Itoa(int(budget.ID)), map[string]any{"name": "Renamed"}, http.StatusOK)
	var updated api.Budget
	decodeJSON(t, w, &updated)
	if updated.Id != int(budget.ID) || updated.Name != "Renamed" {
		t.Fatalf("updated = %+v, want id %d name %q", updated, budget.ID, "Renamed")
	}
}

func TestDeleteBudgetBudgetId(t *testing.T) {
	ts := setupTestServer(t)
	budget := ts.newBudget(t, "Home Budget")
	w := ts.doReq(t, "DELETE", "/api/v1/budget/"+strconv.Itoa(int(budget.ID)), nil, http.StatusNoContent)
	if w.Body.Len() != 0 {
		t.Fatalf("delete body = %q, want empty", w.Body.String())
	}
}

func TestGetBudgetBudgetIdMonth(t *testing.T) {
	ts := setupTestServer(t)
	budget := ts.newBudget(t, "Home Budget")
	w := ts.doReq(t, "GET", "/api/v1/budget/"+strconv.Itoa(int(budget.ID))+"/2026-09", nil, http.StatusOK)
	var month api.BudgetMonth
	decodeJSON(t, w, &month)
	if month.Month == "" {
		t.Fatalf("month = %+v, want a non-empty month", month)
	}
}