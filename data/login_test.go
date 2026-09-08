package data_test

import (
	"testing"

	"samuellando.com/YNAFB/data"
)

func TestCreateLogin(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	login, err := queries.CreateLogin(ctx, data.CreateLoginParams{Username: "alice", Password: "pw"})
	if err != nil {
		t.Fatal(err)
	}
	if login.Username != "alice" {
		t.Error("username does not match")
	}
	if login.Password != "pw" {
		t.Error("password does not match")
	}
	if login.ID != 1 {
		t.Error("ID of the first login should be 1")
	}
}

func TestCreateLoginDuplicateUsername(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	if _, err := queries.CreateLogin(ctx, data.CreateLoginParams{Username: "alice", Password: "pw"}); err != nil {
		t.Fatal(err)
	}
	_, err := queries.CreateLogin(ctx, data.CreateLoginParams{Username: "alice", Password: "pw2"})
	if err == nil {
		t.Error("Duplicate username should raise an error")
	}
}

func TestCreateLoginShortUsernameRejected(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	_, err := queries.CreateLogin(ctx, data.CreateLoginParams{Username: "ab", Password: "pw"})
	if err == nil {
		t.Error("A username shorter than 3 characters should be rejected")
	}
}

func TestGetLoginByUsername(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	login := newLogin(t, queries, ctx, "alice")
	got, err := queries.GetLoginByUsername(ctx, data.GetLoginByUsernameParams{Username: "alice"})
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != login.ID {
		t.Error("getting login by username, id does not match")
	}
	if got.Username != "alice" {
		t.Error("getting login by username, username does not match")
	}
	if got.Password != login.Password {
		t.Error("getting login by username, password does not match")
	}
}

func TestGetLoginByUsernameDoesNotExist(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	_, err := queries.GetLoginByUsername(ctx, data.GetLoginByUsernameParams{Username: "nobody"})
	if err == nil {
		t.Fatal("Getting a non existent login by username should fail")
	}
}

func TestListLogins(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	newLogin(t, queries, ctx, "alice")
	newLogin(t, queries, ctx, "bob")
	logins, err := queries.ListLogins(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(logins) != 2 {
		t.Fatalf("There should be two logins, got %d", len(logins))
	}
	if logins[0].Username != "alice" {
		t.Error("First login should be alice (ordered by id)")
	}
	if logins[1].Username != "bob" {
		t.Error("Second login should be bob (ordered by id)")
	}
}

func TestUpdateLoginPassword(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	login := newLogin(t, queries, ctx, "alice")
	updated, err := queries.UpdateLoginPassword(ctx, data.UpdateLoginPasswordParams{Password: "newpw", ID: login.ID})
	if err != nil {
		t.Fatal(err)
	}
	if updated.ID != login.ID {
		t.Error("ID changed on update")
	}
	if updated.Password != "newpw" {
		t.Error("login password was not updated")
	}
	got, err := queries.GetLoginByUsername(ctx, data.GetLoginByUsernameParams{Username: "alice"})
	if err != nil {
		t.Fatal(err)
	}
	if got.Password != "newpw" {
		t.Error("login password was not updated")
	}
}

func TestDeleteLogin(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	login := newLogin(t, queries, ctx, "alice")
	if err := queries.DeleteLogin(ctx, data.DeleteLoginParams{ID: login.ID}); err != nil {
		t.Fatal(err)
	}
	if _, err := queries.GetLoginByUsername(ctx, data.GetLoginByUsernameParams{Username: "alice"}); err == nil {
		t.Fatal("Deleting a login should make it unretrievable")
	}
}

func TestDeleteLoginCascadesBudgets(t *testing.T) {
	db, queries, ctx := setup(t)
	defer teardown(db)
	login := newLogin(t, queries, ctx, "alice")
	budget := newBudgetForLogin(t, queries, ctx, login.ID, "testBudget")
	if err := queries.DeleteLogin(ctx, data.DeleteLoginParams{ID: login.ID}); err != nil {
		t.Fatal(err)
	}
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM budget WHERE id = ?`, budget.ID).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatal("Deleting a login should cascade delete its budgets")
	}
}
