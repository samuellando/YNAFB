package auth

import "testing"

func TestHashAndCheckPassword(t *testing.T) {
	hash, err := HashPassword("correct horse battery staple")
	if err != nil {
		t.Fatal(err)
	}
	if hash == "correct horse battery staple" {
		t.Fatal("password was stored in plaintext")
	}
	if !CheckPassword("correct horse battery staple", hash) {
		t.Fatal("expected the correct password to match")
	}
	if CheckPassword("wrong password", hash) {
		t.Fatal("expected an incorrect password not to match")
	}
}

func TestHashPasswordProducesUniqueSalts(t *testing.T) {
	a, err := HashPassword("same password")
	if err != nil {
		t.Fatal(err)
	}
	b, err := HashPassword("same password")
	if err != nil {
		t.Fatal(err)
	}
	if a == b {
		t.Fatal("expected distinct hashes for the same password due to salting")
	}
	if !CheckPassword("same password", a) || !CheckPassword("same password", b) {
		t.Fatal("expected both hashes to verify the password")
	}
}

func TestCheckPasswordRejectsInvalidHash(t *testing.T) {
	if CheckPassword("password", "not-a-valid-hash") {
		t.Fatal("expected an invalid hash to fail the check")
	}
}
