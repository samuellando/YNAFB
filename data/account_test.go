package data_test

import (
	"database/sql"
	"testing"

	"samuellando.com/YNAFB/data"
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
	accounts, err := queries.ListAccountsBalances(ctx, data.ListAccountsBalancesParams{LoginID: budget.LoginID, BudgetID: budget.ID})
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
	accounts, err = queries.ListAccountsBalances(ctx, data.ListAccountsBalancesParams{LoginID: budget.LoginID, BudgetID: budget.ID})
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
	accounts, err := queries.ListAccountsBalances(ctx, data.ListAccountsBalancesParams{LoginID: budget.LoginID, BudgetID: budget.ID})
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
	accounts, err = queries.ListAccountsBalances(ctx, data.ListAccountsBalancesParams{LoginID: budget.LoginID, BudgetID: budget.ID})
	if err != nil {
		t.Fatal(err)
	}
	if len(accounts) != 0 {
		t.Fatal("There should be no account after")
	}
}

func TestGetAccountByName(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	account := newAccount(t, queries, ctx, budget, "testaccount")
	nameAccount, err := queries.GetAccountByName(ctx, data.GetAccountByNameParams{
		LoginID:  budget.LoginID,
		Name:     "testaccount",
		BudgetID: budget.ID,
	})
	if err != nil {
		t.Fatal(err)
	}
	if account.ID != nameAccount.ID {
		t.Error("getting account by name, id does not match")
	}
	if nameAccount.BudgetID != budget.ID {
		t.Error("getting account by name, budget id does not match")
	}
	if nameAccount.Name != "testaccount" {
		t.Error("getting account by name, name does not match")
	}
}

func TestGetAccountByNameDoesNotExist(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	_, err := queries.GetAccountByName(ctx, data.GetAccountByNameParams{
		LoginID:  budget.LoginID,
		Name:     "testaccount",
		BudgetID: budget.ID,
	})
	if err == nil {
		t.Fatal("Getting non existent account by name should fail")
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
	newNameAccount, err := queries.GetAccountByName(ctx, data.GetAccountByNameParams{
		LoginID:  budget.LoginID,
		Name:     "newName",
		BudgetID: budget.ID,
	})
	if err != nil {
		t.Fatal(err)
	}
	if account.ID != newNameAccount.ID {
		t.Error("ID changed on update")
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

func TestScopingGetAccountByNameScopedByLogin(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budgetA := newBudget(t, queries, ctx, "budgetA")
	budgetB := newBudget(t, queries, ctx, "budgetB")
	newAccount(t, queries, ctx, budgetB, "shared")

	_, err := queries.GetAccountByName(ctx, data.GetAccountByNameParams{
		LoginID:  budgetA.LoginID,
		BudgetID: budgetB.ID,
		Name:     "shared",
	})
	if err != sql.ErrNoRows {
		t.Fatalf("expected sql.ErrNoRows for another login's account, got %v", err)
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
	accounts, err := queries.ListAccountsBalances(ctx, data.ListAccountsBalancesParams{LoginID: budgetB.LoginID, BudgetID: budgetB.ID})
	if err != nil {
		t.Fatal(err)
	}
	if len(accounts) != 1 {
		t.Fatalf("expected accountB to survive a cross-login delete, got %d accounts", len(accounts))
	}
}

func TestScopingGetAccountBalancesScopedByLogin(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budgetA := newBudget(t, queries, ctx, "budgetA")
	budgetB := newBudget(t, queries, ctx, "budgetB")
	accountB := newAccount(t, queries, ctx, budgetB, "accountB")

	_, err := queries.GetAccountBalances(ctx, data.GetAccountBalancesParams{
		LoginID:  budgetA.LoginID,
		BudgetID: budgetB.ID,
		ID:       accountB.ID,
	})
	if err != sql.ErrNoRows {
		t.Fatalf("expected sql.ErrNoRows for another login's account balances, got %v", err)
	}
}

func TestScopingGetAccountBalanceAsOfScopedByLogin(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budgetA := newBudget(t, queries, ctx, "budgetA")
	budgetB := newBudget(t, queries, ctx, "budgetB")
	accountB := newAccount(t, queries, ctx, budgetB, "accountB")
	payeeB := newPayee(t, queries, ctx, budgetB, "payeeB")
	newTrx(t, queries, ctx, budgetB, accountB, payeeB, mustTime(t, 2026, 1, 1), 0, 8000, "")

	balance, err := queries.GetAccountBalanceAsOf(ctx, data.GetAccountBalanceAsOfParams{
		LoginID:  budgetA.LoginID,
		BudgetID: budgetB.ID,
		ID:       accountB.ID,
		Date:     mustTime(t, 2026, 2, 1),
	})
	if err != nil {
		t.Fatal(err)
	}
	if balance != 0 {
		t.Fatalf("expected 0 balance for another login's account, got %d", balance)
	}
}

func TestScopingListAccountTransactionsScopedByLogin(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budgetA := newBudget(t, queries, ctx, "budgetA")
	budgetB := newBudget(t, queries, ctx, "budgetB")
	accountB := newAccount(t, queries, ctx, budgetB, "accountB")
	payeeB := newPayee(t, queries, ctx, budgetB, "payeeB")
	newTrx(t, queries, ctx, budgetB, accountB, payeeB, mustTime(t, 2026, 1, 1), 1000, 0, "")

	rows, err := queries.ListAccountTransactions(ctx, data.ListAccountTransactionsParams{
		LoginID:  budgetA.LoginID,
		BudgetID: budgetB.ID,
		ID:       accountB.ID,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 0 {
		t.Fatalf("expected 0 transactions for another login's account, got %d", len(rows))
	}
}
