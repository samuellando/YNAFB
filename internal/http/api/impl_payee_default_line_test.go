package api_test

import (
	"database/sql"
	"net/http"
	"strconv"
	"testing"

	"samuellando.com/YNAFB/data"
	"samuellando.com/YNAFB/internal/http/api"
)

func TestGetBudgetBudgetIdPayeePayeeIdDefaultLine(t *testing.T) {
	ts := setupTestServer(t)
	budget := ts.newBudget(t, "Home Budget")
	payee := ts.newPayee(t, budget, "Payee")
	ts.newDefaultLine(t, budget, payee)
	w := ts.doReq(t, "GET", "/api/v1/budget/"+strconv.Itoa(int(budget.ID))+"/payee/"+strconv.Itoa(int(payee.ID))+"/default-line", nil, http.StatusOK)
	var lines []api.PayeeDefaultLine
	decodeJSON(t, w, &lines)
	if len(lines) != 1 {
		t.Fatalf("lines = %+v, want 1 entry", lines)
	}
	if lines[0].PayeeId != int(payee.ID) || lines[0].Income != true || lines[0].Percent != 100 {
		t.Fatalf("lines[0] = %+v, want payee %d income true percent 100", lines[0], payee.ID)
	}
}

func TestPostBudgetBudgetIdPayeePayeeIdDefaultLine(t *testing.T) {	ts := setupTestServer(t)
	budget := ts.newBudget(t, "Home Budget")
	payee := ts.newPayee(t, budget, "Payee")
	w := ts.doReq(t, "POST", "/api/v1/budget/"+strconv.Itoa(int(budget.ID))+"/payee/"+strconv.Itoa(int(payee.ID))+"/default-line",
		map[string]any{"income": true, "percent": 100}, http.StatusOK)
	var line api.PayeeDefaultLine
	decodeJSON(t, w, &line)
	if line.Id == 0 {
		t.Fatalf("line = %+v, want a non-zero id", line)
	}
	if line.PayeeId != int(payee.ID) || line.Income != true || line.Percent != 100 {
		t.Fatalf("line = %+v, want payee %d income true percent 100", line, payee.ID)
	}
}

func TestPutBudgetBudgetIdPayeePayeeIdDefaultLineId(t *testing.T) {
	ts := setupTestServer(t)
	budget := ts.newBudget(t, "Home Budget")
	payee := ts.newPayee(t, budget, "Payee")
	line := ts.newDefaultLine(t, budget, payee)
	w := ts.doReq(t, "PUT", "/api/v1/budget/"+strconv.Itoa(int(budget.ID))+"/payee/"+strconv.Itoa(int(payee.ID))+"/default-line/"+strconv.Itoa(int(line.ID)),
		map[string]any{"income": true, "percent": 50}, http.StatusOK)
	var updated api.PayeeDefaultLine
	decodeJSON(t, w, &updated)
	if updated.Id != int(line.ID) || updated.Percent != 50 {
		t.Fatalf("updated = %+v, want id %d percent 50", updated, line.ID)
	}
}

func TestDeleteBudgetBudgetIdPayeePayeeIdDefaultLineId(t *testing.T) {	ts := setupTestServer(t)
	budget := ts.newBudget(t, "Home Budget")
	payee := ts.newPayee(t, budget, "Payee")
	line := ts.newDefaultLine(t, budget, payee)
	w := ts.doReq(t, "DELETE", "/api/v1/budget/"+strconv.Itoa(int(budget.ID))+"/payee/"+strconv.Itoa(int(payee.ID))+"/default-line/"+strconv.Itoa(int(line.ID)), nil, http.StatusOK)
	if w.Body.Len() != 0 {
		t.Fatalf("delete body = %q, want empty", w.Body.String())
	}
}

func TestGetBudgetBudgetIdPayeePayeeIdDefaultLineExpenseShare(t *testing.T) {
	ts := setupTestServer(t)
	budget := ts.newBudget(t, "Home Budget")
	payee := ts.newPayee(t, budget, "Payee")
	share := ts.newExpenseShare(t, budget, "Trip")
	_, err := ts.queries.CreatePayeeDefaultLine(ts.ctx, data.CreatePayeeDefaultLineParams{
		BudgetID:       budget.ID,
		LoginID:        budget.LoginID,
		PayeeID:        payee.ID,
		ExpenseShareID: sql.NullInt64{Int64: share.ExpenseShareID, Valid: true},
		Income:         false,
		Percent:        100,
	})
	if err != nil {
		t.Fatal(err)
	}
	w := ts.doReq(t, "GET", "/api/v1/budget/"+strconv.Itoa(int(budget.ID))+"/payee/"+strconv.Itoa(int(payee.ID))+"/default-line", nil, http.StatusOK)
	var lines []api.PayeeDefaultLine
	decodeJSON(t, w, &lines)
	if len(lines) != 1 {
		t.Fatalf("lines = %+v, want 1 entry", lines)
	}
	if lines[0].ExpenseShareId == nil || *lines[0].ExpenseShareId != int(share.ExpenseShareID) {
		t.Fatalf("lines[0].ExpenseShareId = %+v, want %d", lines[0].ExpenseShareId, share.ExpenseShareID)
	}
	if lines[0].ExpenseShareName == nil || *lines[0].ExpenseShareName != "Trip" {
		t.Fatalf("lines[0].ExpenseShareName = %+v, want Trip", lines[0].ExpenseShareName)
	}
}