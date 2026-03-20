package badger

import (
	"context"
	"encoding/json"
	"fmt"

	badgerdb "github.com/dgraph-io/badger/v4"

	"github.com/arkcode369/ark-vault/internal/domain"
)

// ChallengeRepo implements ports.ChallengeStore backed by BadgerDB.
type ChallengeRepo struct {
	store *Store
}

// NewChallengeRepo creates a new ChallengeRepo.
func NewChallengeRepo(store *Store) *ChallengeRepo {
	return &ChallengeRepo{store: store}
}

func challengeKey(yearWeek string) string {
	return fmt.Sprintf("gam:challenge:%s", yearWeek)
}

func challengeEntryKey(yearWeek string, telegramID int64) string {
	return fmt.Sprintf("gam:challenge_entry:%s:%d", yearWeek, telegramID)
}

func challengeEntryPrefix(yearWeek string) string {
	return fmt.Sprintf("gam:challenge_entry:%s:", yearWeek)
}

// GetChallenge returns the weekly challenge for the given year-week.
// Returns nil (not an error) when no challenge exists.
func (r *ChallengeRepo) GetChallenge(_ context.Context, yearWeek string) (*domain.WeeklyChallenge, error) {
	var c domain.WeeklyChallenge
	err := r.store.Get(challengeKey(yearWeek), &c)
	if err == badgerdb.ErrKeyNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// SaveChallenge persists a weekly challenge.
func (r *ChallengeRepo) SaveChallenge(_ context.Context, challenge *domain.WeeklyChallenge) error {
	return r.store.Set(challengeKey(challenge.YearWeek), challenge)
}

// GetChallengeEntry returns a member's entry for the given week.
// Returns nil (not an error) when no entry exists.
func (r *ChallengeRepo) GetChallengeEntry(_ context.Context, yearWeek string, telegramID int64) (*domain.ChallengeEntry, error) {
	var e domain.ChallengeEntry
	err := r.store.Get(challengeEntryKey(yearWeek, telegramID), &e)
	if err == badgerdb.ErrKeyNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &e, nil
}

// SaveChallengeEntry persists a challenge entry.
func (r *ChallengeRepo) SaveChallengeEntry(_ context.Context, entry *domain.ChallengeEntry) error {
	return r.store.Set(challengeEntryKey(entry.YearWeek, entry.TelegramID), entry)
}

// GetChallengeEntries returns all entries for the given week.
func (r *ChallengeRepo) GetChallengeEntries(_ context.Context, yearWeek string) ([]domain.ChallengeEntry, error) {
	raw, err := r.store.Scan(challengeEntryPrefix(yearWeek))
	if err != nil {
		return nil, err
	}
	var entries []domain.ChallengeEntry
	for _, b := range raw {
		var e domain.ChallengeEntry
		if err := json.Unmarshal(b, &e); err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	return entries, nil
}
