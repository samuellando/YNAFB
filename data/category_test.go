package data_test

import (
	"database/sql"
	"testing"

	"samuellando.com/YNAFB/data"
)

func TestCreateCategory(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget, err := queries.CreateBudget(ctx, "testBudget")
	if err != nil {
		t.Fatal(err)
	}
	category, err := queries.CreateCategory(ctx, data.CreateCategoryParams{
		Budget:        budget.ID,
		Name:          "testcategory",
		CategoryGroup: sql.NullInt64{},
	})
	if err != nil {
		t.Fatal(err)
	}
	if category.Budget != budget.ID {
		t.Error("budget id does not match")
	}
	if category.Name != "testcategory" {
		t.Error("category name does not match")
	}
	if category.CategoryGroup.Valid {
		t.Error("category group should be null")
	}
	if category.ID != 1 {
		t.Error("ID of the first category should be 1")
	}
}

func TestCreateCategoryWithGroup(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget, err := queries.CreateBudget(ctx, "testBudget")
	if err != nil {
		t.Fatal(err)
	}
	groupID, err := queries.CreateCategoryGroup(ctx, data.CreateCategoryGroupParams{
		Budget: budget.ID,
		Name:   "testgroup",
	})
	if err != nil {
		t.Fatal(err)
	}
	category, err := queries.CreateCategory(ctx, data.CreateCategoryParams{
		Budget:        budget.ID,
		Name:          "testcategory",
		CategoryGroup: sql.NullInt64{Int64: groupID, Valid: true},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !category.CategoryGroup.Valid {
		t.Fatal("category group should be set")
	}
	if category.CategoryGroup.Int64 != groupID {
		t.Error("category group id does not match")
	}
}

func TestCreateCategoryNonExistingBudget(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget, err := queries.CreateBudget(ctx, "testBudget")
	if err != nil {
		t.Fatal(err)
	}
	groupID, err := queries.CreateCategoryGroup(ctx, data.CreateCategoryGroupParams{
		Budget: budget.ID,
		Name:   "testgroup",
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = queries.CreateCategory(ctx, data.CreateCategoryParams{
		Budget:        99,
		Name:          "testcategory",
		CategoryGroup: sql.NullInt64{Int64: groupID, Valid: true},
	})
	if err == nil {
		t.Error("Non existing budget should raise an error")
	}
}

func TestCreateCategoryNonExistingGroup(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget, err := queries.CreateBudget(ctx, "testBudget")
	if err != nil {
		t.Fatal(err)
	}
	_, err = queries.CreateCategory(ctx, data.CreateCategoryParams{
		Budget:        budget.ID,
		Name:          "testcategory",
		CategoryGroup: sql.NullInt64{Int64: 99, Valid: true},
	})
	if err == nil {
		t.Error("Non existing category group should raise an error")
	}
}

func TestDeleteBudgetCascadesCategories(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget, err := queries.CreateBudget(ctx, "testBudget")
	if err != nil {
		t.Fatal(err)
	}
	category, err := queries.CreateCategory(ctx, data.CreateCategoryParams{
		Budget: budget.ID,
		Name:   "testcategory",
	})
	if err != nil {
		t.Fatal(err)
	}
	categories, err := queries.ListCategories(ctx, budget.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(categories) != 1 {
		t.Fatal("There should be one category before")
	}
	err = queries.DeleteBudget(ctx, budget.ID)
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
	budget, err := queries.CreateBudget(ctx, "testBudget")
	if err != nil {
		t.Fatal(err)
	}
	groupID, err := queries.CreateCategoryGroup(ctx, data.CreateCategoryGroupParams{
		Budget: budget.ID,
		Name:   "testgroup",
	})
	if err != nil {
		t.Fatal(err)
	}
	groups, err := queries.ListCategoryGroups(ctx, budget.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(groups) != 1 {
		t.Fatal("There should be one category group before")
	}
	err = queries.DeleteBudget(ctx, budget.ID)
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
	budget, err := queries.CreateBudget(ctx, "testBudget")
	if err != nil {
		t.Fatal(err)
	}
	_, err = queries.CreateCategory(ctx, data.CreateCategoryParams{
		Budget: budget.ID,
		Name:   "",
	})
	if err == nil {
		t.Error("Empty category name should raise an error")
	}
}

func TestCreateCategoryIncomeNameRejected(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget, err := queries.CreateBudget(ctx, "testBudget")
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"income", "Income", "INCOME", "iNcOmE"} {
		_, err := queries.CreateCategory(ctx, data.CreateCategoryParams{
			Budget: budget.ID,
			Name:   name,
		})
		if err == nil {
			t.Errorf("Category name %q should be rejected", name)
		}
	}
}

func TestCreateCategoryDuplicateNameConstraint(t *testing.T) {
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
	_, err = queries.CreateCategory(ctx, data.CreateCategoryParams{
		Budget: budget.ID,
		Name:   "testcategory",
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = queries.CreateCategory(ctx, data.CreateCategoryParams{
		Budget: budget.ID,
		Name:   "testcategory",
	})
	if err == nil {
		t.Error("Should get an error for duplicate category name in same budget")
	}
	_, err = queries.CreateCategory(ctx, data.CreateCategoryParams{
		Budget: budget2.ID,
		Name:   "testcategory",
	})
	if err != nil {
		t.Error("Should not get an error for duplicate category names across budgets")
	}
}

func TestUpdateCategory(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget, err := queries.CreateBudget(ctx, "testBudget")
	if err != nil {
		t.Fatal(err)
	}
	groupID, err := queries.CreateCategoryGroup(ctx, data.CreateCategoryGroupParams{
		Budget: budget.ID,
		Name:   "testgroup",
	})
	if err != nil {
		t.Fatal(err)
	}
	category, err := queries.CreateCategory(ctx, data.CreateCategoryParams{
		Budget:        budget.ID,
		Name:          "testcategory",
		CategoryGroup: sql.NullInt64{},
	})
	if err != nil {
		t.Fatal(err)
	}
	n, err := queries.UpdateCategory(ctx, data.UpdateCategoryParams{
		Name:          "newName",
		CategoryGroup: sql.NullInt64{Int64: groupID, Valid: true},
		ID:            category.ID,
	})
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Error("The number of affected rows should be 1")
	}
	newNameCategory, err := queries.GetCategoryByName(ctx, data.GetCategoryByNameParams{
		Name:   "newName",
		Budget: budget.ID,
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
	budget, err := queries.CreateBudget(ctx, "testBudget")
	if err != nil {
		t.Fatal(err)
	}
	category, err := queries.CreateCategory(ctx, data.CreateCategoryParams{
		Budget: budget.ID,
		Name:   "testcategory",
	})
	if err != nil {
		t.Fatal(err)
	}
	categories, err := queries.ListCategories(ctx, budget.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(categories) != 1 {
		t.Fatal("There should be one category before")
	}
	err = queries.DeleteCategory(ctx, category.ID)
	if err != nil {
		t.Fatal(err)
	}
	categories, err = queries.ListCategories(ctx, budget.ID)
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
	budget, err := queries.CreateBudget(ctx, "testBudget")
	if err != nil {
		t.Fatal(err)
	}
	budget2, err := queries.CreateBudget(ctx, "testBudget2")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := queries.CreateCategory(ctx, data.CreateCategoryParams{
		Budget: budget.ID,
		Name:   "categoryB",
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := queries.CreateCategory(ctx, data.CreateCategoryParams{
		Budget: budget.ID,
		Name:   "categoryA",
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := queries.CreateCategory(ctx, data.CreateCategoryParams{
		Budget: budget2.ID,
		Name:   "otherBudgetCategory",
	}); err != nil {
		t.Fatal(err)
	}
	categories, err := queries.ListCategories(ctx, budget.ID)
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

func TestGetCategoryByName(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget, err := queries.CreateBudget(ctx, "testBudget")
	if err != nil {
		t.Fatal(err)
	}
	category, err := queries.CreateCategory(ctx, data.CreateCategoryParams{
		Budget: budget.ID,
		Name:   "testcategory",
	})
	if err != nil {
		t.Fatal(err)
	}
	nameCategory, err := queries.GetCategoryByName(ctx, data.GetCategoryByNameParams{
		Name:   "testcategory",
		Budget: budget.ID,
	})
	if err != nil {
		t.Fatal(err)
	}
	if category.ID != nameCategory.ID {
		t.Error("getting category by name, id does not match")
	}
	if nameCategory.Budget != budget.ID {
		t.Error("getting category by name, budget id does not match")
	}
	if nameCategory.Name != "testcategory" {
		t.Error("getting category by name, name does not match")
	}
}

func TestGetCategoryByNameDoesNotExist(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget, err := queries.CreateBudget(ctx, "testBudget")
	if err != nil {
		t.Fatal(err)
	}
	_, err = queries.GetCategoryByName(ctx, data.GetCategoryByNameParams{
		Name:   "testcategory",
		Budget: budget.ID,
	})
	if err == nil {
		t.Fatal("Getting non existent category by name should fail")
	}
}

func TestDeleteCategoryGroupSetsCategoryGroupNull(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget, err := queries.CreateBudget(ctx, "testBudget")
	if err != nil {
		t.Fatal(err)
	}
	groupID, err := queries.CreateCategoryGroup(ctx, data.CreateCategoryGroupParams{
		Budget: budget.ID,
		Name:   "testgroup",
	})
	if err != nil {
		t.Fatal(err)
	}
	category, err := queries.CreateCategory(ctx, data.CreateCategoryParams{
		Budget:        budget.ID,
		Name:          "testcategory",
		CategoryGroup: sql.NullInt64{Int64: groupID, Valid: true},
	})
	if err != nil {
		t.Fatal(err)
	}
	err = queries.DeleteCategoryGroup(ctx, groupID)
	if err != nil {
		t.Fatal(err)
	}
	var group sql.NullInt64
	if err := db.QueryRow(`SELECT category_group FROM category WHERE id = ?`, category.ID).Scan(&group); err != nil {
		t.Fatal(err)
	}
	if group.Valid {
		t.Error("Deleting a category group should set category.category_group to null")
	}
}

func TestCreateCategoryGroupEmptyName(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget, err := queries.CreateBudget(ctx, "testBudget")
	if err != nil {
		t.Fatal(err)
	}
	_, err = queries.CreateCategoryGroup(ctx, data.CreateCategoryGroupParams{
		Budget: budget.ID,
		Name:   "",
	})
	if err == nil {
		t.Error("Empty category group name should raise an error")
	}
}

func TestCreateCategoryGroupNonExistingBudget(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	_, err := queries.CreateCategoryGroup(ctx, data.CreateCategoryGroupParams{
		Budget: 99,
		Name:   "testgroup",
	})
	if err == nil {
		t.Error("Non existing budget should raise an error")
	}
}

func TestCreateCategoryGroup(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget, err := queries.CreateBudget(ctx, "testBudget")
	if err != nil {
		t.Fatal(err)
	}
	groupID, err := queries.CreateCategoryGroup(ctx, data.CreateCategoryGroupParams{
		Budget: budget.ID,
		Name:   "testgroup",
	})
	if err != nil {
		t.Fatal(err)
	}
	group, err := queries.GetCategoryGroupByName(ctx, data.GetCategoryGroupByNameParams{
		Name:   "testgroup",
		Budget: budget.ID,
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
	if group.Budget != budget.ID {
		t.Error("category group budget id does not match")
	}
}

func TestCreateCategoryGroupDuplicateName(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget, err := queries.CreateBudget(ctx, "testBudget")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := queries.CreateCategoryGroup(ctx, data.CreateCategoryGroupParams{
		Budget: budget.ID,
		Name:   "testgroup",
	}); err != nil {
		t.Fatal(err)
	}
	_, err = queries.CreateCategoryGroup(ctx, data.CreateCategoryGroupParams{
		Budget: budget.ID,
		Name:   "testgroup",
	})
	if err == nil {
		t.Error("Duplicate category group name should raise an error")
	}
}

func TestGetOrCreateCategoryGroup(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget, err := queries.CreateBudget(ctx, "testBudget")
	if err != nil {
		t.Fatal(err)
	}
	id1, err := queries.GetOrCreateCategoryGroup(ctx, data.GetOrCreateCategoryGroupParams{
		Budget: budget.ID,
		Name:   "testgroup",
	})
	if err != nil {
		t.Fatal(err)
	}
	id2, err := queries.GetOrCreateCategoryGroup(ctx, data.GetOrCreateCategoryGroupParams{
		Budget: budget.ID,
		Name:   "testgroup",
	})
	if err != nil {
		t.Fatal(err)
	}
	if id1 != id2 {
		t.Error("GetOrCreate should return the existing group on conflict")
	}
	id3, err := queries.GetOrCreateCategoryGroup(ctx, data.GetOrCreateCategoryGroupParams{
		Budget: budget.ID,
		Name:   "othergroup",
	})
	if err != nil {
		t.Fatal(err)
	}
	if id3 == id1 {
		t.Error("GetOrCreate with a new name should create a new group")
	}
}

func TestUpdateCategoryGroup(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget, err := queries.CreateBudget(ctx, "testBudget")
	if err != nil {
		t.Fatal(err)
	}
	groupID, err := queries.CreateCategoryGroup(ctx, data.CreateCategoryGroupParams{
		Budget: budget.ID,
		Name:   "testgroup",
	})
	if err != nil {
		t.Fatal(err)
	}
	n, err := queries.UpdateCategoryGroup(ctx, data.UpdateCategoryGroupParams{
		Name: "newName",
		ID:   groupID,
	})
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Error("The number of affected rows should be 1")
	}
	group, err := queries.GetCategoryGroupByName(ctx, data.GetCategoryGroupByNameParams{
		Name:   "newName",
		Budget: budget.ID,
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
	budget, err := queries.CreateBudget(ctx, "testBudget")
	if err != nil {
		t.Fatal(err)
	}
	groupID, err := queries.CreateCategoryGroup(ctx, data.CreateCategoryGroupParams{
		Budget: budget.ID,
		Name:   "testgroup",
	})
	if err != nil {
		t.Fatal(err)
	}
	groups, err := queries.ListCategoryGroups(ctx, budget.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(groups) != 1 {
		t.Fatal("There should be one category group before")
	}
	err = queries.DeleteCategoryGroup(ctx, groupID)
	if err != nil {
		t.Fatal(err)
	}
	groups, err = queries.ListCategoryGroups(ctx, budget.ID)
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
	budget, err := queries.CreateBudget(ctx, "testBudget")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := queries.CreateCategoryGroup(ctx, data.CreateCategoryGroupParams{
		Budget: budget.ID,
		Name:   "groupB",
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := queries.CreateCategoryGroup(ctx, data.CreateCategoryGroupParams{
		Budget: budget.ID,
		Name:   "groupA",
	}); err != nil {
		t.Fatal(err)
	}
	groups, err := queries.ListCategoryGroups(ctx, budget.ID)
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

func TestGetCategoryGroupByName(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	budget, err := queries.CreateBudget(ctx, "testBudget")
	if err != nil {
		t.Fatal(err)
	}
	groupID, err := queries.CreateCategoryGroup(ctx, data.CreateCategoryGroupParams{
		Budget: budget.ID,
		Name:   "testgroup",
	})
	if err != nil {
		t.Fatal(err)
	}
	group, err := queries.GetCategoryGroupByName(ctx, data.GetCategoryGroupByNameParams{
		Name:   "testgroup",
		Budget: budget.ID,
	})
	if err != nil {
		t.Fatal(err)
	}
	if group.ID != groupID {
		t.Error("getting category group by name, id does not match")
	}
	if group.Budget != budget.ID {
		t.Error("getting category group by name, budget id does not match")
	}
	if group.Name != "testgroup" {
		t.Error("getting category group by name, name does not match")
	}
}
