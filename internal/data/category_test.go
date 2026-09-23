package data_test

import (
	"database/sql"
	"testing"

	"samuellando.com/YNAFB/internal/data"
)

func TestCreateCategory(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	category, err := queries.CreateCategory(ctx, data.CreateCategoryParams{
		LoginID:  budget.LoginID,
		BudgetID: budget.ID,
		Name:     "testcategory",
	})
	if err != nil {
		t.Fatal(err)
	}
	if category.BudgetID != budget.ID {
		t.Error("budget id does not match")
	}
	if category.Name != "testcategory" {
		t.Error("category name does not match")
	}
	if category.CategoryGroupID.Valid {
		t.Error("category group should be null")
	}
	if category.ID != 1 {
		t.Error("ID of the first category should be 1")
	}
}

func TestCreateCategoryWithGroup(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	groupID := newCategoryGroup(t, queries, ctx, budget, "testgroup")
	category, err := queries.CreateCategory(ctx, data.CreateCategoryParams{
		LoginID:         budget.LoginID,
		BudgetID:        budget.ID,
		Name:            "testcategory",
		CategoryGroupID: sql.NullInt64{Int64: groupID, Valid: true},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !category.CategoryGroupID.Valid {
		t.Fatal("category group should be set")
	}
	if category.CategoryGroupID.Int64 != groupID {
		t.Error("category group id does not match")
	}
}

func TestCreateCategoryNonExistingBudget(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	groupID := newCategoryGroup(t, queries, ctx, budget, "testgroup")
	_, err := queries.CreateCategory(ctx, data.CreateCategoryParams{
		BudgetID:        99,
		Name:            "testcategory",
		CategoryGroupID: sql.NullInt64{Int64: groupID, Valid: true},
	})
	if err == nil {
		t.Error("Non existing budget should raise an error")
	}
}

func TestCreateCategoryNonExistingGroup(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	_, err := queries.CreateCategory(ctx, data.CreateCategoryParams{
		LoginID:         budget.LoginID,
		BudgetID:        budget.ID,
		Name:            "testcategory",
		CategoryGroupID: sql.NullInt64{Int64: 99, Valid: true},
	})
	if err == nil {
		t.Error("Non existing category group should raise an error")
	}
}

func TestDeleteBudgetCascadesCategories(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	category := newCategory(t, queries, ctx, budget, "testcategory")
	categories, err := queries.ListCategories(ctx, data.ListCategoriesParams{LoginID: budget.LoginID, BudgetID: budget.ID})
	if err != nil {
		t.Fatal(err)
	}
	if len(categories) != 1 {
		t.Fatal("There should be one category before")
	}
	err = queries.DeleteBudget(ctx, data.DeleteBudgetParams{LoginID: budget.LoginID, ID: budget.ID})
	if err != nil {
		t.Fatal(err)
	}
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM category WHERE id = ?`, category.ID).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatal("Deleting a budget should cascade delete its categories")
	}
}

func TestDeleteBudgetCascadesCategoryGroups(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	groupID := newCategoryGroup(t, queries, ctx, budget, "testgroup")
	groups, err := queries.ListCategoryGroups(ctx, data.ListCategoryGroupsParams{LoginID: budget.LoginID, BudgetID: budget.ID})
	if err != nil {
		t.Fatal(err)
	}
	if len(groups) != 1 {
		t.Fatal("There should be one category group before")
	}
	err = queries.DeleteBudget(ctx, data.DeleteBudgetParams{LoginID: budget.LoginID, ID: budget.ID})
	if err != nil {
		t.Fatal(err)
	}
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM category_group WHERE id = ?`, groupID).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatal("Deleting a budget should cascade delete its category groups")
	}
}

func TestCreateCategoryEmptyName(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	_, err := queries.CreateCategory(ctx, data.CreateCategoryParams{
		LoginID:  budget.LoginID,
		BudgetID: budget.ID,
		Name:     "",
	})
	if err == nil {
		t.Error("Empty category name should raise an error")
	}
}

