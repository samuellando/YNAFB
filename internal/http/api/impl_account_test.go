package api_test

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"samuellando.com/YNAFB/internal/http/api"
)

func TestGetBudgetBudgetIdAccount(t *testing.T) {
	ts := setupTestServer(t)
	budget := ts.newBudget(t, "Home Budget")
	account := ts.newAccount(t, budget, "Chequing")
	w := ts.doReq(t, "GET", "/api/v1/budget/"+strconv.Itoa(int(budget.ID))+"/account", nil, http.StatusOK)
	var accounts []api.AccountSummaryDetail
	decodeJSON(t, w, &accounts)
	if len(accounts) != 1 {
		t.Fatalf("accounts = %+v, want 1 entry", accounts)
	}
	if accounts[0].Id != int(account.ID) || accounts[0].Name != "Chequing" {
		t.Fatalf("accounts[0] = %+v, want id %d name %q", accounts[0], account.ID, "Chequing")
	}
	if accounts[0].Balance != 0 || accounts[0].ReconciledBalance != 0 {
		t.Fatalf("accounts[0] balance = %d/%d, want 0/0", accounts[0].Balance, accounts[0].ReconciledBalance)
	}
}

func TestPostBudgetBudgetIdAccount(t *testing.T) {
	ts := setupTestServer(t)
	budget := ts.newBudget(t, "Home Budget")
	w := ts.doReq(t, "POST", "/api/v1/budget/"+strconv.Itoa(int(budget.ID))+"/account", map[string]any{"name": "Chequing"}, http.StatusOK)
	var account api.Account
	decodeJSON(t, w, &account)
	if account.Id == 0 {
		t.Fatalf("account = %+v, want a non-zero id", account)
	}
	if account.Name != "Chequing" {
		t.Fatalf("account name = %q, want %q", account.Name, "Chequing")
	}
}

func TestGetBudgetBudgetIdAccountId(t *testing.T) {
	ts := setupTestServer(t)
	budget := ts.newBudget(t, "Home Budget")
	account := ts.newAccount(t, budget, "Chequing")
	w := ts.doReq(t, "GET", "/api/v1/budget/"+strconv.Itoa(int(budget.ID))+"/account/"+strconv.Itoa(int(account.ID)), nil, http.StatusOK)
	var detail api.AccountDetail
	decodeJSON(t, w, &detail)
	if detail.Summary.Id != int(account.ID) || detail.Summary.Name != "Chequing" {
		t.Fatalf("summary = %+v, want id %d name %q", detail.Summary, account.ID, "Chequing")
	}
	if detail.Summary.Balance != 0 || detail.Summary.ReconciledBalance != 0 {
		t.Fatalf("summary balance = %d/%d, want 0/0", detail.Summary.Balance, detail.Summary.ReconciledBalance)
	}
	if len(detail.Transactions) != 0 {
		t.Fatalf("transactions = %+v, want none", detail.Transactions)
	}
}

func TestPutBudgetBudgetIdAccountId(t *testing.T) {
	ts := setupTestServer(t)
	budget := ts.newBudget(t, "Home Budget")
	account := ts.newAccount(t, budget, "Chequing")
	w := ts.doReq(t, "PUT", "/api/v1/budget/"+strconv.Itoa(int(budget.ID))+"/account/"+strconv.Itoa(int(account.ID)), map[string]any{"name": "Savings"}, http.StatusOK)
	var updated api.Account
	decodeJSON(t, w, &updated)
	if updated.Id != int(account.ID) || updated.Name != "Savings" {
		t.Fatalf("updated = %+v, want id %d name %q", updated, account.ID, "Savings")
	}
}

func TestDeleteBudgetBudgetIdAccountId(t *testing.T) {
	ts := setupTestServer(t)
	budget := ts.newBudget(t, "Home Budget")
	account := ts.newAccount(t, budget, "Chequing")
	w := ts.doReq(t, "DELETE", "/api/v1/budget/"+strconv.Itoa(int(budget.ID))+"/account/"+strconv.Itoa(int(account.ID)), nil, http.StatusOK)
	if w.Body.Len() != 0 {
		t.Fatalf("delete body = %q, want empty", w.Body.String())
	}
}

func TestPostBudgetBudgetIdAccountIdImport(t *testing.T) {
	ts := setupTestServer(t)
	budget := ts.newBudget(t, "Home Budget")
	account := ts.newAccount(t, budget, "Chequing")

	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	fw, err := mw.CreateFormFile("statement", "statement.txt")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := fw.Write([]byte("ynafb-test-statement")); err != nil {
		t.Fatal(err)
	}
	if err := mw.Close(); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest("POST", "/api/v1/budget/"+strconv.Itoa(int(budget.ID))+"/account/"+strconv.Itoa(int(account.ID))+"/import", &buf)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	req.AddCookie(&http.Cookie{Name: "jwt", Value: ts.token})
	w := httptest.NewRecorder()
	ts.mux.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("import: status = %d, body = %q", w.Code, w.Body.String())
	}
	if w.Body.String() != "2" {
		t.Fatalf("import: body = %q, want %q", w.Body.String(), "2")
	}
}

func TestPostBudgetBudgetIdAccountIdReconcile(t *testing.T) {
	ts := setupTestServer(t)
	budget := ts.newBudget(t, "Home Budget")
	account := ts.newAccount(t, budget, "Chequing")
	payee := ts.newPayee(t, budget, "Payee")
	ts.newTrx(t, budget, account, payee)
	w := ts.doReq(t, "POST", "/api/v1/budget/"+strconv.Itoa(int(budget.ID))+"/account/"+strconv.Itoa(int(account.ID))+"/reconcile",
		map[string]any{"date": "2026-09-30T00:00:00Z", "balance": -1000}, http.StatusOK)
	if w.Body.Len() != 0 {
		t.Fatalf("reconcile body = %q, want empty", w.Body.String())
	}
}