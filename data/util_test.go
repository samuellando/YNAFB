package data_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"samuellando.com/YNAFB/data"
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

func newBudget(t *testing.T, queries *data.Queries, ctx context.Context, name string) data.Budget {
	t.Helper()
	b, err := queries.CreateBudget(ctx, name)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func newAccount(t *testing.T, queries *data.Queries, ctx context.Context, budget int64, name string) data.Account {
	t.Helper()
	a, err := queries.CreateAccount(ctx, data.CreateAccountParams{Budget: budget, Name: name})
	if err != nil {
		t.Fatal(err)
	}
	return a
}

func newPayee(t *testing.T, queries *data.Queries, ctx context.Context, budget int64, name string) data.Payee {
	t.Helper()
	p, err := queries.CreatePayee(ctx, data.CreatePayeeParams{Budget: budget, Name: name})
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func newCategory(t *testing.T, queries *data.Queries, ctx context.Context, budget int64, name string) data.Category {
	t.Helper()
	c, err := queries.CreateCategory(ctx, data.CreateCategoryParams{Budget: budget, Name: name})
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func newTransaction(t *testing.T, queries *data.Queries, ctx context.Context, account, payee int64, date types.UnixTime, out, in int64, note string) data.Transaction {
	t.Helper()
	tx, err := queries.CreateTransaction(ctx, data.CreateTransactionParams{
		Date:         date,
		Account:      account,
		Payee:        payee,
		TotalOutflow: out,
		TotalInflow:  in,
		Note:         note,
	})
	if err != nil {
		t.Fatal(err)
	}
	return tx
}

func newCategoryLine(t *testing.T, queries *data.Queries, ctx context.Context, tx int64, category int64, out, in int64) data.TransactionCategory {
	t.Helper()
	tc, err := queries.CreateTransactionCategory(ctx, data.CreateTransactionCategoryParams{
		Transaction: tx,
		Category:    sql.NullInt64{Int64: category, Valid: true},
		Outflow:     out,
		Inflow:      in,
	})
	if err != nil {
		t.Fatal(err)
	}
	return tc
}

func newTransfer(t *testing.T, queries *data.Queries, ctx context.Context, tx int64, otherAccount int64, out, in int64) data.TransactionCategory {
	t.Helper()
	tc, err := queries.CreateTransactionCategory(ctx, data.CreateTransactionCategoryParams{
		Transaction:  tx,
		OtherAccount: sql.NullInt64{Int64: otherAccount, Valid: true},
		Outflow:      out,
		Inflow:       in,
	})
	if err != nil {
		t.Fatal(err)
	}
	return tc
}

func newIncomeLine(t *testing.T, queries *data.Queries, ctx context.Context, tx int64, in int64) data.TransactionCategory {
	t.Helper()
	tc, err := queries.CreateTransactionCategory(ctx, data.CreateTransactionCategoryParams{
		Transaction: tx,
		Income:      true,
		Inflow:      in,
	})
	if err != nil {
		t.Fatal(err)
	}
	return tc
}

func listAccountTransactions(t *testing.T, queries *data.Queries, ctx context.Context, accountID int64) []data.ListAccountTransactionsRow {
	t.Helper()
	rows, err := queries.ListAccountTransactions(ctx, accountID)
	if err != nil {
		t.Fatal(err)
	}
	return rows
}