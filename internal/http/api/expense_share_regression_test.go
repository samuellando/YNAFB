package api_test

import (
	"database/sql"
	"errors"
	"testing"
	"time"
)

func strPtr(s string) *string { return &s }

// End-to-end regression test for expense shares: create, join by code,
// split lines, counterparty categorizations, balances, update, leave.
func TestExpenseShareLifecycle(t *testing.T) {
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
	categoryB, err := budgetB.CreateCategory(ctx, "Food", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create with nil defaultName falls back to name.
	memA, err := budgetA.CreateExpenseShare(ctx, "Trip", "Alice", nil)
	if err != nil {
		t.Fatal(err)
	}
	if memA.Share().DefaultName() != "Trip" {
		t.Fatalf("expected default name %q, got %q", "Trip", memA.Share().DefaultName())
	}
	shareID := memA.Share().ID()

	listed, err := budgetA.ListExpenseShares(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(listed) != 1 || listed[0].ID() != memA.ID() {
		t.Fatalf("expected 1 share membership, got %+v", listed)
	}

	got, err := budgetA.GetExpenseShare(ctx, shareID)
	if err != nil {
		t.Fatal(err)
	}
	if got.ID() != memA.ID() {
		t.Fatalf("expected membership %d, got %d", memA.ID(), got.ID())
	}

	code, err := memA.Share().MintCode(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(code.Code()) < 10 {
		t.Fatalf("expected code length >= 10, got %q", code.Code())
	}
	if !code.Expires().After(time.Now()) {
		t.Fatalf("expected future expiry, got %v", code.Expires())
	}

	if _, err := budgetB.JoinExpenseShare(ctx, "no-such-code", "Trip", "Bob"); err == nil {
		t.Fatal("expected error joining with invalid code")
	}
	memB, err := budgetB.JoinExpenseShare(ctx, code.Code(), "Trip", "Bob")
	if err != nil {
		t.Fatal(err)
	}
	if memB.Share().ID() != shareID {
		t.Fatalf("expected share %d, got %d", shareID, memB.Share().ID())
	}

	members, err := memA.Members(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(members) != 2 {
		t.Fatalf("expected 2 members, got %d", len(members))
	}

	resolved, err := memA.Share().GetMembership(ctx, budgetB.ID())
	if err != nil {
		t.Fatal(err)
	}
	if resolved.ID() != memB.ID() {
		t.Fatalf("expected membership %d, got %d", memB.ID(), resolved.ID())
	}

	// Source transaction in A with a split line tagging B for 4000.
	trx, err := accountA.CreateTransaction(ctx, payeeA, time.Now(), 10000, 0, "dinner")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := trx.AddLine(ctx, nil, nil, false, memA.Share(), memB, nil, 4000, 0); err != nil {
		t.Fatal(err)
	}

	// B sees the shared transaction with A's display name.
	bTrxs, err := memB.ListTransactions(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(bTrxs) != 1 {
		t.Fatalf("expected 1 shared transaction, got %d", len(bTrxs))
	}
	view := bTrxs[0]
	if view.SourceBudget().ID() != budgetA.ID() {
		t.Fatalf("expected source budget %d, got %d", budgetA.ID(), view.SourceBudget().ID())
	}
	if view.SourceBudgetDisplayName() != "Alice" {
		t.Fatalf("expected source display %q, got %q", "Alice", view.SourceBudgetDisplayName())
	}
	if view.TotalOutflow() != 10000 {
		t.Fatalf("expected total outflow 10000, got %d", view.TotalOutflow())
	}

	if _, err := memB.GetTransaction(ctx, trx.ID()); err != nil {
		t.Fatal(err)
	}
	if _, err := memA.GetTransaction(ctx, trx.ID()); err != nil {
		t.Fatal(err)
	}
	if _, err := memB.GetTransaction(ctx, 999999); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("expected sql.ErrNoRows, got %v", err)
	}

	// Counterparty categorization in B's own budget.
	cat, err := view.AddCategorization(ctx, categoryB, 4000, 0)
	if err != nil {
		t.Fatal(err)
	}
	if cat.Outflow() != 4000 || cat.Inflow() != 0 {
		t.Fatalf("unexpected categorization amounts %+v", cat)
	}
	catCategory, err := cat.Category()
	if err != nil {
		t.Fatal(err)
	}
	if catCategory.Name() != "Food" {
		t.Fatalf("expected category %q, got %q", "Food", catCategory.Name())
	}
	cats, err := view.MyCategorizations(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(cats) != 1 || cats[0].ID() != cat.ID() {
		t.Fatalf("expected 1 categorization, got %+v", cats)
	}
	if _, err := view.GetCategorization(ctx, cat.ID()); err != nil {
		t.Fatal(err)
	}

	// Balances follow the spec formula; categorizations are neutral.
	totalA, balancesA, err := memA.Balances(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if totalA != 4000 {
		t.Fatalf("expected caller net 4000, got %d", totalA)
	}
	if len(balancesA) != 1 || balancesA[0].Budget.ID() != budgetB.ID() || balancesA[0].Balance != 4000 || balancesA[0].DisplayName != "Bob" {
		t.Fatalf("unexpected pairwise balances %+v", balancesA)
	}
	totalB, balancesB, err := memB.Balances(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if totalB != -4000 {
		t.Fatalf("expected caller net -4000, got %d", totalB)
	}
	if len(balancesB) != 1 || balancesB[0].Balance != -4000 {
		t.Fatalf("unexpected pairwise balances %+v", balancesB)
	}

	if err := cat.Update(ctx, categoryB, 3000, 0); err != nil {
		t.Fatal(err)
	}
	if cat.Outflow() != 3000 {
		t.Fatalf("expected outflow 3000, got %d", cat.Outflow())
	}
	if totalB2, _, err := memB.Balances(ctx); err != nil || totalB2 != -4000 {
		t.Fatalf("categorizations must be balance-neutral: total=%d err=%v", totalB2, err)
	}
	if err := cat.Delete(ctx); err != nil {
		t.Fatal(err)
	}
	cats, err = view.MyCategorizations(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(cats) != 0 {
		t.Fatalf("expected no categorizations after delete, got %d", len(cats))
	}

	// Settlement of 1000 sent by B to A.
	settleTrx, err := accountB.CreateTransaction(ctx, payeeB, time.Now(), 1000, 0, "settle")
	if err != nil {
		t.Fatal(err)
	}
	memAForB, err := memB.Share().GetMembership(ctx, budgetA.ID())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := settleTrx.AddLine(ctx, nil, nil, false, memB.Share(), nil, memAForB, 1000, 0); err != nil {
		t.Fatal(err)
	}
	if totalA2, _, err := memA.Balances(ctx); err != nil || totalA2 != 5000 {
		t.Fatalf("expected caller net 5000 after settlement, got %d err=%v", totalA2, err)
	}
	if totalB3, _, err := memB.Balances(ctx); err != nil || totalB3 != -5000 {
		t.Fatalf("expected caller net -5000 after settlement, got %d err=%v", totalB3, err)
	}
	aTrxs, err := memA.ListTransactions(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(aTrxs) != 2 {
		t.Fatalf("expected 2 shared transactions, got %d", len(aTrxs))
	}

	// Rename and leave.
	if err := memB.Update(ctx, strPtr("Trips"), nil); err != nil {
		t.Fatal(err)
	}
	if memB.Name() != "Trips" || memB.DisplayName() != "Bob" {
		t.Fatalf("unexpected membership %+v", memB)
	}
	if err := memB.Leave(ctx); err != nil {
		t.Fatal(err)
	}
	remaining, err := memA.Members(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(remaining) != 1 {
		t.Fatalf("expected 1 remaining member, got %d", len(remaining))
	}
	empty, err := budgetB.ListExpenseShares(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(empty) != 0 {
		t.Fatalf("expected no shares after leave, got %d", len(empty))
	}
	if _, err := memB.GetTransaction(ctx, trx.ID()); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("expected sql.ErrNoRows after leave, got %v", err)
	}
}
