package exporter

import (
	"context"

	"github.com/arkcode369/ark-vault/internal/domain"
)

// Exporter combines CSV and PDF export into a single struct implementing ports.Exporter.
type Exporter struct {
	csv *CSVExporter
	pdf *PDFExporter
}

// NewExporter creates a combined Exporter.
func NewExporter() *Exporter {
	return &Exporter{
		csv: NewCSVExporter(),
		pdf: NewPDFExporter(),
	}
}

// ExportCSV delegates to CSVExporter.
func (e *Exporter) ExportCSV(ctx context.Context, trades []domain.Trade) ([]byte, error) {
	return e.csv.ExportCSV(ctx, trades)
}

// ExportPDF delegates to PDFExporter.
func (e *Exporter) ExportPDF(ctx context.Context, username string, trades []domain.Trade, stats *domain.Stats) ([]byte, error) {
	return e.pdf.ExportPDF(ctx, username, trades, stats)
}