func TestCreateCategoryIncomeNameRejected(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	for _, name := range []string{"income", "Income", "INCOME", "iNcOmE"} {
		_, err := queries.CreateCategory(ctx, data.CreateCategoryParams{
			LoginID:  budget.LoginID,
			BudgetID: budget.ID,
			Name:     name,
		})
		if err == nil {
			t.Errorf("Category name %q should be rejected", name)
		}
	}
}

func TestCreateCategoryDuplicateNameConstraint(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	budget2 := newBudget(t, queries, ctx, "testBudget2")
	if _, err := queries.CreateCategory(ctx, data.CreateCategoryParams{
		LoginID:  budget.LoginID,
		BudgetID: budget.ID,
		Name:     "testcategory",
	}); err != nil {
		t.Fatal(err)
	}
	_, err := queries.CreateCategory(ctx, data.CreateCategoryParams{
		LoginID:  budget.LoginID,
		BudgetID: budget.ID,
		Name:     "testcategory",
	})
	if err == nil {
		t.Error("Should get an error for duplicate category name in same budget")
	}
	if _, err := queries.CreateCategory(ctx, data.CreateCategoryParams{
		LoginID:  budget2.LoginID,
		BudgetID: budget2.ID,
		Name:     "testcategory",
	}); err != nil {
		t.Error("Should not get an error for duplicate category names across budgets")
	}
}

func TestUpdateCategory(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	groupID := newCategoryGroup(t, queries, ctx, budget, "testgroup")
	category := newCategory(t, queries, ctx, budget, "testcategory")
	updated, err := queries.UpdateCategory(ctx, data.UpdateCategoryParams{
		LoginID:         budget.LoginID,
		Name:            "newName",
		CategoryGroupID: sql.NullInt64{Int64: groupID, Valid: true},
		ID:              category.ID,
		BudgetID:        budget.ID,
	})
	if err != nil {
		t.Fatal(err)
	}
	if updated.ID != category.ID {
		t.Error("ID changed on update")
	}
	if updated.Name != "newName" {
		t.Error("category name was not updated")
	}
	if !updated.CategoryGroupID.Valid || updated.CategoryGroupID.Int64 != groupID {
		t.Error("category group was not updated")
	}
	newNameCategory, err := queries.GetCategory(ctx, data.GetCategoryParams{
		LoginID:  budget.LoginID,
		ID:       category.ID,
		BudgetID: budget.ID,
	})
	if err != nil {
		t.Fatal(err)
	}
	if category.ID != newNameCategory.ID {
		t.Error("ID changed on update")
	}
}

func TestDeleteCategory(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	category := newCategory(t, queries, ctx, budget, "testcategory")
	categories, err := queries.ListCategories(ctx, data.ListCategoriesParams{LoginID: budget.LoginID, BudgetID: budget.ID})
	if err != nil {
		t.Fatal(err)
	}
	if len(categories) != 1 {
		t.Fatal("There should be one category before")
	}
	err = queries.DeleteCategory(ctx, data.DeleteCategoryParams{ID: category.ID, BudgetID: budget.ID, LoginID: budget.LoginID})
	if err != nil {
		t.Fatal(err)
	}
	categories, err = queries.ListCategories(ctx, data.ListCategoriesParams{LoginID: budget.LoginID, BudgetID: budget.ID})
	if err != nil {
		t.Fatal(err)
	}
	if len(categories) != 0 {
		t.Fatal("There should be no category after")
	}
}

