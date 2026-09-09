package api_test

import (
	"net/http"
	"strconv"
	"testing"

	"samuellando.com/YNAFB/internal/http/api"
)

func TestGetBudgetBudgetIdPayee(t *testing.T) {
	ts := setupTestServer(t)
	budget := ts.newBudget(t, "Home Budget")
	payee := ts.newPayee(t, budget, "Payee")
	w := ts.doReq(t, "GET", "/api/v1/budget/"+strconv.Itoa(int(budget.ID))+"/payee", nil, http.StatusOK)
	var payees []api.Payee
	decodeJSON(t, w, &payees)
	if len(payees) != 1 {
		t.Fatalf("payees = %+v, want 1 entry", payees)
	}
	if payees[0].Id != int(payee.ID) || payees[0].Name != "Payee" {
		t.Fatalf("payees[0] = %+v, want id %d name %q", payees[0], payee.ID, "Payee")
	}
}

func TestPostBudgetBudgetIdPayee(t *testing.T) {
	ts := setupTestServer(t)
	budget := ts.newBudget(t, "Home Budget")
	w := ts.doReq(t, "POST", "/api/v1/budget/"+strconv.Itoa(int(budget.ID))+"/payee", map[string]any{"name": "Payee"}, http.StatusOK)
	var payee api.Payee
	decodeJSON(t, w, &payee)
	if payee.Id == 0 {
		t.Fatalf("payee = %+v, want a non-zero id", payee)
	}
	if payee.Name != "Payee" {
		t.Fatalf("payee name = %q, want %q", payee.Name, "Payee")
	}
}

func TestPutBudgetBudgetIdPayeeId(t *testing.T) {
	ts := setupTestServer(t)
	budget := ts.newBudget(t, "Home Budget")
	payee := ts.newPayee(t, budget, "Payee")
	w := ts.doReq(t, "PUT", "/api/v1/budget/"+strconv.Itoa(int(budget.ID))+"/payee/"+strconv.Itoa(int(payee.ID)), map[string]any{"name": "Payee2"}, http.StatusOK)
	var updated api.Payee
	decodeJSON(t, w, &updated)
	if updated.Id != int(payee.ID) || updated.Name != "Payee2" {
		t.Fatalf("updated = %+v, want id %d name %q", updated, payee.ID, "Payee2")
	}
}

func TestDeleteBudgetBudgetIdPayeeId(t *testing.T) {
	ts := setupTestServer(t)
	budget := ts.newBudget(t, "Home Budget")
	payee := ts.newPayee(t, budget, "Payee")
	w := ts.doReq(t, "DELETE", "/api/v1/budget/"+strconv.Itoa(int(budget.ID))+"/payee/"+strconv.Itoa(int(payee.ID)), nil, http.StatusOK)
	if w.Body.Len() != 0 {
		t.Fatalf("delete body = %q, want empty", w.Body.String())
	}
}