package ports

import (
	"context"

	"github.com/arkcode369/ark-vault/internal/domain"
)

// TradeRepository defines persistence operations for trades.
type TradeRepository interface {
	// SaveTrade creates a new trade entry for the given member.
	SaveTrade(ctx context.Context, memberID int64, trade *domain.Trade) error

	// GetTrades returns all trades for a member, ordered by date descending.
	GetTrades(ctx context.Context, memberID int64) ([]domain.Trade, error)

	// GetTradeByID returns a single trade.
	GetTradeByID(ctx context.Context, tradeID string) (*domain.Trade, error)

	// UpdateTrade persists changes to an existing trade.
	UpdateTrade(ctx context.Context, trade *domain.Trade) error
}

// MemberRepository defines persistence operations for members.
type MemberRepository interface {
	// EnsureMember creates the member page in Notion if it doesn't exist,
	// otherwise returns the existing member record.
	EnsureMember(ctx context.Context, m *domain.Member) (*domain.Member, error)

	// GetMember retrieves a member by Telegram ID.
	GetMember(ctx context.Context, telegramID int64) (*domain.Member, error)

	// ListMembers returns all registered members.
	ListMembers(ctx context.Context) ([]domain.Member, error)
}
