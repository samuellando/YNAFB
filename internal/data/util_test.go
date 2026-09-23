package data_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"samuellando.com/YNAFB/internal/data"
	dbutil "samuellando.com/YNAFB/internal/db"
	"samuellando.com/YNAFB/internal/db/types"

	"github.com/pressly/goose/v3"
)

func setup(t *testing.T) (*sql.DB, *data.Queries, context.Context) {
	t.Helper()
	goose.SetLogger(goose.NopLogger())
	db, err := dbutil.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	queries := data.New(db)
	ctx := context.Background()
	return db, queries, ctx
}

func teardown(db *sql.DB) {
	db.Close()
}

func mustTime(t *testing.T, y int, m time.Month, d int) types.UnixTime {
	t.Helper()
	return types.UnixTime{Time: time.Date(y, m, d, 0, 0, 0, 0, time.UTC)}
}

func newLogin(t *testing.T, queries *data.Queries, ctx context.Context, username string) data.Login {
	t.Helper()
	l, err := queries.CreateLogin(ctx, data.CreateLoginParams{Username: username, Password: "pw"})
	if err != nil {
		t.Fatal(err)
	}
	return l
}

func newBudget(t *testing.T, queries *data.Queries, ctx context.Context, name string) data.Budget {
	t.Helper()
	return newBudgetForLogin(t, queries, ctx, newLogin(t, queries, ctx, name).ID, name)
}

