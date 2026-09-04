package data_test

import (
	"samuellando.com/YNAFB/data"
	"testing"
)

func TestCreateAccount(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	account := newAccount(t, queries, ctx, budget.ID, "testaccount")
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
	budget := newBudget(t, queries, ctx, "testBudget")
	_, err := queries.CreateAccount(ctx, data.CreateAccountParams{
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
	budget := newBudget(t, queries, ctx, "testBudget")
	budget2 := newBudget(t, queries, ctx, "testBudget2")
	newAccount(t, queries, ctx, budget.ID, "testaccount")
	_, err := queries.CreateAccount(ctx, data.CreateAccountParams{
		Budget: budget.ID,
		Name:   "testaccount",
	})
	if err == nil {
		t.Error("Should get an error for duplicate account name in same budget")
	}
	newAccount(t, queries, ctx, budget2.ID, "testaccount")
}

func TestDeleteAccount(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	account := newAccount(t, queries, ctx, budget.ID, "testaccount")
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
	budget := newBudget(t, queries, ctx, "testBudget")
	account := newAccount(t, queries, ctx, budget.ID, "testaccount")
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
	budget := newBudget(t, queries, ctx, "testBudget")
	account := newAccount(t, queries, ctx, budget.ID, "testaccount")
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
	budget := newBudget(t, queries, ctx, "testBudget")
	_, err := queries.GetAccountByName(ctx, data.GetAccountByNameParams{
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
	budget := newBudget(t, queries, ctx, "testBudget")
	account := newAccount(t, queries, ctx, budget.ID, "testaccount")
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
