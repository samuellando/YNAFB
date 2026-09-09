package api_test

import (
	"net/http"
	"strconv"
	"testing"

	"samuellando.com/YNAFB/internal/http/api"
)

func TestPostBudgetBudgetIdAccountAccountIdTransaction(t *testing.T) {
	ts := setupTestServer(t)
	budget := ts.newBudget(t, "Home Budget")
	account := ts.newAccount(t, budget, "Chequing")
	payee := ts.newPayee(t, budget, "Payee")
	w := ts.doReq(t, "POST", "/api/v1/budget/"+strconv.Itoa(int(budget.ID))+"/account/"+strconv.Itoa(int(account.ID))+"/transaction",
		map[string]any{"payeeId": payee.ID, "date": "2026-09-01T00:00:00Z", "outflow": 1000, "inflow": 0}, http.StatusOK)
	var trx api.Transaction
	decodeJSON(t, w, &trx)
	if trx.Id == 0 {
		t.Fatalf("trx = %+v, want a non-zero id", trx)
	}
	if trx.AccountId != int(account.ID) || trx.PayeeId != int(payee.ID) {
		t.Fatalf("trx = %+v, want account %d payee %d", trx, account.ID, payee.ID)
	}
	if trx.Outflow != 1000 || trx.Inflow != 0 {
		t.Fatalf("trx = %+v, want outflow 1000 inflow 0", trx)
	}
	if trx.Date == "" {
		t.Fatalf("trx = %+v, want a non-empty date", trx)
	}
}

func TestPutBudgetBudgetIdAccountAccountIdTransactionId(t *testing.T) {
	ts := setupTestServer(t)
	budget := ts.newBudget(t, "Home Budget")
	account := ts.newAccount(t, budget, "Chequing")
	payee := ts.newPayee(t, budget, "Payee")
	trx := ts.newTrx(t, budget, account, payee)
	w := ts.doReq(t, "PUT", "/api/v1/budget/"+strconv.Itoa(int(budget.ID))+"/account/"+strconv.Itoa(int(account.ID))+"/transaction/"+strconv.Itoa(int(trx.ID)),
		map[string]any{"payeeId": payee.ID, "date": "2026-09-01T00:00:00Z", "outflow": 1500, "inflow": 0}, http.StatusOK)
	var updated api.Transaction
	decodeJSON(t, w, &updated)
	if updated.Id != int(trx.ID) || updated.Outflow != 1500 {
		t.Fatalf("updated = %+v, want id %d outflow 1500", updated, trx.ID)
	}
}

func TestDeleteBudgetBudgetIdAccountAccountIdTransactionId(t *testing.T) {
	ts := setupTestServer(t)
	budget := ts.newBudget(t, "Home Budget")
	account := ts.newAccount(t, budget, "Chequing")
	payee := ts.newPayee(t, budget, "Payee")
	trx := ts.newTrx(t, budget, account, payee)
	w := ts.doReq(t, "DELETE", "/api/v1/budget/"+strconv.Itoa(int(budget.ID))+"/account/"+strconv.Itoa(int(account.ID))+"/transaction/"+strconv.Itoa(int(trx.ID)), nil, http.StatusOK)
	if w.Body.Len() != 0 {
		t.Fatalf("delete body = %q, want empty", w.Body.String())
	}
}