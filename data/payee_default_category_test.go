package data_test

import (
	"database/sql"
	"testing"

	"samuellando.com/YNAFB/data"
)

func TestCreatePayeeDefaultCategoryCategory(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	payee := newPayee(t, queries, ctx, budget.ID, "testpayee")
	category := newCategory(t, queries, ctx, budget.ID, "testcategory")
	defaultCategory, err := queries.CreatePayeeDefaultCategory(ctx, data.CreatePayeeDefaultCategoryParams{
		Payee:        payee.ID,
		OtherAccount: sql.NullInt64{},
		Category:     sql.NullInt64{Int64: category.ID, Valid: true},
		Income:       false,
		Percent:      100,
	})
	if err != nil {
		t.Fatal(err)
	}
	if defaultCategory.Payee != payee.ID {
		t.Error("payee id does not match")
	}
	if defaultCategory.OtherAccount.Valid {
		t.Error("other account should be null for a category default")
	}
	if !defaultCategory.Category.Valid || defaultCategory.Category.Int64 != category.ID {
		t.Error("category id does not match")
	}
	if defaultCategory.Income {
		t.Error("income should be false")
	}
	if defaultCategory.Percent != 100 {
		t.Error("percent does not match")
	}
	if defaultCategory.ID != 1 {
		t.Error("ID of the first payee default category should be 1")
	}
}

