package auth

import (
	"fmt"
	"testing"
	"time"
)

func TestGenerateJWTProducesUniqueTokens(t *testing.T) {
	a, err := GenerateJWT("alice")
	if err != nil {
		t.Fatal(err)
	}
	b, err := GenerateJWT("alice")
	if err != nil {
		t.Fatal(err)
	}
	if a == b {
		t.Fatal("expected distinct tokens across generations")
	}
}

func TestParseJWTRoundTrip(t *testing.T) {
	token, err := GenerateJWT("alice")
	if err != nil {
		t.Fatal(err)
	}
	subject, next, err := ParseJWT(token)
	if err != nil {
		t.Fatal(err)
	}
	if subject != "alice" {
		t.Fatalf("expected subject alice, got %q", subject)
	}
	if next == "" || next == token {
		t.Fatal("expected a new, different token to be issued")
	}
	// The rotated token should continue the chain.
	subject2, _, err := ParseJWT(next)
	if err != nil {
		t.Fatal(err)
	}
	if subject2 != "alice" {
		t.Fatalf("expected subject alice, got %q", subject2)
	}
}

func TestParseJWTConsumesToken(t *testing.T) {
	token, err := GenerateJWT("alice")
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := ParseJWT(token); err != nil {
		t.Fatal(err)
	}
	// A token is single-use, so a second parse must be denied.
	if _, _, err := ParseJWT(token); err == nil {
		t.Fatal("expected a reused token to be rejected")
	}
}

func TestParseJWTUnknownToken(t *testing.T) {
	if _, _, err := ParseJWT("not-a-real-token"); err == nil {
		t.Fatal("expected an unknown token to be rejected")
	}
}

func TestParseJWTAfterDevalidate(t *testing.T) {
	token, err := GenerateJWT("alice")
	if err != nil {
		t.Fatal(err)
	}
	DevalidateJWT(token)
	if _, _, err := ParseJWT(token); err == nil {
		t.Fatal("expected a devalidated token to be rejected")
	}
}

func TestParseJWTExpiredToken(t *testing.T) {
	token, err := GenerateJWT("alice")
	if err != nil {
		t.Fatal(err)
	}
	m.Lock()
	validTokens[token] = time.Now().Add(-time.Minute)
	m.Unlock()
	if _, _, err := ParseJWT(token); err == nil {
		t.Fatal("expected an expired token to be rejected")
	}
}

func TestClearExpiredTokens(t *testing.T) {
	expired, err := GenerateJWT("alice")
	if err != nil {
		t.Fatal(err)
	}
	live, err := GenerateJWT("bob")
	if err != nil {
		t.Fatal(err)
	}
	m.Lock()
	validTokens[expired] = time.Now().Add(-time.Minute)
	m.Unlock()

	clearExpiredTokens()

	m.Lock()
	defer m.Unlock()
	if _, ok := validTokens[expired]; ok {
		t.Fatal("expected the expired token to be cleared")
	}
	if _, ok := validTokens[live]; !ok {
		t.Fatal("expected the live token to be retained")
	}
}

func TestParseJWTConcurrent(t *testing.T) {
	const workers = 16
	results := make(chan error, workers)
	for i := 0; i < workers; i++ {
		go func() {
			token, err := GenerateJWT("user")
			if err != nil {
				results <- err
				return
			}
			subject, next, err := ParseJWT(token)
			if err != nil {
				results <- err
				return
			}
			if subject != "user" {
				results <- fmt.Errorf("expected subject user, got %q", subject)
				return
			}
			if next == "" || next == token {
				results <- fmt.Errorf("expected a rotated token")
				return
			}
			results <- nil
		}()
	}
	for i := 0; i < workers; i++ {
		if err := <-results; err != nil {
			t.Fatal(err)
		}
	}
}
