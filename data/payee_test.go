package data_test

import (
	"database/sql"
	"testing"

	"samuellando.com/YNAFB/data"
)

func TestCreatePayee(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	payee, err := queries.CreatePayee(ctx, data.CreatePayeeParams{
		LoginID:  budget.LoginID,
		BudgetID: budget.ID,
		Name:     "testpayee",
	})
	if err != nil {
		t.Fatal(err)
	}
	if payee.BudgetID != budget.ID {
		t.Error("budget id does not match")
	}
	if payee.Name != "testpayee" {
		t.Error("payee name does not match")
	}
	if payee.ID != 1 {
		t.Error("ID of the first payee should be 1")
	}
}

func TestCreatePayeeNonExistingBudget(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	_, err := queries.CreatePayee(ctx, data.CreatePayeeParams{
		BudgetID: 1,
		Name:     "testpayee",
	})
	if err == nil {
		t.Error("Non existing budget should raise an error")
	}
}

func TestCreatePayeeEmptyName(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	_, err := queries.CreatePayee(ctx, data.CreatePayeeParams{
		LoginID:  budget.LoginID,
		BudgetID: budget.ID,
		Name:     "",
	})
	if err == nil {
		t.Error("Empty payee name should raise an error")
	}
}

func TestCreatePayeeDuplicateNameConstraint(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	budget2 := newBudget(t, queries, ctx, "testBudget2")
	if _, err := queries.CreatePayee(ctx, data.CreatePayeeParams{
		LoginID:  budget.LoginID,
		BudgetID: budget.ID,
		Name:     "testpayee",
	}); err != nil {
		t.Fatal(err)
	}
	_, err := queries.CreatePayee(ctx, data.CreatePayeeParams{
		LoginID:  budget.LoginID,
		BudgetID: budget.ID,
		Name:     "testpayee",
	})
	if err == nil {
		t.Error("Should get an error for duplicate payee name in same budget")
	}
	_, err = queries.CreatePayee(ctx, data.CreatePayeeParams{
		LoginID:  budget2.LoginID,
		BudgetID: budget2.ID,
		Name:     "testpayee",
	})
	if err != nil {
		t.Error("Should not get an error for duplicate payee names across budgets")
	}
}

func TestUpdatePayee(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	payee := newPayee(t, queries, ctx, budget, "testpayee")
	updated, err := queries.UpdatePayee(ctx, data.UpdatePayeeParams{
		LoginID:  budget.LoginID,
		Name:     "newName",
		ID:       payee.ID,
		BudgetID: budget.ID,
	})
	if err != nil {
		t.Fatal(err)
	}
	if updated.ID != payee.ID {
		t.Error("ID changed on update")
	}
	if updated.Name != "newName" {
		t.Error("payee name was not updated")
	}
	newNamePayee, err := queries.GetPayeeByName(ctx, data.GetPayeeByNameParams{
		LoginID:  budget.LoginID,
		Name:     "newName",
		BudgetID: budget.ID,
	})
	if err != nil {
		t.Fatal(err)
	}
	if payee.ID != newNamePayee.ID {
		t.Error("ID changed on update")
	}
}

func TestDeletePayee(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	payee := newPayee(t, queries, ctx, budget, "testpayee")
	payees, err := queries.ListPayees(ctx, data.ListPayeesParams{LoginID: budget.LoginID, BudgetID: budget.ID})
	if err != nil {
		t.Fatal(err)
	}
	if len(payees) != 1 {
		t.Fatal("There should be one payee before")
	}
	err = queries.DeletePayee(ctx, data.DeletePayeeParams{ID: payee.ID, BudgetID: budget.ID, LoginID: budget.LoginID})
	if err != nil {
		t.Fatal(err)
	}
	payees, err = queries.ListPayees(ctx, data.ListPayeesParams{LoginID: budget.LoginID, BudgetID: budget.ID})
	if err != nil {
		t.Fatal(err)
	}
	if len(payees) != 0 {
		t.Fatal("There should be no payee after")
	}
}

