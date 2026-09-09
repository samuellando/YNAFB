package api_test

import (
	"net/http"
	"strconv"
	"testing"

	"samuellando.com/YNAFB/internal/http/api"
)

func TestGetBudgetBudgetIdCategoryGroup(t *testing.T) {
	ts := setupTestServer(t)
	budget := ts.newBudget(t, "Home Budget")
	group := ts.newCategoryGroup(t, budget, "Group")
	w := ts.doReq(t, "GET", "/api/v1/budget/"+strconv.Itoa(int(budget.ID))+"/category-group", nil, http.StatusOK)
	var groups []api.CategoryGroup
	decodeJSON(t, w, &groups)
	if len(groups) != 1 {
		t.Fatalf("groups = %+v, want 1 entry", groups)
	}
	if groups[0].Id != int(group) || groups[0].Name != "Group" {
		t.Fatalf("groups[0] = %+v, want id %d name %q", groups[0], group, "Group")
	}
}

func TestPostBudgetBudgetIdCategoryGroup(t *testing.T) {
	ts := setupTestServer(t)
	budget := ts.newBudget(t, "Home Budget")
	w := ts.doReq(t, "POST", "/api/v1/budget/"+strconv.Itoa(int(budget.ID))+"/category-group", map[string]any{"name": "Group"}, http.StatusOK)
	var group api.CategoryGroup
	decodeJSON(t, w, &group)
	if group.Id == 0 {
		t.Fatalf("group = %+v, want a non-zero id", group)
	}
	if group.Name != "Group" {
		t.Fatalf("group name = %q, want %q", group.Name, "Group")
	}
}

func TestPutBudgetBudgetIdCategoryGroupId(t *testing.T) {
	ts := setupTestServer(t)
	budget := ts.newBudget(t, "Home Budget")
	group := ts.newCategoryGroup(t, budget, "Group")
	w := ts.doReq(t, "PUT", "/api/v1/budget/"+strconv.Itoa(int(budget.ID))+"/category-group/"+strconv.Itoa(int(group)), map[string]any{"name": "Group2"}, http.StatusOK)
	var updated api.CategoryGroup
	decodeJSON(t, w, &updated)
	if updated.Id != int(group) || updated.Name != "Group2" {
		t.Fatalf("updated = %+v, want id %d name %q", updated, group, "Group2")
	}
}

func TestDeleteBudgetBudgetIdCategoryGroupId(t *testing.T) {
	ts := setupTestServer(t)
	budget := ts.newBudget(t, "Home Budget")
	group := ts.newCategoryGroup(t, budget, "Group")
	w := ts.doReq(t, "DELETE", "/api/v1/budget/"+strconv.Itoa(int(budget.ID))+"/category-group/"+strconv.Itoa(int(group)), nil, http.StatusOK)
	if w.Body.Len() != 0 {
		t.Fatalf("delete body = %q, want empty", w.Body.String())
	}
}