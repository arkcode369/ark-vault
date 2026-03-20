package formatter

import (
	"fmt"
	"strings"
)

// FormatPrice formats a price with appropriate decimal places.
func FormatPrice(price float64) string {
	// Forex pairs usually 4-5 decimals, gold/indices 2 decimals
	if price >= 100 {
		return fmt.Sprintf("%.2f", price)
	}
	return fmt.Sprintf("%.5f", price)
}

// FormatPips formats pips with sign.
func FormatPips(pips float64) string {
	return fmt.Sprintf("%+.1f", pips)
}

// FormatPercent formats a percentage.
func FormatPercent(pct float64) string {
	return fmt.Sprintf("%.1f%%", pct)
}

// SanitizeSymbol normalises a symbol string.
func SanitizeSymbol(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ToUpper(s)
	s = strings.ReplaceAll(s, "/", "")
	s = strings.ReplaceAll(s, " ", "")
	return s
}
