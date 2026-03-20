package telegram

import (
	"testing"

	"github.com/arkcode369/ark-vault/internal/domain"
)

func TestParseJournalMessage_Valid(t *testing.T) {
	input := `#journal
Pair: EURUSD
Type: BUY
Entry: 1.0850
SL: 1.0800
TP: 1.0950
Result: WIN +50 pips`

	result := ParseJournalMessage(input)
	if result.Err != "" {
		t.Fatalf("unexpected error: %s", result.Err)
	}

	tr := result.Trade
	if tr.Symbol != "EURUSD" {
		t.Errorf("symbol: got %q, want EURUSD", tr.Symbol)
	}
	if tr.Direction != domain.DirBuy {
		t.Errorf("direction: got %q, want BUY", tr.Direction)
	}
	if tr.EntryPrice != 1.0850 {
		t.Errorf("entry: got %f, want 1.0850", tr.EntryPrice)
	}
	if tr.StopLoss != 1.0800 {
		t.Errorf("sl: got %f, want 1.0800", tr.StopLoss)
	}
	if tr.TakeProfit != 1.0950 {
		t.Errorf("tp: got %f, want 1.0950", tr.TakeProfit)
	}
	if tr.Status != domain.StatusWin {
		t.Errorf("status: got %q, want WIN", tr.Status)
	}
	if tr.ResultPips != 50 {
		t.Errorf("pips: got %f, want 50", tr.ResultPips)
	}
}

func TestParseJournalMessage_GoldSell(t *testing.T) {
	input := `#journal
Pair: XAUUSD
Type: SELL
Entry: 2345.50
SL: 2360.00
TP: 2320.00
Result: LOSS -145 pips`

	result := ParseJournalMessage(input)
	if result.Err != "" {
		t.Fatalf("unexpected error: %s", result.Err)
	}

	tr := result.Trade
	if tr.Symbol != "XAUUSD" {
		t.Errorf("symbol: got %q", tr.Symbol)
	}
	if tr.Direction != domain.DirSell {
		t.Errorf("direction: got %q", tr.Direction)
	}
	if tr.Status != domain.StatusLoss {
		t.Errorf("status: got %q", tr.Status)
	}
	if tr.ResultPips != -145 {
		t.Errorf("pips: got %f, want -145", tr.ResultPips)
	}
}

func TestParseJournalMessage_OpenTrade(t *testing.T) {
	input := `#journal
Pair: BTCUSD
Type: BUY
Entry: 65000
SL: 63000
TP: 70000`

	result := ParseJournalMessage(input)
	if result.Err != "" {
		t.Fatalf("unexpected error: %s", result.Err)
	}
	if result.Trade.Status != "" {
		t.Errorf("status should be empty for open trade, got %q", result.Trade.Status)
	}
}

func TestParseJournalMessage_MissingFields(t *testing.T) {
	input := `#journal
Pair: EURUSD
Type: BUY`

	result := ParseJournalMessage(input)
	if result.Err == "" {
		t.Fatal("expected error for missing fields")
	}
	if result.Trade != nil {
		t.Fatal("should not return trade on error")
	}
}

func TestParseJournalMessage_NoTag(t *testing.T) {
	input := `Pair: EURUSD
Type: BUY
Entry: 1.0850
SL: 1.0800
TP: 1.0950`

	result := ParseJournalMessage(input)
	if result.Err == "" {
		t.Fatal("expected error for missing #journal tag")
	}
}

func TestParseJournalMessage_CaseInsensitive(t *testing.T) {
	input := `#JOURNAL
pair: eurusd
type: buy
entry: 1.0850
sl: 1.0800
tp: 1.0950`

	result := ParseJournalMessage(input)
	if result.Err != "" {
		t.Fatalf("unexpected error: %s", result.Err)
	}
	if result.Trade.Symbol != "EURUSD" {
		t.Errorf("symbol: got %q", result.Trade.Symbol)
	}
	if result.Trade.Direction != domain.DirBuy {
		t.Errorf("direction: got %q", result.Trade.Direction)
	}
}

func TestParseJournalMessage_BreakEven(t *testing.T) {
	input := `#journal
Pair: GBPJPY
Type: BUY
Entry: 189.50
SL: 189.00
TP: 190.50
Result: BE`

	result := ParseJournalMessage(input)
	if result.Err != "" {
		t.Fatalf("unexpected error: %s", result.Err)
	}
	if result.Trade.Status != domain.StatusBE {
		t.Errorf("status: got %q, want BE", result.Trade.Status)
	}
}

func TestParseJournalMessage_WithSlash(t *testing.T) {
	input := `#journal
Pair: EUR/USD
Type: BUY
Entry: 1.0850
SL: 1.0800
TP: 1.0950`

	result := ParseJournalMessage(input)
	if result.Err != "" {
		t.Fatalf("unexpected error: %s", result.Err)
	}
	if result.Trade.Symbol != "EURUSD" {
		t.Errorf("symbol: got %q, want EURUSD (slash stripped)", result.Trade.Symbol)
	}
}
