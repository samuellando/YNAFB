package importer

import (
	"testing"
	"time"
)

func TestParsePeriod(t *testing.T) {
	p, err := parsePeriod("Jul 15 — Aug 14, 2026")
	if err != nil {
		t.Fatalf("parsePeriod error: %v", err)
	}
	wantStart := time.Date(2026, time.July, 15, 0, 0, 0, 0, time.UTC)
	wantEnd := time.Date(2026, time.August, 14, 0, 0, 0, 0, time.UTC)
	if !p.start.Equal(wantStart) || !p.end.Equal(wantEnd) {
		t.Fatalf("got %v..%v, want %v..%v", p.start, p.end, wantStart, wantEnd)
	}
}

func TestParsePeriodYearWrap(t *testing.T) {
	p, err := parsePeriod("Dec 15 — Jan 14, 2026")
	if err != nil {
		t.Fatalf("parsePeriod error: %v", err)
	}
	if p.start.Year() != 2025 {
		t.Fatalf("start year = %d, want 2025", p.start.Year())
	}
	if p.end.Year() != 2026 {
		t.Fatalf("end year = %d, want 2026", p.end.Year())
	}
}

func TestPeriodResolveDate(t *testing.T) {
	p, err := parsePeriod("Jul 15 — Aug 14, 2026")
	if err != nil {
		t.Fatalf("parsePeriod error: %v", err)
	}

	got := p.resolveDate(time.July, 16)
	want := time.Date(2026, time.July, 16, 0, 0, 0, 0, time.UTC)
	if !got.Equal(want) {
		t.Fatalf("resolveDate = %v, want %v", got, want)
	}
}
