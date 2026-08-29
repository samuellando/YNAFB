package visa

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

func isMinusSign(r rune) bool {
	switch r {
	case '-', '\u2010', '\u2011', '\u2012', '\u2013', '\u2014', '\u2015', '\u2212':
		return true
	}
	return false
}

// parseMoney converts a money string such as "$1,234.56", "–$99.00", or
// "(1,234.56)" into an amount in cents. A leading minus, en/em dash, or
// surrounding parentheses indicate a negative value.
func parseMoney(s string) (int64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, fmt.Errorf("empty amount")
	}

	negative := false

	if strings.HasPrefix(s, "(") && strings.HasSuffix(s, ")") {
		negative = true
		s = strings.TrimSpace(s[1 : len(s)-1])
	}

	// Strip a leading sign.
	runes := []rune(s)
	i := 0
	for i < len(runes) && (isMinusSign(runes[i]) || runes[i] == '+') {
		if isMinusSign(runes[i]) {
			negative = true
		}
		i++
	}
	s = strings.TrimSpace(string(runes[i:]))

	// Remove currency symbols, thousands separators, and stray letters/spaces.
	var cleaned strings.Builder
	for _, r := range s {
		switch {
		case r == '$' || r == '€' || r == '£' || r == ',' || r == ' ':
			continue
		case unicode.IsLetter(r):
			continue
		default:
			cleaned.WriteRune(r)
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

	cents := wholeCents*100 + fracCents
	if negative {
		cents = -cents
	}
	return cents, nil
}
