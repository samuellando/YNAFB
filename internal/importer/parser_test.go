package importer

import (
	"testing"

	"samuellando.com/YNAFB/internal/importer/testparser"
)

func init() {
	Register(testparser.Parser{})
}

func TestParseDispatchesToMatchingParser(t *testing.T) {
	stmt, err := Parse([]byte("ynafb-test-statement"))
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	if stmt.Institution != "Test Bank" {
		t.Fatalf("institution = %q, want %q", stmt.Institution, "Test Bank")
	}
	if len(stmt.Entries) != 2 {
		t.Fatalf("entries = %d, want 2", len(stmt.Entries))
	}
	if stmt.Entries[0].Payee != "Test Merchant" || stmt.Entries[0].Outflow != 1234 {
		t.Fatalf("first entry = %+v", stmt.Entries[0])
	}
	if stmt.Entries[1].Payee != "Incoming Transfer" || stmt.Entries[1].Inflow != 5000 {
		t.Fatalf("second entry = %+v", stmt.Entries[1])
	}
}

func TestParseUnknownFormat(t *testing.T) {
	_, err := Parse([]byte("this is not a statement"))
	if err == nil {
		t.Fatal("expected error for unknown statement format")
	}
}
