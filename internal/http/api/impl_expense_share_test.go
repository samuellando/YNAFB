package api_test

import (
	"database/sql"
	"strconv"
	"testing"
	"time"

	"samuellando.com/YNAFB/data"
	"samuellando.com/YNAFB/internal/db/types"
	"samuellando.com/YNAFB/internal/http/api"
)

func TestPutExpenseShareTrxSplits(t *testing.T) {
	ts := setupTestServer(t)
	budgetA := ts.newBudget(t, "Alice")
	budgetB := ts.newBudget(t, "Bob")
	budgetC := ts.newBudget(t, "Carol")
	share := ts.newExpenseShare(t, budgetA, "Trip")
	for _, b := range []data.Budget{budgetB, budgetC} {
		if _, err := ts.queries.JoinExpenseShare(ts.ctx, data.JoinExpenseShareParams{
			ExpenseShareID: share.ExpenseShareID,
			LoginID:        b.LoginID,
			BudgetID:       b.ID,
			DisplayName:    b.Name,
		}); err != nil {
			t.Fatal(err)
		}
	}
	account := ts.newAccount(t, budgetA, "Chequing")
	payee := ts.newPayee(t, budgetA, "Store")
	trx := ts.newTrx(t, budgetA, account, payee)
	if _, err := ts.queries.CreateTrxLine(ts.ctx, data.CreateTrxLineParams{
		TrxID:          trx.ID,
		ExpenseShareID: sql.NullInt64{Int64: share.ExpenseShareID, Valid: true},
		Income:         false,
		Outflow:        1000,
		Inflow:         0,
		BudgetID:       budgetA.ID,
		LoginID:        budgetA.LoginID,
	}); err != nil {
		t.Fatal(err)
	}
	sharePath := "/api/v1/budget/" + strconv.Itoa(int(budgetA.ID)) +
		"/expense-share/" + strconv.Itoa(int(share.ExpenseShareID))

	var published api.ExpenseShareTrx
	w := ts.doReq(t, "POST", sharePath+"/trx", map[string]any{"trxId": trx.ID}, 201)
	decodeJSON(t, w, &published)

	splitsPath := sharePath + "/trx/" + strconv.Itoa(published.Id) + "/splits"
	put := func(splits []map[string]any, want int) api.ExpenseShareTrxDetail {
		t.Helper()
		w := ts.doReq(t, "PUT", splitsPath, map[string]any{"splits": splits}, want)
		var detail api.ExpenseShareTrxDetail
		if want == 200 {
			decodeJSON(t, w, &detail)
		}
		return detail
	}
	split := func(budget data.Budget, outflow int) map[string]any {
		return map[string]any{"budgetId": budget.ID, "outflow": outflow, "inflow": 0}
	}

	// Only non-publisher splits are submitted; the publisher (Alice) keeps the
	// remainder automatically (here total == requested, so 0). A zero split is
	// allowed for members who are not paying (Carol).
	detail := put([]map[string]any{split(budgetB, 1000), split(budgetC, 0)}, 200)
	if len(detail.Splits) != 3 {
		t.Fatalf("len(splits) = %d, want 3", len(detail.Splits))
	}
	got := map[int]int{}
	defaults := map[int]bool{}
	for _, s := range detail.Splits {
		got[s.BudgetId] = s.SplitOutflow
		defaults[s.BudgetId] = s.IsDefault
	}
	if got[int(budgetB.ID)] != 1000 || got[int(budgetC.ID)] != 0 || got[int(budgetA.ID)] != 0 {
		t.Fatalf("splits = %+v, want {A:0 B:1000 C:0}", got)
	}
	if defaults[int(budgetB.ID)] || defaults[int(budgetC.ID)] {
		t.Fatalf("non-publisher splits should be stored, defaults = %+v", defaults)
	}

	// The publishing budget's split is managed automatically.
	put([]map[string]any{split(budgetA, 0), split(budgetB, 600), split(budgetC, 400)}, 500)
	// Totals must match the requested amount.
	put([]map[string]any{split(budgetB, 600), split(budgetC, 300)}, 500)
	// Every non-publisher member needs a split.
	put([]map[string]any{split(budgetB, 1000)}, 500)
	// Unknown budgets are rejected.
	put([]map[string]any{split(budgetB, 600), map[string]any{"budgetId": 9999, "outflow": 400, "inflow": 0}}, 500)
	// Inflow amounts are rejected on an outflow transaction.
	put([]map[string]any{
		{"budgetId": budgetB.ID, "outflow": 0, "inflow": 600},
		split(budgetC, 400),
	}, 500)

	// Stored splits survive a re-read.
	var reread api.ExpenseShareDetail
	w = ts.doReq(t, "GET", sharePath, nil, 200)
	decodeJSON(t, w, &reread)
	if len(reread.Transactions) != 1 || len(reread.Transactions[0].Splits) != 3 {
		t.Fatalf("reread transactions = %+v", reread.Transactions)
	}

	// A zero split (Carol's) has nothing to categorize.
	carolCategory := ts.newCategory(t, budgetC, "Groceries")
	zeroLinesPath := "/api/v1/budget/" + strconv.Itoa(int(budgetC.ID)) +
		"/expense-share/" + strconv.Itoa(int(share.ExpenseShareID)) +
		"/trx/" + strconv.Itoa(published.Id) + "/lines"
	ts.doReq(t, "PUT", zeroLinesPath,
		map[string]any{"lines": []map[string]any{{"categoryId": carolCategory.ID, "outflow": 100, "inflow": 0}}}, 500)
}

