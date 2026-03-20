package service

import (
	"context"
	"fmt"
	"sort"

	"github.com/arkcode369/ark-vault/internal/domain"
	"github.com/arkcode369/ark-vault/internal/ports"
)

// LeaderboardEntry represents one member's ranking.
type LeaderboardEntry struct {
	Username   string
	FirstName  string
	TelegramID int64
	WinRate    float64
	TotalPips  float64
	TotalTrades int
}

// LeaderboardService computes community rankings.
type LeaderboardService struct {
	trades  ports.TradeRepository
	members ports.MemberRepository
}

// NewLeaderboardService creates a new LeaderboardService.
func NewLeaderboardService(tr ports.TradeRepository, mr ports.MemberRepository) *LeaderboardService {
	return &LeaderboardService{trades: tr, members: mr}
}

// GetLeaderboard returns top N members sorted by the given metric.
// metric: "winrate" or "pips". minTrades is the minimum trade count to qualify.
func (s *LeaderboardService) GetLeaderboard(ctx context.Context, metric string, topN, minTrades int) ([]LeaderboardEntry, error) {
	members, err := s.members.ListMembers(ctx)
	if err != nil {
		return nil, fmt.Errorf("list members: %w", err)
	}

	var entries []LeaderboardEntry
	for _, m := range members {
		trades, err := s.trades.GetTrades(ctx, m.TelegramID)
		if err != nil {
			continue
		}
		stats := domain.CalculateStats(trades)
		if stats.TotalTrades < minTrades {
			continue
		}
		entries = append(entries, LeaderboardEntry{
			Username:    m.Username,
			FirstName:   m.FirstName,
			TelegramID:  m.TelegramID,
			WinRate:     stats.WinRate,
			TotalPips:   stats.TotalPips,
			TotalTrades: stats.TotalTrades,
		})
	}

	switch metric {
	case "pips":
		sort.Slice(entries, func(i, j int) bool {
			return entries[i].TotalPips > entries[j].TotalPips
		})
	default: // winrate
		sort.Slice(entries, func(i, j int) bool {
			return entries[i].WinRate > entries[j].WinRate
		})
	}

	if topN > 0 && len(entries) > topN {
		entries = entries[:topN]
	}
	return entries, nil
}
