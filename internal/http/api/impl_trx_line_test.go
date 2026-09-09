package api_test

import (
	"net/http"
	"strconv"
	"testing"

	"samuellando.com/YNAFB/internal/http/api"
)

func TestPostBudgetBudgetIdAccountAccountIdTransactionTrxIdLine(t *testing.T) {
	ts := setupTestServer(t)
	budget := ts.newBudget(t, "Home Budget")
	account := ts.newAccount(t, budget, "Chequing")
	payee := ts.newPayee(t, budget, "Payee")
	trx := ts.newTrx(t, budget, account, payee)
	w := ts.doReq(t, "POST", "/api/v1/budget/"+strconv.Itoa(int(budget.ID))+"/account/"+strconv.Itoa(int(account.ID))+"/transaction/"+strconv.Itoa(int(trx.ID))+"/line",
		map[string]any{"income": true, "outflow": 0, "inflow": 500}, http.StatusOK)
	var line api.TrxLine
	decodeJSON(t, w, &line)
	if line.Id == 0 {
		t.Fatalf("line = %+v, want a non-zero id", line)
	}
	if line.TrxId != int(trx.ID) || line.Income != true || line.Outflow != 0 || line.Inflow != 500 {
		t.Fatalf("line = %+v, want trx %d income true outflow 0 inflow 500", line, trx.ID)
	}
}

func TestPutBudgetBudgetIdAccountAccountIdTransactionTrxIdLineId(t *testing.T) {
	ts := setupTestServer(t)
	budget := ts.newBudget(t, "Home Budget")
	account := ts.newAccount(t, budget, "Chequing")
	payee := ts.newPayee(t, budget, "Payee")
	trx := ts.newTrx(t, budget, account, payee)
	line := ts.newTrxLine(t, budget, trx)
	w := ts.doReq(t, "PUT", "/api/v1/budget/"+strconv.Itoa(int(budget.ID))+"/account/"+strconv.Itoa(int(account.ID))+"/transaction/"+strconv.Itoa(int(trx.ID))+"/line/"+strconv.Itoa(int(line.ID)),
		map[string]any{"income": true, "outflow": 0, "inflow": 700}, http.StatusOK)
	var updated api.TrxLine
	decodeJSON(t, w, &updated)
	if updated.Id != int(line.ID) || updated.Inflow != 700 {
		t.Fatalf("updated = %+v, want id %d inflow 700", updated, line.ID)
	}
}

func TestDeleteBudgetBudgetIdAccountAccountIdTransactionTrxIdLineId(t *testing.T) {
	ts := setupTestServer(t)
	budget := ts.newBudget(t, "Home Budget")
	account := ts.newAccount(t, budget, "Chequing")
	payee := ts.newPayee(t, budget, "Payee")
	trx := ts.newTrx(t, budget, account, payee)
	line := ts.newTrxLine(t, budget, trx)
	w := ts.doReq(t, "DELETE", "/api/v1/budget/"+strconv.Itoa(int(budget.ID))+"/account/"+strconv.Itoa(int(account.ID))+"/transaction/"+strconv.Itoa(int(trx.ID))+"/line/"+strconv.Itoa(int(line.ID)), nil, http.StatusOK)
	if w.Body.Len() != 0 {
		t.Fatalf("delete body = %q, want empty", w.Body.String())
	}
}