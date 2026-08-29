package checking

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

func isThousandsSeparator(r rune) bool {
	return unicode.IsSpace(r) || r == ',' || r == '\u00A0' || r == '\u202F' || r == '\u2009' || r == '\u2007'
}

// parseAmount converts a French-formatted, unsigned amount such as
// "3 584.59" or "47.78" into cents. The thousands separator is a space
// (possibly non-breaking) and the decimal separator is a period.
func parseAmount(s string) (int64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, fmt.Errorf("empty amount")
	}

	var cleaned strings.Builder
	for _, r := range s {
		switch {
		case r == '.':
			cleaned.WriteRune(r)
		case r >= '0' && r <= '9':
			cleaned.WriteRune(r)
		case isThousandsSeparator(r):
			continue
		default:
			// currency symbol or stray letter
			continue
		}
	}

	digits := cleaned.String()
	whole := digits
	frac := ""
	if idx := strings.Index(digits, "."); idx >= 0 {
		whole = digits[:idx]
		frac = digits[idx+1:]
		if strings.Contains(frac, ".") {
			return 0, fmt.Errorf("invalid amount %q", s)
		}
	}

	if whole == "" {
		whole = "0"
	}
	if frac == "" {
		frac = "0"
	}

	wholeCents, err := strconv.ParseInt(whole, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid amount %q: %w", s, err)
	}

	for len(frac) < 2 {
		frac += "0"
	}
	if len(frac) > 2 {
		frac = frac[:2]
	}
	fracCents, err := strconv.ParseInt(frac, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid amount %q: %w", s, err)
	}

	return wholeCents*100 + fracCents, nil
}
