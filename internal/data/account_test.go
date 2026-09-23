package data_test

import (
	"database/sql"
	"testing"

	"samuellando.com/YNAFB/internal/data"
)

func TestCreateAccount(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	account, err := queries.CreateAccount(ctx, data.CreateAccountParams{
		LoginID:  budget.LoginID,
		BudgetID: budget.ID,
		Name:     "testaccount",
	})
	if err != nil {
		t.Fatal(err)
	}
	if account.BudgetID != budget.ID {
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
		BudgetID: 1,
		Name:     "testaccount",
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
		LoginID:  budget.LoginID,
		BudgetID: budget.ID,
		Name:     "",
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
	if _, err := queries.CreateAccount(ctx, data.CreateAccountParams{
		LoginID:  budget.LoginID,
		BudgetID: budget.ID,
		Name:     "testaccount",
	}); err != nil {
		t.Fatal(err)
	}
	_, err := queries.CreateAccount(ctx, data.CreateAccountParams{
		LoginID:  budget.LoginID,
		BudgetID: budget.ID,
		Name:     "testaccount",
	})
	if err == nil {
		t.Error("Should get an error for duplicate account name in same budget")
	}
	_, err = queries.CreateAccount(ctx, data.CreateAccountParams{
		LoginID:  budget2.LoginID,
		BudgetID: budget2.ID,
		Name:     "testaccount",
	})
	if err != nil {
		t.Error("Should not get an error for duplicate account names across budgets")
	}
}

func TestDeleteAccount(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	account := newAccount(t, queries, ctx, budget, "testaccount")
	accounts, err := queries.ListAccounts(ctx, data.ListAccountsParams{LoginID: budget.LoginID, BudgetID: budget.ID})
	if err != nil {
		t.Fatal(err)
	}
	if len(accounts) != 1 {
		t.Fatal("There should be one account before")
	}
	err = queries.DeleteAccount(ctx, data.DeleteAccountParams{ID: account.ID, BudgetID: budget.ID, LoginID: budget.LoginID})
	if err != nil {
		t.Fatal(err)
	}
	accounts, err = queries.ListAccounts(ctx, data.ListAccountsParams{LoginID: budget.LoginID, BudgetID: budget.ID})
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
	newAccount(t, queries, ctx, budget, "testaccount")
	accounts, err := queries.ListAccounts(ctx, data.ListAccountsParams{LoginID: budget.LoginID, BudgetID: budget.ID})
	if err != nil {
		t.Fatal(err)
	}
	if len(accounts) != 1 {
		t.Fatal("There should be one account before")
	}
	err = queries.DeleteBudget(ctx, data.DeleteBudgetParams{LoginID: budget.LoginID, ID: budget.ID})
	if err != nil {
		t.Fatal(err)
	}
	accounts, err = queries.ListAccounts(ctx, data.ListAccountsParams{LoginID: budget.LoginID, BudgetID: budget.ID})
	if err != nil {
		t.Fatal(err)
	}
	if len(accounts) != 0 {
		t.Fatal("There should be no account after")
	}
}

func TestUpdateAccount(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	account := newAccount(t, queries, ctx, budget, "testaccount")
	updated, err := queries.UpdateAccount(ctx, data.UpdateAccountParams{
		LoginID:  budget.LoginID,
		Name:     "newName",
		ID:       account.ID,
		BudgetID: budget.ID,
	})
	if err != nil {
		t.Fatal(err)
	}
	if updated.ID != account.ID {
		t.Error("ID changed on update")
	}
	if updated.Name != "newName" {
		t.Error("account name was not updated")
	}
}

func TestUpdateAccountWrongBudgetAffectsNothing(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	otherBudget := newBudget(t, queries, ctx, "otherBudget")
	account := newAccount(t, queries, ctx, budget, "testaccount")
	_, err := queries.UpdateAccount(ctx, data.UpdateAccountParams{
		LoginID:  otherBudget.LoginID,
		Name:     "newName",
		ID:       account.ID,
		BudgetID: otherBudget.ID,
	})
	if err != sql.ErrNoRows {
		t.Fatalf("Updating an account with a mismatched budget should return sql.ErrNoRows, got %v", err)
	}
}

func TestScopingCreateAccountUsesLoginBudget(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budgetA := newBudget(t, queries, ctx, "budgetA")
	budgetB := newBudget(t, queries, ctx, "budgetB")

	// Creating an account for budgetB's id while passing budgetA's login_id must
	// fail (no budget matches both login_id and id).
	_, err := queries.CreateAccount(ctx, data.CreateAccountParams{
		BudgetID: budgetB.ID,
		LoginID:  budgetA.LoginID,
		Name:     "sneaky",
	})
	if err == nil {
		t.Fatal("expected creating an account for another login's budget to fail")
	}
}

func TestScopingUpdateAccountScopedByLogin(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budgetA := newBudget(t, queries, ctx, "budgetA")
	budgetB := newBudget(t, queries, ctx, "budgetB")
	accountB := newAccount(t, queries, ctx, budgetB, "accountB")

	// Attempting to update budgetB's account using budgetA's login must be a no-op.
	_, err := queries.UpdateAccount(ctx, data.UpdateAccountParams{
		Name:     "hacked",
		ID:       accountB.ID,
		LoginID:  budgetA.LoginID,
		BudgetID: budgetB.ID,
	})
	if err != sql.ErrNoRows {
		t.Fatalf("expected sql.ErrNoRows when updating across logins, got %v", err)
	}
}

func TestScopingDeleteAccountScopedByLogin(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budgetA := newBudget(t, queries, ctx, "budgetA")
	budgetB := newBudget(t, queries, ctx, "budgetB")
	newAccount(t, queries, ctx, budgetA, "accountA")
	accountB := newAccount(t, queries, ctx, budgetB, "accountB")

	err := queries.DeleteAccount(ctx, data.DeleteAccountParams{
		ID:       accountB.ID,
		BudgetID: budgetB.ID,
		LoginID:  budgetA.LoginID,
	})
	if err != nil {
		t.Fatal(err)
	}
	accounts, err := queries.ListAccounts(ctx, data.ListAccountsParams{LoginID: budgetB.LoginID, BudgetID: budgetB.ID})
	if err != nil {
		t.Fatal(err)
	}
	if len(accounts) != 1 {
		t.Fatalf("expected accountB to survive a cross-login delete, got %d accounts", len(accounts))
	}
}

func TestGetAccount(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	account := newAccount(t, queries, ctx, budget, "testaccount")
	got, err := queries.GetAccount(ctx, data.GetAccountParams{
		LoginID:  budget.LoginID,
		BudgetID: budget.ID,
		ID:       account.ID,
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != account.ID {
		t.Error("getting account, id does not match")
	}
	if got.BudgetID != budget.ID {
		t.Error("getting account, budget id does not match")
	}
	if got.Name != "testaccount" {
		t.Error("getting account, name does not match")
	}
	if got.LoginID != budget.LoginID {
		t.Error("getting account, login id does not match")
	}
	if got.BudgetName != "testBudget" {
		t.Error("getting account, budget name does not match")
	}
}

func TestGetAccountDoesNotExist(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	_, err := queries.GetAccount(ctx, data.GetAccountParams{
		LoginID:  budget.LoginID,
		BudgetID: budget.ID,
		ID:       99,
	})
	if err != sql.ErrNoRows {
		t.Fatalf("Getting non existent account should return sql.ErrNoRows, got %v", err)
	}
}

func TestListAccounts(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	budget2 := newBudget(t, queries, ctx, "testBudget2")
	newAccount(t, queries, ctx, budget, "accountB")
	newAccount(t, queries, ctx, budget, "accountA")
	newAccount(t, queries, ctx, budget2, "otherBudgetAccount")
	accounts, err := queries.ListAccounts(ctx, data.ListAccountsParams{LoginID: budget.LoginID, BudgetID: budget.ID})
	if err != nil {
		t.Fatal(err)
	}
	if len(accounts) != 2 {
		t.Fatal("There should be two accounts in the budget")
	}
	if accounts[0].Name != "accountA" {
		t.Error("First account should be accountA (ordered by name)")
	}
	if accounts[1].Name != "accountB" {
		t.Error("Second account should be accountB (ordered by name)")
	}
	if accounts[0].LoginID != budget.LoginID {
		t.Error("listed account login id does not match")
	}
	if accounts[0].BudgetName != "testBudget" {
		t.Error("listed account budget name does not match")
	}
}

func TestScopingGetAccountScopedByLogin(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budgetA := newBudget(t, queries, ctx, "budgetA")
	budgetB := newBudget(t, queries, ctx, "budgetB")
	accountB := newAccount(t, queries, ctx, budgetB, "accountB")

	_, err := queries.GetAccount(ctx, data.GetAccountParams{
		LoginID:  budgetA.LoginID,
		BudgetID: budgetB.ID,
		ID:       accountB.ID,
	})
	if err != sql.ErrNoRows {
		t.Fatalf("expected sql.ErrNoRows for another login's account, got %v", err)
	}
}

func TestScopingListAccountsScopedByLogin(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budgetA := newBudget(t, queries, ctx, "budgetA")
	budgetB := newBudget(t, queries, ctx, "budgetB")
	newAccount(t, queries, ctx, budgetB, "accountB")

	rows, err := queries.ListAccounts(ctx, data.ListAccountsParams{
		LoginID:  budgetA.LoginID,
		BudgetID: budgetB.ID,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 0 {
		t.Fatalf("expected 0 accounts for another login's budget, got %d", len(rows))
	}
}
