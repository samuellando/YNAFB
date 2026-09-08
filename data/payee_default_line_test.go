package data_test

import (
	"database/sql"
	"testing"

	"samuellando.com/YNAFB/data"
)

func TestCreatePayeeDefaultLineCategory(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	payee := newPayee(t, queries, ctx, budget, "testpayee")
	category := newCategory(t, queries, ctx, budget, "testcategory")
	defaultCategory, err := queries.CreatePayeeDefaultLine(ctx, data.CreatePayeeDefaultLineParams{
		BudgetID:      budget.ID,
		LoginID:       budget.LoginID,
		PayeeID:       payee.ID,
		DestAccountID: sql.NullInt64{},
		CategoryID:    sql.NullInt64{Int64: category.ID, Valid: true},
		Income:        false,
		Percent:       100,
	})
	if err != nil {
		t.Fatal(err)
	}
	if defaultCategory.PayeeID != payee.ID {
		t.Error("payee id does not match")
	}
	if defaultCategory.DestAccountID.Valid {
		t.Error("dest account should be null for a category default")
	}
	if !defaultCategory.CategoryID.Valid || defaultCategory.CategoryID.Int64 != category.ID {
		t.Error("category id does not match")
	}
	if defaultCategory.Income {
		t.Error("income should be false")
	}
	if defaultCategory.Percent != 100 {
		t.Error("percent does not match")
	}
	if defaultCategory.ID != 1 {
		t.Error("ID of the first payee default line should be 1")
	}
}

