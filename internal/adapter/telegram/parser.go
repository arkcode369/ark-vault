package telegram

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/arkcode369/ark-vault/internal/domain"
)

var (
	journalTag = regexp.MustCompile(`(?i)^#journal`)
	fieldRe    = regexp.MustCompile(`(?i)^(pair|symbol|type|direction|entry|sl|stoploss|stop\s*loss|tp|takeprofit|take\s*profit|result|notes?)\s*[:=]\s*(.+)$`)
	resultRe   = regexp.MustCompile(`(?i)(WIN|LOSS|BE)\s*([+-]?\d+\.?\d*)?\s*(pips)?`)
)

// ParseResult holds the output of parsing a #journal message.
type ParseResult struct {
	Trade *domain.Trade
	Err   string // user-friendly error message if parsing failed
}

// ParseJournalMessage extracts trade data from a structured text message.
func ParseJournalMessage(text string) ParseResult {
	lines := strings.Split(strings.TrimSpace(text), "\n")
	if len(lines) == 0 || !journalTag.MatchString(lines[0]) {
		return ParseResult{Err: "Message harus diawali dengan #journal"}
	}

	t := &domain.Trade{}
	parsed := make(map[string]bool)

	for _, line := range lines[1:] {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		matches := fieldRe.FindStringSubmatch(line)
		if matches == nil {
			continue
		}
		key := strings.ToLower(matches[1])
		val := strings.TrimSpace(matches[2])

		switch {
		case key == "pair" || key == "symbol":
			t.Symbol = strings.ToUpper(strings.ReplaceAll(val, "/", ""))
			parsed["symbol"] = true
		case key == "type" || key == "direction":
			dir := strings.ToUpper(val)
			if dir == "BUY" || dir == "LONG" {
				t.Direction = domain.DirBuy
			} else if dir == "SELL" || dir == "SHORT" {
				t.Direction = domain.DirSell
			}
			parsed["direction"] = true
		case key == "entry":
			if v, err := strconv.ParseFloat(val, 64); err == nil {
				t.EntryPrice = v
				parsed["entry"] = true
			}
		case key == "sl" || strings.HasPrefix(key, "stop"):
			if v, err := strconv.ParseFloat(val, 64); err == nil {
				t.StopLoss = v
				parsed["sl"] = true
			}
		case key == "tp" || strings.HasPrefix(key, "take"):
			if v, err := strconv.ParseFloat(val, 64); err == nil {
				t.TakeProfit = v
				parsed["tp"] = true
			}
		case key == "result":
			parseResult(val, t)
			parsed["result"] = true
		case strings.HasPrefix(key, "note"):
			t.Notes = val
		}
	}

	// Validate required fields
	var missing []string
	if !parsed["symbol"] {
		missing = append(missing, "Pair/Symbol")
	}
	if !parsed["direction"] {
		missing = append(missing, "Type/Direction")
	}
	if !parsed["entry"] {
		missing = append(missing, "Entry")
	}
	if !parsed["sl"] {
		missing = append(missing, "SL")
	}
	if !parsed["tp"] {
		missing = append(missing, "TP")
	}

	if len(missing) > 0 {
		return ParseResult{
			Err: "Field berikut belum diisi: " + strings.Join(missing, ", ") +
				"\n\nFormat yang benar:\n#journal\nPair: EURUSD\nType: BUY\nEntry: 1.0850\nSL: 1.0800\nTP: 1.0950\nResult: WIN +50 pips",
		}
	}

	return ParseResult{Trade: t}
}

func parseResult(val string, t *domain.Trade) {
	m := resultRe.FindStringSubmatch(val)
	if m == nil {
		return
	}
	switch strings.ToUpper(m[1]) {
	case "WIN":
		t.Status = domain.StatusWin
	case "LOSS":
		t.Status = domain.StatusLoss
	case "BE":
		t.Status = domain.StatusBE
	}
	if m[2] != "" {
		if pips, err := strconv.ParseFloat(m[2], 64); err == nil {
			t.ResultPips = pips
		}
	}
}
