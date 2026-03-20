package ports

import (
	"context"

	"github.com/arkcode369/ark-vault/internal/domain"
)

// Exporter generates downloadable files from trade data.
type Exporter interface {
	// ExportCSV produces a CSV byte slice for the given trades.
	ExportCSV(ctx context.Context, trades []domain.Trade) ([]byte, error)

	// ExportPDF produces a PDF byte slice with stats summary and trade history.
	ExportPDF(ctx context.Context, username string, trades []domain.Trade, stats *domain.Stats) ([]byte, error)
}
