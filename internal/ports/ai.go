package ports

import "context"

// AIAnalyzer provides AI-powered trade analysis.
type AIAnalyzer interface {
	// AnalyzeTrades sends trade data to AI and returns insights in Bahasa Indonesia.
	AnalyzeTrades(ctx context.Context, prompt string) (string, error)
}