func TestListCategories(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	budget2 := newBudget(t, queries, ctx, "testBudget2")
	newCategory(t, queries, ctx, budget, "categoryB")
	newCategory(t, queries, ctx, budget, "categoryA")
	newCategory(t, queries, ctx, budget2, "otherBudgetCategory")
	categories, err := queries.ListCategories(ctx, data.ListCategoriesParams{LoginID: budget.LoginID, BudgetID: budget.ID})
	if err != nil {
		t.Fatal(err)
	}
	if len(categories) != 2 {
		t.Fatal("There should be two categories in the budget")
	}
	if categories[0].Name != "categoryA" {
		t.Error("First category should be categoryA (ordered by name)")
	}
	if categories[1].Name != "categoryB" {
		t.Error("Second category should be categoryB (ordered by name)")
	}
}

func TestDeleteCategoryGroupNullsCategories(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	groupID := newCategoryGroup(t, queries, ctx, budget, "testgroup")
	category, err := queries.CreateCategory(ctx, data.CreateCategoryParams{
		LoginID:         budget.LoginID,
		BudgetID:        budget.ID,
		Name:            "testcategory",
		CategoryGroupID: sql.NullInt64{Int64: groupID, Valid: true},
	})
	if err != nil {
		t.Fatal(err)
	}
	err = queries.DeleteCategoryGroup(ctx, data.DeleteCategoryGroupParams{ID: groupID, BudgetID: budget.ID, LoginID: budget.LoginID})
	if err != nil {
		t.Fatal(err)
	}
	var group sql.NullInt64
	if err := db.QueryRow(`SELECT category_group_id FROM category WHERE id = ?`, category.ID).Scan(&group); err != nil {
		t.Fatal(err)
	}
	if group.Valid {
		t.Error("Deleting a category group should null its categories' group, not delete them")
	}
}

func TestCreateCategoryGroupEmptyName(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	_, err := queries.CreateCategoryGroup(ctx, data.CreateCategoryGroupParams{
		LoginID:  budget.LoginID,
		BudgetID: budget.ID,
		Name:     "",
	})
	if err == nil {
		t.Error("Empty category group name should raise an error")
	}
}

func TestCreateCategoryGroupNonExistingBudget(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	_, err := queries.CreateCategoryGroup(ctx, data.CreateCategoryGroupParams{
		BudgetID: 99,
		Name:     "testgroup",
	})
	if err == nil {
		t.Error("Non existing budget should raise an error")
	}
}

func TestCreateCategoryGroup(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	groupID, err := queries.CreateCategoryGroup(ctx, data.CreateCategoryGroupParams{
		LoginID:  budget.LoginID,
		BudgetID: budget.ID,
		Name:     "testgroup",
	})
	if err != nil {
		t.Fatal(err)
	}
	group, err := queries.GetCategoryGroup(ctx, data.GetCategoryGroupParams{
		LoginID:  budget.LoginID,
		ID:       groupID,
		BudgetID: budget.ID,
	})
	if err != nil {
		t.Fatal(err)
	}
	if group.ID != groupID {
		t.Error("category group id does not match")
	}
	if group.Name != "testgroup" {
		t.Error("category group name does not match")
	}
	if group.BudgetID != budget.ID {
		t.Error("category group budget id does not match")
	}
}

func TestCreateCategoryGroupDuplicateName(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	if _, err := queries.CreateCategoryGroup(ctx, data.CreateCategoryGroupParams{
		LoginID:  budget.LoginID,
		BudgetID: budget.ID,
		Name:     "testgroup",
	}); err != nil {
		t.Fatal(err)
	}
	_, err := queries.CreateCategoryGroup(ctx, data.CreateCategoryGroupParams{
		LoginID:  budget.LoginID,
		BudgetID: budget.ID,
		Name:     "testgroup",
	})
	if err == nil {
		t.Error("Duplicate category group name should raise an error")
	}
}