func TestCreatePayeeDefaultCategoryOtherAccount(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	payee := newPayee(t, queries, ctx, budget.ID, "testpayee")
	account := newAccount(t, queries, ctx, budget.ID, "testaccount")
	defaultCategory, err := queries.CreatePayeeDefaultCategory(ctx, data.CreatePayeeDefaultCategoryParams{
		Payee:        payee.ID,
		OtherAccount: sql.NullInt64{Int64: account.ID, Valid: true},
		Category:     sql.NullInt64{},
		Income:       false,
		Percent:      100,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !defaultCategory.OtherAccount.Valid || defaultCategory.OtherAccount.Int64 != account.ID {
		t.Error("other account id does not match")
	}
	if defaultCategory.Category.Valid {
		t.Error("category should be null for a transfer default")
	}
}

func TestCreatePayeeDefaultCategoryIncome(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	payee := newPayee(t, queries, ctx, budget.ID, "testpayee")
	defaultCategory, err := queries.CreatePayeeDefaultCategory(ctx, data.CreatePayeeDefaultCategoryParams{
		Payee:        payee.ID,
		OtherAccount: sql.NullInt64{},
		Category:     sql.NullInt64{},
		Income:       true,
		Percent:      100,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !defaultCategory.Income {
		t.Error("income should be true")
	}
	if defaultCategory.Category.Valid {
		t.Error("category should be null for income")
	}
	if defaultCategory.OtherAccount.Valid {
		t.Error("other account should be null for income")
	}
}

func TestCreatePayeeDefaultCategoryXorViolation(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	payee := newPayee(t, queries, ctx, budget.ID, "testpayee")
	account := newAccount(t, queries, ctx, budget.ID, "testaccount")
	category := newCategory(t, queries, ctx, budget.ID, "testcategory")
	_, err := queries.CreatePayeeDefaultCategory(ctx, data.CreatePayeeDefaultCategoryParams{
		Payee:        payee.ID,
		OtherAccount: sql.NullInt64{Int64: account.ID, Valid: true},
		Category:     sql.NullInt64{Int64: category.ID, Valid: true},
		Income:       false,
		Percent:      100,
	})
	if err == nil {
		t.Error("Setting both other_account and category should raise an error")
	}
	_, err = queries.CreatePayeeDefaultCategory(ctx, data.CreatePayeeDefaultCategoryParams{
		Payee:        payee.ID,
		OtherAccount: sql.NullInt64{},
		Category:     sql.NullInt64{},
		Income:       false,
		Percent:      100,
	})
	if err == nil {
		t.Error("Setting neither other_account, category nor income should raise an error")
	}
}

func TestCreatePayeeDefaultCategoryNonExistingPayee(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	category := newCategory(t, queries, ctx, budget.ID, "testcategory")
	_, err := queries.CreatePayeeDefaultCategory(ctx, data.CreatePayeeDefaultCategoryParams{
		Payee:        99,
		OtherAccount: sql.NullInt64{},
		Category:     sql.NullInt64{Int64: category.ID, Valid: true},
		Income:       false,
		Percent:      100,
	})
	if err == nil {
		t.Error("Non existing payee should raise an error")
	}
}

func TestCreatePayeeDefaultCategoryNonExistingOtherAccount(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	payee := newPayee(t, queries, ctx, budget.ID, "testpayee")
	_, err := queries.CreatePayeeDefaultCategory(ctx, data.CreatePayeeDefaultCategoryParams{
		Payee:        payee.ID,
		OtherAccount: sql.NullInt64{Int64: 99, Valid: true},
		Category:     sql.NullInt64{},
		Income:       false,
		Percent:      100,
	})
	if err == nil {
		t.Error("Non existing other account should raise an error")
	}
}

func TestCreatePayeeDefaultCategoryNonExistingCategory(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	payee := newPayee(t, queries, ctx, budget.ID, "testpayee")
	_, err := queries.CreatePayeeDefaultCategory(ctx, data.CreatePayeeDefaultCategoryParams{
		Payee:        payee.ID,
		OtherAccount: sql.NullInt64{},
		Category:     sql.NullInt64{Int64: 99, Valid: true},
		Income:       false,
		Percent:      100,
	})
	if err == nil {
		t.Error("Non existing category should raise an error")
	}
}

func TestDeletePayeeCascadesPayeeDefaultCategories(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	payee := newPayee(t, queries, ctx, budget.ID, "testpayee")
	category := newCategory(t, queries, ctx, budget.ID, "testcategory")
	defaultCategory := newCategoryDefault(t, queries, ctx, payee.ID, category.ID, 100)
	defaults, err := queries.ListPayeeDefaultCategoriesByPayee(ctx, payee.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(defaults) != 1 {
		t.Fatal("There should be one default category before")
	}
	err = queries.DeletePayee(ctx, payee.ID)
	if err != nil {
		t.Fatal(err)
	}
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM payee_default_category WHERE id = ?`, defaultCategory.ID).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatal("Deleting a payee should cascade delete its default categories")
	}
}

func TestDeleteAccountCascadesPayeeDefaultCategories(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	payee := newPayee(t, queries, ctx, budget.ID, "testpayee")
	account := newAccount(t, queries, ctx, budget.ID, "testaccount")
	newTransferDefault(t, queries, ctx, payee.ID, account.ID, 100)
	err := queries.DeleteAccount(ctx, account.ID)
	if err != nil {
		t.Fatal(err)
	}
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM payee_default_category WHERE other_account = ?`, account.ID).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatal("Deleting an account should cascade delete default categories referencing it")
	}
}

func TestDeleteCategoryCascadesPayeeDefaultCategories(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	payee := newPayee(t, queries, ctx, budget.ID, "testpayee")
	category := newCategory(t, queries, ctx, budget.ID, "testcategory")
	newCategoryDefault(t, queries, ctx, payee.ID, category.ID, 100)
	err := queries.DeleteCategory(ctx, category.ID)
	if err != nil {
		t.Fatal(err)
	}
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM payee_default_category WHERE category = ?`, category.ID).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatal("Deleting a category should cascade delete default categories referencing it")
	}
}

func TestCreatePayeeDefaultCategoryPercentNegativeRejected(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	payee := newPayee(t, queries, ctx, budget.ID, "testpayee")
	category := newCategory(t, queries, ctx, budget.ID, "testcategory")
	_, err := queries.CreatePayeeDefaultCategory(ctx, data.CreatePayeeDefaultCategoryParams{
		Payee:        payee.ID,
		OtherAccount: sql.NullInt64{},
		Category:     sql.NullInt64{Int64: category.ID, Valid: true},
		Income:       false,
		Percent:      -1,
	})
	if err == nil {
		t.Error("A negative percent should be rejected")
	}
}

func TestCreatePayeeDefaultCategoryPercentOver100Rejected(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	payee := newPayee(t, queries, ctx, budget.ID, "testpayee")
	category := newCategory(t, queries, ctx, budget.ID, "testcategory")
	_, err := queries.CreatePayeeDefaultCategory(ctx, data.CreatePayeeDefaultCategoryParams{
		Payee:        payee.ID,
		OtherAccount: sql.NullInt64{},
		Category:     sql.NullInt64{Int64: category.ID, Valid: true},
		Income:       false,
		Percent:      101,
	})
	if err == nil {
		t.Error("A percent over 100 should be rejected")
	}
}

func TestCreatePayeeDefaultCategoryPercentZeroAllowed(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	payee := newPayee(t, queries, ctx, budget.ID, "testpayee")
	category := newCategory(t, queries, ctx, budget.ID, "testcategory")
	defaultCategory, err := queries.CreatePayeeDefaultCategory(ctx, data.CreatePayeeDefaultCategoryParams{
		Payee:        payee.ID,
		OtherAccount: sql.NullInt64{},
		Category:     sql.NullInt64{Int64: category.ID, Valid: true},
		Income:       false,
		Percent:      0,
	})
	if err != nil {
		t.Fatal(err)
	}
	if defaultCategory.Percent != 0 {
		t.Error("percent should be 0")
	}
}

func TestUpdatePayeeDefaultCategoryPercentOver100Rejected(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	payee := newPayee(t, queries, ctx, budget.ID, "testpayee")
	category := newCategory(t, queries, ctx, budget.ID, "testcategory")
	defaultCategory := newCategoryDefault(t, queries, ctx, payee.ID, category.ID, 100)
	_, err := queries.UpdatePayeeDefaultCategory(ctx, data.UpdatePayeeDefaultCategoryParams{
		Payee:        payee.ID,
		OtherAccount: sql.NullInt64{},
		Category:     sql.NullInt64{Int64: category.ID, Valid: true},
		Income:       false,
		Percent:      101,
		ID:           defaultCategory.ID,
	})
	if err == nil {
		t.Error("Updating a percent over 100 should be rejected")
	}
}

func TestUpdatePayeeDefaultCategory(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	payee := newPayee(t, queries, ctx, budget.ID, "testpayee")
	category := newCategory(t, queries, ctx, budget.ID, "testcategory")
	account := newAccount(t, queries, ctx, budget.ID, "testaccount")
	defaultCategory := newCategoryDefault(t, queries, ctx, payee.ID, category.ID, 100)
	n, err := queries.UpdatePayeeDefaultCategory(ctx, data.UpdatePayeeDefaultCategoryParams{
		Payee:        payee.ID,
		OtherAccount: sql.NullInt64{Int64: account.ID, Valid: true},
		Category:     sql.NullInt64{},
		Income:       false,
		Percent:      50,
		ID:           defaultCategory.ID,
	})
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Error("The number of affected rows should be 1")
	}
	var otherAccount sql.NullInt64
	var percent int64
	if err := db.QueryRow(`SELECT other_account, percent FROM payee_default_category WHERE id = ?`, defaultCategory.ID).Scan(&otherAccount, &percent); err != nil {
		t.Fatal(err)
	}
	if !otherAccount.Valid || otherAccount.Int64 != account.ID {
		t.Error("payee default category was not updated to a transfer")
	}
	if percent != 50 {
		t.Error("payee default category percent was not updated")
	}
}

func TestDeletePayeeDefaultCategory(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	payee := newPayee(t, queries, ctx, budget.ID, "testpayee")
	category := newCategory(t, queries, ctx, budget.ID, "testcategory")
	defaultCategory := newCategoryDefault(t, queries, ctx, payee.ID, category.ID, 100)
	defaults, err := queries.ListPayeeDefaultCategoriesByPayee(ctx, payee.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(defaults) != 1 {
		t.Fatal("There should be one default category before")
	}
	err = queries.DeletePayeeDefaultCategory(ctx, defaultCategory.ID)
	if err != nil {
		t.Fatal(err)
	}
	defaults, err = queries.ListPayeeDefaultCategoriesByPayee(ctx, payee.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(defaults) != 0 {
		t.Fatal("There should be no default category after")
	}
}

func TestDeletePayeeDefaultCategoriesByPayee(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	payee := newPayee(t, queries, ctx, budget.ID, "testpayee")
	category := newCategory(t, queries, ctx, budget.ID, "testcategory")
	for i := 0; i < 2; i++ {
		newCategoryDefault(t, queries, ctx, payee.ID, category.ID, 100)
	}
	defaults, err := queries.ListPayeeDefaultCategoriesByPayee(ctx, payee.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(defaults) != 2 {
		t.Fatal("There should be two default categories before")
	}
	err = queries.DeletePayeeDefaultCategoriesByPayee(ctx, payee.ID)
	if err != nil {
		t.Fatal(err)
	}
	defaults, err = queries.ListPayeeDefaultCategoriesByPayee(ctx, payee.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(defaults) != 0 {
		t.Fatal("There should be no default category after")
	}
}

func TestListPayeeDefaultCategoriesByPayee(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	payee := newPayee(t, queries, ctx, budget.ID, "testpayee")
	payee2 := newPayee(t, queries, ctx, budget.ID, "otherpayee")
	category := newCategory(t, queries, ctx, budget.ID, "testcategory")
	account := newAccount(t, queries, ctx, budget.ID, "testaccount")
	categoryDefault := newCategoryDefault(t, queries, ctx, payee.ID, category.ID, 100)
	transferDefault := newTransferDefault(t, queries, ctx, payee.ID, account.ID, 50)
	newCategoryDefault(t, queries, ctx, payee2.ID, category.ID, 100)
	defaults, err := queries.ListPayeeDefaultCategoriesByPayee(ctx, payee.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(defaults) != 2 {
		t.Fatal("There should be two default categories for the payee")
	}
	if defaults[0].ID != categoryDefault.ID {
		t.Error("First default should be the category one (ordered by id)")
	}
	if !defaults[0].Category.Valid || defaults[0].Category.Int64 != category.ID {
		t.Error("first default category id does not match")
	}
	if defaults[0].CategoryName.String != "testcategory" {
		t.Error("first default category name does not match")
	}
	if defaults[0].OtherAccount.Valid {
		t.Error("first default other account should be null")
	}
	if defaults[1].ID != transferDefault.ID {
		t.Error("Second default should be the transfer one (ordered by id)")
	}
	if !defaults[1].OtherAccount.Valid || defaults[1].OtherAccount.Int64 != account.ID {
		t.Error("second default other account id does not match")
	}
	if defaults[1].OtherAccountName.String != "testaccount" {
		t.Error("second default other account name does not match")
	}
	if defaults[1].Category.Valid {
		t.Error("second default category should be null")
	}
	if defaults[1].Percent != 50 {
		t.Error("second default percent does not match")
	}
}