func TestDeleteExpenseShareTrx(t *testing.T) {
	ts := setupTestServer(t)
	budgetA := ts.newBudget(t, "Alice")
	budgetB := ts.newBudget(t, "Bob")
	share := ts.newExpenseShare(t, budgetA, "Trip")
	if _, err := ts.queries.JoinExpenseShare(ts.ctx, data.JoinExpenseShareParams{
		ExpenseShareID: share.ExpenseShareID,
		LoginID:        budgetB.LoginID,
		BudgetID:       budgetB.ID,
		DisplayName:    "Bob",
	}); err != nil {
		t.Fatal(err)
	}
	account := ts.newAccount(t, budgetA, "Chequing")
	payee := ts.newPayee(t, budgetA, "Store")
	trx := ts.newTrx(t, budgetA, account, payee)
	if _, err := ts.queries.CreateTrxLine(ts.ctx, data.CreateTrxLineParams{
		TrxID:          trx.ID,
		ExpenseShareID: sql.NullInt64{Int64: share.ExpenseShareID, Valid: true},
		Income:         false,
		Outflow:        1000,
		Inflow:         0,
		BudgetID:       budgetA.ID,
		LoginID:        budgetA.LoginID,
	}); err != nil {
		t.Fatal(err)
	}
	var published api.ExpenseShareTrx
	w := ts.doReq(t, "POST",
		"/api/v1/budget/"+strconv.Itoa(int(budgetA.ID))+
			"/expense-share/"+strconv.Itoa(int(share.ExpenseShareID))+"/trx",
		map[string]any{"trxId": trx.ID}, 201)
	decodeJSON(t, w, &published)

	// Unknown ids are rejected.
	bobTrxPath := func(id any) string {
		return "/api/v1/budget/" + strconv.Itoa(int(budgetB.ID)) +
			"/expense-share/" + strconv.Itoa(int(share.ExpenseShareID)) +
			"/trx/" + strconv.Itoa(id.(int))
	}
	ts.doReq(t, "DELETE", bobTrxPath(9999), nil, 500)

	// Any member may delete, including a non-publisher.
	ts.doReq(t, "DELETE", bobTrxPath(published.Id), nil, 204)

	// The published trx is gone from the share.
	var reread api.ExpenseShareDetail
	w = ts.doReq(t, "GET",
		"/api/v1/budget/"+strconv.Itoa(int(budgetA.ID))+
			"/expense-share/"+strconv.Itoa(int(share.ExpenseShareID)),
		nil, 200)
	decodeJSON(t, w, &reread)
	if len(reread.Transactions) != 0 {
		t.Fatalf("transactions after delete = %+v, want empty", reread.Transactions)
	}
	if reread.Summary.TransactionCount != 0 {
		t.Fatalf("transactionCount = %d, want 0", reread.Summary.TransactionCount)
	}

	// Deleting again fails: the row no longer exists.
	ts.doReq(t, "DELETE", bobTrxPath(published.Id), nil, 500)

	// The source local transaction survives the share delete.
	trxs, err := ts.queries.ListTrxs(ts.ctx, data.ListTrxsParams{
		LoginID:  budgetA.LoginID,
		BudgetID: budgetA.ID,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(trxs) != 1 || trxs[0].ID != trx.ID {
		t.Fatalf("source trxs = %+v, want the original trx %d", trxs, trx.ID)
	}
}

func TestPutExpenseShareTrxLines(t *testing.T) {
	ts := setupTestServer(t)
	budgetA := ts.newBudget(t, "Alice")
	budgetB := ts.newBudget(t, "Bob")
	share := ts.newExpenseShare(t, budgetA, "Trip")
	if _, err := ts.queries.JoinExpenseShare(ts.ctx, data.JoinExpenseShareParams{
		ExpenseShareID: share.ExpenseShareID,
		LoginID:        budgetB.LoginID,
		BudgetID:       budgetB.ID,
		DisplayName:    "Bob",
	}); err != nil {
		t.Fatal(err)
	}
	category := ts.newCategory(t, budgetB, "Groceries")
	account := ts.newAccount(t, budgetA, "Chequing")
	payee := ts.newPayee(t, budgetA, "Store")
	trx := ts.newTrx(t, budgetA, account, payee)
	if _, err := ts.queries.CreateTrxLine(ts.ctx, data.CreateTrxLineParams{
		TrxID:          trx.ID,
		ExpenseShareID: sql.NullInt64{Int64: share.ExpenseShareID, Valid: true},
		Income:         false,
		Outflow:        1000,
		Inflow:         0,
		BudgetID:       budgetA.ID,
		LoginID:        budgetA.LoginID,
	}); err != nil {
		t.Fatal(err)
	}
	var published api.ExpenseShareTrx
	w := ts.doReq(t, "POST",
		"/api/v1/budget/"+strconv.Itoa(int(budgetA.ID))+
			"/expense-share/"+strconv.Itoa(int(share.ExpenseShareID))+"/trx",
		map[string]any{"trxId": trx.ID}, 201)
	decodeJSON(t, w, &published)

	linesPath := "/api/v1/budget/" + strconv.Itoa(int(budgetB.ID)) +
		"/expense-share/" + strconv.Itoa(int(share.ExpenseShareID)) +
		"/trx/" + strconv.Itoa(published.Id) + "/lines"
	put := func(lines []map[string]any, want int) api.ExpenseShareTrxDetail {
		t.Helper()
		w := ts.doReq(t, "PUT", linesPath, map[string]any{"lines": lines}, want)
		var detail api.ExpenseShareTrxDetail
		if want == 200 {
			decodeJSON(t, w, &detail)
		}
		return detail
	}
	own := func(detail api.ExpenseShareTrxDetail) api.ExpenseShareTrxSplit {
		t.Helper()
		for _, s := range detail.Splits {
			if s.BudgetId == int(budgetB.ID) {
				return s
			}
		}
		t.Fatalf("no own split in %+v", detail.Splits)
		return api.ExpenseShareTrxSplit{}
	}

	// No stored splits exist yet: saving lines materializes Bob's default
	// split (1000) and attaches the line to it.
	detail := put([]map[string]any{
		{"categoryId": category.ID, "outflow": 700, "inflow": 0},
		{"categoryId": category.ID, "outflow": 300, "inflow": 0},
	}, 200)
	got := own(detail)
	if got.SplitOutflow != 1000 || got.IsDefault {
		t.Fatalf("own split = %+v, want stored 1000", got)
	}
	if got.Lines == nil || len(*got.Lines) != 2 {
		t.Fatalf("own lines = %+v, want 2 lines", got.Lines)
	}

	// Re-saving replaces the lines.
	detail = put([]map[string]any{
		{"categoryId": category.ID, "outflow": 1000, "inflow": 0},
	}, 200)
	if got := own(detail); got.Lines == nil || len(*got.Lines) != 1 {
		t.Fatalf("own lines after replace = %+v, want 1 line", got.Lines)
	}

	// An empty list clears the categorization but keeps the split.
	detail = put([]map[string]any{}, 200)
	if got := own(detail); got.Lines == nil || len(*got.Lines) != 0 {
		t.Fatalf("own lines after clear = %+v, want empty", got.Lines)
	}

	// The publisher categorizes through the source transaction.
	pubPath := "/api/v1/budget/" + strconv.Itoa(int(budgetA.ID)) +
		"/expense-share/" + strconv.Itoa(int(share.ExpenseShareID)) +
		"/trx/" + strconv.Itoa(published.Id) + "/lines"
	ts.doReq(t, "PUT", pubPath,
		map[string]any{"lines": []map[string]any{{"categoryId": category.ID, "outflow": 1000, "inflow": 0}}}, 500)
	// Unknown categories, wrong sides, and ambiguous targets are rejected.
	put([]map[string]any{{"categoryId": 9999, "outflow": 1000, "inflow": 0}}, 500)
	put([]map[string]any{{"categoryId": category.ID, "outflow": 0, "inflow": 1000}}, 500)
	put([]map[string]any{{"outflow": 1000, "inflow": 0}}, 500)
}

func TestExpenseShareBalances(t *testing.T) {
	setup := func(t *testing.T, names ...string) (*testServer, []data.Budget, data.BudgetExpenseShare) {
		t.Helper()
		ts := setupTestServer(t)
		budgets := make([]data.Budget, 0, len(names))
		for _, n := range names {
			budgets = append(budgets, ts.newBudget(t, n))
		}
		share := ts.newExpenseShare(t, budgets[0], "Trip")
		for _, b := range budgets[1:] {
			if _, err := ts.queries.JoinExpenseShare(ts.ctx, data.JoinExpenseShareParams{
				ExpenseShareID: share.ExpenseShareID,
				LoginID:        b.LoginID,
				BudgetID:       b.ID,
				DisplayName:    b.Name,
			}); err != nil {
				t.Fatal(err)
			}
		}
		return ts, budgets, share
	}
	var trxSeq int
	// publish creates a trx with the given totals and a single share line for
	// the requested amount, then publishes it to the share.
	publish := func(t *testing.T, ts *testServer, share data.BudgetExpenseShare, pub data.Budget, totalOut, totalIn, reqOut, reqIn int64) api.ExpenseShareTrx {
		t.Helper()
		trxSeq++
		account := ts.newAccount(t, pub, "Chequing "+strconv.Itoa(trxSeq))
		payee := ts.newPayee(t, pub, "Store")
		trx, err := ts.queries.CreateTrx(ts.ctx, data.CreateTrxParams{
			BudgetID:     pub.ID,
			LoginID:      pub.LoginID,
			AccountID:    account.ID,
			PayeeID:      payee.ID,
			Date:         types.UnixTime{Time: time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)},
			TotalOutflow: totalOut,
			TotalInflow:  totalIn,
			Note:         "",
		})
		if err != nil {
			t.Fatal(err)
		}
		if _, err := ts.queries.CreateTrxLine(ts.ctx, data.CreateTrxLineParams{
			TrxID:          trx.ID,
			ExpenseShareID: sql.NullInt64{Int64: share.ExpenseShareID, Valid: true},
			Income:         false,
			Outflow:        reqOut,
			Inflow:         reqIn,
			BudgetID:       pub.ID,
			LoginID:        pub.LoginID,
		}); err != nil {
			t.Fatal(err)
		}
		w := ts.doReq(t, "POST",
			"/api/v1/budget/"+strconv.Itoa(int(pub.ID))+
				"/expense-share/"+strconv.Itoa(int(share.ExpenseShareID))+"/trx",
			map[string]any{"trxId": trx.ID}, 201)
		var published api.ExpenseShareTrx
		decodeJSON(t, w, &published)
		return published
	}
	get := func(t *testing.T, ts *testServer, viewer data.Budget, share data.BudgetExpenseShare) api.ExpenseShareDetail {
		t.Helper()
		w := ts.doReq(t, "GET",
			"/api/v1/budget/"+strconv.Itoa(int(viewer.ID))+
				"/expense-share/"+strconv.Itoa(int(share.ExpenseShareID)),
			nil, 200)
		var detail api.ExpenseShareDetail
		decodeJSON(t, w, &detail)
		return detail
	}
	check := func(t *testing.T, detail api.ExpenseShareDetail, budgets []data.Budget, want map[string]int, wantOrder []string) {
		t.Helper()
		got := map[string]int{}
		var order []string
		sum := 0
		for _, b := range detail.Summary.Balances {
			got[b.DisplayName] = b.Balance
			order = append(order, b.DisplayName)
			sum += b.Balance
		}
		if len(got) != len(budgets) {
			t.Fatalf("balances = %+v, want one row per member (%d)", got, len(budgets))
		}
		for name, amount := range want {
			if got[name] != amount {
				t.Fatalf("balances = %+v, want %+v", got, want)
			}
		}
		for i, name := range wantOrder {
			if order[i] != name {
				t.Fatalf("balance order = %+v, want %+v", order, wantOrder)
			}
		}
		if sum != 0 {
			t.Fatalf("balances sum = %d, want 0 (got %+v)", sum, got)
		}
		if detail.Summary.MemberCount != len(budgets) {
			t.Fatalf("memberCount = %d, want %d", detail.Summary.MemberCount, len(budgets))
		}
	}

	t.Run("default outflow split", func(t *testing.T) {
		ts, budgets, share := setup(t, "Alice", "Bob")
		publish(t, ts, share, budgets[0], 1000, 0, 1000, 0)
		want := map[string]int{"Alice": 1000, "Bob": -1000}
		check(t, get(t, ts, budgets[0], share), budgets, want, []string{"Alice", "Bob"})
		// Every member sees the same balances.
		check(t, get(t, ts, budgets[1], share), budgets, want, []string{"Alice", "Bob"})
	})

	t.Run("stored splits", func(t *testing.T) {
		ts, budgets, share := setup(t, "Alice", "Bob", "Carol")
		published := publish(t, ts, share, budgets[0], 1000, 0, 1000, 0)
		splitsPath := "/api/v1/budget/" + strconv.Itoa(int(budgets[0].ID)) +
			"/expense-share/" + strconv.Itoa(int(share.ExpenseShareID)) +
			"/trx/" + strconv.Itoa(published.Id) + "/splits"
		ts.doReq(t, "PUT", splitsPath, map[string]any{"splits": []map[string]any{
			{"budgetId": budgets[1].ID, "outflow": 600, "inflow": 0},
			{"budgetId": budgets[2].ID, "outflow": 400, "inflow": 0},
		}}, 200)
		check(t, get(t, ts, budgets[0], share), budgets,
			map[string]int{"Alice": 1000, "Bob": -600, "Carol": -400},
			[]string{"Alice", "Carol", "Bob"})
	})

	t.Run("leftover cents go to the lowest budget id", func(t *testing.T) {
		ts, budgets, share := setup(t, "Alice", "Bob", "Carol", "Dave")
		publish(t, ts, share, budgets[0], 1000, 0, 1000, 0)
		check(t, get(t, ts, budgets[0], share), budgets,
			map[string]int{"Alice": 1000, "Bob": -334, "Carol": -333, "Dave": -333},
			[]string{"Alice", "Carol", "Dave", "Bob"})
	})

	t.Run("inflow flips the sign", func(t *testing.T) {
		ts, budgets, share := setup(t, "Alice", "Bob")
		publish(t, ts, share, budgets[0], 0, 1000, 0, 1000)
		check(t, get(t, ts, budgets[0], share), budgets,
			map[string]int{"Alice": -1000, "Bob": 1000},
			[]string{"Bob", "Alice"})
	})

	t.Run("publisher keeps the unrequested remainder", func(t *testing.T) {
		ts, budgets, share := setup(t, "Alice", "Bob")
		publish(t, ts, share, budgets[0], 1200, 0, 1000, 0)
		check(t, get(t, ts, budgets[0], share), budgets,
			map[string]int{"Alice": 1000, "Bob": -1000},
			[]string{"Alice", "Bob"})
	})

	t.Run("empty share lists zero balances", func(t *testing.T) {
		ts, budgets, share := setup(t, "Alice", "Bob")
		detail := get(t, ts, budgets[0], share)
		check(t, detail, budgets,
			map[string]int{"Alice": 0, "Bob": 0},
			[]string{"Alice", "Bob"})
		if detail.Summary.TransactionCount != 0 {
			t.Fatalf("transactionCount = %d, want 0", detail.Summary.TransactionCount)
		}
	})
}
