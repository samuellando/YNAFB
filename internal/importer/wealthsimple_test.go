package importer

import (
	"path/filepath"
	"testing"
	"time"
)

func TestParseWealthsimpleStatement(t *testing.T) {
	stmt, err := ImportFile(filepath.Join("..", "..", "transactions.pdf"))
	if err != nil {
		t.Fatalf("import: %v", err)
	}

	if stmt.Institution != "Wealthsimple" {
		t.Fatalf("institution = %q", stmt.Institution)
	}
	if stmt.Account != "8807" {
		t.Fatalf("account = %q, want 8807", stmt.Account)
	}

	want := time.Date(2026, time.July, 15, 0, 0, 0, 0, time.UTC)
	if !stmt.Start.Equal(want) {
		t.Fatalf("start = %v, want %v", stmt.Start, want)
	}

	if len(stmt.Entries) != 51 {
		t.Fatalf("got %d entries, want 51", len(stmt.Entries))
	}

	first := stmt.Entries[0]
	if first.Type != "payment" {
		t.Fatalf("first type = %q", first.Type)
	}
	if first.Inflow != 133090 || first.Outflow != 0 {
		t.Fatalf("first amounts = out %d in %d", first.Outflow, first.Inflow)
	}
	if first.Payee != "From chequing account" {
		t.Fatalf("first payee = %q", first.Payee)
	}
	if first.TransDate.Year() != 2026 || first.TransDate.Month() != time.July || first.TransDate.Day() != 16 {
		t.Fatalf("first date = %v", first.TransDate)
	}

	jeju := findEntry(t, stmt, "JEJU AIR 4438169296")
	if jeju == nil {
		t.Fatal("missing JEJU AIR entry")
	}
	if jeju.Outflow != 43736 {
		t.Fatalf("jeju outflow = %d", jeju.Outflow)
	}
	if jeju.Note == "" {
		t.Fatal("expected FX note on JEJU AIR entry")
	}
	if jeju.TransDate.Day() != 21 || jeju.PostedDate.Day() != 23 {
		t.Fatalf("jeju dates = trans %v posted %v", jeju.TransDate, jeju.PostedDate)
	}

	cremes := findEntry(t, stmt, "CREMES BOBOULES")
	if cremes == nil || cremes.Outflow != 2662 {
		t.Fatalf("cremes entry missing or wrong: %+v", cremes)
	}

	last := stmt.Entries[len(stmt.Entries)-1]
	if last.Payee != "GREENSPOT RESTAURANT" || last.Outflow != 4967 {
		t.Fatalf("last entry = %+v", last)
	}
}

func findEntry(t *testing.T, stmt Statement, payee string) *Entry {
	t.Helper()
	for i := range stmt.Entries {
		if stmt.Entries[i].Payee == payee {
			return &stmt.Entries[i]
		}
	}
	return nil
}
