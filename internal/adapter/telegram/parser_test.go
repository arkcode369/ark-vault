package telegram

import (
	"testing"

	"github.com/arkcode369/ark-vault/internal/domain"
)

func TestParseJournalMessage_Valid(t *testing.T) {
	input := `#journal
Pair: XAUUSD
Type: BUY
RR: +2
Session: London
Confluence: FVG + OB mitigation on 15m`

	result := ParseJournalMessage(input)
	if result.Err != "" {
		t.Fatalf("unexpected error: %s", result.Err)
	}

	tr := result.Trade
	if tr.Symbol != "XAUUSD" {
		t.Errorf("symbol: got %q, want XAUUSD", tr.Symbol)
	}
	if tr.Direction != domain.DirBuy {
		t.Errorf("direction: got %q, want BUY", tr.Direction)
	}
	if tr.ResultRR != 2 {
		t.Errorf("rr: got %f, want 2", tr.ResultRR)
	}
	if tr.Status != domain.StatusWin {
		t.Errorf("status: got %q, want WIN", tr.Status)
	}
	if tr.TimeWindow != domain.SessionLondon {
		t.Errorf("session: got %q, want London", tr.TimeWindow)
	}
	if tr.Confluence != "FVG + OB mitigation on 15m" {
		t.Errorf("confluence: got %q", tr.Confluence)
	}
}

func TestParseJournalMessage_LossRR(t *testing.T) {
	input := `#journal
Pair: EURUSD
Type: SELL
RR: -1
Session: NY AM`

	result := ParseJournalMessage(input)
	if result.Err != "" {
		t.Fatalf("unexpected error: %s", result.Err)
	}

	tr := result.Trade
	if tr.Symbol != "EURUSD" {
		t.Errorf("symbol: got %q", tr.Symbol)
	}
	if tr.Direction != domain.DirSell {
		t.Errorf("direction: got %q", tr.Direction)
	}
	if tr.Status != domain.StatusLoss {
		t.Errorf("status: got %q", tr.Status)
	}
	if tr.ResultRR != -1 {
		t.Errorf("rr: got %f, want -1", tr.ResultRR)
	}
	if tr.TimeWindow != domain.SessionNYAM {
		t.Errorf("session: got %q, want NY AM", tr.TimeWindow)
	}
}

func TestParseJournalMessage_OpenTrade(t *testing.T) {
	input := `#journal
Pair: BTCUSD
Type: BUY`

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
RR: +2`

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
RR: +2`

	result := ParseJournalMessage(input)
	if result.Err == "" {
		t.Fatal("expected error for missing #journal tag")
	}
}

func TestParseJournalMessage_CaseInsensitive(t *testing.T) {
	input := `#JOURNAL
pair: eurusd
type: buy`

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
RR: 0`

	result := ParseJournalMessage(input)
	if result.Err != "" {
		t.Fatalf("unexpected error: %s", result.Err)
	}
	if result.Trade.Status != domain.StatusBE {
		t.Errorf("status: got %q, want BE", result.Trade.Status)
	}
	if result.Trade.ResultRR != 0 {
		t.Errorf("rr: got %f, want 0", result.Trade.ResultRR)
	}
}

func TestParseJournalMessage_WithSlash(t *testing.T) {
	input := `#journal
Pair: EUR/USD
Type: BUY`

	result := ParseJournalMessage(input)
	if result.Err != "" {
		t.Fatalf("unexpected error: %s", result.Err)
	}
	if result.Trade.Symbol != "EURUSD" {
		t.Errorf("symbol: got %q, want EURUSD (slash stripped)", result.Trade.Symbol)
	}
}

func TestParseJournalMessage_StatusFormat(t *testing.T) {
	input := `#journal
Pair: XAUUSD
Type: SELL
RR: WIN 2RR`

	result := ParseJournalMessage(input)
	if result.Err != "" {
		t.Fatalf("unexpected error: %s", result.Err)
	}
	if result.Trade.Status != domain.StatusWin {
		t.Errorf("status: got %q, want WIN", result.Trade.Status)
	}
	if result.Trade.ResultRR != 2 {
		t.Errorf("rr: got %f, want 2", result.Trade.ResultRR)
	}
}

func TestParseJournalMessage_LossStatusFormat(t *testing.T) {
	input := `#journal
Pair: EURUSD
Type: BUY
RR: LOSS 1RR`

	result := ParseJournalMessage(input)
	if result.Err != "" {
		t.Fatalf("unexpected error: %s", result.Err)
	}
	if result.Trade.Status != domain.StatusLoss {
		t.Errorf("status: got %q, want LOSS", result.Trade.Status)
	}
	if result.Trade.ResultRR != -1 {
		t.Errorf("rr: got %f, want -1", result.Trade.ResultRR)
	}
}
