package exporter

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"

	"github.com/arkcode369/ark-vault/internal/domain"
)

// CSVExporter implements ports.Exporter for CSV output.
type CSVExporter struct{}

// NewCSVExporter creates a CSVExporter.
func NewCSVExporter() *CSVExporter {
	return &CSVExporter{}
}

// ExportCSV produces a CSV byte slice for the given trades.
func (e *CSVExporter) ExportCSV(_ context.Context, trades []domain.Trade) ([]byte, error) {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)

	// Header
	header := []string{
		"Date", "Asset Type", "Symbol", "Direction",
		"Status", "Result RR", "Time Window",
		"Confluence", "Notes",
	}
	if err := w.Write(header); err != nil {
		return nil, err
	}

	for _, t := range trades {
		row := []string{
			t.Date.Format("2006-01-02"),
			t.AssetType.String(),
			t.Symbol,
			string(t.Direction),
			string(t.Status),
			fmt.Sprintf("%.2f", t.ResultRR),
			string(t.TimeWindow),
			t.Confluence,
			t.Notes,
		}
		if err := w.Write(row); err != nil {
			return nil, err
		}
	}

	w.Flush()
	return buf.Bytes(), w.Error()
}
