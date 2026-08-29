package mastercard

import (
	"fmt"
	"strconv"
	"strings"
)

// parseAmount converts an unsigned amount such as "12.50" or "1,695.53" into
// cents. The thousands separator is a comma and the decimal is a period.
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
		case r == ',':
			continue
		default:
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

// parseCreditAmount parses an amount that may carry a "CR" suffix, reporting
// whether it is a credit.
func parseCreditAmount(s string) (cents int64, isCredit bool, err error) {
	s = strings.TrimSpace(s)
	if strings.HasSuffix(s, "CR") {
		isCredit = true
		s = strings.TrimSpace(strings.TrimSuffix(s, "CR"))
	}
	cents, err = parseAmount(s)
	return cents, isCredit, err
}