func newBudgetForLogin(t *testing.T, queries *data.Queries, ctx context.Context, loginID int64, name string) data.Budget {
	t.Helper()
	b, err := queries.CreateBudget(ctx, data.CreateBudgetParams{LoginID: loginID, Name: name})
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func newAccount(t *testing.T, queries *data.Queries, ctx context.Context, budget data.Budget, name string) data.Account {
	t.Helper()
	a, err := queries.CreateAccount(ctx, data.CreateAccountParams{BudgetID: budget.ID, LoginID: budget.LoginID, Name: name})
	if err != nil {
		t.Fatal(err)
	}
	return data.Account{ID: a.ID, BudgetID: a.BudgetID, Name: a.Name}
}

func newPayee(t *testing.T, queries *data.Queries, ctx context.Context, budget data.Budget, name string) data.Payee {
	t.Helper()
	p, err := queries.CreatePayee(ctx, data.CreatePayeeParams{BudgetID: budget.ID, LoginID: budget.LoginID, Name: name})
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func newCategory(t *testing.T, queries *data.Queries, ctx context.Context, budget data.Budget, name string) data.Category {
	t.Helper()
	c, err := queries.CreateCategory(ctx, data.CreateCategoryParams{BudgetID: budget.ID, LoginID: budget.LoginID, Name: name})
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func newCategoryGroup(t *testing.T, queries *data.Queries, ctx context.Context, budget data.Budget, name string) int64 {
	t.Helper()
	id, err := queries.CreateCategoryGroup(ctx, data.CreateCategoryGroupParams{BudgetID: budget.ID, LoginID: budget.LoginID, Name: name})
	if err != nil {
		t.Fatal(err)
	}
	return id
}

func newGoal(t *testing.T, queries *data.Queries, ctx context.Context, budget data.Budget, category int64, goalType string, start types.UnixTime, end types.NullUnixTime, amount int64) data.Goal {
	t.Helper()
	g, err := queries.CreateGoal(ctx, data.CreateGoalParams{
		BudgetID:   budget.ID,
		LoginID:    budget.LoginID,
		Type:       goalType,
		StartDate:  start,
		EndDate:    end,
		CategoryID: category,
		Amount:     amount,
	})
	if err != nil {
		t.Fatal(err)
	}
	return g
}

func newAllocation(t *testing.T, queries *data.Queries, ctx context.Context, loginID, budget, category int64, month types.UnixTime, amount int64) data.Allocation {
	t.Helper()
	a, err := queries.SetAllocation(ctx, data.SetAllocationParams{
		BudgetID:   budget,
		CategoryID: category,
		Month:      month,
		Amount:     amount,
		LoginID:    loginID,
	})
	if err != nil {
		t.Fatal(err)
	}
	return a
}

func newTrx(t *testing.T, queries *data.Queries, ctx context.Context, budget data.Budget, account data.Account, payee data.Payee, date types.UnixTime, out, in int64, note string) data.Trx {
	t.Helper()
	tx, err := queries.CreateTrx(ctx, data.CreateTrxParams{
		BudgetID:     budget.ID,
		LoginID:      budget.LoginID,
		AccountID:    account.ID,
		PayeeID:      payee.ID,
		Date:         date,
		TotalOutflow: out,
		TotalInflow:  in,
		Note:         note,
	})
	if err != nil {
		t.Fatal(err)
	}
	return tx
}

func newCategoryLine(t *testing.T, queries *data.Queries, ctx context.Context, budget data.Budget, tx data.Trx, category data.Category, out, in int64) data.TrxLine {
	t.Helper()
	tc, err := queries.CreateTrxLine(ctx, data.CreateTrxLineParams{
		BudgetID:   budget.ID,
		LoginID:    budget.LoginID,
		TrxID:      tx.ID,
		CategoryID: sql.NullInt64{Int64: category.ID, Valid: true},
		Outflow:    out,
		Inflow:     in,
	})
	if err != nil {
		t.Fatal(err)
	}
	return tc
}

func newTransfer(t *testing.T, queries *data.Queries, ctx context.Context, budget data.Budget, tx data.Trx, otherAccount data.Account, out, in int64) data.TrxLine {
	t.Helper()
	tc, err := queries.CreateTrxLine(ctx, data.CreateTrxLineParams{
		BudgetID:      budget.ID,
		LoginID:       budget.LoginID,
		TrxID:         tx.ID,
		DestAccountID: sql.NullInt64{Int64: otherAccount.ID, Valid: true},
		Outflow:       out,
		Inflow:        in,
	})
	if err != nil {
		t.Fatal(err)
	}
	return tc
}

func newIncomeLine(t *testing.T, queries *data.Queries, ctx context.Context, budget data.Budget, tx data.Trx, in int64) data.TrxLine {
	t.Helper()
	tc, err := queries.CreateTrxLine(ctx, data.CreateTrxLineParams{
		BudgetID: budget.ID,
		LoginID:  budget.LoginID,
		TrxID:    tx.ID,
		Income:   true,
		Inflow:   in,
	})
	if err != nil {
		t.Fatal(err)
	}
	return tc
}

func newCategoryDefault(t *testing.T, queries *data.Queries, ctx context.Context, budget data.Budget, payee data.Payee, category data.Category, percent int64) data.PayeeDefaultLine {
	t.Helper()
	pdl, err := queries.CreatePayeeDefaultLine(ctx, data.CreatePayeeDefaultLineParams{
		BudgetID:   budget.ID,
		LoginID:    budget.LoginID,
		PayeeID:    payee.ID,
		CategoryID: sql.NullInt64{Int64: category.ID, Valid: true},
		Percent:    percent,
	})
	if err != nil {
		t.Fatal(err)
	}
	return pdl
}

func newTransferDefault(t *testing.T, queries *data.Queries, ctx context.Context, budget data.Budget, payee data.Payee, otherAccount data.Account, percent int64) data.PayeeDefaultLine {
	t.Helper()
	pdl, err := queries.CreatePayeeDefaultLine(ctx, data.CreatePayeeDefaultLineParams{
		BudgetID:      budget.ID,
		LoginID:       budget.LoginID,
		PayeeID:       payee.ID,
		DestAccountID: sql.NullInt64{Int64: otherAccount.ID, Valid: true},
		Percent:       percent,
	})
	if err != nil {
		t.Fatal(err)
	}
	return pdl
}

func newLoginBudget(t *testing.T, queries *data.Queries, ctx context.Context, username string) (int64, data.Budget) {
	t.Helper()
	login := newLogin(t, queries, ctx, username)
	return login.ID, newBudgetForLogin(t, queries, ctx, login.ID, username)
}
