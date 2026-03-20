package domain

import (
	"errors"
	"time"
)

// Direction of a trade.
type Direction string

const (
	DirBuy  Direction = "BUY"
	DirSell Direction = "SELL"
)

// TradeStatus captures the lifecycle state of a trade.
type TradeStatus string

const (
	StatusOpen TradeStatus = "OPEN"
	StatusWin  TradeStatus = "WIN"
	StatusLoss TradeStatus = "LOSS"
	StatusBE   TradeStatus = "BE" // break-even
)

// Trade is the core domain entity for a single trade journal entry.
type Trade struct {
	ID          string // Notion page ID of the trade entry
	MemberID    int64  // Telegram user ID
	Date        time.Time
	AssetType   AssetType
	Symbol      string
	Direction   Direction
	EntryPrice  float64
	StopLoss    float64
	TakeProfit  float64
	ClosePrice  float64
	ResultPips  float64
	RRRatio     float64
	Status      TradeStatus
	Notes       string
	ScreenshotURL string // URL after uploaded to Notion
}

// Validate performs basic checks on the trade before persisting.
func (t *Trade) Validate() error {
	if t.Symbol == "" {
		return errors.New("symbol is required")
	}
	if t.Direction != DirBuy && t.Direction != DirSell {
		return errors.New("direction must be BUY or SELL")
	}
	if t.EntryPrice <= 0 {
		return errors.New("entry price must be positive")
	}
	if t.StopLoss <= 0 {
		return errors.New("stop loss must be positive")
	}
	if t.TakeProfit <= 0 {
		return errors.New("take profit must be positive")
	}
	return nil
}

// AutoDetectAsset sets AssetType from Symbol if not already set.
func (t *Trade) AutoDetectAsset() {
	if !t.AssetType.IsValid() || t.AssetType == "" {
		t.AssetType = DetectAssetType(t.Symbol)
	}
}
