package domain

import "testing"

func TestCalculateStats(t *testing.T) {
	trades := []Trade{
		{Status: StatusWin, ResultRR: 2.0, AssetType: AssetForex},
		{Status: StatusWin, ResultRR: 1.5, AssetType: AssetForex},
		{Status: StatusLoss, ResultRR: -1.0, AssetType: AssetGold},
		{Status: StatusWin, ResultRR: 3.0, AssetType: AssetGold},
		{Status: StatusBE, ResultRR: 0, AssetType: AssetForex},
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

	// Total RR: 2.0+1.5-1.0+3.0 = 5.5
	if s.TotalRR != 5.5 {
		t.Errorf("total RR: got %.1f, want 5.5", s.TotalRR)
	}

	if s.BestRR != 3.0 {
		t.Errorf("best: got %.1f, want 3.0", s.BestRR)
	}
	if s.WorstRR != -1.0 {
		t.Errorf("worst: got %.1f, want -1.0", s.WorstRR)
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
