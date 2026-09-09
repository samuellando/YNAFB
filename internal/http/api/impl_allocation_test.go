package api_test

import (
	"net/http"
	"strconv"
	"testing"

	"samuellando.com/YNAFB/internal/http/api"
)

func TestPutBudgetBudgetIdCategoryIdAllocationMonth(t *testing.T) {
	ts := setupTestServer(t)
	budget := ts.newBudget(t, "Home Budget")
	category := ts.newCategory(t, budget, "Groceries")
	w := ts.doReq(t, "PUT", "/api/v1/budget/"+strconv.Itoa(int(budget.ID))+"/category/"+strconv.Itoa(int(category.ID))+"/allocation/2026-09",
		map[string]any{"amount": 1000}, http.StatusOK)
	var allocation api.Allocation
	decodeJSON(t, w, &allocation)
	if allocation.Id == 0 {
		t.Fatalf("allocation = %+v, want a non-zero id", allocation)
	}
	if allocation.CategoryId != int(category.ID) || allocation.Amount != 1000 {
		t.Fatalf("allocation = %+v, want category %d amount 1000", allocation, category.ID)
	}
	if allocation.Month == "" {
		t.Fatalf("allocation = %+v, want a non-empty month", allocation)
	}
}