func TestUpdateCategoryGroup(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	groupID := newCategoryGroup(t, queries, ctx, budget, "testgroup")
	updated, err := queries.UpdateCategoryGroup(ctx, data.UpdateCategoryGroupParams{
		LoginID:  budget.LoginID,
		Name:     "newName",
		ID:       groupID,
		BudgetID: budget.ID,
	})
	if err != nil {
		t.Fatal(err)
	}
	if updated.ID != groupID {
		t.Error("ID changed on update")
	}
	if updated.Name != "newName" {
		t.Error("category group name was not updated")
	}
	group, err := queries.GetCategoryGroup(ctx, data.GetCategoryGroupParams{
		LoginID:  budget.LoginID,
		ID:       groupID,
		BudgetID: budget.ID,
	})
	if err != nil {
		t.Fatal(err)
	}
	if groupID != group.ID {
		t.Error("ID changed on update")
	}
}

func TestDeleteCategoryGroup(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	groupID := newCategoryGroup(t, queries, ctx, budget, "testgroup")
	groups, err := queries.ListCategoryGroups(ctx, data.ListCategoryGroupsParams{LoginID: budget.LoginID, BudgetID: budget.ID})
	if err != nil {
		t.Fatal(err)
	}
	if len(groups) != 1 {
		t.Fatal("There should be one category group before")
	}
	err = queries.DeleteCategoryGroup(ctx, data.DeleteCategoryGroupParams{ID: groupID, BudgetID: budget.ID, LoginID: budget.LoginID})
	if err != nil {
		t.Fatal(err)
	}
	groups, err = queries.ListCategoryGroups(ctx, data.ListCategoryGroupsParams{LoginID: budget.LoginID, BudgetID: budget.ID})
	if err != nil {
		t.Fatal(err)
	}
	if len(groups) != 0 {
		t.Fatal("There should be no category group after")
	}
}

func TestListCategoryGroups(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	newCategoryGroup(t, queries, ctx, budget, "groupB")
	newCategoryGroup(t, queries, ctx, budget, "groupA")
	groups, err := queries.ListCategoryGroups(ctx, data.ListCategoryGroupsParams{LoginID: budget.LoginID, BudgetID: budget.ID})
	if err != nil {
		t.Fatal(err)
	}
	if len(groups) != 2 {
		t.Fatal("There should be two category groups")
	}
	if groups[0].Name != "groupA" {
		t.Error("First category group should be groupA (ordered by name)")
	}
	if groups[1].Name != "groupB" {
		t.Error("Second category group should be groupB (ordered by name)")
	}
}