func TestCreatePayeeDefaultLineDestAccount(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	payee := newPayee(t, queries, ctx, budget, "testpayee")
	account := newAccount(t, queries, ctx, budget, "testaccount")
	defaultCategory, err := queries.CreatePayeeDefaultLine(ctx, data.CreatePayeeDefaultLineParams{
		BudgetID:      budget.ID,
		LoginID:       budget.LoginID,
		PayeeID:       payee.ID,
		DestAccountID: sql.NullInt64{Int64: account.ID, Valid: true},
		CategoryID:    sql.NullInt64{},
		Income:        false,
		Percent:       100,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !defaultCategory.DestAccountID.Valid || defaultCategory.DestAccountID.Int64 != account.ID {
		t.Error("dest account id does not match")
	}
	if defaultCategory.CategoryID.Valid {
		t.Error("category should be null for a transfer default")
	}
}

func TestCreatePayeeDefaultLineIncome(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	payee := newPayee(t, queries, ctx, budget, "testpayee")
	defaultCategory, err := queries.CreatePayeeDefaultLine(ctx, data.CreatePayeeDefaultLineParams{
		BudgetID:      budget.ID,
		LoginID:       budget.LoginID,
		PayeeID:       payee.ID,
		DestAccountID: sql.NullInt64{},
		CategoryID:    sql.NullInt64{},
		Income:        true,
		Percent:       100,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !defaultCategory.Income {
		t.Error("income should be true")
	}
	if defaultCategory.CategoryID.Valid {
		t.Error("category should be null for income")
	}
	if defaultCategory.DestAccountID.Valid {
		t.Error("dest account should be null for income")
	}
}

func TestCreatePayeeDefaultLineXorViolation(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	payee := newPayee(t, queries, ctx, budget, "testpayee")
	account := newAccount(t, queries, ctx, budget, "testaccount")
	category := newCategory(t, queries, ctx, budget, "testcategory")
	_, err := queries.CreatePayeeDefaultLine(ctx, data.CreatePayeeDefaultLineParams{
		BudgetID:      budget.ID,
		LoginID:       budget.LoginID,
		PayeeID:       payee.ID,
		DestAccountID: sql.NullInt64{Int64: account.ID, Valid: true},
		CategoryID:    sql.NullInt64{Int64: category.ID, Valid: true},
		Income:        false,
		Percent:       100,
	})
	if err == nil {
		t.Error("Setting both dest_account and category should raise an error")
	}
	_, err = queries.CreatePayeeDefaultLine(ctx, data.CreatePayeeDefaultLineParams{
		BudgetID:      budget.ID,
		LoginID:       budget.LoginID,
		PayeeID:       payee.ID,
		DestAccountID: sql.NullInt64{},
		CategoryID:    sql.NullInt64{},
		Income:        false,
		Percent:       100,
	})
	if err == nil {
		t.Error("Setting neither dest_account, category nor income should raise an error")
	}
}

func TestCreatePayeeDefaultLineNonExistingPayee(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	category := newCategory(t, queries, ctx, budget, "testcategory")
	_, err := queries.CreatePayeeDefaultLine(ctx, data.CreatePayeeDefaultLineParams{
		BudgetID:      budget.ID,
		LoginID:       budget.LoginID,
		PayeeID:       99,
		DestAccountID: sql.NullInt64{},
		CategoryID:    sql.NullInt64{Int64: category.ID, Valid: true},
		Income:        false,
		Percent:       100,
	})
	if err == nil {
		t.Error("Non existing payee should raise an error")
	}
}

func TestCreatePayeeDefaultLineNonExistingDestAccount(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	payee := newPayee(t, queries, ctx, budget, "testpayee")
	_, err := queries.CreatePayeeDefaultLine(ctx, data.CreatePayeeDefaultLineParams{
		BudgetID:      budget.ID,
		LoginID:       budget.LoginID,
		PayeeID:       payee.ID,
		DestAccountID: sql.NullInt64{Int64: 99, Valid: true},
		CategoryID:    sql.NullInt64{},
		Income:        false,
		Percent:       100,
	})
	if err == nil {
		t.Error("Non existing dest account should raise an error")
	}
}

func TestCreatePayeeDefaultLineNonExistingCategory(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	payee := newPayee(t, queries, ctx, budget, "testpayee")
	_, err := queries.CreatePayeeDefaultLine(ctx, data.CreatePayeeDefaultLineParams{
		BudgetID:      budget.ID,
		LoginID:       budget.LoginID,
		PayeeID:       payee.ID,
		DestAccountID: sql.NullInt64{},
		CategoryID:    sql.NullInt64{Int64: 99, Valid: true},
		Income:        false,
		Percent:       100,
	})
	if err == nil {
		t.Error("Non existing category should raise an error")
	}
}

func TestDeletePayeeCascadesPayeeDefaultLines(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	payee := newPayee(t, queries, ctx, budget, "testpayee")
	category := newCategory(t, queries, ctx, budget, "testcategory")
	defaultLine := newCategoryDefault(t, queries, ctx, budget, payee, category, 100)
	defaults, err := queries.ListPayeeDefaultLinesByPayee(ctx, data.ListPayeeDefaultLinesByPayeeParams{LoginID: budget.LoginID, PayeeID: payee.ID, BudgetID: budget.ID})
	if err != nil {
		t.Fatal(err)
	}
	if len(defaults) != 1 {
		t.Fatal("There should be one default line before")
	}
	err = queries.DeletePayee(ctx, data.DeletePayeeParams{ID: payee.ID, BudgetID: budget.ID, LoginID: budget.LoginID})
	if err != nil {
		t.Fatal(err)
	}
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM payee_default_line WHERE id = ?`, defaultLine.ID).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatal("Deleting a payee should cascade delete its default lines")
	}
}

func TestDeleteAccountCascadesPayeeDefaultLines(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	payee := newPayee(t, queries, ctx, budget, "testpayee")
	account := newAccount(t, queries, ctx, budget, "testaccount")
	newTransferDefault(t, queries, ctx, budget, payee, account, 100)
	err := queries.DeleteAccount(ctx, data.DeleteAccountParams{ID: account.ID, BudgetID: budget.ID, LoginID: budget.LoginID})
	if err != nil {
		t.Fatal(err)
	}
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM payee_default_line WHERE dest_account_id = ?`, account.ID).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatal("Deleting an account should cascade delete default lines referencing it")
	}
}

func TestDeleteCategoryCascadesPayeeDefaultLines(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	payee := newPayee(t, queries, ctx, budget, "testpayee")
	category := newCategory(t, queries, ctx, budget, "testcategory")
	newCategoryDefault(t, queries, ctx, budget, payee, category, 100)
	err := queries.DeleteCategory(ctx, data.DeleteCategoryParams{ID: category.ID, BudgetID: budget.ID, LoginID: budget.LoginID})
	if err != nil {
		t.Fatal(err)
	}
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM payee_default_line WHERE category_id = ?`, category.ID).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatal("Deleting a category should cascade delete default lines referencing it")
	}
}

func TestCreatePayeeDefaultLinePercentNegativeRejected(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	payee := newPayee(t, queries, ctx, budget, "testpayee")
	category := newCategory(t, queries, ctx, budget, "testcategory")
	_, err := queries.CreatePayeeDefaultLine(ctx, data.CreatePayeeDefaultLineParams{
		BudgetID:      budget.ID,
		LoginID:       budget.LoginID,
		PayeeID:       payee.ID,
		DestAccountID: sql.NullInt64{},
		CategoryID:    sql.NullInt64{Int64: category.ID, Valid: true},
		Income:        false,
		Percent:       -1,
	})
	if err == nil {
		t.Error("A negative percent should be rejected")
	}
}

func TestCreatePayeeDefaultLinePercentOver100Rejected(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	payee := newPayee(t, queries, ctx, budget, "testpayee")
	category := newCategory(t, queries, ctx, budget, "testcategory")
	_, err := queries.CreatePayeeDefaultLine(ctx, data.CreatePayeeDefaultLineParams{
		BudgetID:      budget.ID,
		LoginID:       budget.LoginID,
		PayeeID:       payee.ID,
		DestAccountID: sql.NullInt64{},
		CategoryID:    sql.NullInt64{Int64: category.ID, Valid: true},
		Income:        false,
		Percent:       101,
	})
	if err == nil {
		t.Error("A percent over 100 should be rejected")
	}
}

func TestCreatePayeeDefaultLinePercentZeroRejected(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	payee := newPayee(t, queries, ctx, budget, "testpayee")
	category := newCategory(t, queries, ctx, budget, "testcategory")
	_, err := queries.CreatePayeeDefaultLine(ctx, data.CreatePayeeDefaultLineParams{
		BudgetID:      budget.ID,
		LoginID:       budget.LoginID,
		PayeeID:       payee.ID,
		DestAccountID: sql.NullInt64{},
		CategoryID:    sql.NullInt64{Int64: category.ID, Valid: true},
		Income:        false,
		Percent:       0,
	})
	if err == nil {
		t.Error("A zero percent should be rejected")
	}
}

func TestUpdatePayeeDefaultLinePercentOver100Rejected(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	payee := newPayee(t, queries, ctx, budget, "testpayee")
	category := newCategory(t, queries, ctx, budget, "testcategory")
	defaultLine := newCategoryDefault(t, queries, ctx, budget, payee, category, 100)
	_, err := queries.UpdatePayeeDefaultLine(ctx, data.UpdatePayeeDefaultLineParams{
		PayeeID:       payee.ID,
		DestAccountID: sql.NullInt64{},
		CategoryID:    sql.NullInt64{Int64: category.ID, Valid: true},
		Income:        false,
		Percent:       101,
		ID:            defaultLine.ID,
		BudgetID:      budget.ID,
		LoginID:       budget.LoginID,
	})
	if err == nil {
		t.Error("Updating a percent over 100 should be rejected")
	}
}

func TestUpdatePayeeDefaultLine(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	payee := newPayee(t, queries, ctx, budget, "testpayee")
	category := newCategory(t, queries, ctx, budget, "testcategory")
	account := newAccount(t, queries, ctx, budget, "testaccount")
	defaultLine := newCategoryDefault(t, queries, ctx, budget, payee, category, 100)
	updated, err := queries.UpdatePayeeDefaultLine(ctx, data.UpdatePayeeDefaultLineParams{
		PayeeID:       payee.ID,
		DestAccountID: sql.NullInt64{Int64: account.ID, Valid: true},
		CategoryID:    sql.NullInt64{},
		Income:        false,
		Percent:       50,
		ID:            defaultLine.ID,
		BudgetID:      budget.ID,
		LoginID:       budget.LoginID,
	})
	if err != nil {
		t.Fatal(err)
	}
	if updated.ID != defaultLine.ID {
		t.Error("ID changed on update")
	}
	if !updated.DestAccountID.Valid || updated.DestAccountID.Int64 != account.ID {
		t.Error("payee default line was not updated to a transfer")
	}
	if updated.Percent != 50 {
		t.Error("payee default line percent was not updated")
	}
	var destAccount sql.NullInt64
	var percent int64
	if err := db.QueryRow(`SELECT dest_account_id, percent FROM payee_default_line WHERE id = ?`, defaultLine.ID).Scan(&destAccount, &percent); err != nil {
		t.Fatal(err)
	}
	if !destAccount.Valid || destAccount.Int64 != account.ID {
		t.Error("payee default line was not updated to a transfer")
	}
	if percent != 50 {
		t.Error("payee default line percent was not updated")
	}
}

func TestDeletePayeeDefaultLine(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	payee := newPayee(t, queries, ctx, budget, "testpayee")
	category := newCategory(t, queries, ctx, budget, "testcategory")
	defaultLine := newCategoryDefault(t, queries, ctx, budget, payee, category, 100)
	defaults, err := queries.ListPayeeDefaultLinesByPayee(ctx, data.ListPayeeDefaultLinesByPayeeParams{LoginID: budget.LoginID, PayeeID: payee.ID, BudgetID: budget.ID})
	if err != nil {
		t.Fatal(err)
	}
	if len(defaults) != 1 {
		t.Fatal("There should be one default line before")
	}
	err = queries.DeletePayeeDefaultLine(ctx, data.DeletePayeeDefaultLineParams{ID: defaultLine.ID, BudgetID: budget.ID, LoginID: budget.LoginID})
	if err != nil {
		t.Fatal(err)
	}
	defaults, err = queries.ListPayeeDefaultLinesByPayee(ctx, data.ListPayeeDefaultLinesByPayeeParams{LoginID: budget.LoginID, PayeeID: payee.ID, BudgetID: budget.ID})
	if err != nil {
		t.Fatal(err)
	}
	if len(defaults) != 0 {
		t.Fatal("There should be no default line after")
	}
}

func TestDeletePayeeDefaultLinesByPayee(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	payee := newPayee(t, queries, ctx, budget, "testpayee")
	category := newCategory(t, queries, ctx, budget, "testcategory")
	for i := 0; i < 2; i++ {
		newCategoryDefault(t, queries, ctx, budget, payee, category, 100)
	}
	defaults, err := queries.ListPayeeDefaultLinesByPayee(ctx, data.ListPayeeDefaultLinesByPayeeParams{LoginID: budget.LoginID, PayeeID: payee.ID, BudgetID: budget.ID})
	if err != nil {
		t.Fatal(err)
	}
	if len(defaults) != 2 {
		t.Fatal("There should be two default lines before")
	}
	err = queries.DeletePayeeDefaultLinesByPayee(ctx, data.DeletePayeeDefaultLinesByPayeeParams{LoginID: budget.LoginID, PayeeID: payee.ID, BudgetID: budget.ID})
	if err != nil {
		t.Fatal(err)
	}
	defaults, err = queries.ListPayeeDefaultLinesByPayee(ctx, data.ListPayeeDefaultLinesByPayeeParams{LoginID: budget.LoginID, PayeeID: payee.ID, BudgetID: budget.ID})
	if err != nil {
		t.Fatal(err)
	}
	if len(defaults) != 0 {
		t.Fatal("There should be no default line after")
	}
}

func TestListPayeeDefaultLinesByPayee(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	payee := newPayee(t, queries, ctx, budget, "testpayee")
	payee2 := newPayee(t, queries, ctx, budget, "otherpayee")
	category := newCategory(t, queries, ctx, budget, "testcategory")
	account := newAccount(t, queries, ctx, budget, "testaccount")
	categoryDefault := newCategoryDefault(t, queries, ctx, budget, payee, category, 100)
	transferDefault := newTransferDefault(t, queries, ctx, budget, payee, account, 50)
	newCategoryDefault(t, queries, ctx, budget, payee2, category, 100)
	defaults, err := queries.ListPayeeDefaultLinesByPayee(ctx, data.ListPayeeDefaultLinesByPayeeParams{LoginID: budget.LoginID, PayeeID: payee.ID, BudgetID: budget.ID})
	if err != nil {
		t.Fatal(err)
	}
	if len(defaults) != 2 {
		t.Fatal("There should be two default lines for the payee")
	}
	if defaults[0].ID != categoryDefault.ID {
		t.Error("First default should be the category one (ordered by id)")
	}
	if !defaults[0].CategoryID.Valid || defaults[0].CategoryID.Int64 != category.ID {
		t.Error("first default category id does not match")
	}
	if defaults[0].CategoryName.String != "testcategory" {
		t.Error("first default category name does not match")
	}
	if defaults[0].DestAccountID.Valid {
		t.Error("first default dest account should be null")
	}
	if defaults[1].ID != transferDefault.ID {
		t.Error("Second default should be the transfer one (ordered by id)")
	}
	if !defaults[1].DestAccountID.Valid || defaults[1].DestAccountID.Int64 != account.ID {
		t.Error("second default dest account id does not match")
	}
	if defaults[1].DestAccountName.String != "testaccount" {
		t.Error("second default dest account name does not match")
	}
	if defaults[1].CategoryID.Valid {
		t.Error("second default category should be null")
	}
	if defaults[1].Percent != 50 {
		t.Error("second default percent does not match")
	}
}

func TestListPayeeDefaultLinesByPayeeScopedToBudget(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget1 := newBudget(t, queries, ctx, "budget1")
	budget2 := newBudget(t, queries, ctx, "budget2")
	payee1 := newPayee(t, queries, ctx, budget1, "payee1")
	payee2 := newPayee(t, queries, ctx, budget2, "payee2")
	category1 := newCategory(t, queries, ctx, budget1, "cat")
	category2 := newCategory(t, queries, ctx, budget2, "cat")
	newCategoryDefault(t, queries, ctx, budget1, payee1, category1, 100)
	newCategoryDefault(t, queries, ctx, budget2, payee2, category2, 100)
	defaults, err := queries.ListPayeeDefaultLinesByPayee(ctx, data.ListPayeeDefaultLinesByPayeeParams{LoginID: budget1.LoginID, PayeeID: payee1.ID, BudgetID: budget1.ID})
	if err != nil {
		t.Fatal(err)
	}
	if len(defaults) != 1 {
		t.Fatalf("There should be one default line for payee1, got %d", len(defaults))
	}
	if defaults[0].PayeeID != payee1.ID {
		t.Error("payee1 returned the wrong default line")
	}
}

func TestScopingCreatePayeeDefaultLineScopedByLogin(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budgetA := newBudget(t, queries, ctx, "budgetA")
	budgetB := newBudget(t, queries, ctx, "budgetB")
	payeeB := newPayee(t, queries, ctx, budgetB, "payeeB")
	categoryB := newCategory(t, queries, ctx, budgetB, "catB")

	_, err := queries.CreatePayeeDefaultLine(ctx, data.CreatePayeeDefaultLineParams{
		PayeeID:       payeeB.ID,
		DestAccountID: sql.NullInt64{},
		CategoryID:    sql.NullInt64{Int64: categoryB.ID, Valid: true},
		Income:        false,
		Percent:       100,
		BudgetID:      budgetB.ID,
		LoginID:       budgetA.LoginID,
	})
	if err == nil {
		t.Fatal("expected creating a default line for another login's budget to fail")
	}
}

func TestScopingDeletePayeeDefaultLineScopedByLogin(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budgetA := newBudget(t, queries, ctx, "budgetA")
	budgetB := newBudget(t, queries, ctx, "budgetB")
	payeeB := newPayee(t, queries, ctx, budgetB, "payeeB")
	categoryB := newCategory(t, queries, ctx, budgetB, "catB")
	lineB := newCategoryDefault(t, queries, ctx, budgetB, payeeB, categoryB, 100)

	err := queries.DeletePayeeDefaultLine(ctx, data.DeletePayeeDefaultLineParams{
		ID:       lineB.ID,
		BudgetID: budgetB.ID,
		LoginID:  budgetA.LoginID,
	})
	if err != nil {
		t.Fatal(err)
	}
	lines, err := queries.ListPayeeDefaultLinesByPayee(ctx, data.ListPayeeDefaultLinesByPayeeParams{LoginID: budgetB.LoginID, PayeeID: payeeB.ID, BudgetID: budgetB.ID})
	if err != nil {
		t.Fatal(err)
	}
	if len(lines) != 1 {
		t.Fatalf("expected lineB to survive a cross-login delete, got %d lines", len(lines))
	}
}

func TestScopingDeletePayeeDefaultLinesByPayeeScopedByLogin(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budgetA := newBudget(t, queries, ctx, "budgetA")
	budgetB := newBudget(t, queries, ctx, "budgetB")
	payeeB := newPayee(t, queries, ctx, budgetB, "payeeB")
	categoryB := newCategory(t, queries, ctx, budgetB, "catB")
	newCategoryDefault(t, queries, ctx, budgetB, payeeB, categoryB, 100)

	err := queries.DeletePayeeDefaultLinesByPayee(ctx, data.DeletePayeeDefaultLinesByPayeeParams{
		PayeeID:  payeeB.ID,
		BudgetID: budgetB.ID,
		LoginID:  budgetA.LoginID,
	})
	if err != nil {
		t.Fatal(err)
	}
	lines, err := queries.ListPayeeDefaultLinesByPayee(ctx, data.ListPayeeDefaultLinesByPayeeParams{LoginID: budgetB.LoginID, PayeeID: payeeB.ID, BudgetID: budgetB.ID})
	if err != nil {
		t.Fatal(err)
	}
	if len(lines) != 1 {
		t.Fatalf("expected the default line to survive a cross-login delete, got %d lines", len(lines))
	}
}

func TestScopingListPayeeDefaultLinesScopedByLogin(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budgetA := newBudget(t, queries, ctx, "budgetA")
	budgetB := newBudget(t, queries, ctx, "budgetB")
	payeeB := newPayee(t, queries, ctx, budgetB, "payeeB")
	categoryB := newCategory(t, queries, ctx, budgetB, "catB")
	newCategoryDefault(t, queries, ctx, budgetB, payeeB, categoryB, 100)

	rows, err := queries.ListPayeeDefaultLinesByPayee(ctx, data.ListPayeeDefaultLinesByPayeeParams{
		LoginID:  budgetA.LoginID,
		PayeeID:  payeeB.ID,
		BudgetID: budgetB.ID,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 0 {
		t.Fatalf("expected 0 default lines for another login's budget, got %d", len(rows))
	}
}

func TestScopingUpdatePayeeDefaultLineScopedByLogin(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budgetA := newBudget(t, queries, ctx, "budgetA")
	budgetB := newBudget(t, queries, ctx, "budgetB")
	payeeB := newPayee(t, queries, ctx, budgetB, "payeeB")
	categoryB := newCategory(t, queries, ctx, budgetB, "catB")
	lineB := newCategoryDefault(t, queries, ctx, budgetB, payeeB, categoryB, 100)

	_, err := queries.UpdatePayeeDefaultLine(ctx, data.UpdatePayeeDefaultLineParams{
		PayeeID:       payeeB.ID,
		DestAccountID: sql.NullInt64{},
		CategoryID:    sql.NullInt64{Int64: categoryB.ID, Valid: true},
		Income:        false,
		Percent:       50,
		ID:            lineB.ID,
		BudgetID:      budgetB.ID,
		LoginID:       budgetA.LoginID,
	})
	if err != sql.ErrNoRows {
		t.Fatalf("expected sql.ErrNoRows when updating across logins, got %v", err)
	}
}
