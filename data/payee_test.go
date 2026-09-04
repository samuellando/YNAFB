package data_test

import (
	"testing"

	"samuellando.com/YNAFB/data"
)

func TestCreatePayee(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget, err := queries.CreateBudget(ctx, "testBudget")
	if err != nil {
		t.Fatal(err)
	}
	payee, err := queries.CreatePayee(ctx, data.CreatePayeeParams{
		Budget: budget.ID,
		Name:   "testpayee",
	})
	if err != nil {
		t.Fatal(err)
	}
	if payee.Budget != budget.ID {
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
		Budget: 1,
		Name:   "testpayee",
	})
	if err == nil {
		t.Error("Non existing budget should raise an error")
	}
}

func TestCreatePayeeEmptyName(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget, err := queries.CreateBudget(ctx, "testBudget")
	if err != nil {
		t.Fatal(err)
	}
	_, err = queries.CreatePayee(ctx, data.CreatePayeeParams{
		Budget: budget.ID,
		Name:   "",
	})
	if err == nil {
		t.Error("Empty payee name should raise an error")
	}
}

func TestCreatePayeeDuplicateNameConstraint(t *testing.T) {
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
	_, err = queries.CreatePayee(ctx, data.CreatePayeeParams{
		Budget: budget.ID,
		Name:   "testpayee",
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = queries.CreatePayee(ctx, data.CreatePayeeParams{
		Budget: budget.ID,
		Name:   "testpayee",
	})
	if err == nil {
		t.Error("Should get an error for duplicate payee name in same budget")
	}
	_, err = queries.CreatePayee(ctx, data.CreatePayeeParams{
		Budget: budget2.ID,
		Name:   "testpayee",
	})
	if err != nil {
		t.Error("Should not get an error for duplicate payee names across budgets")
	}
}

func TestUpdatePayee(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget, err := queries.CreateBudget(ctx, "testBudget")
	if err != nil {
		t.Fatal(err)
	}
	payee, err := queries.CreatePayee(ctx, data.CreatePayeeParams{
		Budget: budget.ID,
		Name:   "testpayee",
	})
	if err != nil {
		t.Fatal(err)
	}
	n, err := queries.UpdatePayee(ctx, data.UpdatePayeeParams{
		Name: "newName",
		ID:   payee.ID,
	})
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Error("The number of affected rows should be 1")
	}
	newNamePayee, err := queries.GetPayeeByName(ctx, data.GetPayeeByNameParams{
		Name:   "newName",
		Budget: budget.ID,
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
	budget, err := queries.CreateBudget(ctx, "testBudget")
	if err != nil {
		t.Fatal(err)
	}
	payee, err := queries.CreatePayee(ctx, data.CreatePayeeParams{
		Budget: budget.ID,
		Name:   "testpayee",
	})
	if err != nil {
		t.Fatal(err)
	}
	payees, err := queries.ListPayees(ctx, budget.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(payees) != 1 {
		t.Fatal("There should be one payee before")
	}
	err = queries.DeletePayee(ctx, payee.ID)
	if err != nil {
		t.Fatal(err)
	}
	payees, err = queries.ListPayees(ctx, budget.ID)
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
	budget, err := queries.CreateBudget(ctx, "testBudget")
	if err != nil {
		t.Fatal(err)
	}
	payee, err := queries.CreatePayee(ctx, data.CreatePayeeParams{
		Budget: budget.ID,
		Name:   "testpayee",
	})
	if err != nil {
		t.Fatal(err)
	}
	payees, err := queries.ListPayees(ctx, budget.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(payees) != 1 {
		t.Fatal("There should be one payee before")
	}
	err = queries.DeleteBudget(ctx, budget.ID)
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
	budget, err := queries.CreateBudget(ctx, "testBudget")
	if err != nil {
		t.Fatal(err)
	}
	budget2, err := queries.CreateBudget(ctx, "testBudget2")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := queries.CreatePayee(ctx, data.CreatePayeeParams{
		Budget: budget.ID,
		Name:   "payeeB",
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := queries.CreatePayee(ctx, data.CreatePayeeParams{
		Budget: budget.ID,
		Name:   "payeeA",
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := queries.CreatePayee(ctx, data.CreatePayeeParams{
		Budget: budget2.ID,
		Name:   "otherBudgetPayee",
	}); err != nil {
		t.Fatal(err)
	}
	payees, err := queries.ListPayees(ctx, budget.ID)
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
	budget, err := queries.CreateBudget(ctx, "testBudget")
	if err != nil {
		t.Fatal(err)
	}
	payee, err := queries.CreatePayee(ctx, data.CreatePayeeParams{
		Budget: budget.ID,
		Name:   "testpayee",
	})
	if err != nil {
		t.Fatal(err)
	}
	namePayee, err := queries.GetPayeeByName(ctx, data.GetPayeeByNameParams{
		Name:   "testpayee",
		Budget: budget.ID,
	})
	if err != nil {
		t.Fatal(err)
	}
	if payee.ID != namePayee.ID {
		t.Error("getting payee by name, id does not match")
	}
	if namePayee.Budget != budget.ID {
		t.Error("getting payee by name, budget id does not match")
	}
	if namePayee.Name != "testpayee" {
		t.Error("getting payee by name, name does not match")
	}
}

func TestGetPayeeByNameDoesNotExist(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget, err := queries.CreateBudget(ctx, "testBudget")
	if err != nil {
		t.Fatal(err)
	}
	_, err = queries.GetPayeeByName(ctx, data.GetPayeeByNameParams{
		Name:   "testpayee",
		Budget: budget.ID,
	})
	if err == nil {
		t.Fatal("Getting non existent payee by name should fail")
	}
}
