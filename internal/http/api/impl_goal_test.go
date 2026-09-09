package api_test

import (
	"net/http"
	"strconv"
	"testing"

	"samuellando.com/YNAFB/internal/http/api"
)

func TestGetBudgetBudgetIdCategoryCategoryIdGoal(t *testing.T) {
	ts := setupTestServer(t)
	budget := ts.newBudget(t, "Home Budget")
	category := ts.newCategory(t, budget, "Groceries")
	ts.newGoal(t, budget, category)
	w := ts.doReq(t, "GET", "/api/v1/budget/"+strconv.Itoa(int(budget.ID))+"/category/"+strconv.Itoa(int(category.ID))+"/goal", nil, http.StatusOK)
	var goal api.Goal
	decodeJSON(t, w, &goal)
	if goal.Id == 0 {
		t.Fatalf("goal = %+v, want a non-zero id", goal)
	}
	if goal.CategoryId != int(category.ID) || goal.Amount != 5000 {
		t.Fatalf("goal = %+v, want category %d amount 5000", goal, category.ID)
	}
	if goal.StartMonth == "" {
		t.Fatalf("goal = %+v, want a non-empty start month", goal)
	}
}

func TestPostBudgetBudgetIdCategoryCategoryIdGoal(t *testing.T) {
	ts := setupTestServer(t)
	budget := ts.newBudget(t, "Home Budget")
	category := ts.newCategory(t, budget, "Groceries")
	w := ts.doReq(t, "POST", "/api/v1/budget/"+strconv.Itoa(int(budget.ID))+"/category/"+strconv.Itoa(int(category.ID))+"/goal",
		map[string]any{"type": "monthly", "startMonth": "2026-09", "amount": 5000}, http.StatusOK)
	var goal api.Goal
	decodeJSON(t, w, &goal)
	if goal.Id == 0 {
		t.Fatalf("goal = %+v, want a non-zero id", goal)
	}
	if goal.CategoryId != int(category.ID) || goal.Amount != 5000 {
		t.Fatalf("goal = %+v, want category %d amount 5000", goal, category.ID)
	}
}

func TestPutBudgetBudgetIdCategoryCategoryIdGoal(t *testing.T) {
	ts := setupTestServer(t)
	budget := ts.newBudget(t, "Home Budget")
	category := ts.newCategory(t, budget, "Groceries")
	ts.newGoal(t, budget, category)
	w := ts.doReq(t, "PUT", "/api/v1/budget/"+strconv.Itoa(int(budget.ID))+"/category/"+strconv.Itoa(int(category.ID))+"/goal",
		map[string]any{"type": "save", "startMonth": "2026-09", "endMonth": "2026-12", "amount": 10000}, http.StatusOK)
	var goal api.Goal
	decodeJSON(t, w, &goal)
	if goal.Amount != 10000 || goal.StartMonth == "" {
		t.Fatalf("goal = %+v, want amount 10000 and a non-empty start month", goal)
	}
}

func TestDeleteBudgetBudgetIdCategoryCategoryIdGoal(t *testing.T) {
	ts := setupTestServer(t)
	budget := ts.newBudget(t, "Home Budget")
	category := ts.newCategory(t, budget, "Groceries")
	ts.newGoal(t, budget, category)
	w := ts.doReq(t, "DELETE", "/api/v1/budget/"+strconv.Itoa(int(budget.ID))+"/category/"+strconv.Itoa(int(category.ID))+"/goal", nil, http.StatusOK)
	if w.Body.Len() != 0 {
		t.Fatalf("delete body = %q, want empty", w.Body.String())
	}
}