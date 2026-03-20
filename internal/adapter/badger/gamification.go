package badger

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	badgerdb "github.com/dgraph-io/badger/v4"

	"github.com/arkcode369/ark-vault/internal/domain"
)

// GamificationRepo implements ports.GamificationStore backed by BadgerDB.
type GamificationRepo struct {
	store *Store
}

// NewGamificationRepo creates a new GamificationRepo.
func NewGamificationRepo(store *Store) *GamificationRepo {
	return &GamificationRepo{store: store}
}

func profileKey(id int64) string   { return fmt.Sprintf("gam:profile:%d", id) }
func streakKey(id int64) string    { return fmt.Sprintf("gam:streak:%d", id) }
func xpLogKey(id int64, ts int64) string {
	return fmt.Sprintf("gam:xp_log:%d:%d:%d", id, ts, rand.Int63n(1000000))
}
func xpLogPrefix(id int64) string  { return fmt.Sprintf("gam:xp_log:%d:", id) }

// GetProfile returns the gamification profile for the given Telegram ID.
// Returns nil (not an error) when the profile does not exist yet.
func (r *GamificationRepo) GetProfile(_ context.Context, telegramID int64) (*domain.GamificationProfile, error) {
	var p domain.GamificationProfile
	err := r.store.Get(profileKey(telegramID), &p)
	if err == badgerdb.ErrKeyNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// SaveProfile persists the gamification profile.
func (r *GamificationRepo) SaveProfile(_ context.Context, profile *domain.GamificationProfile) error {
	return r.store.Set(profileKey(profile.TelegramID), profile)
}

// GetStreak returns the streak data for the given Telegram ID.
// Returns nil (not an error) when no streak data exists yet.
func (r *GamificationRepo) GetStreak(_ context.Context, telegramID int64) (*domain.StreakData, error) {
	var s domain.StreakData
	err := r.store.Get(streakKey(telegramID), &s)
	if err == badgerdb.ErrKeyNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// SaveStreak persists the streak data.
func (r *GamificationRepo) SaveStreak(_ context.Context, streak *domain.StreakData) error {
	return r.store.Set(streakKey(streak.TelegramID), streak)
}

// AppendXPEvent records a single XP event keyed by member ID and unix timestamp.
func (r *GamificationRepo) AppendXPEvent(_ context.Context, event *domain.XPEvent) error {
	key := xpLogKey(event.TelegramID, event.Timestamp.UnixNano())
	return r.store.Set(key, event)
}

// GetXPEvents returns all XP events for the member since the given time.
func (r *GamificationRepo) GetXPEvents(_ context.Context, telegramID int64, since time.Time) ([]domain.XPEvent, error) {
	raw, err := r.store.Scan(xpLogPrefix(telegramID))
	if err != nil {
		return nil, err
	}
	var events []domain.XPEvent
	for _, b := range raw {
		var ev domain.XPEvent
		if err := json.Unmarshal(b, &ev); err != nil {
			return nil, err
		}
		if !ev.Timestamp.Before(since) {
			events = append(events, ev)
		}
	}
	return events, nil
}
