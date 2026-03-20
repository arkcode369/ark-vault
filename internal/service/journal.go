package service

import (
	"context"
	"fmt"
	"time"

	"github.com/arkcode369/ark-vault/internal/domain"
	"github.com/arkcode369/ark-vault/internal/ports"
)

// JournalService orchestrates trade recording.
type JournalService struct {
	trades  ports.TradeRepository
	members ports.MemberRepository
	images  ports.ImageStore
}

// NewJournalService creates a new JournalService.
func NewJournalService(tr ports.TradeRepository, mr ports.MemberRepository, is ports.ImageStore) *JournalService {
	return &JournalService{trades: tr, members: mr, images: is}
}

// RecordTrade validates, enriches, and persists a new trade.
func (s *JournalService) RecordTrade(ctx context.Context, memberTgID int64, username, firstName string, trade *domain.Trade) error {
	trade.AutoDetectAsset()
	if err := trade.Validate(); err != nil {
		return fmt.Errorf("validation: %w", err)
	}
	if trade.Date.IsZero() {
		trade.Date = time.Now().UTC()
	}
	if trade.Status == "" {
		trade.Status = domain.StatusOpen
	}
	trade.MemberID = memberTgID

	// Ensure member exists in Notion
	member := &domain.Member{
		TelegramID: memberTgID,
		Username:   username,
		FirstName:  firstName,
		JoinDate:   time.Now().UTC(),
	}
	member, err := s.members.EnsureMember(ctx, member)
	if err != nil {
		return fmt.Errorf("ensure member: %w", err)
	}

	// Save trade
	if err := s.trades.SaveTrade(ctx, memberTgID, trade); err != nil {
		return fmt.Errorf("save trade: %w", err)
	}

	return nil
}

// UploadScreenshot attaches a screenshot to an existing trade entry.
func (s *JournalService) UploadScreenshot(ctx context.Context, tradeID string, filename string, data []byte) error {
	if s.images == nil {
		return nil // image store not configured
	}
	_, err := s.images.Upload(ctx, tradeID, filename, data)
	return err
}

// GetMemberStats retrieves all trades for a member and calculates stats.
func (s *JournalService) GetMemberStats(ctx context.Context, memberTgID int64) (*domain.Stats, error) {
	trades, err := s.trades.GetTrades(ctx, memberTgID)
	if err != nil {
		return nil, fmt.Errorf("get trades: %w", err)
	}
	if len(trades) == 0 {
		return nil, nil
	}
	stats := domain.CalculateStats(trades)
	return &stats, nil
}

// CloseTrade updates a trade with close price and result.
func (s *JournalService) CloseTrade(ctx context.Context, tradeID string, closePrice float64, resultPips float64, status domain.TradeStatus) error {
	trade, err := s.trades.GetTradeByID(ctx, tradeID)
	if err != nil {
		return fmt.Errorf("get trade: %w", err)
	}
	if trade.Status != domain.StatusOpen {
		return fmt.Errorf("trade is already closed (status: %s)", trade.Status)
	}
	trade.ClosePrice = closePrice
	trade.ResultPips = resultPips
	trade.Status = status
	if trade.StopLoss != 0 && trade.EntryPrice != 0 {
		risk := abs(trade.EntryPrice - trade.StopLoss)
		if risk > 0 {
			trade.RRRatio = abs(resultPips) / risk
			if status == domain.StatusLoss {
				trade.RRRatio = -trade.RRRatio
			}
		}
	}
	return s.trades.UpdateTrade(ctx, trade)
}

func abs(f float64) float64 {
	if f < 0 {
		return -f
	}
	return f
}