func TestScopingCategoryGroupScopedByLogin(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budgetA := newBudget(t, queries, ctx, "budgetA")
	budgetB := newBudget(t, queries, ctx, "budgetB")
	newCategoryGroup(t, queries, ctx, budgetB, "group")

	rows, err := queries.ListCategoryGroups(ctx, data.ListCategoryGroupsParams{
		LoginID:  budgetA.LoginID,
		BudgetID: budgetB.ID,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 0 {
		t.Fatalf("expected 0 category groups for another login's budget, got %d", len(rows))
	}
}

func TestScopingCreateCategoryScopedByLogin(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budgetA := newBudget(t, queries, ctx, "budgetA")
	budgetB := newBudget(t, queries, ctx, "budgetB")

	_, err := queries.CreateCategory(ctx, data.CreateCategoryParams{
		LoginID:  budgetA.LoginID,
		BudgetID: budgetB.ID,
		Name:     "sneaky",
	})
	if err == nil {
		t.Fatal("expected creating a category for another login's budget to fail")
	}
}

func TestScopingDeleteCategoryScopedByLogin(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budgetA := newBudget(t, queries, ctx, "budgetA")
	budgetB := newBudget(t, queries, ctx, "budgetB")
	categoryB := newCategory(t, queries, ctx, budgetB, "catB")

	err := queries.DeleteCategory(ctx, data.DeleteCategoryParams{
		ID:       categoryB.ID,
		BudgetID: budgetB.ID,
		LoginID:  budgetA.LoginID,
	})
	if err != nil {
		t.Fatal(err)
	}
	categories, err := queries.ListCategories(ctx, data.ListCategoriesParams{LoginID: budgetB.LoginID, BudgetID: budgetB.ID})
	if err != nil {
		t.Fatal(err)
	}
	if len(categories) != 1 {
		t.Fatalf("expected categoryB to survive a cross-login delete, got %d categories", len(categories))
	}
}

func TestScopingListCategoriesScopedByLogin(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budgetA := newBudget(t, queries, ctx, "budgetA")
	budgetB := newBudget(t, queries, ctx, "budgetB")
	newCategory(t, queries, ctx, budgetB, "catB")

	rows, err := queries.ListCategories(ctx, data.ListCategoriesParams{
		LoginID:  budgetA.LoginID,
		BudgetID: budgetB.ID,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 0 {
		t.Fatalf("expected 0 categories for another login's budget, got %d", len(rows))
	}
}

func TestScopingUpdateCategoryScopedByLogin(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budgetA := newBudget(t, queries, ctx, "budgetA")
	budgetB := newBudget(t, queries, ctx, "budgetB")
	categoryB := newCategory(t, queries, ctx, budgetB, "catB")

	_, err := queries.UpdateCategory(ctx, data.UpdateCategoryParams{
		Name:            "hacked",
		CategoryGroupID: sql.NullInt64{},
		ID:              categoryB.ID,
		BudgetID:        budgetB.ID,
		LoginID:         budgetA.LoginID,
	})
	if err != sql.ErrNoRows {
		t.Fatalf("expected sql.ErrNoRows when updating across logins, got %v", err)
	}
}

func TestScopingCreateCategoryGroupScopedByLogin(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budgetA := newBudget(t, queries, ctx, "budgetA")
	budgetB := newBudget(t, queries, ctx, "budgetB")

	_, err := queries.CreateCategoryGroup(ctx, data.CreateCategoryGroupParams{
		LoginID:  budgetA.LoginID,
		BudgetID: budgetB.ID,
		Name:     "sneaky",
	})
	if err == nil {
		t.Fatal("expected creating a category group for another login's budget to fail")
	}
}

func TestScopingDeleteCategoryGroupScopedByLogin(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budgetA := newBudget(t, queries, ctx, "budgetA")
	budgetB := newBudget(t, queries, ctx, "budgetB")
	groupB := newCategoryGroup(t, queries, ctx, budgetB, "groupB")

	err := queries.DeleteCategoryGroup(ctx, data.DeleteCategoryGroupParams{
		ID:       groupB,
		BudgetID: budgetB.ID,
		LoginID:  budgetA.LoginID,
	})
	if err != nil {
		t.Fatal(err)
	}
	groups, err := queries.ListCategoryGroups(ctx, data.ListCategoryGroupsParams{LoginID: budgetB.LoginID, BudgetID: budgetB.ID})
	if err != nil {
		t.Fatal(err)
	}
	if len(groups) != 1 {
		t.Fatalf("expected groupB to survive a cross-login delete, got %d groups", len(groups))
	}
}

func TestScopingUpdateCategoryGroupScopedByLogin(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budgetA := newBudget(t, queries, ctx, "budgetA")
	budgetB := newBudget(t, queries, ctx, "budgetB")
	groupB := newCategoryGroup(t, queries, ctx, budgetB, "groupB")

	_, err := queries.UpdateCategoryGroup(ctx, data.UpdateCategoryGroupParams{
		Name:     "hacked",
		ID:       groupB,
		BudgetID: budgetB.ID,
		LoginID:  budgetA.LoginID,
	})
	if err != sql.ErrNoRows {
		t.Fatalf("expected sql.ErrNoRows when updating across logins, got %v", err)
	}
}

func TestGetCategory(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	groupID := newCategoryGroup(t, queries, ctx, budget, "testgroup")
	category, err := queries.CreateCategory(ctx, data.CreateCategoryParams{
		LoginID:         budget.LoginID,
		BudgetID:        budget.ID,
		Name:            "testcategory",
		CategoryGroupID: sql.NullInt64{Int64: groupID, Valid: true},
	})
	if err != nil {
		t.Fatal(err)
	}
	ungrouped := newCategory(t, queries, ctx, budget, "ungrouped")

	got, err := queries.GetCategory(ctx, data.GetCategoryParams{
		LoginID:  budget.LoginID,
		BudgetID: budget.ID,
		ID:       category.ID,
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != category.ID {
		t.Error("getting category, id does not match")
	}
	if got.Name != "testcategory" {
		t.Error("getting category, name does not match")
	}
	if !got.GroupID.Valid || got.GroupID.Int64 != groupID {
		t.Error("getting category, group id does not match")
	}
	if got.GroupName.String != "testgroup" {
		t.Error("getting category, group name does not match")
	}

	got, err = queries.GetCategory(ctx, data.GetCategoryParams{
		LoginID:  budget.LoginID,
		BudgetID: budget.ID,
		ID:       ungrouped.ID,
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.GroupID.Valid {
		t.Error("getting ungrouped category, group id should be null")
	}
}

func TestGetCategoryDoesNotExist(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	_, err := queries.GetCategory(ctx, data.GetCategoryParams{
		LoginID:  budget.LoginID,
		BudgetID: budget.ID,
		ID:       99,
	})
	if err != sql.ErrNoRows {
		t.Fatalf("Getting non existent category should return sql.ErrNoRows, got %v", err)
	}
}

func TestGetCategoryGroup(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	groupID := newCategoryGroup(t, queries, ctx, budget, "testgroup")
	group, err := queries.GetCategoryGroup(ctx, data.GetCategoryGroupParams{
		LoginID:  budget.LoginID,
		BudgetID: budget.ID,
		ID:       groupID,
	})
	if err != nil {
		t.Fatal(err)
	}
	if group.ID != groupID {
		t.Error("getting category group, id does not match")
	}
	if group.BudgetID != budget.ID {
		t.Error("getting category group, budget id does not match")
	}
	if group.Name != "testgroup" {
		t.Error("getting category group, name does not match")
	}
}

func TestGetCategoryGroupDoesNotExist(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget := newBudget(t, queries, ctx, "testBudget")
	_, err := queries.GetCategoryGroup(ctx, data.GetCategoryGroupParams{
		LoginID:  budget.LoginID,
		BudgetID: budget.ID,
		ID:       99,
	})
	if err != sql.ErrNoRows {
		t.Fatalf("Getting non existent category group should return sql.ErrNoRows, got %v", err)
	}
}

func TestScopingGetCategoryScopedByLogin(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budgetA := newBudget(t, queries, ctx, "budgetA")
	budgetB := newBudget(t, queries, ctx, "budgetB")
	categoryB := newCategory(t, queries, ctx, budgetB, "catB")

	_, err := queries.GetCategory(ctx, data.GetCategoryParams{
		LoginID:  budgetA.LoginID,
		BudgetID: budgetB.ID,
		ID:       categoryB.ID,
	})
	if err != sql.ErrNoRows {
		t.Fatalf("expected sql.ErrNoRows for another login's category, got %v", err)
	}
}

func TestScopingGetCategoryGroupScopedByLogin(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budgetA := newBudget(t, queries, ctx, "budgetA")
	budgetB := newBudget(t, queries, ctx, "budgetB")
	groupB := newCategoryGroup(t, queries, ctx, budgetB, "groupB")

	_, err := queries.GetCategoryGroup(ctx, data.GetCategoryGroupParams{
		LoginID:  budgetA.LoginID,
		BudgetID: budgetB.ID,
		ID:       groupB,
	})
	if err != sql.ErrNoRows {
		t.Fatalf("expected sql.ErrNoRows for another login's category group, got %v", err)
	}
}
