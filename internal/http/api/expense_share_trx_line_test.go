package api_test

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"samuellando.com/YNAFB/internal/http/api"
)

// Share refs on transaction lines must round-trip as member budget ids:
// POST/PUT resolve them, PUT preserves them (previously stripped), and the
// account detail includes them (previously omitted).
func TestExpenseShareTrxLineRefs(t *testing.T) {
	ts := setupTestServer(t)
	ctx := ts.ctx

	budgetA, err := ts.domain.CreateBudget(ctx, int(ts.loginID), "Budget A")
	if err != nil {
		t.Fatal(err)
	}
	budgetB, err := ts.domain.CreateBudget(ctx, int(ts.loginID), "Budget B")
	if err != nil {
		t.Fatal(err)
	}
	accountA, err := budgetA.CreateAccount(ctx, "Chequing A")
	if err != nil {
		t.Fatal(err)
	}
	payeeA, err := budgetA.CreatePayee(ctx, "Restaurant")
	if err != nil {
		t.Fatal(err)
	}
	accountB, err := budgetB.CreateAccount(ctx, "Chequing B")
	if err != nil {
		t.Fatal(err)
	}
	payeeB, err := budgetB.CreatePayee(ctx, "Alice")
	if err != nil {
		t.Fatal(err)
	}

	memA, err := budgetA.CreateExpenseShare(ctx, "Trip", "Alice", nil)
	if err != nil {
		t.Fatal(err)
	}
	shareID := memA.Share().ID()
	code, err := memA.Share().MintCode(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := budgetB.JoinExpenseShare(ctx, code.Code(), "Trip", "Bob"); err != nil {
		t.Fatal(err)
	}

	trx, err := accountA.CreateTransaction(ctx, payeeA, time.Now(), 10000, 0, "dinner")
	if err != nil {
		t.Fatal(err)
	}
	linePath := fmt.Sprintf("/api/v1/budget/%d/account/%d/transaction/%d/line",
		budgetA.ID(), accountA.ID(), trx.ID())

	// Split line tagging B: response ids must be the global share id and B's
	// budget id so the frontend can round-trip them into later requests.
	w := ts.doReq(t, "POST", linePath, map[string]any{
		"expenseShareId": shareID,
		"splitBudgetId":  budgetB.ID(),
		"outflow":        4000,
		"inflow":         0,
	}, http.StatusOK)
	var created api.TrxLine
	decodeJSON(t, w, &created)
	if created.ExpenseShareId == nil || *created.ExpenseShareId != shareID {
		t.Fatalf("expected expenseShareId %d, got %+v", shareID, created.ExpenseShareId)
	}
	if created.SplitBudgetId == nil || *created.SplitBudgetId != budgetB.ID() {
		t.Fatalf("expected splitBudgetId %d, got %+v", budgetB.ID(), created.SplitBudgetId)
	}

	// Settlement-only POST (no splitBudgetId) must not panic.
	settleTrx, err := accountB.CreateTransaction(ctx, payeeB, time.Now(), 1000, 0, "settle")
	if err != nil {
		t.Fatal(err)
	}
	settlePath := fmt.Sprintf("/api/v1/budget/%d/account/%d/transaction/%d/line",
		budgetB.ID(), accountB.ID(), settleTrx.ID())
	w = ts.doReq(t, "POST", settlePath, map[string]any{
		"expenseShareId": shareID,
		"destBudgetId":   budgetA.ID(),
		"outflow":        1000,
		"inflow":         0,
	}, http.StatusOK)
	var settled api.TrxLine
	decodeJSON(t, w, &settled)
	if settled.DestBudgetId == nil || *settled.DestBudgetId != budgetA.ID() {
		t.Fatalf("expected destBudgetId %d, got %+v", budgetA.ID(), settled.DestBudgetId)
	}

	// PUT preserves share refs while updating amounts.
	w = ts.doReq(t, "PUT", fmt.Sprintf("%s/%d", linePath, created.Id), map[string]any{
		"expenseShareId": shareID,
		"splitBudgetId":  budgetB.ID(),
		"outflow":        5000,
		"inflow":         0,
	}, http.StatusOK)
	var updated api.TrxLine
	decodeJSON(t, w, &updated)
	if updated.Outflow != 5000 {
		t.Fatalf("expected outflow 5000, got %d", updated.Outflow)
	}
	if updated.ExpenseShareId == nil || *updated.ExpenseShareId != shareID ||
		updated.SplitBudgetId == nil || *updated.SplitBudgetId != budgetB.ID() {
		t.Fatalf("PUT stripped share refs: %+v", updated)
	}

	// Account detail exposes the refs with display names.
	w = ts.doReq(t, "GET", fmt.Sprintf("/api/v1/budget/%d/account/%d",
		budgetA.ID(), accountA.ID()), nil, http.StatusOK)
	var detail api.AccountDetail
	decodeJSON(t, w, &detail)
	found := false
	for _, tr := range detail.Transactions {
		for _, line := range tr.TransactionLines {
			if line.LineId == created.Id {
				found = true
				if line.ExpenseShareId == nil || *line.ExpenseShareId != shareID {
					t.Fatalf("missing expenseShareId in %+v", line)
				}
				if line.SplitBudgetId == nil || *line.SplitBudgetId != budgetB.ID() {
					t.Fatalf("missing splitBudgetId in %+v", line)
				}
				if line.SplitBudgetDisplayName == nil || *line.SplitBudgetDisplayName != "Bob" {
					t.Fatalf("missing split display name in %+v", line)
				}
			}
		}
	}
	if !found {
		t.Fatalf("created line %d not found in account detail", created.Id)
	}

	// Share detail exposes the split breakdown with member budget ids.
	w = ts.doReq(t, "GET", fmt.Sprintf("/api/v1/budget/%d/expense-share/%d",
		budgetA.ID(), shareID), nil, http.StatusOK)
	var shareDetail api.ExpenseShareDetail
	decodeJSON(t, w, &shareDetail)
	shareFound := false
	for _, shareTrx := range shareDetail.Transactions {
		if shareTrx.TrxId != trx.ID() {
			continue
		}
		shareFound = true
		if shareTrx.SplitLines == nil || len(*shareTrx.SplitLines) != 1 {
			t.Fatalf("expected 1 split line, got %+v", shareTrx.SplitLines)
		}
		split := (*shareTrx.SplitLines)[0]
		if split.SplitBudgetId != budgetB.ID() {
			t.Fatalf("expected splitBudgetId %d, got %d", budgetB.ID(), split.SplitBudgetId)
		}
		if split.SplitBudgetDisplayName != "Bob" {
			t.Fatalf("expected split display %q, got %q", "Bob", split.SplitBudgetDisplayName)
		}
		if split.Outflow != 5000 {
			t.Fatalf("expected split outflow 5000, got %d", split.Outflow)
		}
	}
	if !shareFound {
		t.Fatalf("transaction %d not found in share detail", trx.ID())
	}
}
