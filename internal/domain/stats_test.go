package domain

import "testing"

func TestCalculateStats(t *testing.T) {
	trades := []Trade{
		{Status: StatusWin, ResultPips: 50, RRRatio: 2.0, AssetType: AssetForex},
		{Status: StatusWin, ResultPips: 30, RRRatio: 1.5, AssetType: AssetForex},
		{Status: StatusLoss, ResultPips: -25, RRRatio: -1.0, AssetType: AssetGold},
		{Status: StatusWin, ResultPips: 80, RRRatio: 3.0, AssetType: AssetGold},
		{Status: StatusBE, ResultPips: 0, AssetType: AssetForex},
		{Status: StatusOpen, AssetType: AssetCrypto},
	}

	s := CalculateStats(trades)

	if s.TotalTrades != 6 {
		t.Errorf("total: got %d, want 6", s.TotalTrades)
	}
	if s.Wins != 3 {
		t.Errorf("wins: got %d, want 3", s.Wins)
	}
	if s.Losses != 1 {
		t.Errorf("losses: got %d, want 1", s.Losses)
	}
	if s.BreakEvens != 1 {
		t.Errorf("be: got %d, want 1", s.BreakEvens)
	}
	if s.OpenTrades != 1 {
		t.Errorf("open: got %d, want 1", s.OpenTrades)
	}

	// Win rate: 3 wins / (3+1) closed = 75%
	if s.WinRate != 75.0 {
		t.Errorf("win rate: got %.1f, want 75.0", s.WinRate)
	}

	// Total pips: 50+30-25+80 = 135
	if s.TotalPips != 135 {
		t.Errorf("total pips: got %.1f, want 135", s.TotalPips)
	}

	if s.BestPips != 80 {
		t.Errorf("best: got %.1f, want 80", s.BestPips)
	}
	if s.WorstPips != -25 {
		t.Errorf("worst: got %.1f, want -25", s.WorstPips)
	}
}

func TestCalculateStats_Empty(t *testing.T) {
	s := CalculateStats(nil)
	if s.TotalTrades != 0 {
		t.Errorf("expected 0 trades, got %d", s.TotalTrades)
	}
	if s.WinRate != 0 {
		t.Errorf("expected 0 win rate, got %f", s.WinRate)
	}
}

func TestDetectAssetType(t *testing.T) {
	tests := []struct {
		symbol string
		want   AssetType
	}{
		{"EURUSD", AssetForex},
		{"XAUUSD", AssetGold},
		{"NAS100", AssetIndices},
		{"BTCUSD", AssetCrypto},
		{"GBPJPY", AssetForex},
		{"GOLD", AssetGold},
		{"unknown", AssetForex}, // fallback
	}

	for _, tt := range tests {
		got := DetectAssetType(tt.symbol)
		if got != tt.want {
			t.Errorf("DetectAssetType(%q) = %q, want %q", tt.symbol, got, tt.want)
		}
	}
}
