package api_test

import (
	"net/http"
	"strconv"
	"testing"

	"samuellando.com/YNAFB/internal/http/api"
)

func TestGetBudgetBudgetIdCategory(t *testing.T) {
	ts := setupTestServer(t)
	budget := ts.newBudget(t, "Home Budget")
	category := ts.newCategory(t, budget, "Groceries")
	w := ts.doReq(t, "GET", "/api/v1/budget/"+strconv.Itoa(int(budget.ID))+"/category", nil, http.StatusOK)
	var categories []api.Category
	decodeJSON(t, w, &categories)
	if len(categories) != 1 {
		t.Fatalf("categories = %+v, want 1 entry", categories)
	}
	if categories[0].Id != int(category.ID) || categories[0].Name != "Groceries" {
		t.Fatalf("categories[0] = %+v, want id %d name %q", categories[0], category.ID, "Groceries")
	}
}

func TestPostBudgetBudgetIdCategory(t *testing.T) {
	ts := setupTestServer(t)
	budget := ts.newBudget(t, "Home Budget")
	w := ts.doReq(t, "POST", "/api/v1/budget/"+strconv.Itoa(int(budget.ID))+"/category", map[string]any{"name": "Groceries"}, http.StatusOK)
	var category api.Category
	decodeJSON(t, w, &category)
	if category.Id == 0 {
		t.Fatalf("category = %+v, want a non-zero id", category)
	}
	if category.Name != "Groceries" {
		t.Fatalf("category name = %q, want %q", category.Name, "Groceries")
	}
}

func TestPutBudgetBudgetIdCategoryId(t *testing.T) {
	ts := setupTestServer(t)
	budget := ts.newBudget(t, "Home Budget")
	category := ts.newCategory(t, budget, "Groceries")
	w := ts.doReq(t, "PUT", "/api/v1/budget/"+strconv.Itoa(int(budget.ID))+"/category/"+strconv.Itoa(int(category.ID)), map[string]any{"name": "Food"}, http.StatusOK)
	var updated api.Category
	decodeJSON(t, w, &updated)
	if updated.Id != int(category.ID) || updated.Name != "Food" {
		t.Fatalf("updated = %+v, want id %d name %q", updated, category.ID, "Food")
	}
}

func TestDeleteBudgetBudgetIdCategoryId(t *testing.T) {
	ts := setupTestServer(t)
	budget := ts.newBudget(t, "Home Budget")
	category := ts.newCategory(t, budget, "Groceries")
	w := ts.doReq(t, "DELETE", "/api/v1/budget/"+strconv.Itoa(int(budget.ID))+"/category/"+strconv.Itoa(int(category.ID)), nil, http.StatusOK)
	if w.Body.Len() != 0 {
		t.Fatalf("delete body = %q, want empty", w.Body.String())
	}
}