func TestDeleteBudgetCascadesPayees(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	payee := newPayee(t, queries, ctx, budget, "testpayee")
	payees, err := queries.ListPayees(ctx, data.ListPayeesParams{LoginID: budget.LoginID, BudgetID: budget.ID})
	if err != nil {
		t.Fatal(err)
	}
	if len(payees) != 1 {
		t.Fatal("There should be one payee before")
	}
	err = queries.DeleteBudget(ctx, data.DeleteBudgetParams{LoginID: budget.LoginID, ID: budget.ID})
	if err != nil {
		t.Fatal(err)
	}
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM payee WHERE id = ?`, payee.ID).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatal("Deleting a budget should cascade delete its payees")
	}
}

func TestListPayees(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	budget2 := newBudget(t, queries, ctx, "testBudget2")
	newPayee(t, queries, ctx, budget, "payeeB")
	newPayee(t, queries, ctx, budget, "payeeA")
	newPayee(t, queries, ctx, budget2, "otherBudgetPayee")
	payees, err := queries.ListPayees(ctx, data.ListPayeesParams{LoginID: budget.LoginID, BudgetID: budget.ID})
	if err != nil {
		t.Fatal(err)
	}
	if len(payees) != 2 {
		t.Fatal("There should be two payees in the budget")
	}
	if payees[0].Name != "payeeA" {
		t.Error("First payee should be payeeA (ordered by name)")
	}
	if payees[1].Name != "payeeB" {
		t.Error("Second payee should be payeeB (ordered by name)")
	}
}

func TestGetPayeeByName(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	payee := newPayee(t, queries, ctx, budget, "testpayee")
	namePayee, err := queries.GetPayeeByName(ctx, data.GetPayeeByNameParams{
		LoginID:  budget.LoginID,
		Name:     "testpayee",
		BudgetID: budget.ID,
	})
	if err != nil {
		t.Fatal(err)
	}
	if payee.ID != namePayee.ID {
		t.Error("getting payee by name, id does not match")
	}
	if namePayee.BudgetID != budget.ID {
		t.Error("getting payee by name, budget id does not match")
	}
	if namePayee.Name != "testpayee" {
		t.Error("getting payee by name, name does not match")
	}
}

func TestGetPayeeByNameDoesNotExist(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	_, err := queries.GetPayeeByName(ctx, data.GetPayeeByNameParams{
		LoginID:  budget.LoginID,
		Name:     "testpayee",
		BudgetID: budget.ID,
	})
	if err == nil {
		t.Fatal("Getting non existent payee by name should fail")
	}
}

func TestScopingPayeeScopedByLogin(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budgetA := newBudget(t, queries, ctx, "budgetA")
	budgetB := newBudget(t, queries, ctx, "budgetB")
	newPayee(t, queries, ctx, budgetB, "payeeB")

	rows, err := queries.ListPayees(ctx, data.ListPayeesParams{
		LoginID:  budgetA.LoginID,
		BudgetID: budgetB.ID,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 0 {
		t.Fatalf("expected 0 payees for another login's budget, got %d", len(rows))
	}
}

func TestScopingCreatePayeeScopedByLogin(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budgetA := newBudget(t, queries, ctx, "budgetA")
	budgetB := newBudget(t, queries, ctx, "budgetB")

	_, err := queries.CreatePayee(ctx, data.CreatePayeeParams{
		LoginID:  budgetA.LoginID,
		BudgetID: budgetB.ID,
		Name:     "sneaky",
	})
	if err == nil {
		t.Fatal("expected creating a payee for another login's budget to fail")
	}
}

func TestScopingDeletePayeeScopedByLogin(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budgetA := newBudget(t, queries, ctx, "budgetA")
	budgetB := newBudget(t, queries, ctx, "budgetB")
	payeeB := newPayee(t, queries, ctx, budgetB, "payeeB")

	err := queries.DeletePayee(ctx, data.DeletePayeeParams{
		ID:       payeeB.ID,
		BudgetID: budgetB.ID,
		LoginID:  budgetA.LoginID,
	})
	if err != nil {
		t.Fatal(err)
	}
	payees, err := queries.ListPayees(ctx, data.ListPayeesParams{LoginID: budgetB.LoginID, BudgetID: budgetB.ID})
	if err != nil {
		t.Fatal(err)
	}
	if len(payees) != 1 {
		t.Fatalf("expected payeeB to survive a cross-login delete, got %d payees", len(payees))
	}
}

func TestScopingGetPayeeByNameScopedByLogin(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budgetA := newBudget(t, queries, ctx, "budgetA")
	budgetB := newBudget(t, queries, ctx, "budgetB")
	newPayee(t, queries, ctx, budgetB, "shared")

	_, err := queries.GetPayeeByName(ctx, data.GetPayeeByNameParams{
		LoginID:  budgetA.LoginID,
		BudgetID: budgetB.ID,
		Name:     "shared",
	})
	if err != sql.ErrNoRows {
		t.Fatalf("expected sql.ErrNoRows for another login's payee, got %v", err)
	}
}

func TestScopingUpdatePayeeScopedByLogin(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budgetA := newBudget(t, queries, ctx, "budgetA")
	budgetB := newBudget(t, queries, ctx, "budgetB")
	payeeB := newPayee(t, queries, ctx, budgetB, "payeeB")

	_, err := queries.UpdatePayee(ctx, data.UpdatePayeeParams{
		Name:     "hacked",
		ID:       payeeB.ID,
		BudgetID: budgetB.ID,
		LoginID:  budgetA.LoginID,
	})
	if err != sql.ErrNoRows {
		t.Fatalf("expected sql.ErrNoRows when updating across logins, got %v", err)
	}
}
