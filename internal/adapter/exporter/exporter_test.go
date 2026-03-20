package exporter

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/arkcode369/ark-vault/internal/domain"
)

func TestCSVExport(t *testing.T) {
	exp := NewCSVExporter()
	trades := []domain.Trade{
		{
			Date:       time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC),
			AssetType:  domain.AssetForex,
			Symbol:     "EURUSD",
			Direction:  domain.DirBuy,
			EntryPrice: 1.085,
			StopLoss:   1.08,
			TakeProfit: 1.095,
			ResultPips: 50,
			Status:     domain.StatusWin,
		},
	}

	data, err := exp.ExportCSV(context.Background(), trades)
	if err != nil {
		t.Fatal(err)
	}

	csv := string(data)
	if !bytes.Contains([]byte(csv), []byte("EURUSD")) {
		t.Error("CSV should contain EURUSD")
	}
	if !bytes.Contains([]byte(csv), []byte("BUY")) {
		t.Error("CSV should contain BUY")
	}
	if !bytes.Contains([]byte(csv), []byte("WIN")) {
		t.Error("CSV should contain WIN")
	}
}

func TestPDFExport(t *testing.T) {
	exp := NewPDFExporter()
	trades := []domain.Trade{
		{
			Date:       time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC),
			AssetType:  domain.AssetForex,
			Symbol:     "EURUSD",
			Direction:  domain.DirBuy,
			EntryPrice: 1.085,
			StopLoss:   1.08,
			TakeProfit: 1.095,
			ResultPips: 50,
			RRRatio:    2.0,
			Status:     domain.StatusWin,
		},
	}
	stats := &domain.Stats{
		TotalTrades: 1,
		Wins:        1,
		WinRate:     100,
		TotalPips:   50,
		BestPips:    50,
		AvgRR:       2.0,
	}

	data, err := exp.ExportPDF(context.Background(), "testuser", trades, stats)
	if err != nil {
		t.Fatal(err)
	}

	// Check it's a valid PDF (starts with %PDF-)
	if !bytes.HasPrefix(data, []byte("%PDF-")) {
		t.Error("output should be a valid PDF")
	}
	// Check it ends with %%EOF
	if !bytes.Contains(data, []byte("%%EOF")) {
		t.Error("PDF should contain EOF marker")
	}
}
