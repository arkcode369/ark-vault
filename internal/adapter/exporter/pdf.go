package exporter

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/arkcode369/ark-vault/internal/domain"
)

// PDFExporter generates a simple text-based PDF report.
// Uses a minimal PDF generator without external dependencies.
type PDFExporter struct{}

// NewPDFExporter creates a PDFExporter.
func NewPDFExporter() *PDFExporter {
	return &PDFExporter{}
}

// ExportPDF produces a minimal PDF byte slice for the given trades and stats.
func (e *PDFExporter) ExportPDF(_ context.Context, username string, trades []domain.Trade, stats *domain.Stats) ([]byte, error) {
	var content strings.Builder

	content.WriteString(fmt.Sprintf("ARK Vault — Trade Journal Report\n"))
	content.WriteString(fmt.Sprintf("Member: @%s\n", username))
	content.WriteString(fmt.Sprintf("Generated: %s\n", trades[0].Date.Format("2006-01-02")))
	content.WriteString(strings.Repeat("=", 50) + "\n\n")

	// Stats summary
	if stats != nil {
		content.WriteString("PERFORMANCE SUMMARY\n")
		content.WriteString(strings.Repeat("-", 30) + "\n")
		content.WriteString(fmt.Sprintf("Total Trades:    %d\n", stats.TotalTrades))
		content.WriteString(fmt.Sprintf("Win/Loss/BE:     %d / %d / %d\n", stats.Wins, stats.Losses, stats.BreakEvens))
		content.WriteString(fmt.Sprintf("Win Rate:        %.1f%%\n", stats.WinRate))
		content.WriteString(fmt.Sprintf("Total Pips:      %+.1f\n", stats.TotalPips))
		content.WriteString(fmt.Sprintf("Avg RR:          %.2f\n", stats.AvgRR))
		content.WriteString(fmt.Sprintf("Best Trade:      %+.1f pips\n", stats.BestPips))
		content.WriteString(fmt.Sprintf("Worst Trade:     %+.1f pips\n", stats.WorstPips))
		content.WriteString(fmt.Sprintf("Max Win Streak:  %d\n\n", stats.MaxWinStrk))
	}

	// Trade list
	content.WriteString("TRADE HISTORY\n")
	content.WriteString(strings.Repeat("-", 30) + "\n")
	for i, t := range trades {
		content.WriteString(fmt.Sprintf("\n#%d  %s  %s %s\n", i+1, t.Date.Format("2006-01-02"), t.Symbol, t.Direction))
		content.WriteString(fmt.Sprintf("    Entry: %g  SL: %g  TP: %g\n", t.EntryPrice, t.StopLoss, t.TakeProfit))
		if t.ClosePrice != 0 {
			content.WriteString(fmt.Sprintf("    Close: %g\n", t.ClosePrice))
		}
		content.WriteString(fmt.Sprintf("    Status: %s", t.Status))
		if t.ResultPips != 0 {
			content.WriteString(fmt.Sprintf("  Pips: %+.1f  RR: %.2f", t.ResultPips, t.RRRatio))
		}
		content.WriteString("\n")
	}

	// Generate minimal valid PDF
	return generateMinimalPDF(content.String()), nil
}

// generateMinimalPDF creates a bare-minimum valid PDF with text content.
func generateMinimalPDF(text string) []byte {
	var buf bytes.Buffer

	lines := strings.Split(text, "\n")

	// PDF Header
	buf.WriteString("%PDF-1.4\n")

	// Object 1: Catalog
	obj1Offset := buf.Len()
	buf.WriteString("1 0 obj\n<< /Type /Catalog /Pages 2 0 R >>\nendobj\n")

	// Object 2: Pages
	obj2Offset := buf.Len()
	buf.WriteString("2 0 obj\n<< /Type /Pages /Kids [3 0 R] /Count 1 >>\nendobj\n")

	// Object 4: Font
	obj4Offset := buf.Len()
	buf.WriteString("4 0 obj\n<< /Type /Font /Subtype /Type1 /BaseFont /Courier >>\nendobj\n")

	// Object 5: Stream content
	var stream bytes.Buffer
	stream.WriteString("BT\n/F1 9 Tf\n")
	y := 780.0
	for _, line := range lines {
		if y < 40 {
			break // simple single-page limit
		}
		escaped := pdfEscape(line)
		stream.WriteString(fmt.Sprintf("1 0 0 1 40 %.0f Tm\n(%s) Tj\n", y, escaped))
		y -= 12
	}
	stream.WriteString("ET\n")

	obj5Offset := buf.Len()
	streamBytes := stream.Bytes()
	buf.WriteString(fmt.Sprintf("5 0 obj\n<< /Length %d >>\nstream\n", len(streamBytes)))
	buf.Write(streamBytes)
	buf.WriteString("\nendstream\nendobj\n")

	// Object 3: Page
	obj3Offset := buf.Len()
	buf.WriteString("3 0 obj\n<< /Type /Page /Parent 2 0 R /MediaBox [0 0 595 842] /Contents 5 0 R /Resources << /Font << /F1 4 0 R >> >> >>\nendobj\n")

	// Cross-reference table
	xrefOffset := buf.Len()
	buf.WriteString("xref\n0 6\n")
	buf.WriteString("0000000000 65535 f \n")
	buf.WriteString(fmt.Sprintf("%010d 00000 n \n", obj1Offset))
	buf.WriteString(fmt.Sprintf("%010d 00000 n \n", obj2Offset))
	buf.WriteString(fmt.Sprintf("%010d 00000 n \n", obj3Offset))
	buf.WriteString(fmt.Sprintf("%010d 00000 n \n", obj4Offset))
	buf.WriteString(fmt.Sprintf("%010d 00000 n \n", obj5Offset))

	// Trailer
	buf.WriteString(fmt.Sprintf("trailer\n<< /Size 6 /Root 1 0 R >>\nstartxref\n%d\n%%%%EOF\n", xrefOffset))

	return buf.Bytes()
}

func pdfEscape(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "(", "\\(")
	s = strings.ReplaceAll(s, ")", "\\)")
	return s
}
