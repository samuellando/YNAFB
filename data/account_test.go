package data_test

import (
	"context"
	"database/sql"
	"samuellando.com/YNAFB/data"
	"samuellando.com/YNAFB/internal/db"
	"testing"

	"github.com/pressly/goose/v3"
)

func setup(t *testing.T) (*sql.DB, *data.Queries, context.Context) {
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

func TestCreateAccount(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget, err := queries.CreateBudget(ctx, "testBudget")
	if err != nil {
		t.Fatal(err)
	}
	account, err := queries.CreateAccount(ctx, data.CreateAccountParams{
		Budget: budget.ID,
		Name:   "testaccount",
	})
	if err != nil {
		t.Fatal(err)
	}
	if account.Budget != budget.ID {
		t.Error("budget id does not match")
	}
	if account.Name != "testaccount" {
		t.Error("account name does not match")
	}
	if account.ID != 1 {
		t.Error("ID of the first account should be 1")
	}
}

func TestCreateAccountNonExistingBudget(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	_, err := queries.CreateAccount(ctx, data.CreateAccountParams{
		Budget: 1,
		Name:   "testaccount",
	})
	if err == nil {
		t.Error("Non existing budget should raise an error")
	}
}

func TestCreateAccountEmptyName(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget, err := queries.CreateBudget(ctx, "testBudget")
	if err != nil {
		t.Fatal(err)
	}
	_, err = queries.CreateAccount(ctx, data.CreateAccountParams{
		Budget: budget.ID,
		Name:   "",
	})
	if err == nil {
		t.Error("Empty account name should raise an error")
	}
}

func TestCreateAccountDuplicateNameConstraint(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget, err := queries.CreateBudget(ctx, "testBudget")
	if err != nil {
		t.Fatal(err)
	}
	budget2, err := queries.CreateBudget(ctx, "testBudget2")
	if err != nil {
		t.Fatal(err)
	}
	_, err = queries.CreateAccount(ctx, data.CreateAccountParams{
		Budget: budget.ID,
		Name:   "testaccount",
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = queries.CreateAccount(ctx, data.CreateAccountParams{
		Budget: budget.ID,
		Name:   "testaccount",
	})
	if err == nil {
		t.Error("Should get an error for duplicate account name in same budget")
	}
	_, err = queries.CreateAccount(ctx, data.CreateAccountParams{
		Budget: budget2.ID,
		Name:   "testaccount",
	})
	if err != nil {
		t.Error("Should not get an error for duplicate account names accorss budgets")
	}
}

func TestDeleteAccount(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget, err := queries.CreateBudget(ctx, "testBudget")
	if err != nil {
		t.Fatal(err)
	}
	account, err := queries.CreateAccount(ctx, data.CreateAccountParams{
		Budget: budget.ID,
		Name:   "testaccount",
	})
	if err != nil {
		t.Fatal(err)
	}
	accounts, err := queries.ListAccountsBalances(ctx, budget.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(accounts) != 1 {
		t.Fatal("There should be one account before")
	}
	err = queries.DeleteAccount(ctx, account.ID)
	if err != nil {
		t.Fatal(err)
	}
	accounts, err = queries.ListAccountsBalances(ctx, budget.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(accounts) != 0 {
		t.Fatal("There should be no account after")
	}
}

func TestDeleteBudgetCascades(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget, err := queries.CreateBudget(ctx, "testBudget")
	if err != nil {
		t.Fatal(err)
	}
	account, err := queries.CreateAccount(ctx, data.CreateAccountParams{
		Budget: budget.ID,
		Name:   "testaccount",
	})
	if err != nil {
		t.Fatal(err)
	}
	accounts, err := queries.ListAccountsBalances(ctx, budget.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(accounts) != 1 {
		t.Fatal("There should be one account before")
	}
	err = queries.DeleteBudget(ctx, account.ID)
	if err != nil {
		t.Fatal(err)
	}
	accounts, err = queries.ListAccountsBalances(ctx, budget.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(accounts) != 0 {
		t.Fatal("There should be no account after")
	}
}

func TestGetAccountByNamwe(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget, err := queries.CreateBudget(ctx, "testBudget")
	if err != nil {
		t.Fatal(err)
	}
	account, err := queries.CreateAccount(ctx, data.CreateAccountParams{
		Budget: budget.ID,
		Name:   "testaccount",
	})
	if err != nil {
		t.Fatal(err)
	}
	nameAccount, err := queries.GetAccountByName(ctx, data.GetAccountByNameParams{
		Name:   "testaccount",
		Budget: budget.ID,
	})
	if err != nil {
		t.Fatal(err)
	}
	if account.ID != nameAccount.ID {
		t.Error("getting accoun t by name, id does not match")
	}
	if nameAccount.Budget != budget.ID {
		t.Error("getting accoun t by name, budget id does not match")
	}
	if nameAccount.Name != "testaccount" {
		t.Error("getting accoun t by name, name does not match")
	}
}

func TestGetAccountByNamweDoesNotExist(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget, err := queries.CreateBudget(ctx, "testBudget")
	if err != nil {
		t.Fatal(err)
	}
	_, err = queries.GetAccountByName(ctx, data.GetAccountByNameParams{
		Name:   "testaccount",
		Budget: budget.ID,
	})
	if err == nil {
		t.Fatal("Getting non existent account by name should fail")
	}
}

func TestUpodateAccount(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget, err := queries.CreateBudget(ctx, "testBudget")
	if err != nil {
		t.Fatal(err)
	}
	account, err := queries.CreateAccount(ctx, data.CreateAccountParams{
		Budget: budget.ID,
		Name:   "testaccount",
	})
	if err != nil {
		t.Fatal(err)
	}
	n, err := queries.UpdateAccount(ctx, data.UpdateAccountParams{
		Name: "newName",
		ID:   account.ID,
	})
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Error("The number of affected rows should be 1")
	}
	newNameAccount, err := queries.GetAccountByName(ctx, data.GetAccountByNameParams{
		Name:   "newName",
		Budget: budget.ID,
	})
	if err != nil {
		t.Fatal(err)
	}
	if account.ID != newNameAccount.ID {
		t.Error("ID changed on update")
	}
}
