package ports

import (
	"context"
	"time"

	"github.com/arkcode369/ark-vault/internal/domain"
)

// GamificationStore provides storage for gamification data.
type GamificationStore interface {
	GetProfile(ctx context.Context, telegramID int64) (*domain.GamificationProfile, error)
	SaveProfile(ctx context.Context, profile *domain.GamificationProfile) error
	GetStreak(ctx context.Context, telegramID int64) (*domain.StreakData, error)
	SaveStreak(ctx context.Context, streak *domain.StreakData) error
	AppendXPEvent(ctx context.Context, event *domain.XPEvent) error
	GetXPEvents(ctx context.Context, telegramID int64, since time.Time) ([]domain.XPEvent, error)
}
