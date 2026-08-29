package importer

import "testing"

func TestParseMoney(t *testing.T) {
	tests := []struct {
		in   string
		want int64
	}{
		{"$26.62", 2662},
		{"$1,330.90", 133090},
		{"–$1,330.90", -133090},
		{"-$4,000.00", -400000},
		{"$5,593.53", 559353},
		{"$0.00", 0},
		{"$1,699.48", 169948},
		{"(1,234.56)", -123456},
		{"+$12.34", 1234},
		{"$0.5", 50},
		{"$100", 10000},
	}
	for _, tt := range tests {
		got, err := ParseMoney(tt.in)
		if err != nil {
			t.Fatalf("ParseMoney(%q) error: %v", tt.in, err)
		}
		if got != tt.want {
			t.Errorf("ParseMoney(%q) = %d, want %d", tt.in, got, tt.want)
		}
	}
}

func TestExtractTrailingAmount(t *testing.T) {
	cents, details, ok := extractTrailingAmount("CREMES BOBOULES$26.62")
	if !ok || cents != 2662 || details != "CREMES BOBOULES" {
		t.Fatalf("got (%d, %q, %v)", cents, details, ok)
	}

	cents, details, ok = extractTrailingAmount("From chequing account–$1,330.90")
	if !ok || cents != -133090 || details != "From chequing account" {
		t.Fatalf("got (%d, %q, %v)", cents, details, ok)
	}

	cents, details, ok = extractTrailingAmount("JEJU AIR 4438169296")
	if ok {
		t.Fatalf("expected no amount, got %d %q", cents, details)
	}
}
